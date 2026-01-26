package live

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raghavyuva/nixopus-api/internal/mover"
)

const (
	// UI dimensions
	boxWidth           = 55
	logsViewportWidth  = 51
	logsViewportHeight = 20
	logLineWidth       = 49
	headerReserveSpace = 10

	// Refresh interval
	tickInterval = time.Second

	// Animation
	loadingDotCycle = 4
)

// Color codes for terminal styling
const (
	colorPrimary = "63"  // Purple/magenta for primary elements
	colorSuccess = "42"  // Green for success states
	colorWarning = "220" // Yellow for warnings/loading
	colorError   = "196" // Red for errors
	colorInfo    = "39"  // Cyan for informational text
	colorMuted   = "240" // Gray for muted/secondary text
	colorLogText = "245" // Light gray for log text
)

// Key bindings
const (
	keyToggleLogs      = "l"
	keyToggleShortcuts = "s"
	keyQuit            = "q"
	keyScrollUp        = "up"
	keyScrollDown      = "down"
	keyPageUp          = "pgup"
	keyPageDown        = "pgdown"
	keyHome            = "home"
	keyEnd             = "end"
)

// Deployment status strings
const (
	statusDeployed  = "deployed"
	statusBuilding  = "building"
	statusDeploying = "deploying"
	statusCloning   = "cloning"
	statusFailed    = "failed"
	statusPending   = "pending"
	statusError     = "error"
	statusUnknown   = "unknown"
)

// Model represents the bubbletea model for the live session UI.
// It manages the state and rendering of the terminal user interface.
type Model struct {
	tracker        *mover.Tracker
	quitting       bool
	isInitializing bool
	initStartTime  time.Time
	showLogs       bool // Toggle for logs view
	showShortcuts  bool // Toggle for shortcuts view
	logsViewport   viewport.Model
}

// NewModel creates a new bubbletea model with initialized viewport.
func NewModel(tracker *mover.Tracker) Model {
	vp := viewport.New(logsViewportWidth, logsViewportHeight)
	return Model{
		tracker:        tracker,
		isInitializing: true,
		initStartTime:  time.Now(),
		logsViewport:   vp,
	}
}

// Init initializes the model and starts the tick command for periodic updates.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the model state.
// It processes window resize events, keyboard input, and periodic ticks.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case TickMsg:
		return m.handleTick()
	}

	return m, nil
}

// handleWindowResize updates viewport dimensions when window size changes.
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	if m.showLogs {
		m.logsViewport.Width = logsViewportWidth
		m.logsViewport.Height = msg.Height - headerReserveSpace
	}
	return m, nil
}

// handleKeyPress processes keyboard input for navigation and commands.
func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle scrolling when in logs view
	if m.showLogs {
		if cmd := m.handleLogsScrolling(key); cmd != nil {
			return m, cmd
		}
	}

	// Handle global commands
	switch key {
	case "ctrl+c", keyQuit:
		m.quitting = true
		return m, tea.Quit
	case keyToggleLogs, strings.ToUpper(keyToggleLogs):
		return m.toggleLogsView()
	case keyToggleShortcuts, strings.ToUpper(keyToggleShortcuts):
		return m.toggleShortcutsView()
	}

	return m, nil
}

// handleLogsScrolling processes scroll commands in logs view.
func (m Model) handleLogsScrolling(key string) tea.Cmd {
	switch key {
	case keyScrollUp, "k":
		m.logsViewport.LineUp(1)
		return nil
	case keyScrollDown, "j":
		m.logsViewport.LineDown(1)
		return nil
	case keyPageUp:
		m.logsViewport.PageUp()
		return nil
	case keyPageDown, " ":
		m.logsViewport.PageDown()
		return nil
	case keyHome, "g":
		m.logsViewport.GotoTop()
		return nil
	case keyEnd, "G":
		m.logsViewport.GotoBottom()
		return nil
	}
	return nil
}

// toggleLogsView switches between logs view and status view.
func (m Model) toggleLogsView() (Model, tea.Cmd) {
	m.showLogs = !m.showLogs
	m.showShortcuts = false

	if m.showLogs {
		statusInfo := m.tracker.GetStatusInfo()
		m.logsViewport.SetContent(m.renderLogsContent(statusInfo))
		m.logsViewport.GotoBottom() // Start at bottom to show latest logs
	}

	return m, nil
}

