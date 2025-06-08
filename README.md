# Pomo 🍅

A terminal-based Pomodoro timer with a clean, intuitive interface.

## Features

- 25-minute focus sessions ("Focusing" state)
- 5-minute break sessions ("Recharging" state)
- configurable session durations
- Terminal bell sound notification
- Interactive progress bar
- Help system (press '?' to toggle)
- Clean, minimal interface with emoji support

## Installation

### With the Pomo Homebrew tap

```shell
brew tap maxplumley/pomo
brew install pomo
```

### Or from source

Ensure you have Go (and Git) installed on your system, then run:

```shell
git clone https://github.com/maxplumley/pomo.git
cd pomo
go build .
```

Either run `./pomo` from the directory you cloned it to, or add it to your PATH to run from anywhere.

## Usage

Start Pomo:

```shell
pomo
```

### Key Bindings
- `f`: Start focus session
- `b`: Start break session
- `e`: End current session
- `q`: Quit Pomo
- `?`: Toggle help view

## Configuration

The Pomodoro timer can be configured using either a `~/.pomo/config.yaml` file or environment variables.

### Using YAML Configuration

Pomo can be configured by creating a `~/.pomo/config.yaml` file that sets custom values for features such as focus and break durations.

```yaml
# Pomodoro Timer Configuration
focus_duration: 25m
break_duration: 5m
sound_enabled: true
```

### Using Environment Variables

You can also configure Pomo via environment variables:

- `POMO_FOCUS_DURATION`: Duration of focus sessions (e.g., "25m")
- `POMO_BREAK_DURATION`: Duration of break sessions (e.g., "5m")
- `POMO_SOUND_ENABLED`: Enable/disable terminal bell sound (true/false)

Environment variables take precedence over values in the config file.

