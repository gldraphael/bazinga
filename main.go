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
	width     int
	height    int
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
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress.Width = m.width - 20
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

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
		m.timer += 100 * time.Millisecond

		if m.state == stateAnimating {
			if m.timer >= 8*time.Second {
				m.state = stateFinal
				m.timer = 0 // Reset timer for the final screen wait
				return m, tick()
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

		if m.state == stateFinal {
			if m.timer >= 5*time.Second {
				m.state = stateInput
				m.textInput.Reset()
				m.percent = 0
				m.timer = 0
				return m, nil
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

	var content string

	switch m.state {
	case stateInput:
		content = fmt.Sprintf(
			"%s\n\n%s",
			m.textInput.View(),
			helpStyle.Render("Press Enter to continue • Press q or Ctrl+C to quit"),
		)
	case stateAnimating:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Processing your genius..."),
			m.progress.ViewAs(m.percent),
			helpStyle.Render("Wait for it..."),
		)
	case stateFinal:
		content = fmt.Sprintf(
			"%s\n\n%s",
			bazingaStyle.Render("BAZINGA!"),
			helpStyle.Render("Starting over soon..."),
		)
	}

	// Center the content on the screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
