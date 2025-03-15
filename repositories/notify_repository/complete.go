package notify_repository

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (r *Repository) CompleteTimer(ctx context.Context, id uint64, chatID int64) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&Timer{}).
		Where("id = ? and chat_id = ? and completed_at is null", id, chatID).
		Updates(map[string]any{"completed_at": sql.NullTime{Time: time.Now().UTC(), Valid: true}})

	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to complete timer")
	}

	return result.RowsAffected == 1, nil
}

func (r *Repository) GetTimerForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (*Timer, error) {
	var timer Timer

	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&timer).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timer")
	}

	return &timer, nil
}

func (r *Repository) SaveTx(ctx context.Context, tx *gorm.DB, timer *Timer) error {
	err := tx.WithContext(ctx).Save(timer).Error
	if err != nil {
		return errors.Wrap(err, "failed to save timer")
	}

	return nil
}
