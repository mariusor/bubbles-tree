package tree

import "testing"

func Test_width(t *testing.T) {
	tests := []struct {
		name string
		arg  Symbols
		want int
	}{
		{
			name: "normal symbols",
			arg:  normalSymbols,
			want: 3,
		},
		{
			name: "thick symbols",
			arg:  thickSymbols,
			want: 3,
		},
		{
			name: "normal edge symbols",
			arg:  normalEdgeSymbols,
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := width(tt.arg); got != tt.want {
				t.Errorf("width() = %v, want %v", got, tt.want)
			}
		})
	}
}
