package logincmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoginModel represents the bubbletea model for the login command
type LoginModel struct {
	quitting bool
	done     bool

	// Steps tracking
	currentStep int
	totalSteps  int
	stepMessage string
	errorMsg    string

	// Results
	verificationURL string
	userCode        string
}

// NewLoginModel creates a new login model
func NewLoginModel() LoginModel {
	return LoginModel{
		totalSteps: 6, // Request code, Display URL, Display code, Open browser, Poll, Save token
	}
}

// Init initializes the model
func (m LoginModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case LoginStepMsg:
		m.currentStep = msg.Step
		m.stepMessage = msg.Message
		// Extract verification URL and user code from messages
		if msg.Step == 1 && msg.Message != "" {
			// Message format: "Visit: {url}"
			m.verificationURL = msg.Message
		}
		if msg.Step == 2 && msg.Message != "" {
			// Message format: "Enter code: {code}"
			m.userCode = msg.Message
		}
		return m, nil

	case LoginSuccessMsg:
		m.done = true
		return m, nil

	case LoginErrorMsg:
		m.errorMsg = msg.Error
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m LoginModel) View() string {
	if m.quitting {
		if m.errorMsg != "" {
			return fmt.Sprintf("\n  %s %s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"), lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errorMsg))
		}
		return "\n  Stopping...\n\n"
	}

	// Banner
	banner := m.renderBanner()

	// Progress steps
	progress := m.renderProgress()

	// Success message if done
	var successMsg string
	if m.done {
		successMsg = m.renderSuccess()
	}

	// Combine all views (no centering, left-aligned like init command)
	content := lipgloss.JoinVertical(lipgloss.Left, banner, "", progress)
	if successMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", successMsg)
	}

	return content
}

// renderBanner renders the banner
func (m LoginModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(55)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render("Nixopus Login")

	return bannerStyle.Render(title)
}

// renderProgress renders the progress steps
func (m LoginModel) renderProgress() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(55).
		MaxWidth(55)

	steps := []string{
		"Requesting device authorization...",
		"Displaying verification URL...",
		"Displaying user code...",
		"Opening browser...",
		"Waiting for authorization...",
		"Saving access token...",
	}

	lines := []string{}
	for i, step := range steps {
		var indicator string
		var stepText string
		if i < m.currentStep {
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
			stepText = step
		} else if i == m.currentStep {
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⏳")
			stepText = step
		} else {
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")
			stepText = step
		}
		lines = append(lines, fmt.Sprintf("%s %s", indicator, stepText))
	}

	if m.stepMessage != "" {
		lines = append(lines, "")
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(m.stepMessage)
		lines = append(lines, msg)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// renderSuccess renders the success message
func (m LoginModel) renderSuccess() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(55)

	contentWidth := 51

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓ Login successful!"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Access token saved to: .nixopus"),
		"",
	}

	nextStepText := "You can now use Nixopus commands"
	nextStep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Width(contentWidth).
		Render(nextStepText)
	lines = append(lines, nextStep)

	lines = append(lines, "")
	initText := "Run 'nixopus live' to initialize and start deployment"
	initLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(contentWidth).
		Render(initText)
	lines = append(lines, initLine)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// LoginProgram wraps the bubbletea program for the login command
type LoginProgram struct {
	program *tea.Program
}

// NewLoginProgram creates a new bubbletea program for the login command
func NewLoginProgram() *LoginProgram {
	model := NewLoginModel()
	p := tea.NewProgram(model)

	return &LoginProgram{
		program: p,
	}
}

// Start starts the program and returns when it exits
func (p *LoginProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program
func (p *LoginProgram) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program
func (p *LoginProgram) Quit() {
	p.program.Quit()
}

// LoginStepMsg is sent when a step progresses
type LoginStepMsg struct {
	Step    int
	Message string
}

// LoginSuccessMsg is sent when login completes successfully
type LoginSuccessMsg struct{}

// LoginErrorMsg is sent when login fails
type LoginErrorMsg struct {
	Error string
}
