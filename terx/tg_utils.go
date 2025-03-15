package terx

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"html"
)

func (r *Context) BuildReply(text string) tgbotapi.MessageConfig {
	if r.Chat == nil {
		panic(errors.New("chat is nil"))
	}
	msgReq := tgbotapi.NewMessage(r.Chat.ID, text)
	if r.Update.Message != nil {
		msgReq.ReplyToMessageID = r.Update.Message.MessageID
	}
	msgReq.ParseMode = tgbotapi.ModeHTML

	return msgReq
}

func (r *Context) Reply(text string) error {
	_, err := r.ReplyWithMessage(text)
	return err
}

func (r *Context) Replyf(text string, a ...any) error {
	_, err := r.ReplyWithMessage(fmt.Sprintf(text, a...))
	return err
}

func (r *Context) TryReply(text string) {
	_, err := r.ReplyWithMessage(text)
	if err != nil {
		r.LogWithUpdate().
			Error().Stack().
			Err(err).
			Str("text", text).
			Msg("failed.to.reply")
	}
}

func (r *Context) ReplyWithMessage(text string) (tgbotapi.Message, error) {
	msgReq := r.BuildReply(text)

	msg, err := r.Terx.Bot.Send(msgReq)
	if err != nil {
		return tgbotapi.Message{}, errors.Wrap(err, "failed to send message")
	}

	return msg, nil
}

func (r *Context) EditMsgText(msg *tgbotapi.Message, text string) error {
	if text == msg.Text {
		r.LogWithUpdate().
			Warn().
			Str("text", text).
			Msg("attempt.to.edit.unmodified.msg")

		return nil
	}

	editMessageTextReq := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	editMessageTextReq.ParseMode = tgbotapi.ModeHTML

	newMsg, err := r.Terx.Bot.Send(editMessageTextReq)
	if err != nil {
		return errors.Wrap(err, "failed to edit message")
	}

	*msg = newMsg

	return nil
}

func (r *Context) ReplyWithClientErr(err error) error {
	if err == nil {
		return nil
	}

	r.LogWithUpdate().
		Warn().Err(err).
		Msg("client.error")

	return r.Reply(fmt.Sprintf("Bad request: <code>%s</code>", html.EscapeString(err.Error())))
}

func (r *Context) TryReplyOnErr(err error) {
	if err == nil {
		return
	}

	err = r.Reply(fmt.Sprintf("Error occurred: %s", err.Error()))
	if err != nil {
		zerolog.Ctx(r.Ctx).Error().Stack().Err(err).Msg("failed.to.try.reply.on.err")
	}
}
