package live

import (
	"fmt"
	"strings"

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

// Print writes text without a trailing newline. Use for inline prompts.
func (w *Writer) Print(text string) {
	w.term.Print(text)
}

// Println writes a line with newline.
func (w *Writer) Println(text string) {
	w.term.Println(text)
}

// Header prints bold text. Used once at startup for the banner.
func (w *Writer) Header(text string) {
	w.term.Println(fmt.Sprintf("  %s", boldStyle.Render(text)))
}

// UserInput prints what the user typed, styled as "you: text".
func (w *Writer) UserInput(text string) {
	w.term.Println(fmt.Sprintf("\n%s %s\n", dimStyle.Render("you:"), text))
}

// codeStyle for Dockerfile/code blocks.
var codeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

// ApprovalProposal renders a structured Dockerfile approval UI: header, summary,
// validation, suggestions, and the proposed Dockerfile in a readable code block.
func (w *Writer) ApprovalProposal(summary string, validationScore int, suggestions []string, dockerfile string) {
	w.Blank()
	w.term.Println(fmt.Sprintf("  %s", boldStyle.Render("━━ Dockerfile proposal ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")))
	w.Blank()

	if summary != "" {
		if validationScore > 0 && validationScore < 70 {
			w.term.Println(fmt.Sprintf("  %s %s", yellowStyle.Render("!"), summary))
		} else if validationScore == 0 {
			w.term.Println(fmt.Sprintf("  %s %s", yellowStyle.Render("!"), summary))
		} else {
			w.term.Println(fmt.Sprintf("  %s %s", greenStyle.Render(">"), summary))
		}
	}
	if validationScore >= 0 {
		scoreLabel := fmt.Sprintf("Validation score: %d/100", validationScore)
		if validationScore < 50 {
			w.term.Println(fmt.Sprintf("  %s %s", dimStyle.Render("  "), yellowStyle.Render(scoreLabel)))
		} else {
			w.term.Println(fmt.Sprintf("  %s %s", dimStyle.Render("  "), dimStyle.Render(scoreLabel)))
		}
	}
	if len(suggestions) > 0 {
		w.Blank()
		w.term.Println(fmt.Sprintf("  %s", dimStyle.Render("Suggestions:")))
		for _, s := range suggestions {
			w.term.Println(fmt.Sprintf("  %s %s", dimStyle.Render("  •"), s))
		}
	}
	if dockerfile != "" {
		w.Blank()
		w.term.Println(fmt.Sprintf("  %s", dimStyle.Render("Proposed Dockerfile:")))
		for _, line := range strings.Split(dockerfile, "\n") {
			w.term.Println(fmt.Sprintf("  %s %s", dimStyle.Render("│"), codeStyle.Render(line)))
		}
	}
	w.Blank()
	w.term.Print(fmt.Sprintf("  %s ", boldStyle.Render("Approve deployment?")))
	w.term.Print(fmt.Sprintf("%s ", dimStyle.Render("[y/N]:")))
}
