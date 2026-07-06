package tui

import (
	"testing"

	"github.com/tomyan/sumi/runtime/layout"
)

// F3c: FrameLog — append-only frames (sumi components) streamed into
// the inline zone; archiving hands their rows to scrollback via
// ReleaseTop and unmounts them.

func frameComp(content string, disposed *[]string) *Component {
	return &Component{
		Tree: &layout.Input{Kind: layout.KindBox, Tag: "root", CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{
				{Kind: layout.KindText, Content: content, CursorCol: -1, CursorRow: -1},
			}},
		Dispose: func() { *disposed = append(*disposed, content) },
	}
}

func TestFrameLogAppendMountsUnderHost(t *testing.T) {
	// Given
	var disposed []string
	l := NewFrameLog()

	// When
	id1 := l.Append(frameComp("one", &disposed))
	id2 := l.Append(frameComp("two", &disposed))

	// Then
	if id1 != 1 || id2 != 2 {
		t.Errorf("ids = %d, %d, want 1, 2", id1, id2)
	}
	if len(l.Host.Children) != 2 {
		t.Fatalf("host children = %d, want 2", len(l.Host.Children))
	}
	if got := l.LiveFrames(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("live = %v, want [1 2]", got)
	}
}

func TestFrameLogArchiveIsCumulativeFromTop(t *testing.T) {
	// Given: three frames with laid-out heights 2, 3, 1.
	var disposed []string
	var released []int
	l := NewFrameLog()
	l.ReleaseTop = func(n int) { released = append(released, n) }
	l.Append(frameComp("one", &disposed))
	id2 := l.Append(frameComp("two", &disposed))
	l.Append(frameComp("three", &disposed))
	heights := []int{2, 3, 1}
	for i, c := range l.Host.Children {
		c.LastH = heights[i]
	}

	// When: archiving frame 2 archives frame 1 too.
	l.Archive(id2)

	// Then
	if len(released) != 1 || released[0] != 5 {
		t.Errorf("released = %v, want [5] (2+3)", released)
	}
	if len(disposed) != 2 || disposed[0] != "one" || disposed[1] != "two" {
		t.Errorf("disposed = %v, want [one two]", disposed)
	}
	if got := l.LiveFrames(); len(got) != 1 || got[0] != 3 {
		t.Errorf("live = %v, want [3]", got)
	}
	if len(l.Host.Children) != 1 {
		t.Errorf("host children = %d, want 1", len(l.Host.Children))
	}
}

func TestFrameLogRemoveDisposesWithoutRelease(t *testing.T) {
	// Given
	var disposed []string
	var released []int
	l := NewFrameLog()
	l.ReleaseTop = func(n int) { released = append(released, n) }
	l.Append(frameComp("one", &disposed))
	id2 := l.Append(frameComp("two", &disposed))

	// When: redaction — rows are cleared, not archived.
	l.Remove(id2)

	// Then
	if len(released) != 0 {
		t.Errorf("released = %v, want none", released)
	}
	if len(disposed) != 1 || disposed[0] != "two" {
		t.Errorf("disposed = %v, want [two]", disposed)
	}
	if got := l.LiveFrames(); len(got) != 1 || got[0] != 1 {
		t.Errorf("live = %v, want [1]", got)
	}
}

func TestFrameLogAppendAfterArchive(t *testing.T) {
	// Given
	var disposed []string
	l := NewFrameLog()
	l.ReleaseTop = func(int) {}
	id1 := l.Append(frameComp("one", &disposed))
	l.Archive(id1)

	// When
	id2 := l.Append(frameComp("two", &disposed))

	// Then: ids keep increasing; the new frame is live.
	if id2 != 2 {
		t.Errorf("id = %d, want 2", id2)
	}
	if got := l.LiveFrames(); len(got) != 1 || got[0] != 2 {
		t.Errorf("live = %v, want [2]", got)
	}
}
