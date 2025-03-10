package notify_repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Timer struct {
	ID          uuid.UUID `gorm:"primary_key"`
	CreatedAt   time.Time
	CompletedAt sql.NullTime

	ChatID       int64
	Text         string
	NotifyAt     time.Time
	NotifyPeriod time.Duration
}

func (r *Repository) CreateTimer(
	ctx context.Context,
	chatID int64,
	text string,
	dur time.Duration,
	period time.Duration,
) (*Timer, error) {
	now := time.Now().UTC()
	timer := &Timer{
		ID:           uuid.New(),
		ChatID:       chatID,
		Text:         text,
		NotifyAt:     now.Add(dur),
		CreatedAt:    now,
		NotifyPeriod: period,
	}

	err := r.db.WithContext(ctx).Save(timer).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to save timer")
	}

	return timer, nil
}

func (r *Repository) CompleteTimer(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&Timer{}).
		Updates(map[string]any{"completed_at": sql.NullTime{Time: time.Now().UTC(), Valid: true}}).
		Where("id = ?", id).Error
	if err != nil {
		return errors.Wrap(err, "failed to complete timer")
	}

	return nil
}

func (r *Repository) GetIncompleteTimers(ctx context.Context) ([]Timer, error) {
	var timers []Timer

	err := r.db.WithContext(ctx).Find(&timers).Where("completed_at is null").Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to select")
	}

	return timers, nil
}
