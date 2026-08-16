package service

import "strings"

func normalizeCatalogRoutePath(value string) string {
	path := strings.TrimSpace(value)
	if path == "" || !strings.HasPrefix(path, "/") {
		return ""
	}
	if path != "/" {
		path = "/" + strings.Trim(path, "/")
	}
	return path
}
