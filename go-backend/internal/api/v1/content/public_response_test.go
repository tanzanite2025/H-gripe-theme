package content

import (
	"encoding/json"
	"strings"
	"testing"

	postdomain "commerce-platform/internal/domain/post"
	"commerce-platform/internal/service"
)

func TestPublicPostFromDomainCanonicalizesFeaturedImageURL(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	post := postdomain.Post{
		ID:          1,
		Title:       "Media Post",
		Slug:        "media-post",
		FeaturedImg: "http://media.internal:8080/uploads/posts/cover.webp",
		Status:      "published",
	}

	payload, err := json.Marshal(PublicPostFromDomain(post, resolver))
	if err != nil {
		t.Fatalf("marshal public post: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "media.internal") {
		t.Fatalf("public post response exposes internal media origin: %s", body)
	}
	if !strings.Contains(body, `"featured_image":"https://shop.example.test/uploads/posts/cover.webp"`) {
		t.Fatalf("public post response missing canonical featured image: %s", body)
	}
}
