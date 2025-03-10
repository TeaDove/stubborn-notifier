package tg_bot_service

import (
	"regexp"
	"stubborn-notifier/repositories/notify_repository"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

type TimerJob struct {
	timer notify_repository.Timer
	done  chan struct{}
}

var timerRegexp = must_utils.Must(regexp.Compile(`^in (?P<Dur>.+) with (?P<Text>.+)$`))

func (r *Context) Timer() error {
	groups := timerRegexp.FindStringSubmatch(r.text)
	if groups == nil {
		return r.replyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s", r.text, timerRegexp.String()),
		)
	}

	dur, err := time.ParseDuration(groups[timerRegexp.SubexpIndex("Dur")])
	if err != nil {
		return r.replyWithClientErr(errors.Wrap(err, "bad duration"))
	}

	text := groups[timerRegexp.SubexpIndex("Text")]

	timer, err := r.presentation.notifyRepository.CreateTimer(r.ctx, r.chat.ID, text, dur, time.Second*10)
	if err != nil {
		return errors.Wrap(err, "failed to create timer")
	}

	zerolog.Ctx(r.ctx).Info().Interface("timer", timer).Msg("timer.saved")

	return r.reply("Успешно!")
}

// func (r *Context) NotifyText(text string) {
//	_, err := r.presentation.bot.Send(tgbotapi.NewMessage(r.update.Message.Chat.ID, text))
//	if err != nil {
//		zerolog.Ctx(r.ctx).Error().Stack().Err(err).Msg("failed.to.sent.notify")
//	}
//
//	zerolog.Ctx(r.ctx).Info().Str("text", text).Msg("notify.sent")
//}
