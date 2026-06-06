package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"ota-server/backend/internal/config"
	"ota-server/backend/internal/db"
	"ota-server/backend/internal/server"
	"ota-server/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	pg, err := db.OpenPostgres(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("init postgres failed: %v", err)
	}
	defer pg.Close()

	q := store.New(pg)
	log.Printf("ota-worker started, rabbitmq=%s postgres=%s", cfg.RabbitMQ.URL, cfg.Postgres.Host)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	if err := runWorkerCycle(pg, q, cfg); err != nil {
		log.Printf("worker cycle failed: %v", err)
	}

	for {
		<-ticker.C
		if err := runWorkerCycle(pg, q, cfg); err != nil {
			log.Printf("worker cycle failed: %v", err)
		}
	}
}

func runWorkerCycle(pg *sql.DB, q *store.Queries, cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := pg.ExecContext(ctx, `
INSERT INTO t_task_stats (task_id, total_count, success_count, failed_count, failure_rate, error_distribution, snapshot_time)
SELECT
  ur.task_id,
  COUNT(*)::integer AS total_count,
  SUM(CASE WHEN lower(ur.status) IN ('success', 'succeeded') THEN 1 ELSE 0 END)::integer AS success_count,
  SUM(CASE WHEN lower(ur.status) LIKE 'fail%' OR lower(ur.status) = 'error' THEN 1 ELSE 0 END)::integer AS failed_count,
  CASE WHEN COUNT(*) = 0 THEN 0
       ELSE ROUND((SUM(CASE WHEN lower(ur.status) LIKE 'fail%' OR lower(ur.status) = 'error' THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric), 4)
  END AS failure_rate,
  '{}'::jsonb AS error_distribution,
  NOW() AS snapshot_time
FROM t_upgrade_record ur
GROUP BY ur.task_id
`); err != nil {
		return err
	}

	rows, err := pg.QueryContext(ctx, `
UPDATE t_release_task t
SET state = 'Paused'
FROM (
  SELECT
    ur.task_id,
    CASE WHEN COUNT(*) = 0 THEN 0
         ELSE SUM(CASE WHEN lower(ur.status) LIKE 'fail%' OR lower(ur.status) = 'error' THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric
    END AS failure_rate
  FROM t_upgrade_record ur
  GROUP BY ur.task_id
) s
WHERE t.task_id = s.task_id
  AND t.state = 'Running'
  AND s.failure_rate > t.failure_threshold
RETURNING t.task_id, t.failure_threshold, s.failure_rate
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	paused := 0
	for rows.Next() {
		var taskID string
		var threshold, failureRate float64
		if err := rows.Scan(&taskID, &threshold, &failureRate); err != nil {
			return err
		}
		paused++
		if err := server.UpsertTaskFailureAlert(ctx, q, taskID, failureRate, threshold); err != nil {
			log.Printf("create failure alert for %s: %v", taskID, err)
		}
		server.EmitTaskPausedWebhookAsync(cfg, taskID, "failure_rate_exceeded", failureRate, threshold)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	deleted := int64(0)
	if cfg.Worker.TaskStatsRetentionHours > 0 {
		cleanupRes, err := pg.ExecContext(ctx, `
DELETE FROM t_task_stats
WHERE snapshot_time < NOW() - make_interval(hours => $1)
`, cfg.Worker.TaskStatsRetentionHours)
		if err != nil {
			return err
		}
		deleted, _ = cleanupRes.RowsAffected()
	}

	log.Printf("worker heartbeat: stats snapshot done, auto-pause affected=%d, cleaned_stats=%d", paused, deleted)
	return nil
}
