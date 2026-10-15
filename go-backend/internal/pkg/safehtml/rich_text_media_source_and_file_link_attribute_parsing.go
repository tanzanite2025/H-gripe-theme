package safehtml

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type RichTextMediaAttributeReference struct {
	URL                 *string
	ElementAndAttribute string
	ExpectedMediaType   string
}

type RichTextMediaReferenceDocument struct {
	nodes        []*html.Node
	original     string
	originalURLs []string
	References   []RichTextMediaAttributeReference
}

// Parse the same media attributes supported by Sanitize. Text, alt labels and
// coincidental URL substrings are not media placements.
func ParseRichTextMediaSourceAndFileLinkAttributes(value string) (*RichTextMediaReferenceDocument, error) {
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: "div"})
	if err != nil {
		return nil, fmt.Errorf("parse media references in rich text: %w", err)
	}
	document := &RichTextMediaReferenceDocument{nodes: nodes, original: value}
	var collectRichTextMediaAttributes func(*html.Node)
	collectRichTextMediaAttributes = func(node *html.Node) {
		if node.Type == html.ElementNode && removedTags[node.Data] {
			return
		}
		for index := range node.Attr {
			attribute := &node.Attr[index]
			expectedType, supported := "", false
			switch node.Data + "." + attribute.Key {
			case "img.src", "video.poster":
				expectedType, supported = "image", true
			case "video.src":
				expectedType, supported = "video", true
			case "a.href":
				parsed, parseError := url.Parse(strings.TrimSpace(attribute.Val))
				supported = parseError == nil && (parsed.Scheme == "" || strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https"))
			}
			if !supported || strings.TrimSpace(attribute.Val) == "" {
				continue
			}
			document.References = append(document.References, RichTextMediaAttributeReference{URL: &attribute.Val, ElementAndAttribute: node.Data + "." + attribute.Key, ExpectedMediaType: expectedType})
			document.originalURLs = append(document.originalURLs, attribute.Val)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collectRichTextMediaAttributes(child)
		}
	}
	for _, node := range nodes {
		collectRichTextMediaAttributes(node)
	}
	return document, nil
}

func (document *RichTextMediaReferenceDocument) RenderRichTextWithCanonicalMediaReferences() (string, error) {
	changed := false
	for index, reference := range document.References {
		changed = changed || *reference.URL != document.originalURLs[index]
	}
	if !changed {
		return document.original, nil
	}
	var output strings.Builder
	for _, node := range document.nodes {
		if err := html.Render(&output, node); err != nil {
			return "", err
		}
	}
	return output.String(), nil
}
