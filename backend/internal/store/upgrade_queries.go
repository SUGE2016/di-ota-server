package store

import (
	"context"
	"time"
)

type UpgradeRecord struct {
	ID            int64
	DeviceID      string
	TaskID        string
	Status        string
	CreatedAt     time.Time
	SourceVersion string
	TargetVersion string
	ErrorCode     string
}

func (q *Queries) ListUpgradeRecordsByDevice(ctx context.Context, deviceID string, limit int32) ([]UpgradeRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := q.db.QueryContext(ctx, `
SELECT id, device_id, task_id, status, created_at, source_version, target_version, error_code
FROM t_upgrade_record
WHERE device_id = $1
ORDER BY created_at DESC
LIMIT $2
`, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]UpgradeRecord, 0, limit)
	for rows.Next() {
		var item UpgradeRecord
		if err := rows.Scan(
			&item.ID, &item.DeviceID, &item.TaskID, &item.Status, &item.CreatedAt,
			&item.SourceVersion, &item.TargetVersion, &item.ErrorCode,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