// toggleShortcutsView switches between shortcuts view and status view.
func (m Model) toggleShortcutsView() (Model, tea.Cmd) {
	m.showShortcuts = !m.showShortcuts
	m.showLogs = false
	return m, nil
}

// handleTick processes periodic refresh updates.
func (m Model) handleTick() (Model, tea.Cmd) {
	// Transition from initialization to connected state
	if m.isInitializing {
		statusInfo := m.tracker.GetStatusInfo()
		if statusInfo.ConnectionStatus == mover.ConnectionStatusConnected {
			m.isInitializing = false
		}
	}

	// Update logs viewport content if logs view is open
	if m.showLogs {
		m.updateLogsViewport()
	}

	return m, tickCmd()
}

// updateLogsViewport refreshes the logs viewport content while preserving scroll position.
func (m Model) updateLogsViewport() {
	statusInfo := m.tracker.GetStatusInfo()
	oldContent := m.logsViewport.View()
	newContent := m.renderLogsContent(statusInfo)

	if oldContent != newContent {
		wasAtBottom := m.logsViewport.AtBottom()
		m.logsViewport.SetContent(newContent)
		if wasAtBottom {
			m.logsViewport.GotoBottom()
		}
	}
}

// View renders the complete UI layout.
func (m Model) View() string {
	if m.quitting {
		return "\n  Stopping...\n\n"
	}

	banner := m.renderBanner()

	// Show initialization screen until connected
	if m.isInitializing {
		connectingBox := m.renderConnectingBox()
		return lipgloss.JoinVertical(lipgloss.Left, banner, "", connectingBox)
	}

	// Render main content based on current view mode
	mainContent := m.renderMainContent()

	// Add help card (hidden in shortcuts view)
	var helpCard string
	if !m.showShortcuts {
		helpCard = m.renderHelpCard()
	}

	// Combine all components
	return m.combineViews(banner, mainContent, helpCard)
}

// renderMainContent renders the appropriate main view based on current state.
func (m Model) renderMainContent() string {
	statusInfo := m.tracker.GetStatusInfo()

	switch {
	case m.showShortcuts:
		return m.renderShortcutsView()
	case m.showLogs:
		return m.renderLogsView(statusInfo)
	default:
		return m.renderStatusBox(statusInfo)
	}
}

// combineViews combines banner, main content, and optional help card.
func (m Model) combineViews(banner, mainContent, helpCard string) string {
	components := []string{banner, "", mainContent}
	if helpCard != "" {
		components = append(components, "", helpCard)
	}
	return lipgloss.JoinVertical(lipgloss.Left, components...)
}

// renderBanner renders the top banner with the application title.
func (m Model) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorPrimary)).
		Padding(0, 1).
		Width(boxWidth)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorPrimary)).
		Render("Nixopus Live")

	return bannerStyle.Render(title)
}

// renderStatusBox renders the main status dashboard with connection, sync, and deployment info.
func (m Model) renderStatusBox(info mover.StatusInfo) string {
	boxStyle := m.createBoxStyle(colorMuted)

	lines := m.buildStatusLines(info)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// buildStatusLines constructs the status information lines.
func (m Model) buildStatusLines(info mover.StatusInfo) []string {
	lines := []string{
		fmt.Sprintf("Status: %s", m.formatConnectionStatus(info.ConnectionStatus)),
		fmt.Sprintf("Files synced: %d", info.FilesSynced),
		fmt.Sprintf("Changes detected: %d", info.ChangesDetected),
		fmt.Sprintf("Uptime: %s", m.formatUptime(m.tracker.GetUptime())),
	}

	// Add environment file info if detected
	if info.EnvPath != "" {
		envStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
		lines = append(lines, fmt.Sprintf("Environment: %s", envStyle.Render(info.EnvPath)))
	}

	// Add deployment status section
	lines = append(lines, "")
	lines = append(lines, m.buildDeploymentStatusLines(info)...)

	// Add URL if available
	if info.URL != "" {
		lines = append(lines, "")
		urlLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render("🌐 Your app is live at:")
		urlValue := lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo)).Render(info.URL)
		lines = append(lines, urlLabel, "   "+urlValue)
	}

	return lines
}

