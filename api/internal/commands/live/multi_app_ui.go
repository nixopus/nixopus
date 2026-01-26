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

// MultiAppModel represents the bubbletea model for multi-app live session UI
type MultiAppModel struct {
	tracker        *mover.MultiAppTracker
	quitting       bool
	isInitializing bool
	initStartTime  time.Time
	showLogs       bool
	showShortcuts  bool
	logsViewport   viewport.Model
	selectedApp    string // Currently selected app for logs view
}

// NewMultiAppModel creates a new multi-app bubbletea model
func NewMultiAppModel(tracker *mover.MultiAppTracker) MultiAppModel {
	vp := viewport.New(logsViewportWidth, logsViewportHeight)
	return MultiAppModel{
		tracker:        tracker,
		isInitializing: true,
		initStartTime:  time.Now(),
		logsViewport:   vp,
	}
}

// Init initializes the model and starts the tick command
func (m MultiAppModel) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the model state
func (m MultiAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// handleWindowResize updates viewport dimensions
func (m MultiAppModel) handleWindowResize(msg tea.WindowSizeMsg) (MultiAppModel, tea.Cmd) {
	if m.showLogs {
		m.logsViewport.Width = logsViewportWidth
		m.logsViewport.Height = msg.Height - headerReserveSpace - 10
	}
	return m, nil
}

// handleKeyPress processes keyboard input
func (m MultiAppModel) handleKeyPress(msg tea.KeyMsg) (MultiAppModel, tea.Cmd) {
	key := msg.String()

	if m.showLogs {
		if cmd := m.handleLogsScrolling(key); cmd != nil {
			return m, cmd
		}
	}

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

// handleLogsScrolling processes scroll commands
func (m MultiAppModel) handleLogsScrolling(key string) tea.Cmd {
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

// toggleLogsView switches between logs view and status view
func (m MultiAppModel) toggleLogsView() (MultiAppModel, tea.Cmd) {
	m.showLogs = !m.showLogs
	m.showShortcuts = false

	if m.showLogs {
		sessions := m.tracker.GetSessions()
		if len(sessions) > 0 && m.selectedApp == "" {
			m.selectedApp = sessions[0].Name
		}
		m.updateLogsViewport()
		m.logsViewport.GotoBottom()
	}

	return m, nil
}

// toggleShortcutsView switches shortcuts view
func (m MultiAppModel) toggleShortcutsView() (MultiAppModel, tea.Cmd) {
	m.showShortcuts = !m.showShortcuts
	m.showLogs = false
	return m, nil
}

// handleTick processes periodic refresh updates
func (m MultiAppModel) handleTick() (MultiAppModel, tea.Cmd) {
	sessions := m.tracker.GetSessions()
	
	// Transition from initialization to connected state
	if m.isInitializing {
		allConnected := true
		for _, session := range sessions {
			if session.Status != mover.ConnectionStatusConnected && session.Error == nil {
				allConnected = false
				break
			}
		}
		if allConnected && len(sessions) > 0 {
			m.isInitializing = false
		}
	}

	if m.showLogs {
		m.updateLogsViewport()
	}

	return m, tickCmd()
}

// updateLogsViewport refreshes the logs viewport content
func (m MultiAppModel) updateLogsViewport() {
	sessions := m.tracker.GetSessions()
	var selectedSession *mover.AppSessionInfo
	for _, session := range sessions {
		if session.Name == m.selectedApp {
			selectedSession = session
			break
		}
	}
	if selectedSession == nil && len(sessions) > 0 {
		selectedSession = sessions[0]
		m.selectedApp = sessions[0].Name
	}

	content := m.renderLogsContent(selectedSession)
	wasAtBottom := m.logsViewport.AtBottom()
	m.logsViewport.SetContent(content)
	if wasAtBottom {
		m.logsViewport.GotoBottom()
	}
}

// View renders the complete UI layout
func (m MultiAppModel) View() string {
	if m.quitting {
		return "\n  Stopping all apps...\n\n"
	}

	banner := m.renderBanner()

	if m.isInitializing {
		connectingBox := m.renderConnectingBox()
		return lipgloss.JoinVertical(lipgloss.Left, banner, "", connectingBox)
	}

	mainContent := m.renderMainContent()
	var helpCard string
	if !m.showShortcuts {
		helpCard = m.renderHelpCard()
	}

	return m.combineViews(banner, mainContent, helpCard)
}

// renderMainContent renders the appropriate main view
func (m MultiAppModel) renderMainContent() string {
	switch {
	case m.showShortcuts:
		return m.renderShortcutsView()
	case m.showLogs:
		return m.renderLogsView()
	default:
		return m.renderStatusBox()
	}
}

// combineViews combines banner, main content, and help card
func (m MultiAppModel) combineViews(banner, mainContent, helpCard string) string {
	components := []string{banner, "", mainContent}
	if helpCard != "" {
		components = append(components, "", helpCard)
	}
	return lipgloss.JoinVertical(lipgloss.Left, components...)
}

// renderBanner renders the top banner
func (m MultiAppModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorPrimary)).
		Padding(0, 1).
		Width(boxWidth)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorPrimary)).
		Render("Nixopus Live - All Apps")

	return bannerStyle.Render(title)
}

