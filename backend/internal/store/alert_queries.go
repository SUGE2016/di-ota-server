package store

import (
	"context"
	"encoding/json"
	"time"
)

type AlertRecord struct {
	AlertID      string
	AlertType    string
	Severity     string
	Status       string
	ResourceType string
	ResourceID   string
	Message      string
	Detail       json.RawMessage
	DedupeKey    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ListAlertsParams struct {
	Status   string
	Severity string
	Limit    int32
	Offset   int32
}

func scanAlert(row scanner) (AlertRecord, error) {
	var a AlertRecord
	err := row.Scan(
		&a.AlertID, &a.AlertType, &a.Severity, &a.Status,
		&a.ResourceType, &a.ResourceID, &a.Message, &a.Detail,
		&a.DedupeKey, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (q *Queries) CreateAlert(ctx context.Context, a AlertRecord) (AlertRecord, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO t_alert (
  alert_id, alert_type, severity, status, resource_type, resource_id, message, detail, dedupe_key
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (dedupe_key) DO UPDATE SET
  message = EXCLUDED.message,
  detail = EXCLUDED.detail,
  updated_at = NOW()
RETURNING alert_id, alert_type, severity, status, resource_type, resource_id, message, detail, dedupe_key, created_at, updated_at
`, a.AlertID, a.AlertType, a.Severity, a.Status, a.ResourceType, a.ResourceID, a.Message, a.Detail, a.DedupeKey)
	return scanAlert(row)
}

func (q *Queries) ListAlerts(ctx context.Context, arg ListAlertsParams) ([]AlertRecord, error) {
	if arg.Limit <= 0 || arg.Limit > 200 {
		arg.Limit = 20
	}
	rows, err := q.db.QueryContext(ctx, `
SELECT alert_id, alert_type, severity, status, resource_type, resource_id, message, detail, dedupe_key, created_at, updated_at
FROM t_alert
WHERE ($1 = '' OR status = $1)
  AND ($2 = '' OR severity = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4
`, arg.Status, arg.Severity, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AlertRecord, 0, arg.Limit)
	for rows.Next() {
		item, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (q *Queries) CountAlerts(ctx context.Context, status, severity string) (int64, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM t_alert
WHERE ($1 = '' OR status = $1) AND ($2 = '' OR severity = $2)
`, status, severity)
	var n int64
	err := row.Scan(&n)
	return n, err
}

func (q *Queries) UpdateAlertStatus(ctx context.Context, alertID, status string) (AlertRecord, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE t_alert SET status = $2, updated_at = NOW()
WHERE alert_id = $1
RETURNING alert_id, alert_type, severity, status, resource_type, resource_id, message, detail, dedupe_key, created_at, updated_at
`, alertID, status)
	return scanAlert(row)
}

func (q *Queries) UpdateAlertsStatusByIDs(ctx context.Context, ids []string, status string) (int64, error) {
	var n int64
	for _, id := range ids {
		if _, err := q.UpdateAlertStatus(ctx, id, status); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
