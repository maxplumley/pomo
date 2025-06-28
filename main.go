package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
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
	state         AppState
	startTime     time.Time
	timeRemaining time.Duration
	style         lipgloss.Style
	keys          keyMap
	help          help.Model
	progress      progressBar
	config        Config
	sound         *SoundManager
	numberInput   int
	pomosRequired int
	pomo          int
}

// Initialize progress bar
func (m *Model) initProgress() {
	m.progress.percent = 0
}

// Update progress bar
func (m *Model) updateProgress(duration time.Duration) {
	m.progress.percent = 100 - int((m.timeRemaining.Seconds()/float64(duration.Seconds()))*100)
}

// Key bindings
type keyMap struct {
	Number key.Binding
	Focus  key.Binding
	Break  key.Binding
	Pomo   key.Binding
	End    key.Binding
	Quit   key.Binding
	Help   key.Binding
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
		{k.Focus, k.Break, k.End, k.Pomo}, // first column
		{k.Quit, k.Help},                  // second column
	}
}

var keys = keyMap{
	Number: key.NewBinding(
		key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("", ""),
	),
	Pomo: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("[n]p", "start [n] pomo sessions\n(focus, followed by break session)"),
	),
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

func startFocus(m *Model) tea.Cmd {
	m.startTime = time.Now()
	m.timeRemaining = m.config.FocusDuration
	m.state = AppStateFocus
	m.initProgress()
	return playSound(FocusStart, m.sound)
}

func endFocus(m *Model) tea.Cmd {
	m.timeRemaining = 0
	if m.pomosRequired > 0 {
		return startBreak(m)
	}
	m.state = AppStateWaiting
	return playSound(FocusEnd, m.sound)
}

func cancelFocus(m *Model) tea.Cmd {
	m.timeRemaining = 0
	m.state = AppStateWaiting
	return playSound(FocusCancel, m.sound)
}

func startBreak(m *Model) tea.Cmd {
	m.startTime = time.Now()
	m.timeRemaining = m.config.BreakDuration
	m.state = AppStateBreak
	m.initProgress()
	return playSound(BreakStart, m.sound)
}

func endBreak(m *Model) tea.Cmd {
	m.timeRemaining = 0
	m.pomo = m.pomo + 1
	if m.pomosRequired > 0 && m.pomo <= m.pomosRequired {
		return startFocus(m)
	}
	m.state = AppStateWaiting
	return playSound(BreakEnd, m.sound)
}

func cancelBreak(m *Model) tea.Cmd {
	m.timeRemaining = 0
	m.state = AppStateWaiting
	return playSound(BreakCancel, m.sound)
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
		if key.Matches(msg, keys.Number) {
			i, err := strconv.Atoi(msg.String())
			if err != nil {
				log.Error("parsing number input", err)
			}
			m.numberInput = 10*m.numberInput + i
		}
		if key.Matches(msg, keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
		}
		if key.Matches(msg, keys.Pomo) {
			pomosRequired := m.numberInput
			m.numberInput = 0
			if m.state == AppStateWaiting {
				m.pomosRequired = max(1, pomosRequired)
				m.pomo = 1
				return m, tea.Batch(startFocus(m), tick())
			}
		}
		if key.Matches(msg, keys.Focus) {
			m.pomo = 0
			m.pomosRequired = 0
			m.numberInput = 0
			if m.state == AppStateWaiting {
				return m, tea.Batch(startFocus(m), tick())
			}
		}
		if key.Matches(msg, keys.Break) {
			m.pomo = 0
			m.pomosRequired = 0
			m.numberInput = 0
			if m.state == AppStateWaiting {
				return m, tea.Batch(startBreak(m), tick())
			}
		}
		if key.Matches(msg, keys.End) {
			m.pomo = 0
			m.pomosRequired = 0
			m.numberInput = 0
			if m.state == AppStateFocus {
				return m, tea.Batch(cancelFocus(m), tick())
			} else if m.state == AppStateBreak {
				return m, tea.Batch(cancelBreak(m), tick())
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
		if elapsed >= m.timeRemaining {
			// Session complete, enter waiting state
			if m.state == AppStateFocus {
				return m, tea.Batch(endFocus(m), tick())
			} else if m.state == AppStateBreak {
				return m, tea.Batch(endBreak(m), tick())
			}
		}

		// Update remaining time
		m.timeRemaining = m.timeRemaining - elapsed
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
		minutes := int(math.Ceil(m.timeRemaining.Seconds())) / 60
		seconds := int(math.Ceil(m.timeRemaining.Seconds())) % 60
		timeStr := fmt.Sprintf("%dm %ds", minutes, seconds)
		var s string
		if m.state == AppStateFocus {
			s = "focusing"
		} else {
			s = "recharging"
		}

		barWidth := 50
		if m.pomosRequired > 0 {
			cycle := fmt.Sprintf("%d/%d", m.pomo, m.pomosRequired)
			s = s + strings.Repeat(" ", barWidth-len(s)-len(cycle)) + cycle
		}

		// Create progress bar string
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
