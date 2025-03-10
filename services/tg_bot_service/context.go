package tg_bot_service

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/logger_utils"
	"github.com/teadove/teasutils/utils/redact_utils"
)

type Context struct {
	presentation *Service
	ctx          context.Context

	update   tgbotapi.Update
	text     string
	fulltext string
	command  string

	sentFrom *tgbotapi.User
	chat     *tgbotapi.Chat
}

func (r *Context) addLogCtx() {
	if r.chat != nil && r.chat.Title != "" {
		r.ctx = logger_utils.WithValue(r.ctx, "in", r.chat.Title)
	}

	if r.text != "" {
		r.ctx = logger_utils.WithValue(r.ctx, "text", redact_utils.Trim(r.text))
	}

	if r.sentFrom != nil {
		r.ctx = logger_utils.WithValue(r.ctx, "from", r.sentFrom.String())
	}

	if r.command != "" {
		r.ctx = logger_utils.WithValue(r.ctx, "command", r.command)
	}
}

func (r *Service) makeCtx(ctx context.Context, update *tgbotapi.Update) Context {
	c := Context{
		presentation: r,
		update:       *update,
		chat:         update.FromChat(),
		sentFrom:     update.SentFrom(),
	}

	if update.Message != nil {
		c.fulltext = update.Message.Text
	}

	c.ctx = ctx

	inChat := c.sentFrom != nil && c.chat != nil && c.sentFrom.ID != c.chat.ID
	c.command, c.text = extractCommandAndText(c.fulltext, c.presentation.bot.Self.UserName, inChat)

	c.addLogCtx()

	return c
}

func (r *Context) Log() *zerolog.Logger {
	return zerolog.Ctx(r.ctx)
}

func (r *Context) LogWithUpdate() zerolog.Logger {
	return zerolog.Ctx(r.ctx).With().Interface("update", r.update).Logger()
}
