<img style="width:300px" src = "docs/pomo.png"/>

# Pomo 🍅

A terminal-based Pomodoro timer with a clean, intuitive interface.

![Pomo](docs/pomo.gif)

## Features

- 25-minute focus sessions
- 5-minute break sessions
- configurable session durations and custom sounds
- interactive progress bar

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

Then either run `./pomo` from the directory you cloned it to, or add it to your PATH to run from anywhere.

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

## Configuration

Pomo can be configured using either its YAML configuration file or environment variables.

### Configuration File

Pomo uses a configuration file located at `~/.pomo/config.yaml`. An example configuration file is provided below:

```yaml
focus-duration: 25m
break-duration: 5m
sound:
  enabled: true
  focus-start: https://example.com/focus.mp3
  focus-end: https://example.com/focus_end.mp3
  break-start: https://example.com/break_start.mp3
  break-end: https://example.com/break_end.mp3
  break-cancel: https://example.com/break_cancel.mp3
  focus-cancel: https://example.com/focus_cancel.mp3
```

### Using Environment Variables

You can also configure some Pomo settings via environment variables:

- `POMO_FOCUS_DURATION`: Duration of focus sessions (e.g., "25m")
- `POMO_BREAK_DURATION`: Duration of break sessions (e.g., "5m")
- `POMO_SOUND_ENABLED`: Enable/disable terminal bell sound (true/false)

Environment variables take precedence over values in the config file.

### Sounds

Impress those around you with custom sounds. By default Pomo will play the terminal bell when a session starts or ends, but you can also configure custom sounds for each event via the config file. Custom sounds can be provided as a remote URL or a local path. Sounds from remote URLs are downloaded to `~/.pomo/sounds` and cached there for future use. Only MP3 files are supported. See below for an example config that sets the break end sound to play one of the sample MP3s from the `soundz` directory on the `main` branch of this repository.

```yaml
sound:
  break-end: https://raw.githubusercontent.com/maxplumley/pomo/refs/heads/main/soundz/furby-uhoh.mp3
```

## Debugging 

Logs are written to `~/.pomo/_logs/pomo.log`. The default log level is `warn`, by you can increase log verbosity by passing the `-v` (or `-vv` for debug logging) flag to `pomo`.
