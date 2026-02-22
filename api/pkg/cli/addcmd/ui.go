package addcmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AddModel represents the bubbletea model for the add command
type AddModel struct {
	quitting bool
	done     bool
	errorMsg string

	// Steps tracking
	currentStep int
	totalSteps  int
	stepMessage string

	// Results
	appName  string
	basePath string
}

// NewAddModel creates a new add model
func NewAddModel() AddModel {
	return AddModel{
		totalSteps: 3, // Validate config, Add application, Update config
	}
}

// Init initializes the model
func (m AddModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case AddStepMsg:
		m.currentStep = msg.Step
		m.stepMessage = msg.Message
		return m, nil

	case AddSuccessMsg:
		m.done = true
		m.appName = msg.AppName
		m.basePath = msg.BasePath
		return m, nil

	case AddErrorMsg:
		m.errorMsg = msg.Error
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m AddModel) View() string {
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
func (m AddModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(55)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render("Nixopus Add Application")

	return bannerStyle.Render(title)
}

// renderProgress renders the progress box
func (m AddModel) renderProgress() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(55)

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Adding application to family..."),
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
			stepDesc = "Validating configuration"
		case 1:
			stepDesc = "Adding application to family"
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
func (m AddModel) renderSuccess() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(55)

	contentWidth := 51

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓ Application added successfully!"),
		"",
	}

	if m.appName != "" {
		nameText := fmt.Sprintf("Name: %s", m.appName)
		nameLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(contentWidth).
			Render(nameText)
		lines = append(lines, nameLine)
	}

	if m.basePath != "" {
		pathText := fmt.Sprintf("Base path: %s", m.basePath)
		pathLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(contentWidth).
			Render(pathText)
		lines = append(lines, "", pathLine)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// AddProgram wraps the bubbletea program for the add command
type AddProgram struct {
	program *tea.Program
}

// NewAddProgram creates a new bubbletea program for the add command
func NewAddProgram() *AddProgram {
	model := NewAddModel()
	p := tea.NewProgram(model)

	return &AddProgram{
		program: p,
	}
}

// Start starts the program and returns when it exits
func (p *AddProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program
func (p *AddProgram) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program
func (p *AddProgram) Quit() {
	p.program.Quit()
}

// AddStepMsg is sent when a step progresses
type AddStepMsg struct {
	Step    int
	Message string
}

// AddSuccessMsg is sent when add completes successfully
type AddSuccessMsg struct {
	AppName  string
	BasePath string
}

// AddErrorMsg is sent when add fails
type AddErrorMsg struct {
	Error string
}
