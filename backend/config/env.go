package config

import "github.com/joho/godotenv"

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		Logger.Warn().Msg("Warning: .env file not found, using default environment variables")
	}
}
