package tg_bot_service

import (
	"github.com/pkg/errors"
	"stubborn-notifier/terx"
	"time"
)

func (r *Service) getTimers(c *terx.Context) error {
	timers, err := r.notifyRepository.GetIncompleteTimersForChat(c.Ctx, c.Chat.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get timers")
	}

	for _, timer := range timers {
		err = r.sentTimerDescription(c, &timer)
		if err != nil {
			return errors.Wrap(err, "failed to sent timer description")
		}

		time.Sleep(300 * time.Millisecond)
	}

	return nil
}
