package main

import (
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

var testConfig = Config{
	FocusDuration: 2 * time.Second,
	BreakDuration: 2 * time.Second,
}

func TestFullOutputInteractive(t *testing.T) {
	m := initModel(testConfig, nil, nil)
	tm := teatest.NewTestModel(t, &m, teatest.WithInitialTermSize(300, 100))
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})

	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestFullOutputNonInteractive(t *testing.T) {
	m := initModel(testConfig, nil, &PomConNode{
		focusDuration: DURATION_DEFAULT,
		breakDuration: DURATION_DEFAULT,
	})
	tm := teatest.NewTestModel(t, &m, teatest.WithInitialTermSize(300, 100))
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})

	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}
