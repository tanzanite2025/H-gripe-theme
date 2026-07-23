package faqcontent

import (
	"fmt"
	stdhtml "html"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

var allowedTags = map[string]bool{
	"p":      true,
	"br":     true,
	"strong": true,
	"em":     true,
	"a":      true,
}

var inlineAliasTags = map[string]string{
	"b": "strong",
	"i": "em",
}

var paragraphTags = map[string]bool{
	"div": true,
}

var removedTags = map[string]bool{
	"script": true,
	"style":  true,
	"iframe": true,
	"object": true,
	"embed":  true,
	"form":   true,
	"input":  true,
	"img":    true,
	"video":  true,
	"audio":  true,
}

// SanitizeAnswer normalizes FAQ answer content to the small HTML subset that
// the storefront FAQ renderer supports. Images are intentionally rejected
// here; FAQ images are stored in the dedicated answer image field.
func SanitizeAnswer(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type: html.ElementNode,
		Data: "div",
	})
	if err != nil {
		return "", fmt.Errorf("parse FAQ answer: %w", err)
	}

	var builder strings.Builder
	for _, node := range nodes {
		if err := renderNode(&builder, node); err != nil {
			return "", err
		}
	}

	return strings.TrimSpace(builder.String()), nil
}

func HasVisibleText(value string) bool {
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type: html.ElementNode,
		Data: "div",
	})
	if err != nil {
		return strings.TrimSpace(value) != ""
	}
	for _, node := range nodes {
		if nodeHasVisibleText(node) {
			return true
		}
	}
	return false
}

func nodeHasVisibleText(node *html.Node) bool {
	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data) != ""
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if nodeHasVisibleText(child) {
			return true
		}
	}
	return false
}

func renderNode(builder *strings.Builder, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		escaped := stdhtml.EscapeString(node.Data)
		builder.WriteString(escaped)
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
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if err := renderNode(builder, child); err != nil {
					return err
				}
			}
			builder.WriteString("</p>")
			return nil
		}
		if removedTags[tag] {
			return nil
		}
		if !allowedTags[tag] {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if err := renderNode(builder, child); err != nil {
					return err
				}
			}
			return nil
		}

		builder.WriteByte('<')
		builder.WriteString(tag)
		if tag == "a" {
			href := ""
			for _, attr := range node.Attr {
				if strings.EqualFold(attr.Key, "href") {
					href = strings.TrimSpace(attr.Val)
					break
				}
			}
			if href == "" || !allowedLink(href) {
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if err := renderNode(builder, child); err != nil {
						return err
					}
				}
				return nil
			}
			builder.WriteString(` href="`)
			builder.WriteString(stdhtml.EscapeString(href))
			builder.WriteByte('"')
		}
		builder.WriteByte('>')
		if tag != "br" {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if err := renderNode(builder, child); err != nil {
					return err
				}
			}
			builder.WriteString("</")
			builder.WriteString(tag)
			builder.WriteByte('>')
		}
		return nil
	default:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := renderNode(builder, child); err != nil {
				return err
			}
		}
		return nil
	}
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
