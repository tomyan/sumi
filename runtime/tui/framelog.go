package tui

import "github.com/tomyan/sumi/runtime/layout"

// FrameLog streams an append-only log of frames — each a mounted sumi
// component — into an inline app's live zone. Archiving a frame hands
// its rows to the terminal's native scrollback (zero repaint) and
// disposes the component, so memory tracks the live area rather than
// the transcript. Place Host somewhere in the app tree; frames stack
// under it in block flow.
type FrameLog struct {
	// Host is the container to embed in the app tree.
	Host *layout.Input
	// ReleaseTop hands n top rows of the live zone to scrollback.
	// Wire to App.ReleaseTop for inline apps; nil (or fullscreen)
	// makes Archive dispose-only.
	ReleaseTop func(n int)

	frames []logFrame
	nextID int
}

type logFrame struct {
	id        int
	container *layout.Input
	comp      *Component
}

// NewFrameLog builds an empty log with a fresh host container.
func NewFrameLog() *FrameLog {
	return &FrameLog{
		Host:   &layout.Input{Kind: layout.KindBox, Tag: "div", CursorCol: -1, CursorRow: -1},
		nextID: 1,
	}
}

// Append mounts a component as a new frame at the bottom of the log
// and returns its id. Update a live frame by writing its signals — the
// next render reflects it.
func (l *FrameLog) Append(c *Component) int {
	container := &layout.Input{
		Kind: layout.KindBox, Tag: "div", CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{c.Tree},
	}
	l.Host.Children = append(l.Host.Children, container)
	id := l.nextID
	l.nextID++
	l.frames = append(l.frames, logFrame{id: id, container: container, comp: c})
	return id
}

// Archive hands the rows of frame id — and every frame above it — to
// scrollback, then disposes them. Cumulative from the top: the zone
// can only release its top edge.
func (l *FrameLog) Archive(id int) {
	idx := l.frameIndex(id)
	if idx < 0 {
		return
	}
	rows := 0
	for _, f := range l.frames[:idx+1] {
		rows += f.container.LastH
	}
	if l.ReleaseTop != nil && rows > 0 {
		l.ReleaseTop(rows)
	}
	for _, f := range l.frames[:idx+1] {
		l.dispose(f)
	}
	l.frames = append([]logFrame{}, l.frames[idx+1:]...)
}

// Remove disposes a single frame without archiving — its rows are
// cleared and the zone reflows (for redaction).
func (l *FrameLog) Remove(id int) {
	idx := l.frameIndex(id)
	if idx < 0 {
		return
	}
	l.dispose(l.frames[idx])
	l.frames = append(l.frames[:idx], l.frames[idx+1:]...)
}

// LiveFrames returns the ids still mounted, in order.
func (l *FrameLog) LiveFrames() []int {
	ids := make([]int, len(l.frames))
	for i, f := range l.frames {
		ids[i] = f.id
	}
	return ids
}

func (l *FrameLog) frameIndex(id int) int {
	for i, f := range l.frames {
		if f.id == id {
			return i
		}
	}
	return -1
}

func (l *FrameLog) dispose(f logFrame) {
	if f.comp.Dispose != nil {
		f.comp.Dispose()
	}
	for i, c := range l.Host.Children {
		if c == f.container {
			l.Host.Children = append(l.Host.Children[:i], l.Host.Children[i+1:]...)
			return
		}
	}
}
