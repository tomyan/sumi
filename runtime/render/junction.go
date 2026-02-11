package render

// JunctionChar returns the Unicode box-drawing character for a junction
// where borders extend in the given directions.
// For example, (up=false, right=true, down=true, left=false) returns '┌'.
func JunctionChar(up, right, down, left bool) rune {
	key := junctionKey(up, right, down, left)
	if ch, ok := junctionTable[key]; ok {
		return ch
	}
	return ' '
}

func junctionKey(up, right, down, left bool) uint8 {
	var k uint8
	if up {
		k |= 1
	}
	if right {
		k |= 2
	}
	if down {
		k |= 4
	}
	if left {
		k |= 8
	}
	return k
}

// junctionTable maps direction bitmask to box-drawing character.
// Bits: 1=up, 2=right, 4=down, 8=left.
var junctionTable = map[uint8]rune{
	2 | 4:         '┌', // right+down
	4 | 8:         '┐', // down+left
	1 | 2:         '└', // up+right
	1 | 8:         '┘', // up+left
	1 | 2 | 4:     '├', // up+right+down
	1 | 4 | 8:     '┤', // up+down+left
	2 | 4 | 8:     '┬', // right+down+left
	1 | 2 | 8:     '┴', // up+right+left
	1 | 2 | 4 | 8: '┼', // all four
	2 | 8:         '─', // right+left
	1 | 4:         '│', // up+down
}

// reverseJunctionTable maps box-drawing character back to direction bitmask.
var reverseJunctionTable = func() map[rune]uint8 {
	m := make(map[rune]uint8, len(junctionTable))
	for k, v := range junctionTable {
		m[v] = k
	}
	return m
}()
