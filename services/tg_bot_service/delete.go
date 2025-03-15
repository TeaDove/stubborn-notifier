package tg_bot_service

import (
	"encoding/json"
	"github.com/pkg/errors"
	"stubborn-notifier/terx"
)

type CallbackDataDelete struct {
	ID uint64 `json:"id"`
}

type CallbackData struct {
	Delete *CallbackDataDelete
}

func (r *Service) processCallback(c *terx.Context) error {
	var (
		req CallbackData
		err error
		ok  bool
	)
	err = json.Unmarshal([]byte(c.Update.CallbackData()), &req)
	if err != nil {
		return errors.Wrap(err, "invalid callback data")
	}

	if req.Delete != nil {
		ok, err = r.notifyRepository.CompleteTimer(c.Ctx, req.Delete.ID, c.Chat.ID)
		if err != nil {
			return errors.Wrap(err, "failed to delete")
		}

		if !ok {
			return c.ReplyOnCallback("Timer not found", true)
		}

		return c.ReplyOnCallback("Ok!", false)
	}

	return nil
}
