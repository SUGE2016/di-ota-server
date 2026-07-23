-- Per product_model report-status state machine policy.
-- Missing row => relaxed (default). Only strict models need an explicit row (or mode=strict).

CREATE TABLE IF NOT EXISTS t_product_model_policy (
  product_model VARCHAR(64) PRIMARY KEY,
  report_status_mode VARCHAR(16) NOT NULL DEFAULT 'relaxed'
    CHECK (report_status_mode IN ('relaxed', 'strict')),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_product_model_policy_mode
  ON t_product_model_policy(report_status_mode);
