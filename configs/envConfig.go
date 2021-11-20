package configs

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	EmailPass      string `mapstructure:"EMAIL_PASS"`
	EmailLogin     string `mapstructure:"EMAIL_LOGIN"`
	SmtpPort       int    `mapstructure:"EMAIL_SMTP_PORT"`
	SmtpHost       string `mapstructure:"EMAIL_SMTP_HOST"`
	EmailRecipient string `mapstructure:"EMAIL_RECIPIENT"`
}

func LoadEnvConfig() (Config, error) {
	if appEnv := os.Getenv("APP_ENV"); appEnv != "production" {
		viper.SetConfigName("dev")
		viper.AddConfigPath("../")
	} else {
		viper.SetConfigName("prod")
		viper.AddConfigPath("./")

	}

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, errors.Wrap(err, "failed to load envs")
	}

	viper.SetConfigType("env")
	viper.AutomaticEnv()

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, errors.Wrap(err, "failed to unmarshal envs")
	}

	return config, nil
}
