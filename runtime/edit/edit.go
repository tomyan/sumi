// Package edit provides a text editing model for contenteditable elements.
// Supports insert, delete, navigation, readline shortcuts, undo/redo,
// kill ring with yank cycling, and command history with edit preservation.
package edit

import "unicode"

// snapshot captures a point-in-time editing state for undo/redo.
type snapshot struct {
	value  string
	cursor int
}

// histEntry stores a history entry with its editing state preserved.
type histEntry struct {
	original string     // the originally submitted value
	value    string     // current value (may be edited)
	cursor   int        // cursor position
	undoStack []snapshot // undo state for this entry
	redoStack []snapshot // redo state for this entry
}

// State holds the editing state for a contenteditable element.
type State struct {
	Value  string
	Cursor int // rune offset into Value

	// Undo/redo for current input.
	undoStack []snapshot
	redoStack []snapshot

	// Kill ring.
	killRing    []string
	killRingIdx int
	lastYank    bool // true if the last operation was a yank (for Alt+Y cycling)

	// Command history.
	history []histEntry
	histIdx int    // -1 = current input, 0..len-1 = browsing history
	saved   snapshot // current input saved when entering history
}

// saveUndo pushes the current state onto the undo stack and clears redo.
func (s *State) saveUndo() {
	s.undoStack = append(s.undoStack, snapshot{s.Value, s.Cursor})
	s.redoStack = nil
	s.lastYank = false
}

// pushKill adds text to the kill ring.
func (s *State) pushKill(text string) {
	if text == "" {
		return
	}
	s.killRing = append(s.killRing, text)
	s.killRingIdx = len(s.killRing) - 1
}

// --- Editing operations ---

// Insert inserts a rune at the cursor position.
func (s *State) Insert(ch rune) {
	s.saveUndo()
	r := s.runes()
	s.Value = string(r[:s.Cursor]) + string(ch) + string(r[s.Cursor:])
	s.Cursor++
}

// InsertString inserts a string at the cursor position.
func (s *State) InsertString(text string) {
	if text == "" {
		return
	}
	s.saveUndo()
	r := s.runes()
	s.Value = string(r[:s.Cursor]) + text + string(r[s.Cursor:])
	s.Cursor += len([]rune(text))
}

// InsertNewline inserts a newline at the cursor position.
func (s *State) InsertNewline() {
	s.Insert('\n')
}

// Backspace deletes the rune before the cursor.
func (s *State) Backspace() {
	if s.Cursor <= 0 {
		return
	}
	s.saveUndo()
	r := s.runes()
	s.Value = string(r[:s.Cursor-1]) + string(r[s.Cursor:])
	s.Cursor--
}

// Delete deletes the rune at the cursor.
func (s *State) Delete() {
	r := s.runes()
	if s.Cursor >= len(r) {
		return
	}
	s.saveUndo()
	s.Value = string(r[:s.Cursor]) + string(r[s.Cursor+1:])
}

// --- Navigation ---

// Left moves the cursor one rune left.
func (s *State) Left() {
	if s.Cursor > 0 {
		s.Cursor--
	}
	s.lastYank = false
}

// Right moves the cursor one rune right.
func (s *State) Right() {
	if s.Cursor < len(s.runes()) {
		s.Cursor++
	}
	s.lastYank = false
}

// Home moves the cursor to the start.
func (s *State) Home() {
	s.Cursor = 0
	s.lastYank = false
}

// End moves the cursor to the end.
func (s *State) End() {
	s.Cursor = len(s.runes())
	s.lastYank = false
}

// WordLeft moves the cursor to the start of the previous word.
func (s *State) WordLeft() {
	r := s.runes()
	i := s.Cursor
	for i > 0 && isSpace(r[i-1]) {
		i--
	}
	for i > 0 && !isSpace(r[i-1]) {
		i--
	}
	s.Cursor = i
	s.lastYank = false
}

// WordRight moves the cursor to the end of the next word.
func (s *State) WordRight() {
	r := s.runes()
	i := s.Cursor
	for i < len(r) && isSpace(r[i]) {
		i++
	}
	for i < len(r) && !isSpace(r[i]) {
		i++
	}
	s.Cursor = i
	s.lastYank = false
}

// --- Kill operations (push to kill ring) ---

