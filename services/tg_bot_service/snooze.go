package tg_bot_service

import (
	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/must_utils"
	"regexp"
	"time"
)

var snoozeRegexp = must_utils.Must(regexp.Compile(`^(?P<ID>.+) for (?P<Dur>.+)$`))

func (r *Context) snooze() error {
	groups := snoozeRegexp.FindStringSubmatch(r.text)
	if groups == nil {
		return r.replyWithClientErr(
			errors.Errorf("failed to match request, text: %s, expected: %s", r.text, timerRegexp.String()),
		)
	}

	dur, err := time.ParseDuration(groups[snoozeRegexp.SubexpIndex("Dur")])
	if err != nil {
		return r.replyWithClientErr(errors.Wrap(err, "bad duration"))
	}

	if dur > 15*time.Minute {
		return r.replyWithClientErr(errors.New("duration too far in the future"))
	}

	return r.reply("Snoozed!")
}
