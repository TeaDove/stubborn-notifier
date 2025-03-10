package tg_bot_service

import (
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

var notifyRegexp = must_utils.Must(regexp.Compile(`^at (?P<Cron>.+) about (?P<Text>.+)$`))

func (r *Context) Notify() error {
	groups := notifyRegexp.FindStringSubmatch(r.text)
	if groups == nil {
		return r.replyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s", r.text, notifyRegexp.String()),
		)
	}

	cron := groups[notifyRegexp.SubexpIndex("Cron")]
	text := groups[notifyRegexp.SubexpIndex("Text")]

	_, err := r.presentation.scheduler.Cron(cron).Do(r.NotifyText, text)
	if err != nil {
		return errors.Wrap(err, "failed schedule notify")
	}

	zerolog.Ctx(r.ctx).Info().Str("text", text).Str("cron", cron).Msg("notify.saved")

	return r.reply("Успешно!")
}

func (r *Context) NotifyText(text string) {
	_, err := r.presentation.bot.Send(tgbotapi.NewMessage(r.update.Message.Chat.ID, text))
	if err != nil {
		zerolog.Ctx(r.ctx).Error().Stack().Err(err).Msg("failed.to.sent.notify")
	}

	zerolog.Ctx(r.ctx).Info().Str("text", text).Msg("notify.sent")
}
