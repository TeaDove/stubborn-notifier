package terx

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/logger_utils"
	"github.com/teadove/teasutils/utils/redact_utils"
)

type Context struct {
	Terx *Terx
	Ctx  context.Context

	Text     string
	FullText string
	Command  string

	Update   tgbotapi.Update
	SentFrom *tgbotapi.User
	Chat     *tgbotapi.Chat
}

func (r *Context) addLogCtx() {
	if r.Chat != nil && r.Chat.Title != "" {
		r.Ctx = logger_utils.WithValue(r.Ctx, "in", r.Chat.Title)
	}

	if r.Text != "" {
		r.Ctx = logger_utils.WithValue(r.Ctx, "text", redact_utils.Trim(r.Text))
	}

	if r.SentFrom != nil {
		r.Ctx = logger_utils.WithValue(r.Ctx, "from", r.SentFrom.String())
	}

	if r.Command != "" {
		r.Ctx = logger_utils.WithValue(r.Ctx, "command", r.Command)
	}
}

func (r *Terx) makeCtx(ctx context.Context, update *tgbotapi.Update) *Context {
	c := Context{
		Terx:     r,
		Update:   *update,
		Chat:     update.FromChat(),
		SentFrom: update.SentFrom(),
	}

	if update.Message != nil {
		c.FullText = update.Message.Text
	}

	c.Ctx = ctx

	inChat := c.SentFrom != nil && c.Chat != nil && c.SentFrom.ID != c.Chat.ID
	c.Command, c.Text = extractCommandAndText(c.FullText, r.Bot.Self.UserName, inChat)
	c.Text = strings.TrimSpace(c.Text)

	c.addLogCtx()

	return &c
}

func (r *Context) Log() *zerolog.Logger {
	return zerolog.Ctx(r.Ctx)
}

func (r *Context) LogWithUpdate() *zerolog.Logger {
	logger := zerolog.Ctx(r.Ctx).With().Interface("update", r.Update).Logger()
	return &logger
}
