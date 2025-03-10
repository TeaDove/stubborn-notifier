package tg_bot_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_TGBotPresentation_extractCommandAndTextPrivate_Ok(t *testing.T) {
	t.Parallel()

	command, text := extractCommandAndText("", "username", false)
	assert.Equal(t, "", command)
	assert.Equal(t, "", text)

	command, text = extractCommandAndText("bot", "username", false)
	assert.Equal(t, "", command)
	assert.Equal(t, "bot", text)

	command, text = extractCommandAndText("/", "username", false)
	assert.Equal(t, "", command)
	assert.Equal(t, "/", text)

	command, text = extractCommandAndText("/@", "username", false)
	assert.Equal(t, "", command)
	assert.Equal(t, "/@", text)

	command, text = extractCommandAndText("/start", "username", false)
	assert.Equal(t, "start", command)
	assert.Equal(t, "", text)

	command, text = extractCommandAndText("/start bot", "username", false)
	assert.Equal(t, "start", command)
	assert.Equal(t, "bot", text)

	command, text = extractCommandAndText("/start@username bot", "username", false)
	assert.Equal(t, "start", command)
	assert.Equal(t, "bot", text)
}

func TestUnit_TGBotPresentation_extractCommandAndTextChat_Ok(t *testing.T) {
	t.Parallel()

	command, text := extractCommandAndText("/start@username", "username", true)
	assert.Equal(t, "start", command)
	assert.Equal(t, "", text)

	command, text = extractCommandAndText("/start@username", "other_username", true)
	assert.Equal(t, "", command)
	assert.Equal(t, "/start@username", text)

	command, text = extractCommandAndText("/@username", "username", true)
	assert.Equal(t, "", command)
	assert.Equal(t, "/@username", text)

	command, text = extractCommandAndText("/start@username bot", "username", true)
	assert.Equal(t, "start", command)
	assert.Equal(t, "bot", text)

	command, text = extractCommandAndText("/start", "username", true)
	assert.Equal(t, "", command)
	assert.Equal(t, "/start", text)
}
