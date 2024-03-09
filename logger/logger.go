package logger

import (
	"github.com/rs/zerolog"
	"io"
	"time"
)

func New(stdout io.Writer, stderr io.Writer) zerolog.Logger {
	writer := zerolog.MultiLevelWriter(
		levelWriter{
			Writer: zerolog.ConsoleWriter{Out: stdout, TimeFormat: time.RFC3339},
			Levels: []zerolog.Level{
				zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
			},
		},
		levelWriter{
			Writer: zerolog.ConsoleWriter{Out: stderr, TimeFormat: time.RFC3339},
			Levels: []zerolog.Level{
				zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
			},
		},
	)
	return zerolog.New(writer).With().Timestamp().Logger()
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
