package middleware

import (
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

const trustedEdgeCountryKey = "trusted_edge_country"

var trustedEdgeCountryHeaders = []string{
	"CF-IPCountry",
	"CloudFront-Viewer-Country",
	"X-Vercel-IP-Country",
	"X-Country-Code",
}

// NewTrustedEdgeMetadata accepts edge-provided metadata only when the
// immediate peer belongs to the configured trusted proxy networks.
func NewTrustedEdgeMetadata(trustedProxies []string) (gin.HandlerFunc, error) {
	networks, err := parseTrustedProxyNetworks(trustedProxies)
	if err != nil {
		return nil, err
	}

	return func(c *gin.Context) {
		if !isTrustedProxyPeer(c.Request.RemoteAddr, networks) {
			c.Next()
			return
		}

		if country := normalizeEdgeCountry(firstNonEmptyHeader(c, trustedEdgeCountryHeaders...)); country != "" {
			c.Set(trustedEdgeCountryKey, country)
		}
		c.Next()
	}, nil
}

func TrustedEdgeCountry(c *gin.Context) string {
	if c == nil {
		return ""
	}
	value, exists := c.Get(trustedEdgeCountryKey)
	if !exists {
		return ""
	}
	return normalizeEdgeCountry(fmt.Sprint(value))
}

func parseTrustedProxyNetworks(values []string) ([]*net.IPNet, error) {
	networks := make([]*net.IPNet, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if strings.Contains(value, "/") {
			ip, network, err := net.ParseCIDR(value)
			if err != nil {
				return nil, fmt.Errorf("invalid trusted proxy network %q: %w", value, err)
			}
			network.IP = ip
			networks = append(networks, network)
			continue
		}
		ip := net.ParseIP(value)
		if ip == nil {
			return nil, fmt.Errorf("invalid trusted proxy address %q", value)
		}
		bits := 128
		if ip.To4() != nil {
			ip = ip.To4()
			bits = 32
		}
		networks = append(networks, &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(bits, bits),
		})
	}
	return networks, nil
}

func isTrustedProxyPeer(remoteAddr string, networks []*net.IPNet) bool {
	if len(networks) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func firstNonEmptyHeader(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeEdgeCountry(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || strings.EqualFold(value, "XX") || len(value) != 2 {
		return ""
	}
	return value
}
