package service

import (
	"errors"
	"testing"
)

func TestValidateFAQLocaleUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		existing   string
		requested  string
		wantLocale string
		wantErr    error
	}{
		{
			name:       "same locale",
			existing:   "en",
			requested:  "en",
			wantLocale: "en",
		},
		{
			name:       "same locale alias",
			existing:   "en",
			requested:  "en-US",
			wantLocale: "en",
		},
		{
			name:      "different locale is rejected",
			existing:  "en",
			requested: "fr",
			wantErr:   ErrFAQLocaleImmutable,
		},
		{
			name:      "unsupported requested locale is rejected",
			existing:  "en",
			requested: "zz",
			wantErr:   ErrUnsupportedLocale,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotLocale, err := validateFAQLocaleUpdate(tt.existing, tt.requested)
			if gotLocale != tt.wantLocale {
				t.Fatalf("validateFAQLocaleUpdate() locale = %q, want %q", gotLocale, tt.wantLocale)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateFAQLocaleUpdate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
