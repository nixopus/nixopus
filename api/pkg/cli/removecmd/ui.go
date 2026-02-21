package removecmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RemoveModel represents the bubbletea model for the remove command
type RemoveModel struct {
	quitting bool
	done     bool
	errorMsg string

	// Steps tracking
	currentStep int
	totalSteps  int
	stepMessage string

	// Results
	appName string
}

// NewRemoveModel creates a new remove model
func NewRemoveModel() RemoveModel {
	return RemoveModel{
		totalSteps: 3, // Load config, Delete application, Update config
	}
}

// Init initializes the model
func (m RemoveModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m RemoveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case RemoveStepMsg:
		m.currentStep = msg.Step
		m.stepMessage = msg.Message
		return m, nil

	case RemoveSuccessMsg:
		m.done = true
		m.appName = msg.AppName
		return m, nil

	case RemoveErrorMsg:
		m.errorMsg = msg.Error
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m RemoveModel) View() string {
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
func (m RemoveModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(55)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render("Nixopus Remove Application")

	return bannerStyle.Render(title)
}

// renderProgress renders the progress box
func (m RemoveModel) renderProgress() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(55)

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Removing application from family..."),
		"",
	}

	// Show step progress
	for i := 0; i < m.totalSteps; i++ {
		var stepText string
		if i < m.currentStep {
			stepText = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
		} else if i == m.currentStep {
			stepText = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("●")
		} else {
			stepText = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")
		}

		var stepDesc string
		switch i {
		case 0:
			stepDesc = "Loading configuration"
		case 1:
			stepDesc = "Deleting application"
		case 2:
			stepDesc = "Updating config file"
		}

		lines = append(lines, fmt.Sprintf("  %s  %s", stepText, stepDesc))
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
func (m RemoveModel) renderSuccess() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(55)

	contentWidth := 51

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓ Application removed successfully!"),
		"",
	}

	if m.appName != "" {
		nameText := fmt.Sprintf("Removed: %s", m.appName)
		nameLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(contentWidth).
			Render(nameText)
		lines = append(lines, nameLine)
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Config updated and saved to: .nixopus"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// RemoveProgram wraps the bubbletea program for the remove command
type RemoveProgram struct {
	program *tea.Program
}

// NewRemoveProgram creates a new bubbletea program for the remove command
func NewRemoveProgram() *RemoveProgram {
	model := NewRemoveModel()
	p := tea.NewProgram(model)

	return &RemoveProgram{
		program: p,
	}
}

// Start starts the program and returns when it exits
func (p *RemoveProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program
func (p *RemoveProgram) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program
func (p *RemoveProgram) Quit() {
	p.program.Quit()
}

// RemoveStepMsg is sent when a step progresses
type RemoveStepMsg struct {
	Step    int
	Message string
}

// RemoveSuccessMsg is sent when remove completes successfully
type RemoveSuccessMsg struct {
	AppName string
}

// RemoveErrorMsg is sent when remove fails
type RemoveErrorMsg struct {
	Error string
}