// buildDeploymentStatusLines constructs deployment status information.
func (m Model) buildDeploymentStatusLines(info mover.StatusInfo) []string {
	if info.Deployment == nil {
		checkingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
		return []string{checkingStyle.Render("Deployment: Checking...")}
	}

	lines := []string{
		fmt.Sprintf("Deployment: %s", m.formatDeploymentStatus(info.Deployment.Status)),
	}

	if info.Deployment.Message != "" {
		messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
		lines = append(lines, messageStyle.Render("  "+info.Deployment.Message))
	}

	// Show hint to view logs if available
	if len(info.Deployment.Logs) > 0 {
		lines = append(lines, "")
		logsHint := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorInfo)).
			Render(fmt.Sprintf("Press 'l' to view logs (%d available)", len(info.Deployment.Logs)))
		lines = append(lines, logsHint)
	}

	return lines
}

// formatConnectionStatus formats connection status with appropriate color and icon.
func (m Model) formatConnectionStatus(connStatus mover.ConnectionStatus) string {
	var text, color string

	switch connStatus {
	case mover.ConnectionStatusConnected:
		text = "🟢 Connected"
		color = colorSuccess
	case mover.ConnectionStatusConnecting:
		text = "🟡 Connecting..."
		color = colorWarning
	case mover.ConnectionStatusReconnecting:
		text = "🟡 Reconnecting..."
		color = colorWarning
	case mover.ConnectionStatusDisconnected:
		text = "🔴 Disconnected"
		color = colorError
	default:
		text = "⚪ Unknown"
		color = colorMuted
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// formatUptime formats duration as human-readable uptime string.
func (m Model) formatUptime(d time.Duration) string {
	seconds := int(d.Seconds())

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
	}

	days := hours / 24
	remainingHours := hours % 24
	return fmt.Sprintf("%dd %dh", days, remainingHours)
}

// formatDeploymentStatus formats deployment status with appropriate color and icon.
func (m Model) formatDeploymentStatus(status string) string {
	var text, color string

	switch status {
	case statusDeployed:
		text = "✅ Deployed"
		color = colorSuccess
	case statusBuilding:
		text = "🔨 Building..."
		color = colorWarning
	case statusDeploying:
		text = "🚀 Deploying..."
		color = colorWarning
	case statusCloning:
		text = "📥 Cloning..."
		color = colorWarning
	case statusFailed:
		text = "❌ Failed"
		color = colorError
	case statusPending:
		text = "⏳ Pending"
		color = colorMuted
	case statusError:
		text = "⚠️ Error"
		color = colorError
	case statusUnknown:
		text = "❓ Unknown"
		color = colorMuted
	default:
		text = fmt.Sprintf("⚪ %s", status)
		color = colorMuted
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// renderLogsView renders the full logs view with scrollable viewport.
func (m Model) renderLogsView(info mover.StatusInfo) string {
	boxStyle := m.createBoxStyle(colorMuted)

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Bold(true).
		Render("📋 Deployment Logs ")

	// Update viewport content
	viewportContent := m.renderLogsContent(info)
	m.logsViewport.SetContent(viewportContent)

	footer := m.renderLogsFooter(info)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.logsViewport.View(),
		"",
		footer,
	)

	return boxStyle.Render(content)
}

// renderLogsContent renders the scrollable content for logs with line numbers.
func (m Model) renderLogsContent(info mover.StatusInfo) string {
	if info.Deployment == nil || len(info.Deployment.Logs) == 0 {
		noLogsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
		return noLogsStyle.Render("No logs available yet.")
	}

	var lines []string
	logStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorLogText)).Width(logsViewportWidth)

	for i, logLine := range info.Deployment.Logs {
		wrappedLines := wrapText(logLine, logLineWidth)
		for j, wrappedLine := range wrappedLines {
			if j == 0 {
				// First line with line number
				lines = append(lines, logStyle.Render(fmt.Sprintf("%3d. %s", i+1, wrappedLine)))
			} else {
				// Continuation lines with indentation
				lines = append(lines, logStyle.Render(fmt.Sprintf("     %s", wrappedLine)))
			}
		}
	}

	return strings.Join(lines, "\n")
}

