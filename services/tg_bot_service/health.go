package tg_bot_service

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Service) Health(_ context.Context) error {
	_, err := r.bot.GetMe()
	return errors.Wrap(err, "failed to get me")
}

func (r *Service) Close(_ context.Context) error {
	r.bot.StopReceivingUpdates()
	return nil
}
