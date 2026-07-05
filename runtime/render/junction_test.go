package render

import "testing"

func TestJunctionChar(t *testing.T) {
	tests := []struct {
		name                  string
		up, right, down, left bool
		want                  rune
	}{
		{"right+down = ┌", false, true, true, false, '┌'},
		{"down+left = ┐", false, false, true, true, '┐'},
		{"up+right = └", true, true, false, false, '└'},
		{"up+left = ┘", true, false, false, true, '┘'},
		{"up+right+down = ├", true, true, true, false, '├'},
		{"up+down+left = ┤", true, false, true, true, '┤'},
		{"right+down+left = ┬", false, true, true, true, '┬'},
		{"up+right+left = ┴", true, true, false, true, '┴'},
		{"all four = ┼", true, true, true, true, '┼'},
		{"right+left = ─", false, true, false, true, '─'},
		{"up+down = │", true, false, true, false, '│'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			got := JunctionChar(tt.up, tt.right, tt.down, tt.left)

			// Then
			if got != tt.want {
				t.Errorf("JunctionChar(%v,%v,%v,%v) = %c, want %c",
					tt.up, tt.right, tt.down, tt.left, got, tt.want)
			}
		})
	}
}