// renderLogsFooter renders the footer with scroll information and navigation hints.
func (m Model) renderLogsFooter(info mover.StatusInfo) string {
	var footerLines []string

	if info.Deployment != nil && len(info.Deployment.Logs) > 0 {
		totalLogs := len(info.Deployment.Logs)
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render(fmt.Sprintf("Showing %d logs", totalLogs))
		footerLines = append(footerLines, scrollInfo)
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	footerLines = append(footerLines, hintStyle.Render("↑/↓: scroll  g/G: top/bottom  l: back"))

	return strings.Join(footerLines, "  |  ")
}

// wrapText wraps text to fit within the specified width, breaking long words if necessary.
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			// Break word if it's longer than width
			if len(word) > width {
				for len(word) > width {
					lines = append(lines, word[:width])
					word = word[width:]
				}
				currentLine = word
			} else {
				currentLine = word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// renderHelpCard renders a minimal help card with essential shortcuts.
func (m Model) renderHelpCard() string {
	cardStyle := m.createBoxStyle(colorMuted)

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))

	lines := []string{
		keyStyle.Render("l") + " " + descStyle.Render("logs"),
		keyStyle.Render("Ctrl+C") + " " + descStyle.Render("stop"),
		"",
		hintStyle.Render("Press 's' to see all shortcuts"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return cardStyle.Render(content)
}

// renderShortcutsView renders the full keyboard shortcuts reference view.
func (m Model) renderShortcutsView() string {
	boxStyle := m.createBoxStyle(colorPrimary)

	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true).
			Render("⌨️  Keyboard Shortcuts"),
		"",
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))

	shortcuts := []struct {
		key  string
		desc string
	}{
		{keyToggleLogs, "Toggle logs view"},
		{keyToggleShortcuts, "Toggle shortcuts view"},
		{"Ctrl+C", "Stop"},
		{keyQuit, "Quit"},
	}

	for _, shortcut := range shortcuts {
		keyText := keyStyle.Render(shortcut.key)
		descText := descStyle.Render(shortcut.desc)
		lines = append(lines, fmt.Sprintf("  %s  %s", keyText, descText))
	}

	lines = append(lines, "")
	backHint := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render("Press 's' to go back")
	lines = append(lines, backHint)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// Program wraps the bubbletea program for the live session UI.
type Program struct {
	program *tea.Program
	tracker *mover.Tracker
}

// NewProgram creates a new bubbletea program for the live session.
func NewProgram(tracker *mover.Tracker) *Program {
	model := NewModel(tracker)
	p := tea.NewProgram(model, tea.WithAltScreen())

	return &Program{
		program: p,
		tracker: tracker,
	}
}

// Start starts the program and returns when it exits.
func (p *Program) Start() error {
	_, err := p.program.Run()
	return err
}

// Send sends a message to the program.
func (p *Program) Send(msg tea.Msg) {
	p.program.Send(msg)
}

// Quit quits the program.
func (p *Program) Quit() {
	p.program.Quit()
}

// TickMsg is sent periodically to refresh the view.
type TickMsg struct {
	Time time.Time
}

// tickCmd creates a command that sends a tick message at regular intervals.
func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

// renderConnectingBox renders the connecting/initializing message with animated dots.
func (m Model) renderConnectingBox() string {
	boxStyle := m.createBoxStyle(colorWarning)

	// Animated loading dots based on elapsed time
	dots := m.generateLoadingDots()

	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarning)).
			Bold(true).
			Render("🟡 Establishing connection" + dots),
		"",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render("Setting up your live development session..."),
		"",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render("This will take just a few seconds"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// generateLoadingDots creates animated loading dots based on elapsed time.
func (m Model) generateLoadingDots() string {
	elapsed := time.Since(m.initStartTime)
	dotCount := int(elapsed.Seconds()) % loadingDotCycle
	dots := ""
	for i := 0; i < dotCount; i++ {
		dots += "."
	}
	return dots
}

// createBoxStyle creates a reusable box style with the specified border color.
func (m Model) createBoxStyle(borderColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(boxWidth)
}
