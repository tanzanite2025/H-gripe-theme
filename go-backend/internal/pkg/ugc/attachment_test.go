package ugc

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeUploadImageAttachmentReferencesAcceptsUploadReferences(t *testing.T) {
	input := []string{
		" https://tanzanite.site/uploads/2026/08/02/photo.JPG?cache=1 ",
		"/uploads/2026/08/02/photo.JPG",
		"2026/08/02/other.webp",
	}

	got, err := NormalizeUploadImageAttachmentReferences(input, 4)
	if err != nil {
		t.Fatalf("NormalizeUploadImageAttachmentReferences() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 deduplicated attachments, got %d: %#v", len(got), got)
	}
	if got[0].StorageKey != "2026/08/02/photo.JPG" || got[0].Value != "/uploads/2026/08/02/photo.JPG" {
		t.Fatalf("unexpected first attachment: %#v", got[0])
	}
	if got[1].StorageKey != "2026/08/02/other.webp" || got[1].Value != "/uploads/2026/08/02/other.webp" {
		t.Fatalf("unexpected second attachment: %#v", got[1])
	}
}

func TestNormalizeUploadImageAttachmentReferencesRejectsDangerousReferences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{name: "data url", in: "data:image/png;base64,AAAA", want: ErrAttachmentInvalidURL},
		{name: "javascript url", in: "javascript:alert(1)", want: ErrAttachmentInvalidURL},
		{name: "path traversal", in: "/uploads/2026/../evil.png", want: ErrAttachmentInvalidURL},
		{name: "unsupported extension", in: "/uploads/evil.exe", want: ErrAttachmentInvalidType},
		{name: "non upload absolute path", in: "/assets/photo.jpg", want: ErrAttachmentInvalidURL},
		{name: "encoded traversal", in: "/uploads/2026/%2e%2e/evil.png", want: ErrAttachmentInvalidURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUploadImageAttachmentReferences([]string{tt.in}, 4)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNormalizeUploadImageAttachmentReferencesRejectsTooManyAndTooLong(t *testing.T) {
	_, err := NormalizeUploadImageAttachmentReferences([]string{"/uploads/a.jpg", "/uploads/b.jpg"}, 1)
	if !errors.Is(err, ErrAttachmentTooMany) {
		t.Fatalf("error = %v, want ErrAttachmentTooMany", err)
	}

	_, err = NormalizeUploadImageAttachmentReferences([]string{"/uploads/" + strings.Repeat("a", DefaultAttachmentReferenceMaxBytes) + ".jpg"}, 4)
	if !errors.Is(err, ErrAttachmentTooLong) {
		t.Fatalf("error = %v, want ErrAttachmentTooLong", err)
	}
}
