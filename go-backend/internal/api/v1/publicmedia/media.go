package publicmedia

import (
	"commerce-platform/internal/service"
)

type Resolver = service.PublicMediaURLResolver

func URL(resolver Resolver, value string) string {
	return service.CanonicalPublicMediaURL(resolver, value)
}
