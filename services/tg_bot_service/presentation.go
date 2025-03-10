package tg_bot_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"stubborn-notifier/repositories/notify_repository"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

const onErrorSleepDur = 20 * time.Second

type Service struct {
	bot       *tgbotapi.BotAPI
	scheduler *gocron.Scheduler

	notifyRepository *notify_repository.Repository

	timers   map[uuid.UUID]notify_repository.Timer
	timersMu sync.Mutex
}

func NewService(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	scheduler *gocron.Scheduler,
	notifyRepository *notify_repository.Repository,
) (*Service, error) {
	command := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "notify",
			Description: "Уведомить меня!",
		},
		tgbotapi.BotCommand{
			Command:     "timer",
			Description: "Поставить таймер!",
		},
		tgbotapi.BotCommand{
			Command:     "disable",
			Description: "Отключить уведомляшку",
		},
	)

	_, err := bot.Request(command)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set commands")
	}

	r := &Service{
		bot:              bot,
		scheduler:        scheduler,
		notifyRepository: notifyRepository,
		timers:           make(map[uuid.UUID]notify_repository.Timer, 10),
	}

	err = r.RestartTimers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to restart timers")
	}

	return r, nil
}

func (r *Service) PollerRun(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	// TODO move to settings
	u.Timeout = 10
	updates := r.bot.GetUpdatesChan(u)

	zerolog.Ctx(ctx).Info().Msg("bot.polling.started")

	var wg sync.WaitGroup

	for update := range updates {
		wg.Add(1)

		go must_utils.DoOrLogWithStacktrace(
			func(ctx context.Context) error {
				defer func() {
					err := must_utils.AnyToErr(recover())
					if err == nil {
						return
					}

					zerolog.Ctx(ctx).
						Error().
						Stack().Err(err).
						Interface("update", update).
						Msg("panic.in.process.update")
				}()

				return r.processUpdate(ctx, &wg, &update)
			},
			"error.during.update.process",
		)(ctx)
	}

	wg.Wait()
}

func extractCommandAndText(text string, botUsername string, isChat bool) (string, string) {
	// TODO move to other module
	if len(text) <= 1 || text[0] != '/' || strings.HasPrefix(text, "/@") {
		return "", text
	}

	spaceIdx := strings.Index(text, " ")

	atIdx := strings.Index(text, "@")
	if atIdx == -1 && isChat {
		return "", text
	}

	if atIdx != -1 && (spaceIdx == -1 || spaceIdx > atIdx) {
		var extractedUsername string
		if spaceIdx == -1 {
			extractedUsername = text[atIdx:]
		} else {
			extractedUsername = text[atIdx:spaceIdx]
		}

		if extractedUsername == fmt.Sprintf("@%s", botUsername) {
			if spaceIdx == -1 {
				return text[1:atIdx], ""
			}

			return text[1:atIdx], text[spaceIdx+1:]
		} else {
			return "", text
		}
	}

	if spaceIdx == -1 {
		return text[1:], ""
	} else {
		return text[1:spaceIdx], text[spaceIdx+1:]
	}
}

func (r *Service) runCommand(c *Context) error {
	handlers := map[string]func() error{
		"notify":  c.Notify,
		"timer":   c.Timer,
		"disable": c.Disable,
		"start":   func() error { return c.reply("https://crontab.guru/#0_9_*_*_*") },
	}

	handler, ok := handlers[c.command]
	if !ok {
		return c.replyWithClientErr(errors.New("command not found"))
	}

	err := handler()
	if err != nil {
		return errors.Wrap(err, "failed to execute command")
	}

	zerolog.Ctx(c.ctx).Debug().Msg("update.processed")
	return nil
}

func (r *Service) processCallbackQuery(c *Context) error {
	// Игноринг из-за бага в tgbotapi...
	_, _ = r.bot.Send(tgbotapi.NewCallback(c.update.CallbackQuery.ID, "Processing!"))

	var timerReq struct {
		Timer string `json:"timer"`
	}

	err := json.Unmarshal([]byte(c.update.CallbackQuery.Data), &timerReq)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal callback query data")
	}

	in, err := time.ParseDuration(timerReq.Timer)
	if err != nil {
		return errors.Wrap(err, "failed to parse timer duration")
	}

	timer, err := r.notifyRepository.CreateTimer(c.ctx, c.chat.ID, "Time elaplsed!", in, time.Minute*1)
	if err != nil {
		return errors.Wrap(err, "failed to create timer")
	}

	return c.scheduleTimer(timer)
}

func (r *Service) processUpdate(ctx context.Context, wg *sync.WaitGroup, update *tgbotapi.Update) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wg.Done()

	c := r.makeCtx(ctx, update)

	if update.CallbackQuery != nil {
		err := r.processCallbackQuery(&c)
		if err != nil {
			return errors.Wrap(err, "failed to process callback query")
		}

		return nil
	}

	// TODO set advected commands
	err := r.runCommand(&c)
	if err != nil {
		return errors.Wrap(err, "failed to execute command")
	}

	return nil
}
