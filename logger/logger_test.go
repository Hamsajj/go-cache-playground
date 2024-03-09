package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew_MultiLevelWriter(t *testing.T) {

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	logConfig := LogConfig{
		UseColor:   false,
		ConsoleErr: errOut,
		ConsoleOut: out,
		Debug:      true,
	}
	logger := New(logConfig)

	logger.Info().Msg("this is an info level msg")
	logger.Debug().Msg("this is a debug level msg")
	logger.Warn().Msg("this is a warn level msg")
	logger.Error().Msg("this is an error level msg")

	assertContains(t, out, "INF this is an info level msg")
	assertContains(t, out, "DBG this is a debug level msg")
	assertContains(t, out, "WRN this is a warn level msg")

	assertContains(t, errOut, "ERR this is an error level msg")
}

func assertContains(t *testing.T, got *bytes.Buffer, expected string) {
	gotStr := got.String()
	if !strings.Contains(gotStr, expected) {
		t.Errorf("expected '%s' to contain '%s'", gotStr, expected)
	}
}
