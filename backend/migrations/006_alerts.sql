CREATE TABLE IF NOT EXISTS t_alert (
  alert_id VARCHAR(64) PRIMARY KEY,
  alert_type VARCHAR(64) NOT NULL,
  severity VARCHAR(16) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'open',
  resource_type VARCHAR(32) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  message TEXT NOT NULL,
  detail JSONB NOT NULL DEFAULT '{}'::jsonb,
  dedupe_key VARCHAR(128) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_status_created ON t_alert(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alert_resource ON t_alert(resource_type, resource_id);
