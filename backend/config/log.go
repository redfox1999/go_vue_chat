package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	err := os.MkdirAll("logs", 0755)
	if err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	logFilePath := filepath.Join("logs", "app.log")
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, file)

	Logger = zerolog.New(multi).With().
		Timestamp().
		Caller().
		Logger()

	Logger.Info().Msg("Logger initialized, outputting to console and file")
}
