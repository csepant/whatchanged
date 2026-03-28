package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table renders a simple styled table.
type Table struct {
	Headers []string
	Rows    [][]string
	Widths  []int // optional explicit column widths
}

// Print renders the table to the writer.
func (t *Table) Print(w io.Writer) {
	if len(t.Rows) == 0 {
		return
	}

	// Calculate column widths from content
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	// Override with explicit widths if set
	for i, w := range t.Widths {
		if i < len(widths) && w > 0 {
			widths[i] = w
		}
	}

	gap := "  "

	// Header
	if noColor {
		for i, h := range t.Headers {
			fmt.Fprintf(w, "%-*s%s", widths[i], h, gap)
		}
	} else {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
		for i, h := range t.Headers {
			fmt.Fprintf(w, "%s%s", headerStyle.Render(fmt.Sprintf("%-*s", widths[i], h)), gap)
		}
	}
	fmt.Fprintln(w)

	// Separator
	if noColor {
		for i := range t.Headers {
			fmt.Fprintf(w, "%s%s", strings.Repeat("─", widths[i]), gap)
		}
	} else {
		sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		for i := range t.Headers {
			fmt.Fprintf(w, "%s%s", sepStyle.Render(strings.Repeat("─", widths[i])), gap)
		}
	}
	fmt.Fprintln(w)

	// Rows
	for rowIdx, row := range t.Rows {
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			padded := fmt.Sprintf("%-*s", widths[i], cell)
			if noColor {
				fmt.Fprintf(w, "%s%s", padded, gap)
			} else if rowIdx%2 == 1 {
				fmt.Fprintf(w, "%s%s", lipgloss.NewStyle().Faint(true).Render(padded), gap)
			} else {
				fmt.Fprintf(w, "%s%s", padded, gap)
			}
		}
		fmt.Fprintln(w)
	}
}
