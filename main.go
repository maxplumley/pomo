package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

// Custom message types
type tickMsg struct{}

// Progress bar state
type progressBar struct {
	percent int
}

// AppState represents the current state of the application
type AppState string

const (
	AppStateWaiting AppState = "waiting"
	AppStateFocus   AppState = "focus"
	AppStateBreak   AppState = "break"
)

// Model represents our app's state
type Model struct {
	state     AppState
	startTime time.Time
	remaining time.Duration
	style     lipgloss.Style
	keys      keyMap
	help      help.Model
	progress  progressBar
	config    Config
	sound     *SoundManager
}

// Initialize progress bar
func (m *Model) initProgress() {
	m.progress.percent = 0
}

// Update progress bar
func (m *Model) updateProgress(duration time.Duration) {
	m.progress.percent = 100 - int((m.remaining.Seconds()/float64(duration.Seconds()))*100)
}

// Key bindings
type keyMap struct {
	Focus key.Binding
	Break key.Binding
	End   key.Binding
	Quit  key.Binding
	Help  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Focus, k.Break, k.End}, // first column
		{k.Quit, k.Help},          // second column
	}
}

var keys = keyMap{
	Focus: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "start focus session"),
	),
	Break: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "start break session"),
	),
	End: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "end current session"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q, ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{} })
}

func playSound(soundType SoundType, soundManager *SoundManager) tea.Cmd {
	return func() tea.Msg {
		sound := soundManager.PlaySound(soundType)
		<-sound
		return nil
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width

	case tea.KeyMsg:
		if key.Matches(msg, keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
		}
		if key.Matches(msg, keys.Focus) {
			if m.state == AppStateWaiting {
				m.state = AppStateFocus
				m.startTime = time.Now()
				m.remaining = m.config.FocusDuration
				m.initProgress()
				return m, tea.Batch(playSound(FocusStart, m.sound), tick())
			}
		}
		if key.Matches(msg, keys.Break) {
			if m.state == AppStateWaiting {
				m.state = AppStateBreak
				m.startTime = time.Now()
				m.remaining = m.config.BreakDuration
				m.initProgress()
				return m, tea.Batch(playSound(BreakStart, m.sound), tick())
			}
		}
		if key.Matches(msg, keys.End) {
			if m.state != AppStateWaiting {
				m.state = AppStateWaiting
				if m.state == AppStateFocus {
					return m, tea.Batch(playSound(FocusCancel, m.sound), tick())
				} else {
					return m, tea.Batch(playSound(BreakCancel, m.sound), tick())
				}
			}
			return m, nil
		}
		if key.Matches(msg, keys.Quit) {
			m.sound.Cleanup()
			return m, tea.Quit
		}
		return m, nil
	case tickMsg:
		if m.state == AppStateWaiting {
			return m, nil
		}

		elapsed := time.Since(m.startTime)
		if elapsed >= m.remaining {
			// Session complete, enter waiting state
			if m.state == AppStateFocus {
				m.state = AppStateWaiting
				return m, tea.Batch(playSound(FocusEnd, m.sound), tick())
			} else {
				m.state = AppStateWaiting
				return m, tea.Batch(playSound(BreakEnd, m.sound), tick())
			}
		}

		// Update remaining time
		m.remaining = m.remaining - elapsed
		m.startTime = time.Now()
		// Update progress bar
		if m.state == AppStateFocus {
			m.updateProgress(m.config.FocusDuration)
		} else {
			m.updateProgress(m.config.BreakDuration)
		}

		// Keep the timer running
		return m, tick()
	}
	return m, nil
}

func (m *Model) View() string {
	ui := ""
	if m.state == AppStateWaiting {
		ui = "pomo 🍅\n"
	} else {
		minutes := int(math.Ceil(m.remaining.Seconds())) / 60
		seconds := int(math.Ceil(m.remaining.Seconds())) % 60
		timeStr := fmt.Sprintf("%dm %ds", minutes, seconds)
		var s string
		if m.state == AppStateFocus {
			s = "focusing"
		} else {
			s = "recharging"
		}

		// Create progress bar string
		barWidth := 50
		filled := int(float64(barWidth) * float64(m.progress.percent) / 100.0)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		ui = s + "\n" + bar + " " + timeStr
	}
	return m.style.Render(ui + "\n\n" + m.help.View(m.keys))
}

func main() {
	err := initLogger()
	if err != nil {
		log.Errorf("error initializing logger: %v", err)
		os.Exit(1)
	}

	config, err := loadConfig()
	if err != nil {
		log.Errorf("error loading config: %v", err)
		os.Exit(1)
	}

	// Initialize style
	style := lipgloss.NewStyle().
		Padding(1, 2)

	// Initialize sound manager
	soundManager := NewSoundManager(config.SoundConfig)
	if err := soundManager.Init(); err != nil {
		log.Errorf("error initializing sound manager: %v", err)
		os.Exit(1)
	}

	// Initialize model
	model := Model{
		state:  AppStateWaiting,
		style:  style,
		keys:   keys,
		help:   help.New(),
		config: config,
		sound:  soundManager,
	}

	// Start the program
	p := tea.NewProgram(&model)
	_, err = p.Run()
	if err != nil {
		fmt.Printf("error running program: %v\n", err)
		os.Exit(1)
	}
}
