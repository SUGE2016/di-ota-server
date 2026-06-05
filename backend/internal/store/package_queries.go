package store

import (
	"context"
	"database/sql"
)

type PackageDetail struct {
	PackageID            string
	ProductCode          string
	Version              string
	FileHash             string
	Signature            string
	Status               string
	MinUpgradableVersion string
}

func (q *Queries) GetPackageDetail(ctx context.Context, packageID string) (PackageDetail, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT package_id, product_code, version, file_hash, signature, status, min_upgradable_version
FROM t_package WHERE package_id = $1
`, packageID)
	var p PackageDetail
	err := row.Scan(&p.PackageID, &p.ProductCode, &p.Version, &p.FileHash, &p.Signature, &p.Status, &p.MinUpgradableVersion)
	if err == sql.ErrNoRows {
		return PackageDetail{}, sql.ErrNoRows
	}
	return p, err
}
