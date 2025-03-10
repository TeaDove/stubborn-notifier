package notify_repository

import (
	"context"
	"stubborn-notifier/settings"

	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(ctx context.Context) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(settings.Settings.DB), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open gorm.db")
	}

	err = db.Migrator().AutoMigrate(&Timer{}, &Notify{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to auto migrate")
	}

	return &Repository{db: db}, nil
}
