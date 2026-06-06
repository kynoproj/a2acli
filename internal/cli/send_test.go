package cli

import "testing"

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", nil, ""},
		{"single", []string{"hello"}, "hello"},
		{"multiple", []string{"hello", "world"}, "hello world"},
		{"preserve-empty-between", []string{"a", "", "b"}, "a  b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinArgs(tt.in); got != tt.want {
				t.Errorf("joinArgs(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
