package initcmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InitModel represents the bubbletea model for the init command
type InitModel struct {
	quitting bool
	done     bool

	// Steps tracking
	currentStep int
	totalSteps  int
	stepMessage string
	errorMsg    string

	// Results
	projectID string
	envPath   string
}

// NewInitModel creates a new init model
func NewInitModel() InitModel {
	return InitModel{
		totalSteps: 4, // Validate API key, Parse env (optional), Create project, Save config
	}
}

// Init initializes the model
func (m InitModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case InitStepMsg:
		m.currentStep = msg.Step
		m.stepMessage = msg.Message
		return m, nil

	case InitSuccessMsg:
		m.done = true
		m.projectID = msg.ProjectID
		m.envPath = msg.EnvPath
		return m, nil

	case InitErrorMsg:
		m.errorMsg = msg.Error
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI
func (m InitModel) View() string {
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

	// Combine all views (no centering, left-aligned like live command)
	content := lipgloss.JoinVertical(lipgloss.Left, banner, "", progress)
	if successMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", successMsg)
	}

	return content
}

// renderBanner renders the banner
func (m InitModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(55)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render("Nixopus Init")

	return bannerStyle.Render(title)
}

// renderProgress renders the progress steps
func (m InitModel) renderProgress() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(55).
		MaxWidth(55)

	steps := []string{
		"Validating API key...",
		"Parsing environment variables...",
		"Creating project...",
		"Saving configuration...",
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
func (m InitModel) renderSuccess() string {
	// Box width: 55 chars total
	// Account for: border (2 chars) + padding left (2) + padding right (2) = 6 chars
	// Content width: 55 - 6 = 49 chars
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(55)

	// Content width inside box: 55 - 4 (padding) = 51 chars
	contentWidth := 51

	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓ Initialized successfully!"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Config saved to: .nixopus"),
	}

	if m.projectID != "" {
		domainURL := buildDomainURL(m.projectID)
		if domainURL != "" {
			lines = append(lines, "")
			urlLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("🌐 Your app will be available at:")
			urlValue := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Width(contentWidth).
				Render("   " + domainURL)
			lines = append(lines, urlLabel, urlValue)
		}
	}

	if m.envPath != "" {
		lines = append(lines, "")
		envText := fmt.Sprintf("Environment: %s", m.envPath)
		envLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(contentWidth).
			Render(envText)
		lines = append(lines, envLine)
	}

	lines = append(lines, "")
	// Wrap long text to fit within box width
	nextStepText := "Run 'nixopus live' to start deployment session"
	nextStep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Width(contentWidth).
		Render(nextStepText)
	lines = append(lines, nextStep)

	lines = append(lines, "")
	addAppsText := "To add more apps, use: nixopus add <path> <name>"
	addApps := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(contentWidth).
		Render(addAppsText)
	lines = append(lines, addApps)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// InitProgram wraps the bubbletea program for the init command
type InitProgram struct {
	program *tea.Program
}

// NewInitProgram creates a new bubbletea program for the init command
func NewInitProgram() *InitProgram {
	model := NewInitModel()
	p := tea.NewProgram(model)

	return &InitProgram{
		program: p,
	}
}

// Start starts the program and returns when it exits
func (p *InitProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program
func (p *InitProgram) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program
func (p *InitProgram) Quit() {
	p.program.Quit()
}

// InitStepMsg is sent when a step progresses
type InitStepMsg struct {
	Step    int
	Message string
}

// InitSuccessMsg is sent when init completes successfully
type InitSuccessMsg struct {
	ProjectID string
	EnvPath   string
}

// InitErrorMsg is sent when init fails
type InitErrorMsg struct {
	Error string
}

// buildDomainURL builds the domain URL from project ID
// Format: https://{first-8-chars-of-project-id}.nixopus.com
func buildDomainURL(projectID string) string {
	if projectID == "" || len(projectID) < 8 {
		return ""
	}
	// Take first 8 characters of project ID (UUID format)
	subdomain := projectID[:8]
	return "https://" + subdomain + ".nixopus.com"
}
