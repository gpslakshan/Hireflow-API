package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger configures the global zerolog logger.
// In development: pretty coloured console output.
// In production:  structured JSON output.
func InitLogger(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "production" {
		// JSON output — machine readable
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Pretty console output — human readable in development
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		).With().Timestamp().Caller().Logger()
	}
}
