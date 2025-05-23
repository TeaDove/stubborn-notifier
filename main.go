package main

import (
	"stubborn-notifier/containers/app_container"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/logger_utils"
)

func main() {
	ctx := logger_utils.NewLoggedCtx()

	container, err := app_container.Build(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to build app container"))
	}

	container.Service.Start(ctx)
}
