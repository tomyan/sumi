package layout

// inlineSegment is a piece of one word owned by a single item — words
// can span run boundaries ("foo<strong>bar</strong>" is one word), and
// each run's part becomes its own fragment.
type inlineSegment struct {
	item int
	text []rune
}

// inlineWord is an unbreakable unit: a text word's segments, or an atom
// (inline-block), plus the item owning the collapsed space before it
// (-1 when none).
type inlineWord struct {
	segs     []inlineSegment
	preSpace int
	atom     int // item index of an atom word; -1 for text words
	atomW    int
	atomH    int
}

func (w *inlineWord) width() int {
	if w.atom >= 0 {
		return w.atomW
	}
	n := 0
	for _, s := range w.segs {
		n += len(s.text)
	}
	return n
}

// tokenizeInline collapses whitespace across the item sequence
// (white-space: normal) and splits it into words; atoms are words of
// their own. A whitespace gap belongs to the item where it first
// appears; leading and trailing gaps are dropped at line assembly.
func tokenizeInline(items []inlineItem) []inlineWord {
	t := &tokenizer{current: inlineWord{preSpace: -1, atom: -1}, pendingSpace: -1}
	for idx, item := range items {
		if item.atom {
			t.closeWord()
			t.startWord()
			t.current.atom = idx
			t.current.atomW = item.box.Width
			t.current.atomH = item.box.Height
			t.closeWord()
			continue
		}
		for _, r := range item.text {
			if isInlineSpace(r) {
				t.closeWord()
				if t.pendingSpace < 0 {
					t.pendingSpace = idx
				}
				continue
			}
			if len(t.current.segs) == 0 && t.current.atom < 0 {
				t.startWord()
			}
			t.current.appendRune(idx, r)
		}
	}
	t.closeWord()
	return t.words
}

type tokenizer struct {
	words        []inlineWord
	current      inlineWord
	pendingSpace int
	open         bool
}

// startWord claims any pending space for the word about to be built.
// A leading gap (nothing before it in the flow) is dropped.
func (t *tokenizer) startWord() {
	if t.open {
		return
	}
	t.open = true
	if t.pendingSpace >= 0 {
		if len(t.words) > 0 {
			t.current.preSpace = t.pendingSpace
		}
		t.pendingSpace = -1
	}
}

// closeWord pushes the word under construction, if any.
func (t *tokenizer) closeWord() {
	if t.current.atom >= 0 || len(t.current.segs) > 0 {
		t.words = append(t.words, t.current)
	}
	t.current = inlineWord{preSpace: -1, atom: -1}
	t.open = false
}

func isInlineSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// appendRune adds a rune to the word, starting a new segment when the
// owning item changes.
func (w *inlineWord) appendRune(item int, r rune) {
	last := len(w.segs) - 1
	if last < 0 || w.segs[last].item != item {
		w.segs = append(w.segs, inlineSegment{item: item})
		last++
	}
	w.segs[last].text = append(w.segs[last].text, r)
}

// breakInline flows the items into lines of availW cells and returns
// each item's fragments in container-content coordinates (atoms get one
// empty-text fragment marking their slot). Soft breaks happen at
// collapsed spaces (the breaking space is consumed); words wider than a
// line hard-break at the width; atoms are unbreakable. Line heights
// follow the tallest item on each line.
func breakInline(items []inlineItem, availW int) [][]Fragment {
	lf := &lineFlow{availW: max(availW, 1), perItem: make([][]Fragment, len(items)), lineHeights: []int{0}}
	for _, w := range tokenizeInline(items) {
		lf.placeWord(w)
	}
	lf.resolveLineOffsets()
	return lf.perItem
}

// lineFlow tracks the fill cursor while words are placed onto lines.
// Fragment Y values hold line indices until resolveLineOffsets rewrites
// them to row offsets using the accumulated line heights.
type lineFlow struct {
	availW, cursorX, line int
	lineHeights           []int
	perItem               [][]Fragment
}

// placeWord soft-wraps to the next line when the word (plus its leading
// space) does not fit, then emits the space and the word's content.
func (lf *lineFlow) placeWord(w inlineWord) {
	spaceW := 0
	if w.preSpace >= 0 && lf.cursorX > 0 {
		spaceW = 1
	}
	if lf.cursorX > 0 && lf.cursorX+spaceW+w.width() > lf.availW {
		lf.newLine()
		spaceW = 0 // the breaking space is consumed
	}
	if spaceW == 1 {
		lf.emit(w.preSpace, " ")
	}
	if w.atom >= 0 {
		lf.growLine(w.atomH)
		lf.perItem[w.atom] = []Fragment{{X: lf.cursorX, Y: lf.line}}
		lf.cursorX += w.atomW
		return
	}
	for _, seg := range w.segs {
		lf.placeSegment(seg)
	}
}

// placeSegment emits a word segment, hard-breaking at the line width
// when the segment overruns it.
func (lf *lineFlow) placeSegment(seg inlineSegment) {
	for start := 0; start < len(seg.text); {
		space := lf.availW - lf.cursorX
		if space <= 0 {
			lf.newLine()
			space = lf.availW
		}
		n := min(len(seg.text)-start, space)
		lf.emit(seg.item, string(seg.text[start:start+n]))
		start += n
	}
}

// emit appends text at the cursor to the item's fragments, extending
// the previous fragment when contiguous on the same line.
func (lf *lineFlow) emit(item int, text string) {
	lf.growLine(1)
	frags := lf.perItem[item]
	if n := len(frags) - 1; n >= 0 && frags[n].Y == lf.line && frags[n].X+runeLen(frags[n].Text) == lf.cursorX {
		frags[n].Text += text
	} else {
		lf.perItem[item] = append(frags, Fragment{X: lf.cursorX, Y: lf.line, Text: text})
	}
	lf.cursorX += runeLen(text)
}

func (lf *lineFlow) newLine() {
	lf.line++
	lf.cursorX = 0
	lf.lineHeights = append(lf.lineHeights, 0)
}

// growLine raises the current line's height to at least h.
func (lf *lineFlow) growLine(h int) {
	if lf.lineHeights[lf.line] < h {
		lf.lineHeights[lf.line] = h
	}
}

// resolveLineOffsets rewrites fragment Y values from line indices to
// row offsets (prefix sums of line heights).
func (lf *lineFlow) resolveLineOffsets() {
	offsets := make([]int, len(lf.lineHeights))
	y := 0
	for i, h := range lf.lineHeights {
		offsets[i] = y
		y += h
	}
	for i := range lf.perItem {
		for j := range lf.perItem[i] {
			lf.perItem[i][j].Y = offsets[lf.perItem[i][j].Y]
		}
	}
}
