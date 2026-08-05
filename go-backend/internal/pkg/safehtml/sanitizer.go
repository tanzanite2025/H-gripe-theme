package safehtml

import (
	"fmt"
	stdhtml "html"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var allowedTags = map[string]bool{
	"p":          true,
	"br":         true,
	"strong":     true,
	"em":         true,
	"a":          true,
	"ul":         true,
	"ol":         true,
	"li":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"blockquote": true,
	"code":       true,
	"pre":        true,
	"figure":     true,
	"figcaption": true,
	"img":        true,
	"video":      true,
}

var inlineAliasTags = map[string]string{
	"b": "strong",
	"i": "em",
}

var paragraphTags = map[string]bool{
	"div":     true,
	"section": true,
	"article": true,
}

var removedTags = map[string]bool{
	"script":   true,
	"style":    true,
	"iframe":   true,
	"object":   true,
	"embed":    true,
	"form":     true,
	"input":    true,
	"audio":    true,
	"svg":      true,
	"math":     true,
	"template": true,
}

func Sanitize(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
	})
	if err != nil {
		return "", fmt.Errorf("parse rich html: %w", err)
	}

	var builder strings.Builder
	for _, node := range nodes {
		if err := renderNode(&builder, node); err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(builder.String()), nil
}

func renderNode(builder *strings.Builder, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		builder.WriteString(stdhtml.EscapeString(node.Data))
		return nil
	case html.CommentNode:
		return nil
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		if alias, ok := inlineAliasTags[tag]; ok {
			tag = alias
		}
		if paragraphTags[tag] {
			builder.WriteString("<p>")
			if err := renderChildren(builder, node); err != nil {
				return err
			}
			builder.WriteString("</p>")
			return nil
		}
		if removedTags[tag] {
			return nil
		}
		if !allowedTags[tag] {
			return renderChildren(builder, node)
		}
		if tag == "img" {
			return renderImage(builder, node)
		}
		if tag == "video" {
			return renderVideo(builder, node)
		}

		builder.WriteByte('<')
		builder.WriteString(tag)
		if tag == "a" {
			href := linkHref(node)
			if href == "" || !allowedLink(href) {
				return renderChildren(builder, node)
			}
			builder.WriteString(` href="`)
			builder.WriteString(stdhtml.EscapeString(href))
			builder.WriteByte('"')
			builder.WriteString(` rel="nofollow noopener noreferrer"`)
		}
		builder.WriteByte('>')
		if tag != "br" {
			if err := renderChildren(builder, node); err != nil {
				return err
			}
			builder.WriteString("</")
			builder.WriteString(tag)
			builder.WriteByte('>')
		}
		return nil
	default:
		return renderChildren(builder, node)
	}
}

func renderImage(builder *strings.Builder, node *html.Node) error {
	src := nodeAttr(node, "src")
	if src == "" || !allowedMediaSource(src) {
		return nil
	}

	builder.WriteString(`<img src="`)
	builder.WriteString(stdhtml.EscapeString(src))
	builder.WriteByte('"')
	builder.WriteString(` alt="`)
	builder.WriteString(stdhtml.EscapeString(nodeAttr(node, "alt")))
	builder.WriteByte('"')
	builder.WriteString(` loading="lazy">`)
	return nil
}

func renderVideo(builder *strings.Builder, node *html.Node) error {
	src := nodeAttr(node, "src")
	if src == "" || !allowedMediaSource(src) {
		return nil
	}

	builder.WriteString(`<video src="`)
	builder.WriteString(stdhtml.EscapeString(src))
	builder.WriteByte('"')
	if poster := nodeAttr(node, "poster"); poster != "" && allowedMediaSource(poster) {
		builder.WriteString(` poster="`)
		builder.WriteString(stdhtml.EscapeString(poster))
		builder.WriteByte('"')
	}
	builder.WriteString(` controls preload="metadata"></video>`)
	return nil
}

func renderChildren(builder *strings.Builder, node *html.Node) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := renderNode(builder, child); err != nil {
			return err
		}
	}
	return nil
}

func linkHref(node *html.Node) string {
	return nodeAttr(node, "href")
}

func nodeAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func allowedLink(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto":
		return true
	default:
		return parsed.Scheme == "" && strings.HasPrefix(value, "/")
	}
}

func allowedMediaSource(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return true
	default:
		return parsed.Scheme == "" && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//")
	}
}
