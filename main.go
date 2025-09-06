package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

// Custom message types
type tickMsg struct{}

type endMsg struct{}

// Model represents our app's state
type Model struct {
	ending      bool
	lastTick    time.Time
	interactive bool
	pomCon      PomCon
	style       gloss.Style
	keys        keyMap
	help        help.Model
	config      Config
	sound       *SoundManager
	numberInput int
}

var (
	versionFlag = flag.Bool("version", false, "print pomo version")
	version     = "unknown"
)

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

func start(m *Model, p PomCon) tea.Cmd {
	m.pomCon = p
	current := p.Current(m.config.FocusDuration, m.config.BreakDuration)
	if current.t == FocusSession {
		return playSound(FocusStart, m.sound, nil)
	} else {
		return playSound(BreakStart, m.sound, nil)
	}
}

func end(m *Model) tea.Cmd {
	m.ending = true
	m.sound.Cleanup()
	return tea.Quit
}

func playSound(soundType SoundType, soundManager *SoundManager, endSound tea.Msg) tea.Cmd {
	return func() tea.Msg {
		sound := soundManager.PlaySound(soundType)
		if sound != nil {
			<-sound
		}
		return endSound
	}
}

func (m *Model) Init() tea.Cmd {
	m.lastTick = time.Now()
	if m.pomCon != nil {
		return tea.Sequence(tick(), start(m, m.pomCon))
	}
	return tick()
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
			if m.pomCon == nil {
				var pomCon PomCon = &PomConNode{
					focusDuration: DURATION_DEFAULT,
					breakDuration: DURATION_DEFAULT,
				}
				if pomosRequired > 1 {
					pomCon = &PomConRepeatNode{
						repeat: pomosRequired,
						child:  pomCon,
					}
				}
				return m, start(m, pomCon)
			}
		}
		if key.Matches(msg, keys.Focus) {
			m.numberInput = 0
			if m.pomCon == nil {
				return m, start(m, &PomConNode{
					focusDuration: DURATION_DEFAULT,
					breakDuration: DURATION_NOT_SET,
				})
			}
		}
		if key.Matches(msg, keys.Break) {
			m.numberInput = 0
			if m.pomCon == nil {
				return m, start(m, &PomConNode{
					focusDuration: DURATION_NOT_SET,
					breakDuration: DURATION_DEFAULT,
				})
			}
		}
		if key.Matches(msg, keys.End) {
			m.numberInput = 0
			if m.pomCon == nil {
				return m, nil
			}
			current := m.pomCon.Current(m.config.FocusDuration, m.config.BreakDuration)
			m.pomCon = nil
			var soundEnd tea.Msg = nil
			if !m.interactive {
				m.ending = true
				soundEnd = endMsg{}
			}
			if current.t == FocusSession {
				return m, playSound(FocusCancel, m.sound, soundEnd)
			}
			return m, playSound(BreakCancel, m.sound, soundEnd)
		}
		if key.Matches(msg, keys.Quit) {
			return m, end(m)
		}
		return m, nil
	case tickMsg:
		if m.pomCon == nil {
			m.lastTick = time.Now()
			return m, tick()
		}

		current := m.pomCon.Current(m.config.FocusDuration, m.config.BreakDuration)
		elapsed := time.Since(m.lastTick)
		m.lastTick = time.Now()
		m.pomCon.Tick(elapsed, m.config.FocusDuration, m.config.BreakDuration)
		next := m.pomCon.Current(m.config.FocusDuration, m.config.BreakDuration)

		if m.pomCon.Complete(m.config.FocusDuration, m.config.BreakDuration) {
			m.pomCon = nil
			var soundEnd tea.Msg = nil
			if !m.interactive {
				m.ending = true
				soundEnd = endMsg{}
			}
			if current.t == FocusSession {
				return m, tea.Sequence(tick(), playSound(FocusEnd, m.sound, soundEnd))
			}
			return m, tea.Sequence(tick(), playSound(BreakEnd, m.sound, soundEnd))
		}

		if current.t != next.t {
			if next.t == FocusSession {
				return m, tea.Sequence(tick(), playSound(FocusStart, m.sound, nil))
			}
			return m, tea.Sequence(tick(), playSound(BreakStart, m.sound, nil))
		}

		return m, tick()
	case endMsg:
		return m, end(m)
	}
	return m, nil
}

func (m *Model) View() string {
	ui := ""
	if m.pomCon == nil || m.pomCon.Complete(m.config.FocusDuration, m.config.BreakDuration) {
		ui = "pomo 🍅\n"
		if m.ending {
			ui = ui + "ending...\n"
		}
	} else {
		current := m.pomCon.Current(m.config.FocusDuration, m.config.BreakDuration)

		remaining := current.duration - current.elapsed

		minutes := remaining / time.Minute
		seconds := (remaining % time.Minute) / time.Second
		timeStr := fmt.Sprintf("%dm %ds", minutes, seconds)
		var s string
		if current.t == FocusSession {
			s = "focusing"
		} else {
			s = "recharging"
		}

		barWidth := 50

		printer := NewProgressPrinter(m.config.FocusDuration, m.config.BreakDuration)
		err := Crawl(m.pomCon, printer)
		if err != nil {
			log.Errorf("error printing pomcon %v", err)
		} else {
			cycle := printer.v.b.String()
			s = s + strings.Repeat(" ", barWidth-len(s)-gloss.Width(cycle)) + cycle
		}

		// Create progress bar string
		filled := int(float64(barWidth) * float64(current.elapsed) / float64(current.duration))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		ui = s + "\n" + bar + " " + timeStr
	}
	return m.style.Render(ui + "\n\n" + m.help.View(m.keys))
}

func initModel(config Config, soundManager *SoundManager, pomCon PomCon) Model {
	// Initialize style
	style := gloss.NewStyle().
		Padding(1, 2)

	interactive := true
	if pomCon != nil {
		interactive = false
	}

	model := Model{
		interactive: interactive,
		pomCon:      pomCon,
		style:       style,
		keys:        keys,
		help:        help.New(),
		config:      config,
		sound:       soundManager,
	}
	return model
}

func main() {

	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}
	err := initLogger()
	if err != nil {
		fmt.Println(fmt.Errorf("error initializing logger: %w", err))
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) > 1 {
		fmt.Printf("invalid arguments\n")
		os.Exit(1)
	}

	var pomCon PomCon
	if len(args) == 1 {
		pomCon, err = FromString(args[0])
		if err != nil {
			fmt.Printf("invalid pomcon: %v\n", err)
			os.Exit(1)
		}
	}

	config, err := loadConfig()
	if err != nil {
		log.Errorf("error loading config: %v", err)
		os.Exit(2)
	}

	// Initialize sound manager
	soundManager := NewSoundManager(config.SoundConfig)
	if err := soundManager.Init(); err != nil {
		log.Errorf("error initializing sound manager: %v", err)
		os.Exit(2)
	}

	model := initModel(config, soundManager, pomCon)

	// Start the program
	p := tea.NewProgram(&model)
	_, err = p.Run()
	if err != nil {
		fmt.Printf("error running program: %v\n", err)
		os.Exit(2)
	}
}
