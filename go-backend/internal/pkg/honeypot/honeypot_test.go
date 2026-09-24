package honeypot

import "testing"

func TestFilled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "whitespace", value: " \t\n", want: false},
		{name: "value", value: "https://example.test", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filled(tt.value); got != tt.want {
				t.Fatalf("Filled(%q) = %t, want %t", tt.value, got, tt.want)
			}
		})
	}
}

func TestPolicyModes(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{mode: ModeOff, want: false},
		{mode: ModeShadow, want: false},
		{mode: ModeEnforce, want: true},
		{mode: "", want: true},
		{mode: "unknown", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := NewPolicy(tt.mode).ShouldDrop("filled", "test", "field", "/test"); got != tt.want {
				t.Fatalf("NewPolicy(%q).ShouldDrop() = %t, want %t", tt.mode, got, tt.want)
			}
		})
	}
}
