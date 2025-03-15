package terx

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Terx) Health(_ context.Context) error {
	_, err := r.Bot.GetMe()
	return errors.Wrap(err, "failed to get me")
}

func (r *Terx) Close(_ context.Context) error {
	r.Bot.StopReceivingUpdates()
	return nil
}
