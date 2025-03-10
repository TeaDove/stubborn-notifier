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
	Attempt      uint64
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

	err := r.db.WithContext(ctx).
		Save(timer).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to save timer")
	}

	return timer, nil
}

func (r *Repository) CompleteTimer(ctx context.Context, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&Timer{}).
		Where("id = ?", id).
		Updates(map[string]any{"completed_at": sql.NullTime{Time: time.Now().UTC(), Valid: true}})

	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to complete timer")
	}

	return result.RowsAffected == 1, nil
}

func (r *Repository) IncAttemptsTimer(ctx context.Context, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Exec("update timers set attempt = attempt + 1 where id = ?", id)

	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to complete timer")
	}

	return result.RowsAffected == 1, nil
}

func (r *Repository) GetIncompleteTimers(ctx context.Context) ([]Timer, error) {
	var timers []Timer

	err := r.db.WithContext(ctx).
		Where("completed_at is null").
		Find(&timers).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to select")
	}

	return timers, nil
}

func (r *Repository) GetTimer(ctx context.Context, id uuid.UUID) (*Timer, error) {
	var timer Timer

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&timer).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to select")
	}

	return &timer, nil
}

func (r *Repository) TimerIsCompleted(ctx context.Context, id uuid.UUID) (bool, error) {
	var timer Timer

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&timer).
		Error
	if err != nil {
		return false, errors.Wrap(err, "failed to select")
	}

	return timer.CompletedAt.Valid, nil
}
