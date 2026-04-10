package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateAnimating
	stateFinal
)

type tickMsg time.Time

type model struct {
	textInput textinput.Model
	progress  progress.Model
	state     state
	percent   float64
	timer     time.Duration
	quitting  bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	bazingaStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EE6FF8")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EE6FF8")).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something funny..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	p := progress.New(progress.WithDefaultGradient())

	return model{
		textInput: ti,
		progress:  p,
		state:     stateInput,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.state == stateInput {
				m.state = stateAnimating
				m.timer = 0
				m.percent = 0.1
				return m, tick()
			}
		}

	case tickMsg:
		if m.state == stateAnimating {
			m.timer += 100 * time.Millisecond
			if m.timer >= 2500*time.Millisecond {
				m.state = stateFinal
				return m, nil
			}

			// Chaotic progress: jump between -0.1 and 0.3
			increment := (rand.Float64() * 0.4) - 0.1
			m.percent += increment
			if m.percent < 0 {
				m.percent = 0
			}
			if m.percent > 1.0 {
				m.percent = 1.0
			}

			return m, tick()
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	var cmd tea.Cmd
	if m.state == stateInput {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var s string

	switch m.state {
	case stateInput:
		s = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("BAZINGA! CLI"),
			m.textInput.View(),
			helpStyle.Render("Press Enter to continue • Press q or Ctrl+C to quit"),
		)
	case stateAnimating:
		s = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Processing your genius..."),
			m.progress.ViewAs(m.percent),
			helpStyle.Render("Wait for it..."),
		)
	case stateFinal:
		s = fmt.Sprintf(
			"\n\n%s\n\n%s",
			bazingaStyle.Render("BAZINGA!"),
			helpStyle.Render("Press q to exit"),
		)
	}

	return s + "\n"
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
