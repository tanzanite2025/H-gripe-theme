package localbackup

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

// Reuse the same deployment inputs as the API and storefront, without a DNS
// lookup, an HTTP request or a second domain-binding step. Explicit origins
// remain available for references saved before an application's URL changed.
func (cfg *Configuration) inheritMediaOriginsFromExistingServerAndStorefrontDeploymentURLs() error {
	origins := make([]string, 0, len(cfg.PublicOrigins)+2)
	seen := map[string]bool{}
	appendOrigin := func(value string) error {
		parsed, err := url.Parse(strings.TrimSpace(value))
		if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.Opaque != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
			return errors.New("local backup media origins require absolute HTTP or HTTPS URLs without credentials, queries or fragments")
		}
		origin := parsed.Scheme + "://" + strings.ToLower(parsed.Host)
		if !seen[origin] {
			seen[origin] = true
			origins = append(origins, origin)
		}
		return nil
	}
	for _, origin := range cfg.PublicOrigins {
		if err := appendOrigin(origin); err != nil {
			return err
		}
	}
	for _, name := range []string{"SERVER_BASE_URL", "STOREFRONT_BASE_URL"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			if err := appendOrigin(value); err != nil {
				// Never include a deployment value in an error: invalid input may
				// contain credentials even though valid inputs cannot.
				return errors.New("invalid " + name + " for local backup media origin recognition")
			}
		}
	}
	cfg.PublicOrigins = origins
	return nil
}
