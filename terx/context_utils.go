package terx

import (
	"fmt"
	"strings"
)

func extractCommandAndText(text string, botUsername string, isChat bool) (string, string) {
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

//func (r *Service) runCommand(c *Context) error {
//	handlers := map[string]func() error{
//		"notify":  c.Notify,
//		"timer":   c.Timer,
//		"disable": c.Disable,
//		"start":   func() error { return c.reply("https://crontab.guru/#0_9_*_*_*") },
//	}
//
//	handler, ok := handlers[c.command]
//	if !ok {
//		return c.replyWithClientErr(errors.New("command not found"))
//	}
//
//	err := handler()
//	if err != nil {
//		return errors.Wrap(err, "failed to execute command")
//	}
//
//	zerolog.Ctx(c.ctx).Debug().Msg("update.processed")
//	return nil
//}
