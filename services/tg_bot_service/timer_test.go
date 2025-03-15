package tg_bot_service

import (
	"github.com/stretchr/testify/assert"
	"github.com/teadove/teasutils/utils/must_utils"
	"testing"
	"time"
)

func TestUnit_TimerParser_ParseWhole_Ok(t *testing.T) {
	in := `in 10m about "Buy coffee" every 24h at 10:00`
	req, err := parseIntoRequest(in)
	assert.NoError(t, err)
	assert.Equal(t, "Buy coffee", req.About)
	assert.Equal(t, 10*time.Minute, req.In)
	assert.Equal(t, 24*time.Hour, req.Every)
	assert.Equal(t, must_utils.Must(time.Parse("15:04", "10:00")), req.At)
}
