package logger

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

type LogConfig struct {
	ConsoleOut io.Writer
	ConsoleErr io.Writer
	Debug      bool
	UseColor   bool
}

func New(config LogConfig) zerolog.Logger {

	var consoleOut io.Writer = os.Stdout
	var consoleErr io.Writer = os.Stderr

	if config.ConsoleOut != nil {
		consoleOut = config.ConsoleOut
	}

	if config.ConsoleErr != nil {
		consoleErr = config.ConsoleErr
	}
	writer := zerolog.MultiLevelWriter(
		levelWriter{
			Writer: zerolog.ConsoleWriter{Out: consoleOut, TimeFormat: time.RFC3339, NoColor: !config.UseColor},
			Levels: []zerolog.Level{
				zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
			},
		},
		levelWriter{
			Writer: zerolog.ConsoleWriter{Out: consoleErr, TimeFormat: time.RFC3339, NoColor: !config.UseColor},
			Levels: []zerolog.Level{
				zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
			},
		},
	)
	level := zerolog.InfoLevel
	if config.Debug {
		level = zerolog.DebugLevel
	}
	return zerolog.New(writer).With().Timestamp().Logger().Level(level)
}

type levelWriter struct {
	io.Writer
	Levels []zerolog.Level
}

func (w levelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	for _, l := range w.Levels {
		if l == level {
			return w.Write(p)
		}
	}
	return len(p), nil
}
