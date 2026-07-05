package layout

// inlineSegment is a piece of one word owned by a single run — words can
// span run boundaries ("foo<strong>bar</strong>" is one word), and each
// run's part becomes its own fragment.
type inlineSegment struct {
	run  int
	text []rune
}

// inlineWord is an unbreakable unit: its segments, plus the run owning
// the collapsed space before it (-1 when none).
type inlineWord struct {
	segs     []inlineSegment
	preSpace int
}

func (w *inlineWord) width() int {
	n := 0
	for _, s := range w.segs {
		n += len(s.text)
	}
	return n
}

// tokenizeInline collapses whitespace across the run sequence
// (white-space: normal) and splits it into words. A whitespace gap
// belongs to the run where it first appears; leading and trailing gaps
// are dropped at line assembly.
func tokenizeInline(texts []string) []inlineWord {
	var words []inlineWord
	current := inlineWord{preSpace: -1}
	pendingSpace := -1
	for run, text := range texts {
		for _, r := range text {
			if isInlineSpace(r) {
				if len(current.segs) > 0 {
					words = append(words, current)
					current = inlineWord{preSpace: -1}
				}
				if pendingSpace < 0 {
					pendingSpace = run
				}
				continue
			}
			if len(current.segs) == 0 && pendingSpace >= 0 {
				if len(words) > 0 {
					current.preSpace = pendingSpace
				}
				pendingSpace = -1
			}
			current.appendRune(run, r)
		}
	}
	if len(current.segs) > 0 {
		words = append(words, current)
	}
	return words
}

func isInlineSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// appendRune adds a rune to the word, starting a new segment when the
// owning run changes.
func (w *inlineWord) appendRune(run int, r rune) {
	last := len(w.segs) - 1
	if last < 0 || w.segs[last].run != run {
		w.segs = append(w.segs, inlineSegment{run: run})
		last++
	}
	w.segs[last].text = append(w.segs[last].text, r)
}

// breakInline flows the runs' text into lines of availW cells and
// returns each run's fragments in container-content coordinates.
// Soft breaks happen at collapsed spaces (the breaking space is
// consumed); words wider than a line hard-break at the width.
func breakInline(texts []string, availW int) [][]Fragment {
	lf := &lineFlow{availW: max(availW, 1), perRun: make([][]Fragment, len(texts))}
	for _, w := range tokenizeInline(texts) {
		lf.placeWord(w)
	}
	return lf.perRun
}

// lineFlow tracks the fill cursor while words are placed onto lines.
type lineFlow struct {
	availW, cursorX, lineY int
	perRun                 [][]Fragment
}

// placeWord soft-wraps to the next line when the word (plus its leading
// space) does not fit, then emits the space and the word's segments.
func (lf *lineFlow) placeWord(w inlineWord) {
	spaceW := 0
	if w.preSpace >= 0 && lf.cursorX > 0 {
		spaceW = 1
	}
	if lf.cursorX > 0 && lf.cursorX+spaceW+w.width() > lf.availW {
		lf.lineY++
		lf.cursorX = 0
		spaceW = 0 // the breaking space is consumed
	}
	if spaceW == 1 {
		lf.emit(w.preSpace, " ")
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
			lf.lineY++
			lf.cursorX = 0
			space = lf.availW
		}
		n := min(len(seg.text)-start, space)
		lf.emit(seg.run, string(seg.text[start:start+n]))
		start += n
	}
}

// emit appends text at the cursor to the run's fragments, extending the
// previous fragment when contiguous on the same line.
func (lf *lineFlow) emit(run int, text string) {
	frags := lf.perRun[run]
	if n := len(frags) - 1; n >= 0 && frags[n].Y == lf.lineY && frags[n].X+runeLen(frags[n].Text) == lf.cursorX {
		frags[n].Text += text
	} else {
		lf.perRun[run] = append(frags, Fragment{X: lf.cursorX, Y: lf.lineY, Text: text})
	}
	lf.cursorX += runeLen(text)
}
