package i18n

import "testing"

func TestResolveSupportedLocale(t *testing.T) {
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
		{name: "french region", input: "fr-FR", output: "fr"},
		{name: "portuguese region", input: "pt-BR", output: "pt"},
		{name: "unsupported", input: "zz-ZZ", output: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveSupportedLocale(tt.input); got != tt.output {
				t.Fatalf("resolveSupportedLocale(%q) = %q, want %q", tt.input, got, tt.output)
			}
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		output string
	}{
		{name: "uses first supported locale", header: "de-DE,de;q=0.9,en;q=0.8", output: "de"},
		{name: "maps chinese browser locale", header: "zh-CN,zh;q=0.9,en;q=0.8", output: "zh_cn"},
		{name: "skips unsupported locale", header: "zz-ZZ,fr-FR;q=0.9", output: "fr"},
		{name: "falls back to english", header: "zz-ZZ", output: "en"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := parseAcceptLanguage(tt.header); got != tt.output {
				t.Fatalf("parseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.output)
			}
		})
	}
}

func TestIsValidLocale(t *testing.T) {
	t.Parallel()

	if !isValidLocale("zh-CN") {
		t.Fatal("expected zh-CN to resolve to supported zh_cn locale")
	}

	if isValidLocale("zh-TW") {
		t.Fatal("expected zh-TW to be unsupported until a Nuxt locale exists")
	}
}
