package gallery

import (
	"encoding/json"
	"strings"
	"testing"

	gallerydomain "commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/service"
)

func TestPublicGalleryFromDomainCanonicalizesFirstPartyMediaURLs(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := &gallerydomain.Gallery{
		ID:         1,
		Name:       "Media Gallery",
		Slug:       "media-gallery",
		CoverImage: "http://media.internal:8080/uploads/gallery/cover.webp",
		Images: []gallerydomain.GalleryImage{
			{
				ID:        2,
				URL:       "http://media.internal:8080/uploads/gallery/full.webp",
				Thumbnail: "http://media.internal:8080/uploads/gallery/thumb.webp",
			},
		},
	}

	payload, err := json.Marshal(publicGalleryFromDomain(item, resolver))
	if err != nil {
		t.Fatalf("marshal public gallery: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "media.internal") {
		t.Fatalf("public gallery response exposes internal media origin: %s", body)
	}
	for _, expected := range []string{
		`"cover_image":"https://shop.example.test/uploads/gallery/cover.webp"`,
		`"url":"https://shop.example.test/uploads/gallery/full.webp"`,
		`"thumbnail":"https://shop.example.test/uploads/gallery/thumb.webp"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("public gallery response missing canonical media %s: %s", expected, body)
		}
	}
}
