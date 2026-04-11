package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

type state int
type finalPhase int

const (
	stateInput state = iota
	stateAnimating
	stateFinal
)

const (
	frameDuration             = 50 * time.Millisecond
	processingDuration        = 8 * time.Second
	processingChaosDuration   = 7500 * time.Millisecond
	finalDuration             = 7 * time.Second
	finalWindupDuration       = 400 * time.Millisecond
	finalSlamDuration         = 900 * time.Millisecond
	finalOverreactionDuration = 1400 * time.Millisecond
	bazingaText               = "BAZINGA!"
)

const (
	finalPhaseWindup finalPhase = iota
	finalPhaseSlam
	finalPhaseOverreaction
	finalPhaseSettle
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

	finalPhase        finalPhase
	finalPhaseElapsed time.Duration
	finalScale        float64
	finalVelocity     float64
	finalWobble       float64
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

	slamSpring = harmonica.NewSpring(harmonica.FPS(int(time.Second/frameDuration)), 9.0, 0.24)
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
		if m.progress.Width < 10 {
			m.progress.Width = 10
		}
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
				m.animTick = 0
				return m, tick()
			}
		}

	case tickMsg:
		m.timer += frameDuration

		if m.state == stateAnimating {
			if m.timer >= processingDuration {
				m.enterFinalState()
				return m, tick()
			}

			if m.timer < processingChaosDuration {
				// Keep the fake progress erratic without overshooting too early.
				increment := ((rand.Float64() * 0.4) - 0.1) * 0.5
				m.percent += increment
				if m.percent < 0 {
					m.percent = 0
				}
				if m.percent > 0.9 {
					m.percent = 0.9
				}
			} else {
				m.percent += (1.0 - m.percent) * 0.25
				if m.percent > 1.0 {
					m.percent = 1.0
				}
			}

			return m, tick()
		}

		if m.state == stateFinal {
			m.advanceFinalAnimation()
			m.animTick++
			if m.timer >= finalDuration {
				return m.resetToInput()
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
	return tea.Tick(frameDuration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (p finalPhase) String() string {
	switch p {
	case finalPhaseWindup:
		return "windup"
	case finalPhaseSlam:
		return "slam"
	case finalPhaseOverreaction:
		return "overreaction"
	case finalPhaseSettle:
		return "settle"
	default:
		return "unknown"
	}
}

func (m *model) enterFinalState() {
	m.state = stateFinal
	m.timer = 0
	m.animTick = 0
	m.finalPhase = finalPhaseWindup
	m.finalPhaseElapsed = 0
	m.finalScale = 0.62
	m.finalVelocity = 0
	m.finalWobble = 0
}

func (m model) resetToInput() (tea.Model, tea.Cmd) {
	m.state = stateInput
	m.textInput.Reset()
	m.percent = 0
	m.timer = 0
	m.animTick = 0
	m.finalPhase = finalPhaseWindup
	m.finalPhaseElapsed = 0
	m.finalScale = 0
	m.finalVelocity = 0
	m.finalWobble = 0
	return m, m.textInput.Focus()
}

func (m *model) advanceFinalAnimation() {
	m.finalPhase, m.finalPhaseElapsed = finalPhaseForElapsed(m.timer)
	m.finalScale, m.finalVelocity = slamSpring.Update(
		m.finalScale,
		m.finalVelocity,
		finalScaleTarget(m.finalPhase),
	)

	amplitude := finalWobbleAmplitude(m.finalPhase, m.finalPhaseElapsed, m.width)
	m.finalWobble = math.Sin(float64(m.animTick)*0.8) * amplitude
}

func finalPhaseForElapsed(elapsed time.Duration) (finalPhase, time.Duration) {
	switch {
	case elapsed < finalWindupDuration:
		return finalPhaseWindup, elapsed
	case elapsed < finalWindupDuration+finalSlamDuration:
		return finalPhaseSlam, elapsed - finalWindupDuration
	case elapsed < finalWindupDuration+finalSlamDuration+finalOverreactionDuration:
		return finalPhaseOverreaction, elapsed - finalWindupDuration - finalSlamDuration
	default:
		return finalPhaseSettle, elapsed - finalWindupDuration - finalSlamDuration - finalOverreactionDuration
	}
}

func finalScaleTarget(phase finalPhase) float64 {
	switch phase {
	case finalPhaseWindup:
		return 0.88
	case finalPhaseSlam:
		return 1.85
	case finalPhaseOverreaction:
		return 1.14
	default:
		return 1.0
	}
}

func finalWobbleAmplitude(phase finalPhase, elapsed time.Duration, width int) float64 {
	widthBonus := 0.0
	if width > 110 {
		widthBonus = 1
	}

	switch phase {
	case finalPhaseSlam:
		return 2.5 + widthBonus
	case finalPhaseOverreaction:
		return 1 + widthBonus + 4*(1-phaseProgress(elapsed, finalOverreactionDuration))
	case finalPhaseSettle:
		return 0.4
	default:
		return 0
	}
}

func phaseProgress(elapsed, total time.Duration) float64 {
	if total <= 0 {
		return 1
	}

	progress := float64(elapsed) / float64(total)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func stylizedBazinga(scale float64, width int) string {
	gap := 0
	switch {
	case scale >= 1.65 && width > 80:
		gap = 2
	case scale >= 1.18 && width > 48:
		gap = 1
	}

	if gap == 0 {
		return bazingaText
	}

	spacer := strings.Repeat(" ", gap)
	glyphs := make([]string, 0, len(bazingaText))
	for _, r := range bazingaText {
		glyphs = append(glyphs, string(r))
	}

	return strings.Join(glyphs, spacer)
}

func (m model) finalBannerStyle(color, accentColor string) lipgloss.Style {
	paddingX := 1
	paddingY := 0
	if m.width > 70 {
		paddingX = 3
		paddingY = 1
	}
	if m.width > 110 && m.height > 28 {
		paddingX = 5
		paddingY = 2
	}

	if m.finalPhase == finalPhaseSlam && m.width > 60 {
		paddingX++
	}

	border := lipgloss.RoundedBorder()
	background := lipgloss.Color("#111111")
	foreground := lipgloss.Color(color)
	borderColor := lipgloss.Color(accentColor)

	switch m.finalPhase {
	case finalPhaseWindup:
		border = lipgloss.RoundedBorder()
	case finalPhaseSlam:
		border = lipgloss.ThickBorder()
		if m.animTick%2 == 0 {
			background = lipgloss.Color(color)
			foreground = lipgloss.Color("#101010")
			borderColor = lipgloss.Color("#FFF6AD")
		}
	case finalPhaseOverreaction:
		border = lipgloss.DoubleBorder()
	default:
		border = lipgloss.ThickBorder()
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(foreground).
		Background(background).
		Padding(paddingY, paddingX).
		Border(border).
		BorderForeground(borderColor).
		Align(lipgloss.Center)
}

func (m model) renderWindup() string {
	color := colors[m.animTick%len(colors)]
	dots := strings.Repeat(".", 3+(m.animTick%2))
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color)).
		Background(lipgloss.Color("#101010")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Align(lipgloss.Center).
		Render(dots)
}

func (m model) renderFinalScene() string {
	if m.finalPhase == finalPhaseWindup {
		return m.renderWindup()
	}

	color := colors[m.animTick%len(colors)]
	accentColor := colors[(m.animTick+2)%len(colors)]
	text := stylizedBazinga(m.finalScale, m.width)
	banner := m.finalBannerStyle(color, accentColor).Render(text)
	scene := m.composeImpactScene(banner, accentColor)
	return m.applyWobble(scene)
}

func (m model) composeImpactScene(banner, accentColor string) string {
	accentStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accentColor))
	quietStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))

	top := ""
	left := ""
	right := ""
	bottom := ""

	switch m.finalPhase {
	case finalPhaseSlam:
		if m.animTick%2 == 0 {
			top = accentStyle.Render("*  !  *  !")
			left = accentStyle.Render(">>")
			right = accentStyle.Render("<<")
			bottom = accentStyle.Render("!  *  !  *")
		} else {
			top = accentStyle.Render("! * ! * !")
			left = accentStyle.Render("!!")
			right = accentStyle.Render("!!")
			bottom = accentStyle.Render("*  !  *")
		}
	case finalPhaseOverreaction:
		if m.animTick%3 == 0 {
			top = accentStyle.Render("*   !   *")
			left = accentStyle.Render(">")
			right = accentStyle.Render("<")
			bottom = accentStyle.Render("!  *  !")
		} else {
			top = accentStyle.Render("!  *  !")
			left = accentStyle.Render(">>")
			right = accentStyle.Render("<<")
			bottom = accentStyle.Render("*   *")
		}
	default:
		top = quietStyle.Render("  *   !   *  ")
		left = quietStyle.Render(">")
		right = quietStyle.Render("<")
	}

	middle := banner
	if left != "" || right != "" {
		middle = lipgloss.JoinHorizontal(lipgloss.Center, left, " ", banner, " ", right)
	}

	stageWidth := lipgloss.Width(middle)
	stageWidth = maxInt(stageWidth, lipgloss.Width(top))
	stageWidth = maxInt(stageWidth, lipgloss.Width(bottom))

	lines := make([]string, 0, 3)
	if top != "" {
		lines = append(lines, lipgloss.PlaceHorizontal(stageWidth, lipgloss.Center, top))
	}
	lines = append(lines, lipgloss.PlaceHorizontal(stageWidth, lipgloss.Center, middle))
	if bottom != "" {
		lines = append(lines, lipgloss.PlaceHorizontal(stageWidth, lipgloss.Center, bottom))
	}

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}

func (m model) applyWobble(scene string) string {
	shift := int(math.Round(m.finalWobble))
	wobbleRoom := 4
	if m.width > 100 {
		wobbleRoom = 6
	}

	stageWidth := lipgloss.Width(scene) + wobbleRoom*2
	leftPadding := wobbleRoom + shift
	if leftPadding < 0 {
		leftPadding = 0
	}
	if leftPadding > wobbleRoom*2 {
		leftPadding = wobbleRoom * 2
	}

	shifted := lipgloss.NewStyle().PaddingLeft(leftPadding).Render(scene)
	return lipgloss.PlaceHorizontal(stageWidth, lipgloss.Left, shifted)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
			m.renderFinalScene(),
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
