package ui

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI escape codes from a string.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// VisualWidth calculates the visible display width of a string on terminal,
// taking ANSI escape sequences and wide (East Asian) characters into account.
func VisualWidth(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}

// Table provides a gh-style borderless table printer with ANSI-safe width calculation.
type Table struct {
	headers []string
	rows    [][]string
	indent  string
	padding int
}

// NewTable creates a new Table instance with given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		indent:  "",
		padding: 2,
	}
}

// SetIndent sets the left margin indentation for the table.
func (t *Table) SetIndent(indent string) *Table {
	t.indent = indent
	return t
}

// SetPadding sets the number of spaces between columns.
func (t *Table) SetPadding(padding int) *Table {
	if padding < 1 {
		padding = 1
	}
	t.padding = padding
	return t
}

// AddRow adds a row of cell values to the table.
func (t *Table) AddRow(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Render prints the table to the provided io.Writer (usually os.Stdout).
func (t *Table) Render(w io.Writer) error {
	colCount := len(t.headers)
	for _, row := range t.rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	if colCount == 0 {
		return nil
	}

	// Calculate maximum visual width for each column
	colWidths := make([]int, colCount)
	for i, h := range t.headers {
		w := VisualWidth(h)
		if w > colWidths[i] {
			colWidths[i] = w
		}
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i >= colCount {
				break
			}
			w := VisualWidth(cell)
			if w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	sep := strings.Repeat(" ", t.padding)

	// Print headers (uppercase + Dim)
	if len(t.headers) > 0 {
		var headerLine strings.Builder
		headerLine.WriteString(t.indent)
		for i := 0; i < colCount; i++ {
			h := ""
			if i < len(t.headers) {
				h = strings.ToUpper(t.headers[i])
			}
			visW := VisualWidth(h)
			pad := colWidths[i] - visW
			if pad < 0 {
				pad = 0
			}

			headerLine.WriteString(Dim(h))
			if i < colCount-1 {
				headerLine.WriteString(strings.Repeat(" ", pad))
				headerLine.WriteString(sep)
			}
		}
		headerLine.WriteString("\n")
		if _, err := fmt.Fprint(w, headerLine.String()); err != nil {
			return err
		}
	}

	// Print data rows
	for _, row := range t.rows {
		var rowLine strings.Builder
		rowLine.WriteString(t.indent)
		for i := 0; i < colCount; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			visW := VisualWidth(cell)
			pad := colWidths[i] - visW
			if pad < 0 {
				pad = 0
			}

			rowLine.WriteString(cell)
			if i < colCount-1 {
				rowLine.WriteString(strings.Repeat(" ", pad))
				rowLine.WriteString(sep)
			}
		}
		rowLine.WriteString("\n")
		if _, err := fmt.Fprint(w, rowLine.String()); err != nil {
			return err
		}
	}

	return nil
}

// Print renders the table directly to os.Stdout.
func (t *Table) Print() error {
	return t.Render(os.Stdout)
}
