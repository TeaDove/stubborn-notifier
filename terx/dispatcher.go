package terx

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
	"sync"
)

type ProcessorFunc func(r *Context) error
type FilterFunc func(r *Context) bool

type Handler struct {
	Filters   []FilterFunc
	Processor ProcessorFunc
}

func (r *Terx) PollerRun(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	// TODO move to settings
	u.Timeout = 10
	updates := r.Bot.GetUpdatesChan(u)

	zerolog.Ctx(ctx).
		Info().
		Interface("handlers", len(r.Handlers)).
		Msg("bot.polling.started")

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

func (r *Terx) processUpdate(ctx context.Context, wg *sync.WaitGroup, update *tgbotapi.Update) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wg.Done()

	c := r.makeCtx(ctx, update)
	logger := c.Log()

	for _, handler := range r.Handlers {
		ok := true
		for _, filter := range handler.Filters {
			if !filter(c) {
				ok = false
				break
			}
		}

		if ok {
			err := handler.Processor(c)
			if err != nil {
				err = errors.Wrap(err, "failed to process handler")

				logger.Error().
					Stack().Err(err).
					Type("processor", handler.Processor).
					Interface("update", update).
					Msg("failed.to.process.handler")

				if r.replyWithErr {
					c.TryReplyOnErr(err)
				}
			}

			logger.Debug().Msg("handler.processed")
		}

	}

	return nil
}
