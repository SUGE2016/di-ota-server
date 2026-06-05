package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type DeviceRegistry struct {
	DeviceID           string
	DeviceGroup        string
	ProductModel       string
	HardwareVersion    string
	ProductCode        string
	Tags               json.RawMessage
	CurrentVersion     string
	ReportedVersion    string
	CatalogVersion     string
	CatalogSyncedAt    sql.NullTime
	CatalogSource      string
	EligibilityState   string
	InconsistencyFlags json.RawMessage
	LastSeenAt         sql.NullTime
	RegisteredAt       time.Time
}

const deviceRegistrySelect = `
SELECT device_id, device_group, product_model, hardware_version, product_code, tags,
       current_version, reported_version, catalog_version, catalog_synced_at, catalog_source,
       eligibility_state, inconsistency_flags, last_seen_at, registered_at
FROM t_device
`

func scanDeviceRegistry(row scanner) (DeviceRegistry, error) {
	var d DeviceRegistry
	err := row.Scan(
		&d.DeviceID,
		&d.DeviceGroup,
		&d.ProductModel,
		&d.HardwareVersion,
		&d.ProductCode,
		&d.Tags,
		&d.CurrentVersion,
		&d.ReportedVersion,
		&d.CatalogVersion,
		&d.CatalogSyncedAt,
		&d.CatalogSource,
		&d.EligibilityState,
		&d.InconsistencyFlags,
		&d.LastSeenAt,
		&d.RegisteredAt,
	)
	return d, err
}

type scanner interface {
	Scan(dest ...any) error
}

func (q *Queries) GetDeviceRegistry(ctx context.Context, deviceID string) (DeviceRegistry, error) {
	row := q.db.QueryRowContext(ctx, deviceRegistrySelect+` WHERE device_id = $1`, deviceID)
	d, err := scanDeviceRegistry(row)
	if err != nil {
		return DeviceRegistry{}, err
	}
	return d, nil
}

type InsertDeviceRegistryParams struct {
	DeviceID         string
	DeviceGroup      string
	ProductModel     string
	HardwareVersion  string
	ProductCode      string
	Tags             json.RawMessage
	CatalogVersion   string
	ReportedVersion  string
	CurrentVersion   string
	CatalogSource    string
	InconsistencyFlags json.RawMessage
}

func (q *Queries) InsertDeviceRegistry(ctx context.Context, arg InsertDeviceRegistryParams) (DeviceRegistry, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO t_device (
  device_id, device_group, product_model, hardware_version, product_code, tags,
  current_version, reported_version, catalog_version, catalog_synced_at, catalog_source,
  eligibility_state, inconsistency_flags, last_heartbeat
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10,'active',$11,NOW()
)
RETURNING device_id, device_group, product_model, hardware_version, product_code, tags,
          current_version, reported_version, catalog_version, catalog_synced_at, catalog_source,
          eligibility_state, inconsistency_flags, last_seen_at, registered_at
`, arg.DeviceID, arg.DeviceGroup, arg.ProductModel, arg.HardwareVersion, arg.ProductCode, arg.Tags,
		arg.CurrentVersion, arg.ReportedVersion, arg.CatalogVersion, arg.CatalogSource, arg.InconsistencyFlags)
	return scanDeviceRegistry(row)
}

type UpdateDeviceRegistryParams struct {
	DeviceID           string
	DeviceGroup        string
	ProductModel       string
	HardwareVersion    string
	ProductCode        string
	Tags               json.RawMessage
	CatalogVersion     string
	ReportedVersion    string
	CurrentVersion     string
	CatalogSource      string
	InconsistencyFlags json.RawMessage
}

func (q *Queries) UpdateDeviceRegistry(ctx context.Context, arg UpdateDeviceRegistryParams) (DeviceRegistry, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE t_device SET
  device_group = $2,
  product_model = $3,
  hardware_version = $4,
  product_code = $5,
  tags = $6,
  catalog_version = $7,
  reported_version = $8,
  current_version = $9,
  catalog_synced_at = NOW(),
  catalog_source = $10,
  inconsistency_flags = $11,
  last_heartbeat = NOW()
WHERE device_id = $1
RETURNING device_id, device_group, product_model, hardware_version, product_code, tags,
          current_version, reported_version, catalog_version, catalog_synced_at, catalog_source,
          eligibility_state, inconsistency_flags, last_seen_at, registered_at
`, arg.DeviceID, arg.DeviceGroup, arg.ProductModel, arg.HardwareVersion, arg.ProductCode, arg.Tags,
		arg.CatalogVersion, arg.ReportedVersion, arg.CurrentVersion, arg.CatalogSource, arg.InconsistencyFlags)
	return scanDeviceRegistry(row)
}

func (q *Queries) TouchDeviceReportedVersion(ctx context.Context, deviceID, reportedVersion string) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE t_device SET
  reported_version = $2,
  current_version = $2,
  last_seen_at = NOW(),
  last_heartbeat = NOW()
