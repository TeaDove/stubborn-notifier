package tg_bot_service

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/teadove/teasutils/utils/logger_utils"
	"html"
	"regexp"
	"stubborn-notifier/repositories/notify_repository"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

var timerRegexp = must_utils.Must(regexp.Compile(`^in (?P<Dur>.+) about (?P<Text>.+)$`))

func (r *Context) suggestTimers() error {
	msg := r.buildReply("You need to specify duration and text, or use default timers." +
		"i.e. <code>/timer in 37m about Do math homework</code>")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("in 7m", `{"timer": "7m"}`), tgbotapi.NewInlineKeyboardButtonData("in 13m", `{"timer": "13m"}`)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("in 23m", `{"timer": "23m"}`), tgbotapi.NewInlineKeyboardButtonData("in 47m", `{"timer": "47m"}`)),
	)
	_, err := r.presentation.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

func (r *Context) Timer() error {
	if r.text == "" {
		return r.suggestTimers()
	}

	groups := timerRegexp.FindStringSubmatch(r.text)
	if groups == nil {
		return r.replyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s",
				r.text,
				timerRegexp.String(),
			),
		)
	}

	dur, err := time.ParseDuration(groups[timerRegexp.SubexpIndex("Dur")])
	if err != nil {
		return r.replyWithClientErr(errors.Wrap(err, "bad duration"))
	}

	text := groups[timerRegexp.SubexpIndex("Text")]

	timer, err := r.presentation.notifyRepository.CreateTimer(r.ctx, r.chat.ID, text, dur, time.Minute*1)
	if err != nil {
		return errors.Wrap(err, "failed to create timer")
	}

	return r.scheduleTimer(timer)
}

func (r *Context) scheduleTimer(timer *notify_repository.Timer) error {
	r.presentation.timersMu.Lock()
	defer r.presentation.timersMu.Unlock()
	r.presentation.timers[timer.ID] = *timer

	go r.presentation.notifyTimer(logger_utils.NewLoggedCtx(), timer)

	zerolog.Ctx(r.ctx).Info().Interface("timer", timer).Msg("timer.saved")

	return r.reply(fmt.Sprintf("Timer set!\n\nAt: %s", timer.NotifyAt))
}

func (r *Service) notifyTimer(ctx context.Context, timer *notify_repository.Timer) {
	ctx = logger_utils.WithValue(ctx, "timer_id", timer.ID.String())
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

	msgReq := tgbotapi.NewMessage(timer.ChatID, fmt.Sprintf("%s\n\nTo disable timer, sent:\n <code>/disable %s</code>", html.EscapeString(timer.Text), timer.ID.String()))
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

		_, err = r.bot.Send(msgReq)
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

		zerolog.Ctx(ctx).Info().Str("next", timer.NotifyPeriod.String()).Msg("notification.sent")
		time.Sleep(timer.NotifyPeriod)
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
