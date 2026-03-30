package pagination

import "testing"

func TestGetPageOffset(t *testing.T) {
	tests := []struct {
		page, size, want int64
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
		{1, 100, 0},
		{5, 25, 100},
	}
	for _, tt := range tests {
		got := GetPageOffset(tt.page, tt.size)
		if got != tt.want {
			t.Errorf("GetPageOffset(%d, %d) = %d, want %d", tt.page, tt.size, got, tt.want)
		}
	}
}
