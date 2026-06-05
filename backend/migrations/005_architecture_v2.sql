-- 005_architecture_v2: 设备注册表双版本、任务快照、目录同步审计、待升级 hint

ALTER TABLE t_device ADD COLUMN IF NOT EXISTS reported_version VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS catalog_version VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS catalog_synced_at TIMESTAMPTZ;
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS catalog_source VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS eligibility_state VARCHAR(32) NOT NULL DEFAULT 'active';
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS inconsistency_flags JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE t_device ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;

UPDATE t_device
SET reported_version = current_version,
    catalog_version = current_version
WHERE current_version <> ''
  AND reported_version = ''
  AND catalog_version = '';

CREATE TABLE IF NOT EXISTS t_task_target (
  task_id VARCHAR(64) NOT NULL REFERENCES t_release_task(task_id) ON DELETE CASCADE,
  device_id VARCHAR(64) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (task_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_task_target_device ON t_task_target(device_id);

CREATE TABLE IF NOT EXISTS t_catalog_sync_batch (
  batch_id VARCHAR(64) PRIMARY KEY,
  source VARCHAR(32) NOT NULL,
  mode VARCHAR(32) NOT NULL,
  accepted_count INT NOT NULL DEFAULT 0,
  warned_count INT NOT NULL DEFAULT 0,
  rejected_count INT NOT NULL DEFAULT 0,
  response JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_catalog_sync_created ON t_catalog_sync_batch(created_at DESC);
