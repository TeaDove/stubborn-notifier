package terx

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

type Terx struct {
	Bot      *tgbotapi.BotAPI
	Handlers []Handler

	replyWithErr bool
}

type Config struct {
	Token        string
	ReplyWithErr bool
}

func New(config *Config) (*Terx, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bot client")
	}

	terx := &Terx{
		Bot:          bot,
		Handlers:     make([]Handler, 0),
		replyWithErr: config.ReplyWithErr,
	}

	return terx, nil
}

func (r *Terx) AddHandler(processor ProcessorFunc, filters ...FilterFunc) {
	if processor == nil {
		panic("processor cannot be nil")
	}

	r.Handlers = append(r.Handlers, Handler{
		Filters:   filters,
		Processor: processor,
	})
}
