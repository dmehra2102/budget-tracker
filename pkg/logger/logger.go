package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func New(level string) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}
