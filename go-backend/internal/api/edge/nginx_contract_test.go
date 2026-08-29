package edge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThemeWebNginxProtectsRenderedAndStaticTraffic(t *testing.T) {
	root := findThemeWebRepoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "deployment", "nginx", "theme-web.conf"))
	require.NoError(t, err)

	config := string(content)
	servers := nginxServerBlocks(config)
	require.Len(t, servers, 2, "theme-web must keep storefront and admin server blocks")

	for _, server := range servers {
		require.Contains(t, server, "auth_request /_internal/security/ip-block-check;")
		require.Contains(t, server, "location = /_internal/security/ip-block-check")
		require.Contains(t, server, "internal;")
		require.Contains(t, server, "proxy_pass http://api-ingress:9000/_internal/security/ip-block-check;")
		require.Contains(t, server, "proxy_pass_request_body off;")
		require.Contains(t, server, "proxy_set_header X-Real-IP $remote_addr;")
		require.Contains(t, server, "proxy_set_header X-Forwarded-For $remote_addr;")
		require.Contains(t, server, "location = /healthz")
		require.Contains(t, server, "location /api/")
		require.Contains(t, server, "auth_request off;")
		require.Contains(t, server, "location /")
	}

	storefront := servers[0]
	require.Contains(t, storefront, "location ^~ /_ipx/")
	require.Contains(t, storefront, "location ^~ /uploads/site-logo/")
	require.Contains(t, storefront, "location ^~ /uploads/")
	require.Contains(t, storefront, "location = /api/v1/registration/warranty/claim")
	require.Contains(t, storefront, "location ~ ^/api/v1/(customer-service/attachments|suggestion-feedback/upload|showcase/upload)$")
}

func TestThemeWebNginxAuthCheckIsInternalAndAPIHealthAreExempt(t *testing.T) {
	root := findThemeWebRepoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "deployment", "nginx", "theme-web.conf"))
	require.NoError(t, err)

	config := string(content)
	servers := nginxServerBlocks(config)
	require.Len(t, servers, 2)

	for _, server := range servers {
		authLocation := nginxLocationBlock(server, "location = /_internal/security/ip-block-check")
		require.Contains(t, authLocation, "internal;")
		require.Contains(t, authLocation, "auth_request off;")

		healthLocation := nginxLocationBlock(server, "location = /healthz")
		require.Contains(t, healthLocation, "auth_request off;")

		apiLocation := nginxLocationBlock(server, "location /api/")
		require.Contains(t, apiLocation, "auth_request off;")
	}

	storefront := servers[0]
	for _, prefix := range []string{
		"location = /api/admin/media/assets",
		"location = /api/v1/registration/warranty/claim",
		"location ~ ^/api/v1/(customer-service/attachments|suggestion-feedback/upload|showcase/upload)$",
	} {
		require.Contains(t, nginxLocationBlock(storefront, prefix), "auth_request off;")
	}
	require.Contains(t, nginxLocationBlock(servers[1], "location = /api/admin/media/assets"), "auth_request off;")
}

func findThemeWebRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "deployment", "nginx", "theme-web.conf")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("theme-web repository root not found")
		}
		dir = parent
	}
}

func nginxServerBlocks(config string) []string {
	var blocks []string
	searchFrom := 0
	for {
		start := strings.Index(config[searchFrom:], "\n    server {")
		if start < 0 {
			break
		}
		start += searchFrom + len("\n    server {")
		depth := 1
		for i := start; i < len(config); i++ {
			switch config[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					blocks = append(blocks, config[start:i])
					searchFrom = i + 1
					goto nextServer
				}
			}
		}
		break
	nextServer:
	}
	return blocks
}

func nginxLocationBlock(server, locationPrefix string) string {
	start := strings.Index(server, locationPrefix)
	if start < 0 {
		return ""
	}
	open := strings.Index(server[start:], "{")
	if open < 0 {
		return ""
	}
	open += start
	depth := 1
	for i := open + 1; i < len(server); i++ {
		switch server[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return server[start : i+1]
			}
		}
	}
	return ""
}