// KillToEnd deletes from cursor to end (Ctrl+K). Killed text goes to kill ring.
func (s *State) KillToEnd() {
	r := s.runes()
	killed := string(r[s.Cursor:])
	if killed == "" {
		return
	}
	s.saveUndo()
	s.pushKill(killed)
	s.Value = string(r[:s.Cursor])
}

// KillToStart deletes from start to cursor (Ctrl+U). Killed text goes to kill ring.
func (s *State) KillToStart() {
	r := s.runes()
	killed := string(r[:s.Cursor])
	if killed == "" {
		return
	}
	s.saveUndo()
	s.pushKill(killed)
	s.Value = string(r[s.Cursor:])
	s.Cursor = 0
}

// KillWord deletes the word before the cursor (Ctrl+W). Killed text goes to kill ring.
func (s *State) KillWord() {
	r := s.runes()
	end := s.Cursor
	i := end
	for i > 0 && isSpace(r[i-1]) {
		i--
	}
	for i > 0 && !isSpace(r[i-1]) {
		i--
	}
	killed := string(r[i:end])
	if killed == "" {
		return
	}
	s.saveUndo()
	s.pushKill(killed)
	s.Value = string(r[:i]) + string(r[end:])
	s.Cursor = i
}

// KillWordForward deletes the word after the cursor (Alt+D). Killed text goes to kill ring.
func (s *State) KillWordForward() {
	r := s.runes()
	start := s.Cursor
	i := start
	for i < len(r) && isSpace(r[i]) {
		i++
	}
	for i < len(r) && !isSpace(r[i]) {
		i++
	}
	killed := string(r[start:i])
	if killed == "" {
		return
	}
	s.saveUndo()
	s.pushKill(killed)
	s.Value = string(r[:start]) + string(r[i:])
}

// --- Yank (paste from kill ring) ---

// Yank inserts the most recent kill ring entry at the cursor (Ctrl+Y).
func (s *State) Yank() {
	if len(s.killRing) == 0 {
		return
	}
	s.saveUndo()
	text := s.killRing[s.killRingIdx]
	r := s.runes()
	s.Value = string(r[:s.Cursor]) + text + string(r[s.Cursor:])
	s.Cursor += len([]rune(text))
	s.lastYank = true
}

// YankPop replaces the last yanked text with the previous kill ring entry (Alt+Y).
// Only works immediately after Yank or YankPop.
func (s *State) YankPop() {
	if !s.lastYank || len(s.killRing) < 2 {
		return
	}
	// Remove the previously yanked text.
	prevText := s.killRing[s.killRingIdx]
	prevLen := len([]rune(prevText))
	r := s.runes()
	start := s.Cursor - prevLen
	if start < 0 {
		start = 0
	}
	s.Value = string(r[:start]) + string(r[s.Cursor:])
	s.Cursor = start

	// Cycle to previous kill ring entry.
	s.killRingIdx--
	if s.killRingIdx < 0 {
		s.killRingIdx = len(s.killRing) - 1
	}

	// Insert the new entry.
	text := s.killRing[s.killRingIdx]
	r = s.runes()
	s.Value = string(r[:s.Cursor]) + text + string(r[s.Cursor:])
	s.Cursor += len([]rune(text))
	s.lastYank = true
}

// --- Transpose ---

// TransposeChars swaps the character before the cursor with the one at the cursor (Ctrl+T).
func (s *State) TransposeChars() {
	r := s.runes()
	if len(r) < 2 {
		return
	}
	// At end of line, transpose the two chars before cursor.
	pos := s.Cursor
	if pos >= len(r) {
		pos = len(r) - 1
	}
	if pos < 1 {
		return
	}
	s.saveUndo()
	r[pos-1], r[pos] = r[pos], r[pos-1]
	s.Value = string(r)
	if s.Cursor < len(r) {
		s.Cursor++
	}
}

// --- Word transforms ---

// UppercaseWord converts the word at cursor to uppercase (Alt+U).
func (s *State) UppercaseWord() {
	s.transformWord(unicode.ToUpper)
}

// LowercaseWord converts the word at cursor to lowercase (Alt+L).
func (s *State) LowercaseWord() {
	s.transformWord(unicode.ToLower)
}

