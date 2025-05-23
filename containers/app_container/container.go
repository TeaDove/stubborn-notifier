package app_container

import (
	"context"
	"github.com/teadove/terx/terx"
	"stubborn-notifier/repositories/notify_repository"
	"stubborn-notifier/services/tg_bot_service"
	"stubborn-notifier/settings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
)

type Container struct {
	Service *tg_bot_service.Service
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

	container := &Container{Service: tgBotService}

	return container, nil
}
