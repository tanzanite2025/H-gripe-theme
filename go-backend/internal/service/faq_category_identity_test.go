package service

import (
	"errors"
	"tanzanite/internal/domain/faq"
	"testing"
)

func TestValidateFAQCategoryIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing faq.FAQCategory
		next     faq.FAQCategory
		wantErr  error
	}{
		{
			name: "display fields can change",
			existing: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "en",
				Name:        "Billing",
			},
			next: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "en",
				Name:        "Billing and Charges",
			},
		},
		{
			name: "locale alias keeps identity",
			existing: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "zh_cn",
			},
			next: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "zh-CN",
			},
		},
		{
			name: "category key cannot change",
			existing: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "en",
			},
			next: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "charges",
				Locale:      "en",
			},
			wantErr: ErrFAQCategoryIdentityImmutable,
		},
		{
			name: "page cannot change",
			existing: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "en",
			},
			next: faq.FAQCategory{
				PageID:      "support-shipping",
				CategoryKey: "billing",
				Locale:      "en",
			},
			wantErr: ErrFAQCategoryIdentityImmutable,
		},
		{
			name: "locale cannot change",
			existing: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "en",
			},
			next: faq.FAQCategory{
				PageID:      "support-payment",
				CategoryKey: "billing",
				Locale:      "fr",
			},
			wantErr: ErrFAQCategoryIdentityImmutable,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateFAQCategoryIdentity(&tt.existing, &tt.next)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateFAQCategoryIdentity() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
