package upload

import (
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
)

const SiteLogoSVGDimension = 48

var SiteLogoSVGRule = SVGRule{
	MaxSize:     1 << 20,
	ExactWidth:  SiteLogoSVGDimension,
	ExactHeight: SiteLogoSVGDimension,
}

type SVGRule struct {
	MaxSize     int64
	ExactWidth  int
	ExactHeight int
}

func ValidateSVGFile(file *multipart.FileHeader, rule SVGRule) error {
	if file == nil || file.Size <= 0 {
		return validationError(CodeEmptyFile, "empty_file: uploaded file is empty")
	}
	if rule.MaxSize > 0 && file.Size > rule.MaxSize {
		return validationError(CodeFileTooLarge, "file_too_large: %s exceeds %s", file.Filename, formatBytes(rule.MaxSize))
	}
	if !strings.EqualFold(filepath.Ext(file.Filename), ".svg") {
		return validationError(CodeInvalidType, "invalid_type: %s has an unsupported file extension", file.Filename)
	}

	width, height, err := ReadSVGDimensions(file)
	if err != nil {
		return err
	}
	if rule.ExactWidth > 0 && rule.ExactHeight > 0 && (width != rule.ExactWidth || height != rule.ExactHeight) {
		return validationError(
			CodeInvalidDimensions,
			"invalid_dimensions: %s must be exactly %dx%d (received %dx%d)",
			file.Filename,
			rule.ExactWidth,
			rule.ExactHeight,
			width,
			height,
		)
	}
	return nil
}

func ReadSVGDimensions(file *multipart.FileHeader) (int, int, error) {
	if file == nil || file.Size <= 0 {
		return 0, 0, validationError(CodeEmptyFile, "empty_file: uploaded file is empty")
	}

	src, err := file.Open()
	if err != nil {
		return 0, 0, validationError(CodeInvalidType, "invalid_type: unable to read SVG file")
	}
	defer func() { _ = src.Close() }()

	const maxSVGInspectionBytes = 1 << 20
	payload, err := io.ReadAll(io.LimitReader(src, maxSVGInspectionBytes+1))
	if err != nil {
		return 0, 0, validationError(CodeInvalidType, "invalid_type: unable to read SVG file")
	}
	if len(payload) > maxSVGInspectionBytes {
		return 0, 0, validationError(CodeFileTooLarge, "file_too_large: SVG content exceeds 1MB")
	}

	decoder := xml.NewDecoder(bytes.NewReader(payload))
	var root xml.StartElement
	rootSeen := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG content is not valid XML")
		}

		switch value := token.(type) {
		case xml.Directive:
			return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG directives are not allowed")
		case xml.ProcInst:
			if !strings.EqualFold(strings.TrimSpace(value.Target), "xml") {
				return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG processing instructions are not allowed")
			}
		case xml.StartElement:
			if !rootSeen {
				if !strings.EqualFold(value.Name.Local, "svg") {
					return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG root element is missing")
				}
				root = value
				rootSeen = true
			}
			if !safeSiteLogoSVGElement(value) {
				return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG contains unsupported active content")
			}
		}
	}

	if !rootSeen {
		return 0, 0, validationError(CodeInvalidType, "invalid_type: SVG root element is missing")
	}

	width, height, ok := svgRootDimensions(root)
	if !ok {
		return 0, 0, validationError(CodeInvalidDimensions, "invalid_dimensions: SVG must define a valid width and height")
	}
	return width, height, nil
}

func svgRootDimensions(root xml.StartElement) (int, int, bool) {
	widthValue := svgAttribute(root.Attr, "width")
	heightValue := svgAttribute(root.Attr, "height")
	if widthValue != "" || heightValue != "" {
		width, widthOK := parseSVGLength(widthValue)
		height, heightOK := parseSVGLength(heightValue)
		return width, height, widthOK && heightOK
	}

	viewBox := svgAttribute(root.Attr, "viewBox")
	parts := strings.FieldsFunc(viewBox, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if len(parts) != 4 {
		return 0, 0, false
	}

	width, widthOK := parseSVGNumber(parts[2])
	height, heightOK := parseSVGNumber(parts[3])
	return width, height, widthOK && heightOK
}

func svgAttribute(attributes []xml.Attr, name string) string {
	for _, attribute := range attributes {
		if strings.EqualFold(attribute.Name.Local, name) {
			return strings.TrimSpace(attribute.Value)
		}
	}
	return ""
}

func parseSVGLength(value string) (int, bool) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(normalized, "px") {
		normalized = strings.TrimSpace(strings.TrimSuffix(normalized, "px"))
	}
	return parseSVGNumber(normalized)
}

func parseSVGNumber(value string) (int, bool) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || number <= 0 || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false
	}
	rounded := math.Round(number)
	if math.Abs(number-rounded) > 1e-9 {
		return 0, false
	}
	return int(rounded), true
}

func safeSiteLogoSVGElement(element xml.StartElement) bool {
	switch strings.ToLower(strings.TrimSpace(element.Name.Local)) {
	case "script", "style", "foreignobject", "iframe", "object", "embed", "audio", "video", "animate", "animatemotion", "animatetransform", "set":
		return false
	}

	for _, attribute := range element.Attr {
		name := strings.ToLower(strings.TrimSpace(attribute.Name.Local))
		value := strings.ToLower(strings.TrimSpace(attribute.Value))
		if strings.HasPrefix(name, "on") {
			return false
		}
		if (name == "href" || name == "src") && value != "" && !strings.HasPrefix(value, "#") {
			return false
		}
		if name == "style" && (strings.Contains(value, "url(") || strings.Contains(value, "@import") || strings.Contains(value, "expression(")) {
			return false
		}
	}
	return true
}
