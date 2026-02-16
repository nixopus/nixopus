package live

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Writer provides styled terminal output methods.
// Each method is a thin wrapper around fmt.Fprintf with lipgloss coloring.
type Writer struct {
	term *Terminal
}

// NewWriter creates a writer backed by the given terminal.
func NewWriter(term *Terminal) *Writer {
	return &Writer{term: term}
}

// style helpers — lipgloss auto-disables colors when stdout is not a TTY.
var (
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boldStyle   = lipgloss.NewStyle().Bold(true)
)

// Success prints "  > text" in green. Used for completed steps and agent acknowledgments.
func (w *Writer) Success(text string) {
	w.term.Println(fmt.Sprintf("  %s %s", greenStyle.Render(">"), text))
}

// Error prints "  x text" in red. Used for failures and missing items.
func (w *Writer) Error(text string) {
	w.term.Println(fmt.Sprintf("  %s %s", redStyle.Render("x"), text))
}

// Warning prints "  ! text" in yellow. Used for non-blocking issues.
func (w *Writer) Warning(text string) {
	w.term.Println(fmt.Sprintf("  %s %s", yellowStyle.Render("!"), text))
}

// Progress overwrites the current line with "  . text" in dim.
// Used for in-progress status (connecting, syncing, building).
// Call FinishProgress when the operation completes.
func (w *Writer) Progress(text string) {
	w.term.OverwriteLine(fmt.Sprintf("  %s %s", dimStyle.Render("."), dimStyle.Render(text)))
}

// FinishProgress replaces the in-place progress line with a green success line.
func (w *Writer) FinishProgress(text string) {
	w.term.FinishLine(fmt.Sprintf("  %s %s", greenStyle.Render(">"), text))
}

// Detail prints indented text (4 spaces). Used for agent explanations and multi-line output.
func (w *Writer) Detail(text string) {
	w.term.Println(fmt.Sprintf("    %s", text))
}

// Blank prints an empty line.
func (w *Writer) Blank() {
	w.term.Println("")
}

// Header prints bold text. Used once at startup for the banner.
func (w *Writer) Header(text string) {
	w.term.Println(fmt.Sprintf("  %s", boldStyle.Render(text)))
}

// UserInput prints what the user typed, styled as "you: text".
func (w *Writer) UserInput(text string) {
	w.term.Println(fmt.Sprintf("\n%s %s\n", dimStyle.Render("you:"), text))
}
