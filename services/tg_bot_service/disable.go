package tg_bot_service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
	"regexp"
)

var disableRegexp = must_utils.Must(regexp.Compile(`^(?P<ID>.+)$`))

func (r *Context) Disable() error {
	groups := disableRegexp.FindStringSubmatch(r.text)
	if groups == nil {
		return r.replyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s", r.text, disableRegexp.String()),
		)
	}

	id, err := uuid.Parse(groups[disableRegexp.SubexpIndex("ID")])
	if err != nil {
		return r.replyWithClientErr(errors.Wrap(err, "bad id"))
	}

	ok, err := r.presentation.notifyRepository.CompleteTimer(r.ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to complete timer")
	}

	if !ok {
		return r.replyWithClientErr(errors.New("timer not found or already disabled"))
	}

	timer, err := r.presentation.notifyRepository.GetTimer(r.ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get timer")
	}

	zerolog.Ctx(r.ctx).Info().Interface("timer", timer).Msg("timer.completed")

	return r.reply(fmt.Sprintf("Timer completed in %d attempts!", timer.Attempt))
}
