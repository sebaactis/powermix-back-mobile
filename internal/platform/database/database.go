package database

import (
	"fmt"

	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	switch cfg.Driver {
	case "postgres":
		return gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("driver not supported: %s", cfg.Driver)
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		&voucher.Voucher{},
		&proof.Proof{},
	)
}
