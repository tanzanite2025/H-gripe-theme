package edge

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Run explicitly with the release web image's Nginx base. Both containers use
// synthetic content and a new isolated network, without any business mounts.
func TestProductionNginxPreservesPublicUploadCachingAndPrivateDenialsWithoutDuplicateCacheControlHeaders(t *testing.T) {
	nginxImage := os.Getenv("PRODUCTION_NGINX_UPLOAD_CACHE_BOUNDARY_TEST_IMAGE")
	if nginxImage == "" {
		t.Skip("set PRODUCTION_NGINX_UPLOAD_CACHE_BOUNDARY_TEST_IMAGE to run the isolated real Nginx HTTP boundary test")
	}
	root := findThemeWebRepoRoot(t)
	fixtureDirectory := t.TempDir()
	fixtureConfiguration := "worker_processes 1;\nevents {}\nhttp {\naccess_log off;\nserver {\nlisten 9000;\nlocation = /_internal/security/ip-block-check { return 204; }\n"
	for _, prefix := range []string{"/uploads/", "/uploads/site-logo/"} {
		for _, fixture := range []struct {
			name         string
			status       int
			cacheControl string
		}{
			{"public-original", 200, "public, max-age=31536000, immutable"},
			{"public-derivative", 200, "public, max-age=31536000, immutable"},
			{"public-revalidated", 304, "public, max-age=31536000, immutable"},
			{"protected-original", 404, "private, no-store"},
			{"protected-derivative", 404, "private, no-store"},
		} {
			fixtureConfiguration += fmt.Sprintf("location = %s%s { add_header Cache-Control %q always; return %d; }\n", prefix, fixture.name, fixture.cacheControl, fixture.status)
		}
	}
	fixtureConfiguration += "}\nserver { listen 3000; return 200; }\nserver { listen 8080; return 200; }\n}\n"
	for name, content := range map[string]string{
		"synthetic-upstream.conf":      fixtureConfiguration,
		"storefront-server-names.conf": "server_name upload-cache-boundary.example.test;\n",
		"admin-server-names.conf":      "server_name admin-cache-boundary.example.test;\n",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(fixtureDirectory, name), []byte(content), 0600))
	}

	resourcePrefix := fmt.Sprintf("historical-upload-cache-boundary-%d", time.Now().UnixNano())
	networkName := resourcePrefix + "-network"
	upstreamName := resourcePrefix + "-synthetic-api"
	proxyName := resourcePrefix + "-production-nginx"
	runProductionNginxUploadCacheBoundaryDockerCommand(t, "network", "create", networkName)
	t.Cleanup(func() {
		runProductionNginxUploadCacheBoundaryDockerCommand(t, "network", "rm", networkName)
	})
	runProductionNginxUploadCacheBoundaryDockerCommand(t,
		"run", "-d", "--name", upstreamName, "--network", networkName,
		"--network-alias", "api-ingress", "--network-alias", "storefront", "--network-alias", "admin",
		"--mount", "type=bind,source="+filepath.ToSlash(filepath.Join(fixtureDirectory, "synthetic-upstream.conf"))+",target=/etc/nginx/nginx.conf,readonly",
		"--entrypoint", "nginx", nginxImage, "-g", "daemon off;")
	t.Cleanup(func() {
		runProductionNginxUploadCacheBoundaryDockerCommand(t, "rm", "-f", upstreamName)
	})

	proxyArguments := []string{"run", "-d", "--name", proxyName, "--network", networkName, "-p", "127.0.0.1::8080"}
	for _, mount := range []struct{ source, target string }{
		{filepath.Join(root, "deployment", "nginx", "theme-web.conf"), "/etc/nginx/nginx.conf"},
		{filepath.Join(root, "deployment", "nginx", "trusted-proxies.conf"), "/etc/nginx/trusted-proxies.conf"},
		{filepath.Join(root, "deployment", "nginx", "snippets", "security-headers.conf"), "/etc/nginx/snippets/security-headers.conf"},
		{filepath.Join(fixtureDirectory, "storefront-server-names.conf"), "/etc/nginx/generated-edge/storefront-server-names.conf"},
		{filepath.Join(fixtureDirectory, "admin-server-names.conf"), "/etc/nginx/generated-edge/admin-server-names.conf"},
	} {
		proxyArguments = append(proxyArguments, "--mount", "type=bind,source="+filepath.ToSlash(mount.source)+",target="+mount.target+",readonly")
	}
	proxyArguments = append(proxyArguments, "--entrypoint", "nginx", nginxImage, "-g", "daemon off;")
	runProductionNginxUploadCacheBoundaryDockerCommand(t, proxyArguments...)
	t.Cleanup(func() {
		runProductionNginxUploadCacheBoundaryDockerCommand(t, "rm", "-f", proxyName)
	})
	address := strings.TrimSpace(runProductionNginxUploadCacheBoundaryDockerCommand(t, "port", proxyName, "8080/tcp"))
	baseURL := "http://" + address
	client := &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	deadline := time.Now().Add(20 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		response, err := client.Get(baseURL + "/healthz")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		t.Log(runProductionNginxUploadCacheBoundaryDockerCommand(t, "logs", proxyName))
	}
	require.True(t, ready, "isolated production Nginx did not become healthy")
	for _, prefix := range []string{"/uploads/", "/uploads/site-logo/"} {
		for _, fixture := range []struct {
			name         string
			status       int
			cacheControl string
		}{
			{"public-original", 200, "public, max-age=31536000, immutable"},
			{"public-derivative", 200, "public, max-age=31536000, immutable"},
			{"public-revalidated", 304, "public, max-age=31536000, immutable"},
			{"protected-original", 404, "private, no-store"},
			{"protected-derivative", 404, "private, no-store"},
		} {
			requestVariants := []string{"ordinary"}
			if fixture.status == 404 {
				requestVariants = append(requestVariants, "range", "conditional")
			}
			for _, variant := range requestVariants {
				t.Run(prefix+fixture.name+"/"+variant, func(t *testing.T) {
					request, err := http.NewRequest(http.MethodGet, baseURL+prefix+fixture.name, nil)
					require.NoError(t, err)
					request.Host = "upload-cache-boundary.example.test"
					if variant == "range" {
						request.Header.Set("Range", "bytes=0-9")
					}
					if variant == "conditional" {
						request.Header.Set("If-None-Match", "*")
					}
					response, err := client.Do(request)
					require.NoError(t, err)
					defer response.Body.Close()
					_, err = io.Copy(io.Discard, response.Body)
					require.NoError(t, err)
					require.Equal(t, fixture.status, response.StatusCode)
					require.Equal(t, []string{fixture.cacheControl}, response.Header.Values("Cache-Control"), "Nginx must preserve the API cache policy without appending a contradictory public policy")
					require.Empty(t, response.Header.Get("Location"))
					require.Equal(t, "nosniff", response.Header.Get("X-Content-Type-Options"))
				})
			}
		}
	}
}

func runProductionNginxUploadCacheBoundaryDockerCommand(t *testing.T, arguments ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "docker", arguments...).CombinedOutput()
	require.NoError(t, err, "isolated Nginx Docker command failed: %s", output)
	return string(output)
}