// renderStatusBox renders the main status dashboard for all apps
func (m MultiAppModel) renderStatusBox() string {
	boxStyle := m.createBoxStyle(colorMuted)
	sessions := m.tracker.GetSessions()

	lines := []string{
		fmt.Sprintf("Running %d app(s)", len(sessions)),
		"",
	}

	for _, session := range sessions {
		statusLine := m.formatAppStatus(session)
		lines = append(lines, statusLine)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Uptime: %s", m.formatUptime(m.tracker.GetUptime())))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// formatAppStatus formats a single app's status line
func (m MultiAppModel) formatAppStatus(session *mover.AppSessionInfo) string {
	statusIcon := m.formatConnectionStatusIcon(session.Status)
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorInfo))
	
	name := nameStyle.Render(session.Name)
	status := m.formatConnectionStatus(session.Status)
	
	line := fmt.Sprintf("  %s %s  %s", statusIcon, name, status)
	
	if session.Error != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
		line += "  " + errorStyle.Render("✗ "+session.Error.Error())
	} else {
		line += fmt.Sprintf("  [%d files synced]", session.FilesSynced)
		if session.URL != "" {
			urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
			line += "  " + urlStyle.Render(session.URL)
		}
	}
	
	return line
}

// formatConnectionStatusIcon returns an icon for connection status
func (m MultiAppModel) formatConnectionStatusIcon(status mover.ConnectionStatus) string {
	switch status {
	case mover.ConnectionStatusConnected:
		return "🟢"
	case mover.ConnectionStatusConnecting, mover.ConnectionStatusReconnecting:
		return "🟡"
	case mover.ConnectionStatusDisconnected:
		return "🔴"
	default:
		return "⚪"
	}
}

// formatConnectionStatus formats connection status with color
func (m MultiAppModel) formatConnectionStatus(status mover.ConnectionStatus) string {
	var text, color string

	switch status {
	case mover.ConnectionStatusConnected:
		text = "connected"
		color = colorSuccess
	case mover.ConnectionStatusConnecting:
		text = "connecting..."
		color = colorWarning
	case mover.ConnectionStatusReconnecting:
		text = "reconnecting..."
		color = colorWarning
	case mover.ConnectionStatusDisconnected:
		text = "disconnected"
		color = colorError
	default:
		text = "unknown"
		color = colorMuted
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// formatUptime formats duration as human-readable string
func (m MultiAppModel) formatUptime(d time.Duration) string {
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

// renderLogsView renders the logs view
func (m MultiAppModel) renderLogsView() string {
	boxStyle := m.createBoxStyle(colorMuted)
	sessions := m.tracker.GetSessions()
	
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Bold(true).
		Render("📋 Deployment Logs")

	// App selector
	var appSelector strings.Builder
	if len(sessions) > 1 {
		appSelector.WriteString("App: ")
		for i, session := range sessions {
			if i > 0 {
				appSelector.WriteString(" | ")
			}
			if session.Name == m.selectedApp {
				appSelector.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorPrimary)).Render(session.Name))
			} else {
				appSelector.WriteString(session.Name)
			}
		}
	}

	var selectedSession *mover.AppSessionInfo
	for _, session := range sessions {
		if session.Name == m.selectedApp {
			selectedSession = session
			break
		}
	}
	if selectedSession == nil && len(sessions) > 0 {
		selectedSession = sessions[0]
	}

	viewportContent := m.renderLogsContent(selectedSession)
	m.logsViewport.SetContent(viewportContent)
	footer := m.renderLogsFooter(selectedSession)

	components := []string{header}
	if appSelector.Len() > 0 {
		components = append(components, "", appSelector.String())
	}
	components = append(components, "", m.logsViewport.View(), "", footer)

	content := lipgloss.JoinVertical(lipgloss.Left, components...)
	return boxStyle.Render(content)
}

