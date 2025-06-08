package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	FocusDuration time.Duration
	BreakDuration time.Duration
	SoundEnabled  bool
}

func loadConfig() Config {
	// Get the user's home directory
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return defaultConfig()
	}

	// Create default config if it doesn't exist
	configPath := filepath.Join(usr.HomeDir, ".pomorc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			fmt.Printf("Error creating default config: %v\n", err)
			return defaultConfig()
		}
	}

	// Load the config
	config := defaultConfig()
	if err := parseConfigFile(configPath, &config); err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
	}

	// Apply environment variables if set
	if focus := os.Getenv("POMO_FOCUS_DURATION"); focus != "" {
		if d, err := time.ParseDuration(focus); err == nil {
			config.FocusDuration = d
		}
	}
	if brk := os.Getenv("POMO_BREAK_DURATION"); brk != "" {
		if d, err := time.ParseDuration(brk); err == nil {
			config.BreakDuration = d
		}
	}
	if sound := os.Getenv("POMO_SOUND_ENABLED"); sound != "" {
		config.SoundEnabled = sound == "true"
	}

	return config
}

func defaultConfig() Config {
	return Config{
		FocusDuration: 25 * time.Minute,
		BreakDuration: 5 * time.Minute,
		SoundEnabled:  true,
	}
}

func createDefaultConfig(path string) error {
	defaultConfig := defaultConfig()
	return writeConfigFile(path, defaultConfig)
}

func writeConfigFile(path string, config Config) error {
	// Format durations as minutes for better readability
	focusMins := int(config.FocusDuration.Minutes())
	breakMins := int(config.BreakDuration.Minutes())

	content := fmt.Sprintf(`# Pomodoro Timer Configuration
focus-duration = %dm
break-duration = %dm
sound-enabled = %t
`,
		focusMins,
		breakMins,
		config.SoundEnabled)

	// Ensure the directory exists
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Write file with readable permissions
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	// Write content
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write config content: %v", err)
	}

	// Ensure file is readable
	if err := os.Chmod(path, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %v", err)
	}

	return nil
}

func parseConfigFile(path string, config *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "focus-duration":
			if d, err := time.ParseDuration(value); err == nil {
				config.FocusDuration = d
			}
		case "break-duration":
			if d, err := time.ParseDuration(value); err == nil {
				config.BreakDuration = d
			}
		case "sound-enabled":
			config.SoundEnabled = value == "true"
		}
	}

	return scanner.Err()
}
