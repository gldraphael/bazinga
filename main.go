package main

import (
	"fmt"
	"math"
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
	animTick  int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)

	// Vibrant colors for the Bazinga! color-shifting animation
	colors = []string{"#EE6FF8", "#7D56F4", "#00FFF0", "#FF007A", "#74FF33", "#FFF333"}
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
			if m.timer >= 8000*time.Millisecond {
				m.state = stateFinal
				m.timer = 0 // Reset timer for the final screen wait
				m.animTick = 0
				return m, tick()
			}

			// Refined progress logic:
			// 1. First 7.5s: chaotic jumps up to 0.9.
			// 2. Last 0.5s: smooth finish to 1.0.
			if m.timer < 7500*time.Millisecond {
				// Chaotic progress: jump between -0.1 and 0.3
				increment := (rand.Float64() * 0.4) - 0.1
				m.percent += increment
				if m.percent < 0 {
					m.percent = 0
				}
				if m.percent > 0.9 {
					m.percent = 0.9
				}
			} else {
				// Final 500ms: Smooth finish from current % to 1.0
				remaining := float64(8000*time.Millisecond-m.timer) / float64(500*time.Millisecond)
				if remaining <= 0 {
					m.percent = 1.0
				} else {
					// Linearly interpolate current value to 1.0
					m.percent += (1.0 - m.percent) * 0.2 // Simple ease-out
				}
			}

			return m, tick()
		}

		if m.state == stateFinal {
			m.animTick++
			if m.timer >= 10000*time.Millisecond {
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

func (m model) getBazingaStyle() lipgloss.Style {
	color := colors[m.animTick%len(colors)]
	
	// Responsive font size simulation: adjust padding and border based on terminal size
	paddingX := 2
	paddingY := 1
	if m.width > 120 && m.height > 40 {
		paddingX = 10
		paddingY = 4
	} else if m.width > 80 && m.height > 20 {
		paddingX = 5
		paddingY = 2
	}

	// Throbbing effect: oscillate padding slightly using Sine
	throb := math.Abs(math.Sin(float64(m.animTick)*0.4)) * 2
	paddingX += int(throb)
	paddingY += int(throb / 2)

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color)).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(paddingY, paddingX).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(color)).
		Align(lipgloss.Center)
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
			m.getBazingaStyle().Render("BAZINGA!"),
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
