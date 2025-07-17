package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env              string `json:"env" env:"ENV" env-required:"true"`
	ServerConfigPath string `env:"SERVER_CONFIG_PATH"`
	Database         struct {
		Postgres struct {
			Host           string `json:"host" env:"DATABASE_POSTGRES_HOST" env-required:"true"`
			Username       string `json:"username" env:"DATABASE_POSTGRES_USERNAME" env-required:"true"`
			Password       string `json:"password" env:"DATABASE_POSTGRES_PASSWORD" env-required:"true"`
			Database       string `json:"database" env:"DATABASE_POSTGRES_DATABASE" env-required:"true"`
			Port           string `json:"port" env:"DATABASE_POSTGRES_PORT" env-required:"true"`
			MigrationsPath string `json:"migrations_path" env:"DATABASE_POSTGRES_MIGRATIONS_PATH"`
		} `json:"postgres"`
	} `json:"database"`
	Telegram struct {
		Token string `json:"token" env:"TELEGRAM_TOKEN" env-required:"true"`
	} `json:"telegram"`
	OpenAI struct {
		Token string `json:"token" env:"OPEN_AI_TOKEN" env-required:"true"`
	} `json:"open_ai"`
	InfoChannels struct {
		General string `json:"general" env:"INFO_CHANNELS_GENERAL" env-required:"true"`
		Alarms  string `json:"alarms" env:"INFO_CHANNELS_ALARMS" env-required:"true"`
	} `json:"info_channels"`
}

var cfg Config

func MustSetup() {
	path := os.Getenv("SERVER_CONFIG_PATH")

	var err error
	if path == "" {
		err = cleanenv.ReadEnv(&cfg)
	} else {
		err = cleanenv.ReadConfig(path, &cfg)
	}

	if err != nil {
		panic(err)
	}
}

func IsDev() bool        { return cfg.Env == "dev" }
func IsStaging() bool    { return cfg.Env == "stage" }
func IsProduction() bool { return cfg.Env == "production" }
