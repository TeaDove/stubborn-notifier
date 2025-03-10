package app_container

import (
	"context"
	"stubborn-notifier/repositories/notify_repository"
	"stubborn-notifier/services/tg_bot_service"
	"stubborn-notifier/settings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/teadove/teasutils/utils/di_utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

type Container struct {
	TGBotPresentation *tg_bot_service.Service

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

	// TODO move to settings
	bot, err := tgbotapi.NewBotAPI(settings.Settings.TG.BotToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bot client")
	}

	notifyRepository, err := notify_repository.NewRepository(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create notify repository")
	}

	tgBotService, err := tg_bot_service.NewService(ctx, bot, scheduler, notifyRepository)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tg bot service")
	}

	container := &Container{
		TGBotPresentation: tgBotService,
		healths:           []di_utils.Health{tgBotService},
		closers:           []di_utils.CloserWithContext{tgBotService},
	}

	return container, nil
}
