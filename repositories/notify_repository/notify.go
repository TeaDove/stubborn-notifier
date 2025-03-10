package notify_repository

import (
	"time"

	"github.com/google/uuid"
)

type Notify struct {
	ID        uuid.UUID `gorm:"primary_key"`
	ChatID    int64
	Cron      string
	Text      string
	DoneToday bool
	CreatedAt time.Time
}

func Save() {
}
