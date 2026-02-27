package setenv

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SetEnvModel represents the bubbletea model for the set-env command
type SetEnvModel struct {
	quitting bool
	done     bool
	errorMsg string

	// Results
	envPath string
}

// NewSetEnvModel creates a new set-env model
func NewSetEnvModel() SetEnvModel {
	return SetEnvModel{}
}

// Init initializes the model
func (m SetEnvModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m SetEnvModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case SetEnvSuccessMsg:
		m.done = true
		m.envPath = msg.EnvPath
		return m, nil

	case SetEnvErrorMsg:
		m.errorMsg = msg.Error
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m SetEnvModel) View() string {
	if m.quitting {
		if m.errorMsg != "" {
			return fmt.Sprintf("\n  %s %s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"), lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errorMsg))
		}
		if m.done {
			// Show success message before quitting
			return m.renderSuccess()
		}
		return "\n  Stopping...\n\n"
	}

	// Banner
	banner := m.renderBanner()

	// Success message if done, otherwise show progress
	var content string
	if m.done {
		successMsg := m.renderSuccess()
		content = lipgloss.JoinVertical(lipgloss.Left, banner, "", successMsg)
	} else {
		progress := m.renderProgress()
		content = lipgloss.JoinVertical(lipgloss.Left, banner, "", progress)
	}

	return content
}

// renderBanner renders the banner
func (m SetEnvModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(55)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render("Nixopus Set Env")

	return bannerStyle.Render(title)
}

// renderProgress renders the progress box
func (m SetEnvModel) renderProgress() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(55).
		MaxWidth(55)

	var indicator string
	message := "Setting environment file path..."
	if m.done {
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	} else {
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⏳")
	}

	content := fmt.Sprintf("%s %s", indicator, message)
	return boxStyle.Render(content)
}

// renderSuccess renders the success message
func (m SetEnvModel) renderSuccess() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(55)

	contentWidth := 51

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓ Environment file path set successfully!"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Config updated and saved to: .nixopus"),
	}

	if m.envPath != "" {
		lines = append(lines, "")
		envText := fmt.Sprintf("Environment file: %s", m.envPath)
		envLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Width(contentWidth).
			Render(envText)
		lines = append(lines, envLine)
	}

	lines = append(lines, "")
	nextStepText := "Run 'nixopus live' to start deployment with the new env file"
	nextStep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Width(contentWidth).
		Render(nextStepText)
	lines = append(lines, nextStep)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// SetEnvProgram wraps the bubbletea program for the set-env command
type SetEnvProgram struct {
	program *tea.Program
}

// NewSetEnvProgram creates a new bubbletea program for the set-env command
func NewSetEnvProgram() *SetEnvProgram {
	model := NewSetEnvModel()
	p := tea.NewProgram(model)

	return &SetEnvProgram{
		program: p,
	}
}

// Start starts the program and returns when it exits
func (p *SetEnvProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program
func (p *SetEnvProgram) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program
func (p *SetEnvProgram) Quit() {
	p.program.Quit()
}

// SetEnvSuccessMsg is sent when set-env completes successfully
type SetEnvSuccessMsg struct {
	EnvPath string
}

// SetEnvErrorMsg is sent when set-env fails
type SetEnvErrorMsg struct {
	Error string
}
