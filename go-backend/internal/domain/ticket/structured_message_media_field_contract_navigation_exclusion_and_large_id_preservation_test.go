package ticket

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStructuredMessageMediaFieldsFollowTypeAndExcludeNavigationAndUnrelatedStrings(t *testing.T) {
	for _, scenario := range []struct {
		messageType, metadata string
		paths                 []string
	}{
		{"product", `{"url":"/products/wheel","thumbnail":"/uploads/photo.png","caption":"/uploads/caption.png"}`, []string{"thumbnail"}},
		{"link", `{"url":"/support","thumbnail":"/uploads/photo.png"}`, []string{"thumbnail"}},
		{"order", `{"url":"/orders/1","thumbnail":"/uploads/photo.png"}`, []string{"thumbnail"}},
		{"faq", `{"url":"/faq","answer_image_url":"/uploads/top.png","faq":{"answer_image_url":"/uploads/nested.png"}}`, []string{"answer_image_url", "faq.answer_image_url"}},
		{"video", `{"url":"/uploads/video.mp4","video_url":"/uploads/other.mp4","thumbnail":"/uploads/poster.png"}`, []string{"url", "video_url", "thumbnail"}},
		{"image", `{"url":"/uploads/photo.png","image_url":"/uploads/other.png"}`, []string{"url", "image_url"}},
		{"config_confirm", `{"product":{"url":"/products/1","thumbnail":"/uploads/photo.png"}}`, []string{"product.thumbnail"}},
		{"wheelset_selection_request", `{"product":{"thumbnail":"/uploads/photo.png"}}`, []string{"product.thumbnail"}},
		{"text", `{"url":"/uploads/navigation.png","thumbnail":"/uploads/unused.png"}`, nil},
		{"", `{"url":"/uploads/ambiguous.png","thumbnail":"/uploads/legacy.png"}`, []string{"thumbnail"}},
	} {
		t.Run(scenario.messageType, func(t *testing.T) {
			document, err := ParseStructuredMessageMetadataKnownMediaFields(scenario.messageType, scenario.metadata)
			require.NoError(t, err)
			var paths []string
			for _, field := range document.Fields {
				paths = append(paths, field.Path)
			}
			require.Equal(t, scenario.paths, paths)
		})
	}
}

func TestStructuredMessageMediaRewritingPreservesLargeIdentifiersAndUnknownPayloads(t *testing.T) {
	document, err := ParseStructuredMessageMetadataKnownMediaFields("config_confirm", `{"product":{"id":9007199254740997,"thumbnail":"/uploads/alias.png"},"selection":{"raw":9007199254740999}}`)
	require.NoError(t, err)
	document.Fields[0].URL = "https://media.example.test/uploads/original.png"
	rendered, err := document.RenderMetadataWithCanonicalMediaFields()
	require.NoError(t, err)
	require.Contains(t, rendered, `9007199254740997`)
	require.Contains(t, rendered, `9007199254740999`)
	require.Contains(t, rendered, `https://media.example.test/uploads/original.png`)
	for _, malformed := range []string{`{`, `[]`, `{"product":[]}`, `{"product":{"thumbnail":9}}`} {
		_, err := ParseStructuredMessageMetadataKnownMediaFields("config_confirm", malformed)
		require.Error(t, err)
		require.NotContains(t, err.Error(), malformed)
	}
}
