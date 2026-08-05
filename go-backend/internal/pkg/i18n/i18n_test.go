package i18n

import "testing"

func TestGetLocaleFromPathOnlyReturnsExplicitLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "localized storefront path", path: "/zh_cn/shop", want: "zh_cn"},
		{name: "localized hyphen path", path: "/zh-CN/shop", want: "zh_cn"},
		{name: "default locale path", path: "/en/shop", want: "en"},
		{name: "api path has no locale", path: "/api/v1/products", want: ""},
		{name: "plain storefront path has no locale", path: "/shop", want: ""},
		{name: "root path has no locale", path: "/", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := GetLocaleFromPath(tt.path); got != tt.want {
				t.Fatalf("GetLocaleFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
