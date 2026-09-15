package websitecontent

import (
	stdhtml "html"
	"strings"

	"commerce-platform/internal/pkg/safehtml"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// NormalizeWebsiteBody keeps legacy plain-text settings readable while
// allowing the admin editor to persist a small, safe rich-text HTML subset.
func NormalizeWebsiteBody(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
	if value == "" {
		return "", nil
	}

	if !containsElement(value) {
		return plainTextToHTML(value), nil
	}

	return safehtml.Sanitize(value)
}

func containsElement(value string) bool {
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
	})
	if err != nil {
		return false
	}

	var visit func(*html.Node) bool
	visit = func(node *html.Node) bool {
		if node.Type == html.ElementNode {
			return true
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if visit(child) {
				return true
			}
		}
		return false
	}

	for _, node := range nodes {
		if visit(node) {
			return true
		}
	}
	return false
}

func plainTextToHTML(value string) string {
	paragraphs := strings.Split(value, "\n\n")
	result := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		result = append(result, "<p>"+strings.ReplaceAll(stdhtml.EscapeString(paragraph), "\n", "<br>")+"</p>")
	}
	return strings.Join(result, "")
}
