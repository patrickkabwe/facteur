package migrations

import (
	"citadel/app/models"
	"context"

	"github.com/uptrace/bun"
)

func usersMigrationUp_1716207923(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.User)(nil)).Exec(ctx)
	return err
}

func usersMigrationDown_1716207923(ctx context.Context, db *bun.DB) error {
	_, err := db.NewDropColumn().Model((*models.User)(nil)).Exec(ctx)
	return err
}

func init() {
	Migrations.MustRegister(usersMigrationUp_1716207923, usersMigrationDown_1716207923)
}
