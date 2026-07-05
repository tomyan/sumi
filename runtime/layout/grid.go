package layout

import (
	"strconv"
	"strings"
)

// gridArea is a placed region in track coordinates (0-based, end exclusive).
type gridArea struct {
	colStart, colEnd, rowStart, rowEnd int
}

// layoutGrid lays out children as a CSS grid. Explicit placement via
// grid-column/grid-row/grid-area wins; remaining children auto-flow
// row-by-row (grid-auto-flow: column is not implemented, matching the
// svelterm deviation). Implicit rows size to their tallest item.
func layoutGrid(input *Input, children []*Input, offsetX, offsetY, availW, availH int) []*Box {
	cols := parseTrackList(input.GridTemplateColumns, availW, input.Gap)
	if len(cols) == 0 {
		cols = []int{availW}
	}
	rows := parseTrackList(input.GridTemplateRows, availH, input.Gap)
	areas := parseTemplateAreas(input.GridTemplateAreas)

	placements := placeGridChildren(children, cols, rows, areas)
	rows = growRows(rows, placements)

	// Implicit/auto rows: size to tallest natural item starting in the row.
	colOffsets := trackOffsets(cols, input.Gap)
	rows = sizeAutoRows(rows, children, placements, cols, colOffsets, input.Gap, availH)
	rowOffsets := trackOffsets(rows, input.Gap)

	boxes := make([]*Box, len(children))
	for i, child := range children {
		p := placements[i]
		areaW := spanSize(cols, colOffsets, p.colStart, p.colEnd, input.Gap)
		areaH := spanSize(rows, rowOffsets, p.rowStart, p.rowEnd, input.Gap)
		childBox := layoutNode(child, areaW, areaH)
		childBox.X = offsetX + colOffsets[p.colStart]
		childBox.Y = offsetY + rowOffsets[p.rowStart]
		// Grid items stretch to their area unless explicitly sized.
		if child.Kind != KindText && child.FixedWidth == 0 && child.WidthPct == 0 {
			childBox.Width = areaW
		}
		if child.Kind != KindText && child.FixedHeight == 0 && child.HeightPct == 0 {
			childBox.Height = areaH
		}
		boxes[i] = childBox
	}
	return boxes
}

// placeGridChildren assigns every child a grid area: explicit placement
// first, then row-based auto-flow into free cells.
func placeGridChildren(children []*Input, cols, rows []int, areas map[string]gridArea) []gridArea {
	placements := make([]gridArea, len(children))
	explicit := make([]bool, len(children))
	occ := newOccupancy(len(cols), len(rows))

	for i, child := range children {
		area, ok := explicitPlacement(child, len(cols), areas)
		if !ok {
			continue
		}
		placements[i] = area
		explicit[i] = true
		occ.mark(area)
	}
	for i, child := range children {
		if explicit[i] {
			continue
		}
		span := autoSpan(child, len(cols))
		area := occ.nextFree(span)
		placements[i] = area
		occ.mark(area)
	}
	return placements
}

// explicitPlacement resolves grid-area / grid-column / grid-row.
func explicitPlacement(child *Input, colCount int, areas map[string]gridArea) (gridArea, bool) {
	if child.GridArea != "" {
		if area, ok := areas[child.GridArea]; ok {
			return area, true
		}
	}
	if child.GridColumn == "" && child.GridRow == "" {
		return gridArea{}, false
	}
	colStart, colEnd, colOK := parseGridLine(child.GridColumn, colCount)
	rowStart, rowEnd, rowOK := parseGridLine(child.GridRow, 1<<20)
	if !colOK {
		colStart, colEnd = 0, 1
	}
	if !rowOK {
		return gridArea{}, false // row auto-flows; treat as column-only pin
	}
	return gridArea{colStart, colEnd, rowStart, rowEnd}, colOK && rowOK
}

// parseGridLine parses "2", "1 / 3", or "span 2" into a 0-based half-open
// track range. CSS grid lines are 1-based; end lines are exclusive.
func parseGridLine(v string, max int) (int, int, bool) {
	v = strings.TrimSpace(v)
	if v == "" || v == "auto" {
		return 0, 0, false
	}
	if start, end, found := strings.Cut(v, "/"); found {
		s, err1 := strconv.Atoi(strings.TrimSpace(start))
		e, err2 := strconv.Atoi(strings.TrimSpace(end))
		if err1 != nil || err2 != nil || s < 1 || e <= s {
			return 0, 0, false
		}
		return s - 1, e - 1, true
	}
	if span, ok := strings.CutPrefix(v, "span "); ok {
		n, err := strconv.Atoi(strings.TrimSpace(span))
		if err != nil || n < 1 {
			return 0, 0, false
		}
		return -n, 0, false // negative start signals span-only (auto-flow width)
	}
	s, err := strconv.Atoi(v)
	if err != nil || s < 1 || s > max {
		return 0, 0, false
	}
	return s - 1, s, true
}

