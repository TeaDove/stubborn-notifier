package tg_bot_service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/must_utils"
	"github.com/teadove/terx/terx"
	"gorm.io/gorm"
	"math"
	"regexp"
	"strconv"
	"strings"
	"stubborn-notifier/repositories/notify_repository"
	"time"
)

var disableRegexp = must_utils.Must(regexp.Compile(`^(?P<ID>\d+)$`))

func (r *Service) disable(c *terx.Ctx) error {
	groups := disableRegexp.FindStringSubmatch(c.Text)
	if groups == nil {
		return c.ReplyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s", c.Text, disableRegexp.String()),
		)
	}

	id, err := strconv.ParseUint(groups[disableRegexp.SubexpIndex("ID")], 10, 64)
	if err != nil {
		return c.ReplyWithClientErr(errors.Wrap(err, "bad id"))
	}

	timer, newTimer, err := r.completeAndRescheduleTimer(c, id)
	if err != nil {
		return errors.Wrap(err, "failed to complete timer")
	}

	c.Log().Info().
		Object("timer", timer).
		Object("new_timer", newTimer).
		Msg("timer.completed")

	var text strings.Builder
	if timer.Attempt == 1 {
		text.WriteString("Timer completed!\n")
	} else {
		text.WriteString(fmt.Sprintf("Timer completed in %d attempts with latency of %d minutes\n",
			timer.Attempt,
			int(math.Ceil(time.Now().In(timeZone).Sub(timer.NotifyAt).Minutes())),
		))
	}
	if newTimer != nil {
		err = r.scheduleTimer(c, newTimer)
		if err != nil {
			return errors.Wrap(err, "failed to schedule timer")
		}
	}

	return c.Reply(text.String())
}

func (r *Service) completeAndRescheduleTimer(c *terx.Ctx, id uint64) (*notify_repository.Timer, *notify_repository.Timer, error) {
	var (
		timer    *notify_repository.Timer
		newTimer *notify_repository.Timer
		err      error
	)
	err = r.notifyRepository.DB().Transaction(func(tx *gorm.DB) error {
		timer, err = r.notifyRepository.GetTimerForUpdate(c.Context, tx, id)
		if err != nil {
			return errors.Wrap(err, "failed to get timer for update")
		}

		if timer.CompletedAt.Valid {
			return errors.New("timer is already completed")
		}
		if timer.ChatID != c.Chat.ID {
			return errors.New("invalid chat")
		}

		timer.CompletedAt.Time = time.Now().In(timeZone)
		timer.CompletedAt.Valid = true

		err = r.notifyRepository.SaveTx(c.Context, tx, timer)
		if err != nil {
			return errors.Wrap(err, "failed to save timer for update")
		}

		if timer.Interval.Valid {
			newTimer = &notify_repository.Timer{}
			*newTimer = timer.CopyForNew()

			for newTimer.NotifyAt.Before(time.Now().In(timeZone)) {
				newTimer.NotifyAt = newTimer.NotifyAt.Add(time.Duration(newTimer.Interval.Int64))
			}

			err = r.notifyRepository.SaveTx(c.Context, tx, newTimer)
			if err != nil {
				return errors.Wrap(err, "failed to save timer for update")
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to commit transaction")
	}

	return timer, newTimer, nil
}