// renderLogsContent renders the scrollable content for logs
func (m MultiAppModel) renderLogsContent(session *mover.AppSessionInfo) string {
	if session == nil || session.Deployment == nil || len(session.Deployment.Logs) == 0 {
		noLogsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
		return noLogsStyle.Render("No logs available yet.")
	}

	var lines []string
	logStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorLogText)).Width(logsViewportWidth)

	for i, logLine := range session.Deployment.Logs {
		wrappedLines := wrapText(logLine, logLineWidth)
		for j, wrappedLine := range wrappedLines {
			if j == 0 {
				lines = append(lines, logStyle.Render(fmt.Sprintf("%3d. %s", i+1, wrappedLine)))
			} else {
				lines = append(lines, logStyle.Render(fmt.Sprintf("     %s", wrappedLine)))
			}
		}
	}

	return strings.Join(lines, "\n")
}

// renderLogsFooter renders the footer with scroll information
func (m MultiAppModel) renderLogsFooter(session *mover.AppSessionInfo) string {
	var footerLines []string

	if session != nil && session.Deployment != nil && len(session.Deployment.Logs) > 0 {
		totalLogs := len(session.Deployment.Logs)
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render(fmt.Sprintf("Showing %d logs", totalLogs))
		footerLines = append(footerLines, scrollInfo)
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	footerLines = append(footerLines, hintStyle.Render("↑/↓: scroll  g/G: top/bottom  l: back"))

	return strings.Join(footerLines, "  |  ")
}

// renderHelpCard renders a minimal help card
func (m MultiAppModel) renderHelpCard() string {
	cardStyle := m.createBoxStyle(colorMuted)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo)).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))

	lines := []string{
		keyStyle.Render("l") + " " + descStyle.Render("logs"),
		keyStyle.Render("Ctrl+C") + " " + descStyle.Render("stop all"),
		"",
		hintStyle.Render("Press 's' to see all shortcuts"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return cardStyle.Render(content)
}

// renderShortcutsView renders the full keyboard shortcuts reference
func (m MultiAppModel) renderShortcutsView() string {
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
		{"Ctrl+C", "Stop all apps"},
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

// renderConnectingBox renders the connecting message
func (m MultiAppModel) renderConnectingBox() string {
	boxStyle := m.createBoxStyle(colorWarning)
	dots := m.generateLoadingDots()

	lines := []string{
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarning)).
			Bold(true).
			Render("🟡 Establishing connections" + dots),
		"",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render("Setting up your live development sessions..."),
		"",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render("This will take just a few seconds"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// generateLoadingDots creates animated loading dots
func (m MultiAppModel) generateLoadingDots() string {
	elapsed := time.Since(m.initStartTime)
	dotCount := int(elapsed.Seconds()) % loadingDotCycle
	dots := ""
	for i := 0; i < dotCount; i++ {
		dots += "."
	}
	return dots
}

// createBoxStyle creates a reusable box style
func (m MultiAppModel) createBoxStyle(borderColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(boxWidth)
}

// MultiAppProgram wraps the bubbletea program for multi-app live session
type MultiAppProgram struct {
	program *tea.Program
	tracker *mover.MultiAppTracker
}

// NewMultiAppProgram creates a new bubbletea program for multi-app live session
func NewMultiAppProgram(tracker *mover.MultiAppTracker) *MultiAppProgram {
	model := NewMultiAppModel(tracker)
	p := tea.NewProgram(model, tea.WithAltScreen())

	return &MultiAppProgram{
		program: p,
		tracker: tracker,
	}
}

// Start starts the program and returns when it exits
func (p *MultiAppProgram) Start() error {
	_, err := p.program.Run()
	return err
}

// Quit quits the program
func (p *MultiAppProgram) Quit() {
	p.program.Quit()
}
