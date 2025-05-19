package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
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
