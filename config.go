package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	FocusDuration time.Duration `mapstructure:"focus-duration"`
	BreakDuration time.Duration `mapstructure:"break-duration"`
	SoundConfig   SoundConfig   `mapstructure:"sound"`
}

func loadConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.pomo")
	viper.SetDefault("focus-duration", 25*time.Minute)
	viper.SetDefault("break-duration", 5*time.Minute)
	viper.SetDefault("sound.sound-enabled", true)
	viper.SetEnvPrefix("POMO")
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return config, nil
}
