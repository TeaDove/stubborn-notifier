package settings

import (
	"github.com/teadove/teasutils/utils/logger_utils"
	"github.com/teadove/teasutils/utils/settings_utils"
)

type tgSettings struct {
	BotToken string `env:"BOT_TOKEN" envDefault:"BAD_TOKEN"`
}

type baseSettings struct {
	DB string `env:"DB" envDefault:"./data/db.sqlite"`

	TG tgSettings `envPrefix:"TG__"`
}

// Settings
//nolint: gochecknoglobals // need it
var Settings = settings_utils.MustInitSetting[baseSettings](logger_utils.NewLoggedCtx(), "NOTIFY_", "TG.BotToken")
