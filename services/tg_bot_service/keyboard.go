package tg_bot_service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/teadove/terx/terx"
)

func (r *Service) setKeyboard(c *terx.Ctx) error {
	msg := c.BuildReply("Restored")

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("in 7m"), tgbotapi.NewKeyboardButton("in 13m")},
		[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("in 23m"), tgbotapi.NewKeyboardButton("in 47m")},
		[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(`at 10:20 every 24h about "Daily!"`)},
	)

	_, err := c.Terx.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}