WHERE device_id = $1
`, deviceID, reportedVersion)
	return err
}

func (q *Queries) TouchDeviceLastSeen(ctx context.Context, deviceID, reportedFromRequest string) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE t_device SET
  last_seen_at = NOW(),
  last_heartbeat = NOW(),
  reported_version = CASE WHEN $2 <> '' THEN $2 ELSE reported_version END,
  current_version = CASE WHEN $2 <> '' THEN $2 ELSE current_version END
WHERE device_id = $1
`, deviceID, reportedFromRequest)
	return err
}

func (q *Queries) ListDeviceIDsForTaskSnapshot(ctx context.Context, group, productModel, hardwareVersion string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT device_id FROM t_device
WHERE device_group = $1 AND product_model = $2 AND hardware_version = $3
  AND eligibility_state = 'active'
`, group, productModel, hardwareVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0, 64)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (q *Queries) InsertTaskTarget(ctx context.Context, taskID, deviceID string) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO t_task_target (task_id, device_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING
`, taskID, deviceID)
	return err
}

func (q *Queries) DeviceInTaskTarget(ctx context.Context, taskID, deviceID string) (bool, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT 1 FROM t_task_target WHERE task_id = $1 AND device_id = $2
`, taskID, deviceID)
	var one int
	err := row.Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (q *Queries) ListRunningTasksForDevice(ctx context.Context, deviceID string) ([]TReleaseTask, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT t.task_id, t.package_id, t.target_group, t.product_model, t.hardware_version,
       t.failure_threshold, t.state, t.created_at, t.canary_percent, t.schedule_time, t.force_upgrade
FROM t_release_task t
JOIN t_task_target tt ON tt.task_id = t.task_id AND tt.device_id = $1
WHERE t.state = 'Running'
  AND (t.schedule_time IS NULL OR t.schedule_time <= NOW())
ORDER BY t.force_upgrade DESC, t.created_at DESC
`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TReleaseTask, 0, 4)
	for rows.Next() {
		var i TReleaseTask
		if err := rows.Scan(
			&i.TaskID, &i.PackageID, &i.TargetGroup, &i.ProductModel, &i.HardwareVersion,
			&i.FailureThreshold, &i.State, &i.CreatedAt, &i.CanaryPercent, &i.ScheduleTime, &i.ForceUpgrade,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type CatalogSyncBatch struct {
	BatchID        string
	Source         string
	Mode           string
	AcceptedCount  int32
	WarnedCount    int32
	RejectedCount  int32
	Response       json.RawMessage
	CreatedAt      time.Time
}

func (q *Queries) GetCatalogSyncBatch(ctx context.Context, batchID string) (CatalogSyncBatch, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT batch_id, source, mode, accepted_count, warned_count, rejected_count, response, created_at
FROM t_catalog_sync_batch WHERE batch_id = $1
`, batchID)
	var b CatalogSyncBatch
	err := row.Scan(&b.BatchID, &b.Source, &b.Mode, &b.AcceptedCount, &b.WarnedCount, &b.RejectedCount, &b.Response, &b.CreatedAt)
	return b, err
}

func (q *Queries) InsertCatalogSyncBatch(ctx context.Context, b CatalogSyncBatch) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO t_catalog_sync_batch (
  batch_id, source, mode, accepted_count, warned_count, rejected_count, response
) VALUES ($1,$2,$3,$4,$5,$6,$7)
`, b.BatchID, b.Source, b.Mode, b.AcceptedCount, b.WarnedCount, b.RejectedCount, b.Response)
	return err
}

type PendingUpgradeRow struct {
	DeviceID       string
	TaskID         string
	TargetVersion  string
	PackageID      string
	CanaryPercent  int32
	MinUpgradable  string
}

func (q *Queries) ListPendingUpgrades(ctx context.Context, deviceGroup string, limit, offset int32) ([]PendingUpgradeRow, error) {
	base := `
SELECT d.device_id, t.task_id, p.version, p.package_id, t.canary_percent, p.min_upgradable_version
FROM t_task_target tt
JOIN t_release_task t ON t.task_id = tt.task_id AND t.state = 'Running'
JOIN t_device d ON d.device_id = tt.device_id
JOIN t_package p ON p.package_id = t.package_id
LEFT JOIN t_upgrade_record ur ON ur.device_id = d.device_id AND ur.task_id = t.task_id AND ur.status = 'Success'
WHERE ur.id IS NULL
  AND (t.schedule_time IS NULL OR t.schedule_time <= NOW())
  AND d.eligibility_state = 'active'
`
	var rows *sql.Rows
	var err error
	if deviceGroup != "" {
		rows, err = q.db.QueryContext(ctx, base+`
  AND d.device_group = $1
ORDER BY t.created_at DESC, d.device_id ASC
LIMIT $2 OFFSET $3
`, deviceGroup, limit, offset)
	} else {
		rows, err = q.db.QueryContext(ctx, base+`
ORDER BY t.created_at DESC, d.device_id ASC
LIMIT $1 OFFSET $2
`, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list pending upgrades: %w", err)
	}
	defer rows.Close()
	out := make([]PendingUpgradeRow, 0, limit)
	for rows.Next() {
		var r PendingUpgradeRow
		if err := rows.Scan(&r.DeviceID, &r.TaskID, &r.TargetVersion, &r.PackageID, &r.CanaryPercent, &r.MinUpgradable); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
