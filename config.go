package main

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	FocusDuration time.Duration `mapstructure:"focus_duration"`
	BreakDuration time.Duration `mapstructure:"break_duration"`
	SoundEnabled  bool          `mapstructure:"sound_enabled"`
}

func loadConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.pomo")
	viper.SetDefault("focus_duration", 25*time.Minute)
	viper.SetDefault("break_duration", 5*time.Minute)
	viper.SetDefault("sound_enabled", true)
	viper.SetEnvPrefix("POMO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
