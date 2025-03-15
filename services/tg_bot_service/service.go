package tg_bot_service

import (
	"context"
	"stubborn-notifier/repositories/notify_repository"
	"stubborn-notifier/terx"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const onErrorSleepDur = 20 * time.Second

type Service struct {
	terx      *terx.Terx
	scheduler *gocron.Scheduler

	notifyRepository *notify_repository.Repository

	timers   map[uint64]notify_repository.Timer
	timersMu sync.Mutex
}

func NewService(
	ctx context.Context,
	terxApp *terx.Terx,
	scheduler *gocron.Scheduler,
	notifyRepository *notify_repository.Repository,
) (*Service, error) {
	r := &Service{
		terx:             terxApp,
		scheduler:        scheduler,
		notifyRepository: notifyRepository,
		timers:           make(map[uint64]notify_repository.Timer, 10),
	}

	command := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "disable",
			Description: "Disable notification",
		},
		tgbotapi.BotCommand{
			Command:     "help",
			Description: "Help!",
		},
	)

	_, err := r.terx.Bot.Request(command)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set commands")
	}
	r.terx.AddHandler(r.disable, terx.FilterCommand("disable"))
	r.terx.AddHandler(r.setTimer, terx.FilterIsMessage(), terx.FilterNotCommand())

	err = r.RestartTimers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to restart timers")
	}

	return r, nil
}

func (r *Service) Start(ctx context.Context) {
	r.terx.PollerRun(ctx)
}
