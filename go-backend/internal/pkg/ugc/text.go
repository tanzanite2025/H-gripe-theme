package ugc

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

var ErrTextTooLong = errors.New("text exceeds maximum length")

func PlainText(value string, maxRunes int) (string, error) {
	document, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	appendText(&builder, document)
	result := strings.Join(strings.Fields(builder.String()), " ")
	if maxRunes > 0 && utf8.RuneCountInString(result) > maxRunes {
		return "", ErrTextTooLong
	}
	return result, nil
}

func appendText(builder *strings.Builder, node *html.Node) {
	if node == nil {
		return
	}
	if node.Type == html.TextNode {
		builder.WriteString(node.Data)
		return
	}
	if isDiscardedNode(node) {
		return
	}

	if isBlockNode(node) {
		builder.WriteByte(' ')
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		appendText(builder, child)
	}
	if isBlockNode(node) {
		builder.WriteByte(' ')
	}
}

func isDiscardedNode(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "script", "style", "iframe", "object", "embed", "form", "input", "svg", "math", "template":
		return true
	default:
		return false
	}
}

func isBlockNode(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "p", "br", "li", "ul", "ol", "h2", "h3", "h4", "blockquote", "pre":
		return true
	default:
		return false
	}
}
