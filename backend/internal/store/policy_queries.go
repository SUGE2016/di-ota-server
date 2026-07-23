package store

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const (
	ReportStatusModeRelaxed = "relaxed"
	ReportStatusModeStrict  = "strict"
)

type ProductModelPolicy struct {
	ProductModel     string
	ReportStatusMode string
	UpdatedAt        time.Time
	UpdatedBy        string
}

func NormalizeReportStatusMode(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == ReportStatusModeStrict {
		return ReportStatusModeStrict
	}
	return ReportStatusModeRelaxed
}

// GetReportStatusMode returns relaxed when no policy row exists.
func (q *Queries) GetReportStatusMode(ctx context.Context, productModel string) (string, error) {
	productModel = strings.TrimSpace(productModel)
	if productModel == "" {
		return ReportStatusModeRelaxed, nil
	}
	row := q.db.QueryRowContext(ctx, `
SELECT report_status_mode FROM t_product_model_policy WHERE product_model = $1
`, productModel)
	var mode string
	err := row.Scan(&mode)
	if err == sql.ErrNoRows {
		return ReportStatusModeRelaxed, nil
	}
	if err != nil {
		return "", err
	}
	return NormalizeReportStatusMode(mode), nil
}

func (q *Queries) ListProductModelPolicies(ctx context.Context) ([]ProductModelPolicy, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT product_model, report_status_mode, updated_at, updated_by
FROM t_product_model_policy
ORDER BY product_model ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProductModelPolicy, 0, 16)
	for rows.Next() {
		var p ProductModelPolicy
		if err := rows.Scan(&p.ProductModel, &p.ReportStatusMode, &p.UpdatedAt, &p.UpdatedBy); err != nil {
			return nil, err
		}
		p.ReportStatusMode = NormalizeReportStatusMode(p.ReportStatusMode)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListProductModelsForPolicyUI: distinct models from devices ∪ policy table.
func (q *Queries) ListProductModelsForPolicyUI(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT product_model FROM (
  SELECT DISTINCT product_model FROM t_device WHERE product_model <> ''
  UNION
  SELECT product_model FROM t_product_model_policy
) AS m
ORDER BY product_model ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 32)
	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, err
		}
		out = append(out, model)
	}
	return out, rows.Err()
}

func (q *Queries) UpsertProductModelPolicy(ctx context.Context, productModel, mode, updatedBy string) (ProductModelPolicy, error) {
	productModel = strings.TrimSpace(productModel)
	mode = NormalizeReportStatusMode(mode)
	updatedBy = strings.TrimSpace(updatedBy)
	row := q.db.QueryRowContext(ctx, `
INSERT INTO t_product_model_policy (product_model, report_status_mode, updated_at, updated_by)
VALUES ($1, $2, NOW(), $3)
ON CONFLICT (product_model) DO UPDATE SET
  report_status_mode = EXCLUDED.report_status_mode,
  updated_at = NOW(),
  updated_by = EXCLUDED.updated_by
RETURNING product_model, report_status_mode, updated_at, updated_by
`, productModel, mode, updatedBy)
	var p ProductModelPolicy
	err := row.Scan(&p.ProductModel, &p.ReportStatusMode, &p.UpdatedAt, &p.UpdatedBy)
	if err != nil {
		return ProductModelPolicy{}, err
	}
	p.ReportStatusMode = NormalizeReportStatusMode(p.ReportStatusMode)
	return p, nil
}

func (q *Queries) DeleteProductModelPolicy(ctx context.Context, productModel string) error {
	_, err := q.db.ExecContext(ctx, `
DELETE FROM t_product_model_policy WHERE product_model = $1
`, strings.TrimSpace(productModel))
	return err
}
