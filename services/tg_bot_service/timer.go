package tg_bot_service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/teadove/teasutils/utils/logger_utils"
	"html"
	"strconv"
	"strings"
	"stubborn-notifier/repositories/notify_repository"
	"stubborn-notifier/terx"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var timeZone = time.FixedZone("Europe/Moscow", 3*60*60)

type TimerRequest struct {
	In    sql.Null[time.Duration]
	At    sql.NullTime
	About sql.NullString
	Every sql.Null[time.Duration]
}

func parseField(k, v string, req *TimerRequest) error {
	var err error
	switch k {
	case "in":
		var in time.Duration
		in, err = time.ParseDuration(v)
		req.In = sql.Null[time.Duration]{V: in, Valid: true}
	case "every":
		var every time.Duration
		every, err = time.ParseDuration(v)
		req.Every = sql.Null[time.Duration]{V: every, Valid: true}
	case "at":
		var at time.Time
		at, err = time.ParseInLocation("15:04", v, timeZone)
		req.At = sql.NullTime{Time: at, Valid: true}
	case "about":
		req.About = sql.NullString{String: v, Valid: true}
	default:
		return errors.Errorf("unknown field %s", k)
	}

	if err != nil {
		return errors.Wrapf(err, "parse field %s", k)
	}

	return nil
}

func parseIntoRequest(in string) (TimerRequest, error) {
	r := csv.NewReader(strings.NewReader(in))
	r.Comma = ' '
	record, err := r.Read()
	if err != nil {
		return TimerRequest{}, errors.Wrap(err, "failed to parse as csv")
	}
	if len(record)%2 != 0 {
		return TimerRequest{}, errors.New("uneven number of fields")
	}

	fields := make(map[string]string)
	for i := 0; i < len(record); i += 2 {
		fields[record[i]] = record[i+1]
	}

	var req TimerRequest
	for k, v := range fields {
		err = parseField(strings.ToLower(k), v, &req)
		if err != nil {
			return TimerRequest{}, errors.Wrap(err, "failed to parse field")
		}
	}

	return req, nil
}

func (r *TimerRequest) getAt(now time.Time) (time.Time, error) {
	if r.In.Valid {
		if r.At.Valid {
			return time.Time{}, errors.New(`"at" and "in" cannot be set simultaneously`)
		}
		return now.Add(r.In.V), nil
	}

	if !r.At.Valid {
		return time.Time{}, errors.New(`"at" or "in" should be set`)
	}

	at := time.Date(now.Year(), now.Month(), now.Day(), r.At.Time.Hour(), r.At.Time.Minute(), r.At.Time.Second(), 0, r.At.Time.Location())
	if at.Before(now) {
		at = at.Add(24 * time.Hour)
	}

	return at, nil
}

func (r *Service) setTimer(c *terx.Context) error {
	req, err := parseIntoRequest(c.Text)
	if err != nil {
		return c.ReplyWithClientErr(err)
	}

	at, err := req.getAt(time.Now().In(timeZone))
	if err != nil {
		return c.ReplyWithClientErr(errors.Wrap(err, "failed to get `at`"))
	}

	timer, err := r.notifyRepository.CreateTimer(c.Ctx, c.Chat.ID, req.About, at, sql.Null[time.Duration]{Valid: false})
	if err != nil {
		return errors.Wrap(err, "failed to create timer")
	}

	return r.scheduleTimer(c, timer)
}

func (r *Service) sentTimerDescription(c *terx.Context, timer *notify_repository.Timer) error {
	var text strings.Builder
	text.WriteString(fmt.Sprintf("I'll notify you at %s", timer.NotifyAtStr()))
	if timer.About.Valid {
		text.WriteString(fmt.Sprintf(` about <i>"%s"</i>`, html.EscapeString(timer.About.String)))
	}
	if timer.Interval.Valid {
		text.WriteString(fmt.Sprintf(` every %s`, timer.Interval.V.String()))
	}

	msg := c.BuildReply(text.String())

	callbackData := CallbackData{Delete: &CallbackDataDelete{ID: timer.ID}}
	callbackDataStr, err := json.Marshal(callbackData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal callback data")
	}

	msg.ReplyMarkup =
		tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("delete", string(callbackDataStr)),
			),
		)

	_, err = c.Terx.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

func (r *Service) scheduleTimer(c *terx.Context, timer *notify_repository.Timer) error {
	r.timersMu.Lock()
	defer r.timersMu.Unlock()
	r.timers[timer.ID] = *timer

	go r.notifyTimer(logger_utils.NewLoggedCtx(), timer)

	zerolog.Ctx(c.Ctx).Info().Interface("timer", timer).Msg("timer.saved")

	return r.sentTimerDescription(c, timer)
}

func (r *Service) notifyTimer(ctx context.Context, timer *notify_repository.Timer) {
	ctx = logger_utils.WithValue(ctx, "timer_id", strconv.Itoa(int(timer.ID)))
	zerolog.Ctx(ctx).
		Info().
		Interface("timer", timer).
		Msg("timer.scheduled")

	now := time.Now().UTC()
	if timer.NotifyAt.After(now) {
		sleepDur := timer.NotifyAt.Sub(now)
		zerolog.Ctx(ctx).Debug().Str("dur", sleepDur.String()).Msg("sleeping")
		time.Sleep(sleepDur)
	}

	about := timer.About.String
	if !timer.About.Valid {
		about = "Time ended!"
	}

	msgReq := tgbotapi.NewMessage(
		timer.ChatID,
		fmt.Sprintf("%s\n\nTo disable timer, sent:\n <code>/disable %s</code>",
			html.EscapeString(about),
			strconv.Itoa(int(timer.ID)),
		),
	)
	msgReq.ParseMode = tgbotapi.ModeHTML

	var err error
	zerolog.Ctx(ctx).Info().Msg("notifying")
	for {
		timer, err = r.notifyRepository.GetTimer(ctx, timer.ID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.check.if.timer.is.completed")
			time.Sleep(onErrorSleepDur)
			continue
		}

		if timer.CompletedAt.Valid {
			zerolog.Ctx(ctx).Info().Msg("exiting.notifier")
			r.timersMu.Lock()

			delete(r.timers, timer.ID)

			r.timersMu.Unlock()
			return
		}

		_, err = r.terx.Bot.Send(msgReq)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.sent.message")
			time.Sleep(onErrorSleepDur)
			continue
		}

		_, err = r.notifyRepository.IncAttemptsTimer(ctx, timer.ID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.update.timer")
			time.Sleep(onErrorSleepDur)
			continue
		}

		zerolog.Ctx(ctx).Info().Str("next", (90 * time.Second).String()).Msg("notification.sent")
		time.Sleep(90 * time.Second)
	}
}

func (r *Service) RestartTimers(ctx context.Context) error {
	timers, err := r.notifyRepository.GetIncompleteTimers(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete timers")
	}

	r.timersMu.Lock()
	defer r.timersMu.Unlock()
	for _, timer := range timers {
		_, ok := r.timers[timer.ID]
		if ok {
			continue
		}

		r.timers[timer.ID] = timer
		go r.notifyTimer(logger_utils.NewLoggedCtx(), &timer)
	}

	return nil
}
