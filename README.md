# Pomodoro TUI

A terminal-based Pomodoro timer using the charmbracelet/bubbletea framework.

## Features

- 25-minute focus sessions
- 5-minute break sessions
- Terminal bell sound notification
- Clean, minimal interface

## Installation

1. Ensure you have Go installed on your system
2. Clone this repository
3. Run `go run main.go` in a terminal

## Usage

The application will automatically start a 25-minute focus session. When the time is up, it will:
1. Sound a terminal bell
2. Switch to a 5-minute break session
3. Repeat the cycle

Press 'q' to quit the application.

## Customization

The focus and break durations are defined as constants in the code and can be modified if desired.
