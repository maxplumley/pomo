# Pomodoro TUI 🍅

A modern terminal-based Pomodoro timer with a clean, intuitive interface.

## Features

- 25-minute focus sessions ("Focusing" state)
- 5-minute break sessions ("Recharging" state)
- Terminal bell sound notification
- Interactive progress bar
- Help system (press '?' to toggle)
- Clean, minimal interface with emoji support

## Installation

1. Ensure you have Go installed on your system
2. Clone this repository
3. Run `go run main.go` in a terminal

## Configuration

The Pomodoro timer can be configured using either a `~/.pomorc` file or environment variables.

### Using ~/.pomorc

The first time you run the application, it will create a `~/.pomorc` file with default settings. You can customize it with the following format:

```ini
# Pomodoro Timer Configuration
focus-duration = 25m
break-duration = 5m
sound-enabled = true
```

### Using Environment Variables

You can also configure the timer using environment variables:

- `POMO_FOCUS_DURATION`: Duration of focus sessions (e.g., "25m")
- `POMO_BREAK_DURATION`: Duration of break sessions (e.g., "5m")
- `POMO_SOUND_ENABLED`: Enable/disable terminal bell sound (true/false)

Environment variables take precedence over values in the config file.

## Usage

The application displays a clean interface with the current state (Focusing/Recharging) and a progress bar showing remaining time.

### Key Bindings
- `f`: Start focus session
- `b`: Start break session
- `e`: End current session
- `q`: Quit application
- `?`: Toggle help view

The timer will automatically switch between focus and break sessions, with a terminal bell sound notification when time is up.

## Customization

The focus and break durations are defined as constants in the code and can be modified if desired. The application uses mathematical ceiling for time calculations to ensure accurate timing.
