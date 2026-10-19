package ticket

import (
	"encoding/json"
	"fmt"
	"strings"
)

type StructuredMessageMediaField struct {
	Path              string
	URL               string
	ExpectedMediaType string
}

type StructuredMessageMediaDocument struct {
	Fields   []StructuredMessageMediaField
	metadata map[string]json.RawMessage
	original string
}

// Media dependencies follow the rendered message contract. Navigation URLs,
// captions and arbitrary nested strings are never treated as media addresses.
// Empty historical message types retain known image fields without guessing url.
func ParseStructuredMessageMetadataKnownMediaFields(messageType, metadata string) (*StructuredMessageMediaDocument, error) {
	document := &StructuredMessageMediaDocument{original: metadata}
	var fields []StructuredMessageMediaField
	switch strings.ToLower(strings.TrimSpace(messageType)) {
	case "link", "product", "order":
		fields = []StructuredMessageMediaField{{Path: "thumbnail", ExpectedMediaType: "image"}}
	case "faq":
		fields = []StructuredMessageMediaField{{Path: "answer_image_url", ExpectedMediaType: "image"}, {Path: "faq.answer_image_url", ExpectedMediaType: "image"}}
	case "config_confirm", "wheelset_selection_request":
		fields = []StructuredMessageMediaField{{Path: "product.thumbnail", ExpectedMediaType: "image"}}
	case "image":
		fields = []StructuredMessageMediaField{{Path: "url", ExpectedMediaType: "image"}, {Path: "image_url", ExpectedMediaType: "image"}, {Path: "thumbnail", ExpectedMediaType: "image"}}
	case "video":
		fields = []StructuredMessageMediaField{{Path: "url", ExpectedMediaType: "video"}, {Path: "video_url", ExpectedMediaType: "video"}, {Path: "thumbnail", ExpectedMediaType: "image"}}
	case "":
		fields = []StructuredMessageMediaField{{Path: "thumbnail", ExpectedMediaType: "image"}, {Path: "image_url", ExpectedMediaType: "image"}, {Path: "video_url", ExpectedMediaType: "video"}, {Path: "answer_image_url", ExpectedMediaType: "image"}, {Path: "product.thumbnail", ExpectedMediaType: "image"}, {Path: "faq.answer_image_url", ExpectedMediaType: "image"}}
	default:
		return document, nil
	}
	if strings.TrimSpace(metadata) == "" {
		return document, nil
	}
	if err := json.Unmarshal([]byte(metadata), &document.metadata); err != nil {
		return nil, fmt.Errorf("structured message metadata must be a JSON object")
	}
	for _, field := range fields {
		value, err := readStructuredMessageMediaFieldAtPath(document.metadata, strings.Split(field.Path, "."))
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(value) != "" {
			field.URL = value
			document.Fields = append(document.Fields, field)
		}
	}
	return document, nil
}

func readStructuredMessageMediaFieldAtPath(object map[string]json.RawMessage, path []string) (string, error) {
	raw, exists := object[path[0]]
	if !exists || string(raw) == "null" {
		return "", nil
	}
	if len(path) > 1 {
		var child map[string]json.RawMessage
		if err := json.Unmarshal(raw, &child); err != nil {
			return "", fmt.Errorf("structured message media container %s must be an object", path[0])
		}
		return readStructuredMessageMediaFieldAtPath(child, path[1:])
	}
	var reference string
	if err := json.Unmarshal(raw, &reference); err != nil {
		return "", fmt.Errorf("structured message media field %s must be a string", path[0])
	}
	return reference, nil
}

func (document *StructuredMessageMediaDocument) RenderMetadataWithCanonicalMediaFields() (string, error) {
	changed := false
	for _, field := range document.Fields {
		path := strings.Split(field.Path, ".")
		previous, err := readStructuredMessageMediaFieldAtPath(document.metadata, path)
		if err != nil {
			return "", err
		}
		if previous == field.URL {
			continue
		}
		if err := replaceStructuredMessageMediaFieldAtPath(document.metadata, path, field.URL); err != nil {
			return "", err
		}
		changed = true
	}
	if !changed {
		return document.original, nil
	}
	encoded, err := json.Marshal(document.metadata)
	return string(encoded), err
}

func replaceStructuredMessageMediaFieldAtPath(object map[string]json.RawMessage, path []string, reference string) error {
	if len(path) == 1 {
		encoded, err := json.Marshal(reference)
		object[path[0]] = encoded
		return err
	}
	var child map[string]json.RawMessage
	if err := json.Unmarshal(object[path[0]], &child); err != nil {
		return err
	}
	if err := replaceStructuredMessageMediaFieldAtPath(child, path[1:], reference); err != nil {
		return err
	}
	encoded, err := json.Marshal(child)
	object[path[0]] = encoded
	return err
}
