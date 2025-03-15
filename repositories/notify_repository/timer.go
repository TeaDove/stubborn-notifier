package notify_repository

import (
	"context"
	"database/sql"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/redact_utils"
	"time"

	"github.com/pkg/errors"
)

type Timer struct {
	ID          uint64 `gorm:"primary_key"`
	CreatedAt   time.Time
	CompletedAt sql.NullTime

	ChatID int64

	About    sql.NullString
	NotifyAt time.Time
	Interval sql.Null[time.Duration]
	Attempt  uint64
}

func (r *Timer) NotifyAtStr() string {
	return r.NotifyAt.Format("Jan 02 Mon at 15:04")
}

func (r *Timer) MarshalZerologObject(e *zerolog.Event) {
	if r == nil {
		return
	}

	e.
		Uint64("id", r.ID).
		Int64("chat_id", r.ChatID).
		Time("notify_at", r.NotifyAt)
	if r.About.Valid {
		e.Str("about", redact_utils.Trim(r.About.String))
	}
	if r.CompletedAt.Valid {
		e.Time("completed_at", r.CompletedAt.Time)
	}
}

func (r *Timer) CopyForNew() Timer {
	return Timer{
		ChatID:   r.ChatID,
		About:    r.About,
		NotifyAt: r.NotifyAt.Add(r.Interval.V),
		Interval: r.Interval,
	}
}

func (r *Repository) CreateTimer(
	ctx context.Context,
	chatID int64,
	about sql.NullString,
	at time.Time,
	interval sql.Null[time.Duration],
) (*Timer, error) {
	timer := &Timer{
		ChatID:    chatID,
		About:     about,
		NotifyAt:  at,
		CreatedAt: time.Now(),
		Interval:  interval,
	}

	err := r.db.WithContext(ctx).
		Save(timer).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to save timer")
	}

	return timer, nil
}

func (r *Repository) IncAttemptsTimer(ctx context.Context, id uint64) (bool, error) {
	result := r.db.WithContext(ctx).
		Exec("update timers set attempt = attempt + 1 where id = ?", id)

	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to complete timer")
	}

	return result.RowsAffected == 1, nil
}

func (r *Repository) GetIncompleteTimersForChat(ctx context.Context, chatID int64) ([]Timer, error) {
	var timers []Timer

	err := r.db.WithContext(ctx).
		Where("completed_at is null and chat_id = ?", chatID).
		Find(&timers).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to select")
	}

	return timers, nil
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

func (r *Repository) GetTimer(ctx context.Context, id uint64) (*Timer, error) {
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
