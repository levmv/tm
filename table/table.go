package table

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
)

type Formatter func(string, ...interface{}) string

type Table struct {
	Width int
	Fg    int
	Bg    int
	AltBg int

	header []string
	rows   [][]string
	widths []int

	colsStyles []Style
	rowsStyles []Style
}

type Style struct {
	Fg    int
	Bg    int
	AltBg int
	Mode  int
}

type Row struct {
	Cells []string
	Style Style
}

func colorFormat(str string, fg int, bg int, mode int) string {
	return fmt.Sprintf("\033[%d;38;5;%d;48;5;%dm%s\033[0m", mode, fg, bg, str)
}

func New(header []string, defStyle Style) *Table {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	t := Table{
		Fg:    defStyle.Fg,
		Bg:    defStyle.Bg,
		AltBg: defStyle.AltBg,
		Width: w,
	}
	t.AddRow(header, Style{Fg: t.Fg, Bg: t.Bg})

	for _ = range t.rows[0] {
		t.colsStyles = append(t.colsStyles, Style{})
	}

	return &t
}

func (t *Table) AddRow(row []string, style Style) {
	t.rows = append(t.rows, row)
	t.rowsStyles = append(t.rowsStyles, style)
}

func (t *Table) SetColStyle(colIdx int, style Style) {
	t.colsStyles[colIdx] = style
}

func (t *Table) calcColWidths() {
	colWidths := make([]int, len(t.rows[0]))
	for _, row := range t.rows {
		for j, cell := range row {
			if colWidths[j] < len(cell) {
				colWidths[j] = len(cell)
			}
		}
	}
	widthBudget := t.Width - 2*(len(colWidths)-1)
	curWidth := 0
	for _, v := range colWidths {
		curWidth += v
	}

	if curWidth > widthBudget {
		max := 0
		maxi := 0
		for i := range colWidths {
			if colWidths[i] > max {
				max = colWidths[i]
				maxi = i
			}
		}
		if (curWidth - widthBudget) < (colWidths[maxi] - 15) {
			colWidths[maxi] -= curWidth - widthBudget
		} else {
			colWidths[maxi] = 15
		}
	}
	t.widths = colWidths
}

func (t *Table) rowStyle(row int) Style {
	fg := t.Fg
	bg := t.Bg
	if t.AltBg != 0 && row%2 == 0 {
		bg = t.AltBg
	}

	if t.rowsStyles[row].Fg != 0 {
		fg = t.rowsStyles[row].Fg
	}
	if t.rowsStyles[row].Bg != 0 {
		bg = t.rowsStyles[row].Bg
	}

	return Style{
		Bg: bg,
		Fg: fg,
	}
}

func (t *Table) rowFormatter(row int, val string) string {
	s := t.rowStyle(row)
	return colorFormat(val, s.Fg, s.Bg, 0)
}

func (t *Table) columnFormatter(row int, col int, val string) string {
	s := t.rowStyle(row)

	if t.colsStyles[col].Fg != 0 {
		s.Fg = t.colsStyles[col].Fg
	}
	if t.colsStyles[col].Bg != 0 {
		s.Bg = t.colsStyles[col].Bg
	}
	return colorFormat(val, s.Fg, s.Bg, 0)
}

func (t *Table) Render() {

	t.calcColWidths()

	padStr := strings.Repeat(" ", 2)

	for rIdx, r := range t.rows {
		rows := t.wrapRow(r)

		for _, row := range rows {
			for j, w := range t.widths {
				row[j] = t.columnFormatter(rIdx, j, fmt.Sprintf("%-"+strconv.Itoa(w)+"v", row[j]))
			}
			fmt.Println(strings.Join(row, t.rowFormatter(rIdx, padStr)))
		}
	}
}

func (t *Table) wrapRow(row []string) [][]string {
	var rows [][]string
	rows = append(rows, make([]string, len(row)))

	for colIdx, w := range t.widths {
		if len(row[colIdx]) > w {
			wrapped, linesNum := wrapText(row[colIdx], w)
			if linesNum > len(rows) {
				for i := 0; i <= (linesNum - len(rows)); i++ {
					rows = append(rows, make([]string, len(row)))
				}
			}
			for j := 0; j < linesNum; j++ {
				rows[j][colIdx] = wrapped[j]
			}
		} else {
			rows[0][colIdx] = row[colIdx]
		}
	}
	return rows
}

func wrapText(in string, wrapLen int) ([]string, int) {
	words := strings.Fields(in)
	var lines []string
	var line strings.Builder
	for _, word := range words {
		wordLen := len(word)
		space, spacingLen := spacing(word)

		if line.Len()+wordLen+spacingLen > wrapLen {
			lines = append(lines, line.String())
			line.Reset()
		}
		line.WriteString(word)
		line.WriteString(space)
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines, len(lines)
}

func spacing(word string) (string, int) {
	if len(word) > 0 {
		return " ", 1
	}
	return "", 0
}