// autoSpan returns the column span for an auto-placed child (span n or 1).
func autoSpan(child *Input, colCount int) int {
	if s, ok := strings.CutPrefix(strings.TrimSpace(child.GridColumn), "span "); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n >= 1 {
			if n > colCount {
				return colCount
			}
			return n
		}
	}
	return 1
}

// parseTemplateAreas parses `"a a b" "c c b"` into named rectangles.
func parseTemplateAreas(spec string) map[string]gridArea {
	if spec == "" {
		return nil
	}
	areas := make(map[string]gridArea)
	rows := splitQuoted(spec)
	for r, row := range rows {
		for c, name := range strings.Fields(row) {
			if name == "." {
				continue
			}
			a, seen := areas[name]
			if !seen {
				areas[name] = gridArea{c, c + 1, r, r + 1}
				continue
			}
			if c < a.colStart {
				a.colStart = c
			}
			if c+1 > a.colEnd {
				a.colEnd = c + 1
			}
			if r+1 > a.rowEnd {
				a.rowEnd = r + 1
			}
			areas[name] = a
		}
	}
	return areas
}

// splitQuoted extracts the quoted strings from an areas value.
func splitQuoted(spec string) []string {
	var rows []string
	rest := spec
	for {
		i := strings.IndexByte(rest, '"')
		if i < 0 {
			return rows
		}
		rest = rest[i+1:]
		j := strings.IndexByte(rest, '"')
		if j < 0 {
			return rows
		}
		rows = append(rows, rest[:j])
		rest = rest[j+1:]
	}
}

// occupancy tracks filled grid cells for auto-flow.
type occupancy struct {
	cols  int
	cells map[[2]int]bool // [row, col]
}

func newOccupancy(cols, rows int) *occupancy {
	return &occupancy{cols: cols, cells: make(map[[2]int]bool)}
}

func (o *occupancy) mark(a gridArea) {
	for r := a.rowStart; r < a.rowEnd; r++ {
		for c := a.colStart; c < a.colEnd; c++ {
			o.cells[[2]int{r, c}] = true
		}
	}
}

func (o *occupancy) free(r, c, span int) bool {
	if c+span > o.cols {
		return false
	}
	for i := c; i < c+span; i++ {
		if o.cells[[2]int{r, i}] {
			return false
		}
	}
	return true
}

// nextFree finds the first free cell range scanning rows left-to-right.
func (o *occupancy) nextFree(span int) gridArea {
	if span > o.cols {
		span = o.cols
	}
	for r := 0; ; r++ {
		for c := 0; c+span <= o.cols; c++ {
			if o.free(r, c, span) {
				return gridArea{c, c + span, r, r + 1}
			}
		}
	}
}

// growRows extends the row track list to cover all placements.
func growRows(rows []int, placements []gridArea) []int {
	need := 0
	for _, p := range placements {
		if p.rowEnd > need {
			need = p.rowEnd
		}
	}
	for len(rows) < need {
		rows = append(rows, 0) // implicit auto row, sized later
	}
	return rows
}

// sizeAutoRows gives zero-height (auto/implicit) rows the height of their
// tallest starting item.
func sizeAutoRows(rows []int, children []*Input, placements []gridArea, cols, colOffsets []int, gap, availH int) []int {
	for r := range rows {
		if rows[r] > 0 {
			continue
		}
		maxH := 1
		for i, child := range children {
			p := placements[i]
			if p.rowStart != r {
				continue
			}
			areaW := spanSize(cols, colOffsets, p.colStart, p.colEnd, gap)
			h := layoutNode(child, areaW, availH).Height
			if h > maxH {
				maxH = h
			}
		}
		rows[r] = maxH
	}
	return rows
}

// trackOffsets returns each track's starting offset including gaps.
func trackOffsets(tracks []int, gap int) []int {
	offsets := make([]int, len(tracks)+1)
	cursor := 0
	for i, size := range tracks {
		offsets[i] = cursor
		cursor += size + gap
	}
	offsets[len(tracks)] = cursor
	return offsets
}

// spanSize is the total size of tracks [start, end) including interior gaps.
func spanSize(tracks, offsets []int, start, end, gap int) int {
	if start >= len(tracks) {
		return 0
	}
	if end > len(tracks) {
		end = len(tracks)
	}
	size := 0
	for i := start; i < end; i++ {
		size += tracks[i]
	}
	return size + gap*(end-start-1)
}
