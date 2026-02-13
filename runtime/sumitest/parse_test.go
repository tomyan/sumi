package sumitest

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestParsePlainText(t *testing.T) {
	// Given
	markup := "Hello"

	// When
	rows := Parse(markup)

	// Then
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	assertCells(t, rows[0], "Hello", render.Style{})
}

func TestParseStyledText(t *testing.T) {
	// Given
	markup := "<<green>>Hi<</>>"

	// When
	rows := Parse(markup)

	// Then
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	assertCells(t, rows[0], "Hi", render.Style{FG: render.Color{Name: "green"}})
}

func TestParseMultipleStyles(t *testing.T) {
	// Given
	markup := "<<red>>AB<</>>CD"

	// When
	rows := Parse(markup)

	// Then
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if len(rows[0]) != 4 {
		t.Fatalf("got %d cells, want 4", len(rows[0]))
	}
	redStyle := render.Style{FG: render.Color{Name: "red"}}
	assertCell(t, rows[0][0], 'A', redStyle)
	assertCell(t, rows[0][1], 'B', redStyle)
	assertCell(t, rows[0][2], 'C', render.Style{})
	assertCell(t, rows[0][3], 'D', render.Style{})
}

func TestParseMultipleRows(t *testing.T) {
	// Given
	markup := "Top\nBottom"

	// When
	rows := Parse(markup)

	// Then
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	assertCells(t, rows[0], "Top", render.Style{})
	assertCells(t, rows[1], "Bottom", render.Style{})
}

func TestParseBGColor(t *testing.T) {
	// Given
	markup := "<<bg:red>>X<</>>"

	// When
	rows := Parse(markup)

	// Then
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	assertCell(t, rows[0][0], 'X', render.Style{BG: render.Color{Name: "red"}})
}

func TestParseAllAttributes(t *testing.T) {
	// Given
	markup := "<<bg:yellow,cyan,bold,dim,inverse,italic,strikethrough,underline>>X<</>>"

	// When
	rows := Parse(markup)

	// Then
	want := render.Style{
		FG:            render.Color{Name: "cyan"},
		BG:            render.Color{Name: "yellow"},
		Bold:          true,
		Dim:           true,
		Italic:        true,
		Underline:     true,
		Strikethrough: true,
		Inverse:       true,
	}
	assertCell(t, rows[0][0], 'X', want)
}

func TestParseRoundtrip(t *testing.T) {
	// Given — a buffer with mixed styled content
	buf := render.NewBuffer(10, 2)
	buf.WriteStyledText(0, 0, "Red", render.Style{FG: render.Color{Name: "red"}})
	buf.WriteText(0, 3, " ok")
	buf.WriteStyledText(1, 0, "Bold", render.Style{Bold: true})

	// When — roundtrip through styled text
	markup := buf.ToStyledText()
	rows := Parse(markup)

	// Then — cells should match original
	for r := 0; r < 2; r++ {
		for c := 0; c < len(rows[r]); c++ {
			got := rows[r][c]
			want := buf.Cell(r, c)
			if got.Ch != want.Ch || got.Style != want.Style {
				t.Errorf("row %d col %d: got {%c %v}, want {%c %v}",
					r, c, got.Ch, got.Style, want.Ch, want.Style)
			}
		}
	}
}

func TestParseEmptyString(t *testing.T) {
	// When
	rows := Parse("")

	// Then
	if len(rows) != 0 {
		t.Fatalf("got %d rows, want 0", len(rows))
	}
}

// assertCells checks that all cells in the row have the expected text and style.
func assertCells(t *testing.T, cells []render.Cell, text string, style render.Style) {
	t.Helper()
	runes := []rune(text)
	if len(cells) != len(runes) {
		t.Fatalf("got %d cells, want %d", len(cells), len(runes))
	}
	for i, r := range runes {
		assertCell(t, cells[i], r, style)
	}
}

// assertCell checks a single cell's character and style.
func assertCell(t *testing.T, cell render.Cell, ch rune, style render.Style) {
	t.Helper()
	if cell.Ch != ch {
		t.Errorf("cell.Ch = %c, want %c", cell.Ch, ch)
	}
	if cell.Style != style {
		t.Errorf("cell.Style = %+v, want %+v", cell.Style, style)
	}
}
