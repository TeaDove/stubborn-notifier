package app_container

import (
	"context"
	"stubborn-notifier/repositories/notify_repository"
	"stubborn-notifier/services/tg_bot_service"
	"stubborn-notifier/settings"
	"stubborn-notifier/terx"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/teadove/teasutils/utils/di_utils"

	"github.com/pkg/errors"
)

type Container struct {
	Service *tg_bot_service.Service

	healths []di_utils.Health
	closers []di_utils.CloserWithContext
}

func (r *Container) Healths() []di_utils.Health {
	return r.healths
}

func (r *Container) Closers() []di_utils.CloserWithContext {
	return r.closers
}

func Build(ctx context.Context) (*Container, error) {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()

	notifyRepository, err := notify_repository.NewRepository(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create notify repository")
	}

	appTerx, err := terx.New(&terx.Config{
		Token:        settings.Settings.TG.BotToken,
		ReplyWithErr: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create terx")
	}

	tgBotService, err := tg_bot_service.NewService(ctx, appTerx, scheduler, notifyRepository)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tg bot service")
	}

	container := &Container{
		Service: tgBotService,
		healths: []di_utils.Health{appTerx},
		closers: []di_utils.CloserWithContext{appTerx},
	}

	return container, nil
}
