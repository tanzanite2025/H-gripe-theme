package locales

import "testing"

func TestResolveSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{name: "simplified chinese base", input: "zh", output: "zh_cn"},
		{name: "simplified chinese region", input: "zh-CN", output: "zh_cn"},
		{name: "simplified chinese underscore", input: "zh_CN", output: "zh_cn"},
		{name: "english region", input: "en-US", output: "en"},
		{name: "pidgin region", input: "pcm-NG", output: "pcm"},
		{name: "accept language fragment", input: "fr-FR;q=0.8", output: "fr"},
		{name: "unsupported", input: "zz-ZZ", output: ""},
		{name: "traditional chinese unsupported", input: "zh-TW", output: ""},
		{name: "empty", input: "", output: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ResolveSupported(tt.input); got != tt.output {
				t.Fatalf("ResolveSupported(%q) = %q, want %q", tt.input, got, tt.output)
			}
		})
	}
}

func TestNormalizeDefaultsEmptyToEnglish(t *testing.T) {
	t.Parallel()

	if got := Normalize(""); got != "en" {
		t.Fatalf("Normalize(\"\") = %q, want en", got)
	}
}
