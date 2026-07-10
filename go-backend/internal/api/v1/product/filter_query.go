package product

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseOptionalFloatQuery(c *gin.Context, key string) *float64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &value
}

func parseSpecFilterQuery(c *gin.Context) map[string][]string {
	filters := make(map[string][]string)
	query := c.Request.URL.Query()

	for key, values := range query {
		switch {
		case key == "attributes":
			for _, raw := range values {
				mergeAttributeJSON(filters, raw)
			}
		case strings.HasPrefix(key, "attributes["):
			slug := strings.TrimPrefix(key, "attributes[")
			switch {
			case strings.HasSuffix(slug, "][]"):
				slug = strings.TrimSuffix(slug, "][]")
			case strings.HasSuffix(slug, "]"):
				slug = strings.TrimSuffix(slug, "]")
			}
			appendSpecFilterValues(filters, slug, values)
		case strings.HasPrefix(key, "attributes."):
			slug := strings.TrimPrefix(key, "attributes.")
			slug = strings.TrimSuffix(slug, "[]")
			appendSpecFilterValues(filters, slug, values)
		}
	}

	return filters
}

func mergeAttributeJSON(filters map[string][]string, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}

	var listValues map[string][]string
	if err := json.Unmarshal([]byte(raw), &listValues); err == nil {
		for slug, values := range listValues {
			appendSpecFilterValues(filters, slug, values)
		}
		return
	}

	var singleValues map[string]string
	if err := json.Unmarshal([]byte(raw), &singleValues); err == nil {
		for slug, value := range singleValues {
			appendSpecFilterValues(filters, slug, []string{value})
		}
	}
}

func appendSpecFilterValues(filters map[string][]string, slug string, values []string) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return
	}

	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			filters[slug] = append(filters[slug], value)
		}
	}
}