// CapitalizeWord capitalizes the word at cursor (Alt+C).
func (s *State) CapitalizeWord() {
	r := s.runes()
	if s.Cursor >= len(r) {
		return
	}
	s.saveUndo()
	i := s.Cursor
	// Skip spaces.
	for i < len(r) && isSpace(r[i]) {
		i++
	}
	// Capitalize first char.
	if i < len(r) {
		r[i] = unicode.ToUpper(r[i])
		i++
	}
	// Lowercase rest of word.
	for i < len(r) && !isSpace(r[i]) {
		r[i] = unicode.ToLower(r[i])
		i++
	}
	s.Value = string(r)
	s.Cursor = i
}

func (s *State) transformWord(fn func(rune) rune) {
	r := s.runes()
	if s.Cursor >= len(r) {
		return
	}
	s.saveUndo()
	i := s.Cursor
	for i < len(r) && isSpace(r[i]) {
		i++
	}
	for i < len(r) && !isSpace(r[i]) {
		r[i] = fn(r[i])
		i++
	}
	s.Value = string(r)
	s.Cursor = i
}

// --- Undo / Redo ---

// Undo reverts to the previous state (Ctrl+_ or Ctrl+Z).
func (s *State) Undo() {
	if len(s.undoStack) == 0 {
		return
	}
	s.redoStack = append(s.redoStack, snapshot{s.Value, s.Cursor})
	prev := s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]
	s.Value = prev.value
	s.Cursor = prev.cursor
	s.lastYank = false
}

// Redo reapplies the last undone change (Ctrl+Y after undo, but we use Ctrl+Shift+Z convention).
// Note: in readline, Ctrl+Y is yank. Redo is typically not bound.
// We provide it for programmatic use.
func (s *State) Redo() {
	if len(s.redoStack) == 0 {
		return
	}
	s.undoStack = append(s.undoStack, snapshot{s.Value, s.Cursor})
	next := s.redoStack[len(s.redoStack)-1]
	s.redoStack = s.redoStack[:len(s.redoStack)-1]
	s.Value = next.value
	s.Cursor = next.cursor
	s.lastYank = false
}

// --- History ---

// Submit adds the current value to history and clears the input.
// Returns the submitted value.
func (s *State) Submit() string {
	val := s.Value
	if val != "" {
		s.history = append(s.history, histEntry{
			original: val,
			value:    val,
			cursor:   len([]rune(val)),
		})
	}
	s.Value = ""
	s.Cursor = 0
	s.undoStack = nil
	s.redoStack = nil
	s.histIdx = -1
	s.saved = snapshot{}
	s.lastYank = false
	return val
}

// HistoryUp moves to the previous history entry, preserving edits.
func (s *State) HistoryUp() {
	if len(s.history) == 0 {
		return
	}
	if s.histIdx == -1 {
		// Entering history — save current input.
		s.saved = snapshot{s.Value, s.Cursor}
		s.histIdx = len(s.history) - 1
	} else if s.histIdx > 0 {
		// Save edits to current history entry before moving.
		s.saveHistEdits()
		s.histIdx--
	} else {
		return
	}
	s.restoreHistEntry()
	s.lastYank = false
}

// HistoryDown moves to the next history entry or back to current input.
func (s *State) HistoryDown() {
	if s.histIdx == -1 {
		return
	}
	s.saveHistEdits()
	if s.histIdx < len(s.history)-1 {
		s.histIdx++
		s.restoreHistEntry()
	} else {
		s.histIdx = -1
		s.Value = s.saved.value
		s.Cursor = s.saved.cursor
		s.undoStack = nil
		s.redoStack = nil
	}
	s.lastYank = false
}

// saveHistEdits saves the current editing state back to the history entry.
func (s *State) saveHistEdits() {
	if s.histIdx >= 0 && s.histIdx < len(s.history) {
		s.history[s.histIdx].value = s.Value
		s.history[s.histIdx].cursor = s.Cursor
		s.history[s.histIdx].undoStack = s.undoStack
		s.history[s.histIdx].redoStack = s.redoStack
	}
}

// restoreHistEntry loads a history entry's editing state.
func (s *State) restoreHistEntry() {
	e := s.history[s.histIdx]
	s.Value = e.value
	s.Cursor = e.cursor
	s.undoStack = e.undoStack
	s.redoStack = e.redoStack
}

// Clear resets the value and cursor.
func (s *State) Clear() {
	s.saveUndo()
	s.Value = ""
	s.Cursor = 0
}

// --- Helpers ---

func (s *State) runes() []rune {
	return []rune(s.Value)
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}
