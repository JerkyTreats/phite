package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelFatal = "fatal"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func Debug(msg string, fields ...interface{}) {
	log.Debug().Fields(fields).Msg(msg)
}

func Info(msg string, fields ...interface{}) {
	log.Info().Fields(fields).Msg(msg)
}

func Error(err error, msg string, fields ...interface{}) {
	log.Error().Err(err).Fields(fields).Msg(msg)
}

func Fatal(err error, msg string, fields ...interface{}) {
	log.Fatal().Err(err).Fields(fields).Msg(msg)
}

func WithFields(fields ...interface{}) zerolog.Context {
	return log.With().Fields(fields)
}

// SetLevel sets the global logging level
func SetLevel(level string) error {
	switch level {
	case LevelDebug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case LevelInfo:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case LevelError:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case LevelFatal:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}
