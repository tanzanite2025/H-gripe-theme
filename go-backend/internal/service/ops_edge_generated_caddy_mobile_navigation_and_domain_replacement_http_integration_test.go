package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"
)

// This test runs the actual generated configuration, without external DNS,
// ACME, production ports, or changes to the machine's certificate trust store.
func TestGeneratedCaddyPreservesMobileNavigationAndPostBodiesAcrossDomainReplacement(t *testing.T) {
	caddyExecutable, err := exec.LookPath("caddy")
	if err != nil {
		if os.Getenv("CADDY_EDGE_INTEGRATION_TEST_REQUIRED") == "true" {
			t.Fatal("Caddy is required for edge integration tests in this verification job")
		}
		t.Skip("install Caddy to run generated gateway HTTP/TLS integration coverage")
	}
	for _, domainSuffix := range []string{"original.example.test", "replacement.example.test"} {
		t.Run(domainSuffix, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				fmt.Fprintf(w, "%s %s %s %s", r.Method, r.RequestURI, r.Header.Get("X-Forwarded-Proto"), body)
			}))
			t.Cleanup(upstream.Close)
			repos := newOpsEdgeConfigTestRepositories(t)
			project := createOpsEdgeConfigProject(t, repos, "mobile-gateway-integration", true, ops.ProjectStatusActive)
			createOpsEdgeConfigDomain(t, repos, "shop."+domainSuffix, ops.DomainRoleCanonical, &project.ID, "")
			createOpsEdgeConfigDomain(t, repos, "www."+domainSuffix, ops.DomainRoleAlias, &project.ID, "")
			createOpsEdgeConfigDomain(t, repos, "admin."+domainSuffix, ops.DomainRoleAdmin, &project.ID, "")
			createOpsEdgeConfigDomain(t, repos, "old."+domainSuffix, ops.DomainRoleRedirect, &project.ID, "https://shop."+domainSuffix)
			createOpsEdgeConfigDomain(t, repos, "campaign."+domainSuffix, ops.DomainRoleRedirect, &project.ID, "https://shop."+domainSuffix+"/landing?source=old#intro")
			generated, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
			if err != nil {
				t.Fatal(err)
			}

			httpPort := reserveAvailableGatewayTestPort(t)
			httpsPort := reserveAvailableGatewayTestPort(t)
			storageDirectory := filepath.ToSlash(t.TempDir())
			configuration := fmt.Sprintf("{\n admin off\n persist_config off\n http_port %d\n https_port %d\n skip_install_trust\n storage file_system {\n root %q\n }\n}\n", httpPort, httpsPort, storageDirectory)
			configuration += strings.ReplaceAll(generated.Caddy, "theme-web:8080", upstream.URL)
			for _, route := range generated.Domains {
				configuration = strings.ReplaceAll(configuration, "\n"+route.Domain+" {\n", "\n"+route.Domain+" {\n tls internal\n")
			}
			configurationPath := filepath.Join(t.TempDir(), "generated-mobile-browser-gateway-integration.Caddyfile")
			if err := os.WriteFile(configurationPath, []byte(configuration), 0o600); err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(t.TempDir(), "generated-mobile-browser-gateway-integration.log")
			logFile, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(context.Background(), caddyExecutable, "run", "--config", configurationPath, "--adapter", "caddyfile")
			command.Stdout, command.Stderr = logFile, logFile
			if err := command.Start(); err != nil {
				_ = logFile.Close()
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = command.Process.Kill()
				_ = command.Wait()
				_ = logFile.Close()
				if t.Failed() {
					contents, _ := os.ReadFile(logPath)
					t.Log(string(contents))
				}
			})
			transport := &http.Transport{
				// Only this isolated test gateway uses its private, temporary CA.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
					port := httpPort
					if strings.HasSuffix(address, ":443") {
						port = httpsPort
					}
					return (&net.Dialer{}).DialContext(ctx, network, fmt.Sprintf("127.0.0.1:%d", port))
				},
			}
			t.Cleanup(transport.CloseIdleConnections)
			client := &http.Client{Transport: transport, Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
			ready := false
			for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
				response, err := client.Do(newCaddyEdgeIntegrationRequest(t, http.MethodGet, "https://shop."+domainSuffix+"/", nil))
				if err == nil {
					_ = response.Body.Close()
					if response.StatusCode == http.StatusOK {
						ready = true
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
			if !ready {
				t.Fatal("generated gateway did not become ready on isolated HTTP/TLS ports")
			}

			requestURI := "/zh_cn/tools/a%2Fb?size=700&return=https%3A%2F%2Fexample.com%2Fa"
			clickingUserAgents := []string{
				"Mozilla/5.0 (Linux; Android 13; Pixel 7; wv) AppleWebKit/537.36 Version/4.0 Chrome/120.0.0.0 Mobile Safari/537.36 MicroMessenger/8.0.40",
				"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 Chrome/120.0.0.0 Mobile Safari/537.36 [FB_IAB/FB4A;FBAV/492.0;]",
				"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 Chrome/120.0.0.0 Mobile Safari/537.36 WhatsApp/2.24.20.10",
				"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 Chrome/120.0.0.0 Mobile Safari/537.36 Instagram 340.0",
				"Mozilla/5.0 (iPhone; CPU iPhone OS 17_6 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 MicroMessenger/8.0.40",
				"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
				"TelegramBot (like TwitterBot)",
				"Twitterbot/1.0",
			}
			for _, route := range generated.Domains {
				for _, userAgent := range clickingUserAgents {
					request := newCaddyEdgeIntegrationRequest(t, http.MethodGet, "http://"+route.Domain+requestURI, nil)
					request.Header.Set("User-Agent", userAgent)
					response, err := client.Do(request)
					if err != nil {
						t.Fatal(err)
					}
					_ = response.Body.Close()
					if response.StatusCode != http.StatusMovedPermanently || response.Header.Get("Location") != "https://"+route.Domain+requestURI {
						t.Fatalf("mobile/social GET %s %s: status=%d location=%q", route.Domain, userAgent, response.StatusCode, response.Header.Get("Location"))
					}
				}
				for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions} {
					request := newCaddyEdgeIntegrationRequest(t, method, "http://"+route.Domain+requestURI, strings.NewReader("preserve=body"))
					request.Header.Set("X-Forwarded-Host", "untrusted.example")
					request.Header.Set("X-Forwarded-Proto", "https")
					response, err := client.Do(request)
					if err != nil {
						t.Fatal(err)
					}
					_ = response.Body.Close()
					wantStatus := http.StatusPermanentRedirect
					if method == http.MethodGet || method == http.MethodHead {
						wantStatus = http.StatusMovedPermanently
					}
					if response.StatusCode != wantStatus || response.Header.Get("Location") != "https://"+route.Domain+requestURI {
						t.Fatalf("%s %s: status=%d location=%q", method, route.Domain, response.StatusCode, response.Header.Get("Location"))
					}
				}
				if route.Role == ops.DomainRoleRedirect {
					wantLocation := "https://shop." + domainSuffix + requestURI
					if strings.HasPrefix(route.Domain, "campaign.") {
						wantLocation = route.RedirectTarget
					}
					for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions} {
						request := newCaddyEdgeIntegrationRequest(t, method, "https://"+route.Domain+requestURI, strings.NewReader("preserve=body"))
						response, err := client.Do(request)
						if err != nil {
							t.Fatal(err)
						}
						_ = response.Body.Close()
						wantStatus := http.StatusPermanentRedirect
						if method == http.MethodGet || method == http.MethodHead {
							wantStatus = http.StatusMovedPermanently
						}
						if response.StatusCode != wantStatus || response.Header.Get("Location") != wantLocation {
							t.Fatalf("HTTPS redirect %s %s: status=%d location=%q, want %d %q", method, route.Domain, response.StatusCode, response.Header.Get("Location"), wantStatus, wantLocation)
						}
					}
				} else {
					request := newCaddyEdgeIntegrationRequest(t, http.MethodGet, "https://"+route.Domain+requestURI, nil)
					response, err := client.Do(request)
					if err != nil {
						t.Fatal(err)
					}
					_ = response.Body.Close()
					if response.StatusCode != 200 || response.Header.Get("Location") != "" {
						t.Fatalf("HTTPS should reach upstream without redirect loop: %s status=%d", route.Domain, response.StatusCode)
					}
				}
			}
			followingClient := &http.Client{Transport: transport, Timeout: 5 * time.Second}
			response, err := followingClient.Do(newCaddyEdgeIntegrationRequest(t, http.MethodGet, "http://old."+domainSuffix+requestURI, nil))
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != 200 || string(body) != "GET "+requestURI+" https " {
				t.Fatalf("HTTP old-domain navigation did not pass two 301s to the new host: %d %s", response.StatusCode, body)
			}
			response, err = followingClient.Do(newCaddyEdgeIntegrationRequest(t, http.MethodGet, "http://campaign."+domainSuffix+requestURI, nil))
			if err != nil {
				t.Fatal(err)
			}
			body, _ = io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != 200 || string(body) != "GET /landing?source=old https " {
				t.Fatalf("HTTP explicit-page redirect did not preserve configured destination: %d %s", response.StatusCode, body)
			}
			response, err = followingClient.Do(newCaddyEdgeIntegrationRequest(t, http.MethodPost, "http://old."+domainSuffix+requestURI, strings.NewReader("preserve=body")))
			if err != nil {
				t.Fatal(err)
			}
			body, _ = io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != 200 || string(body) != "POST "+requestURI+" https preserve=body" {
				t.Fatalf("HTTP -> HTTPS -> replacement host lost method/body/URI across 308 redirects: %d %s", response.StatusCode, body)
			}
			response, err = client.Do(newCaddyEdgeIntegrationRequest(t, http.MethodGet, "http://unconfigured.example.test"+requestURI, nil))
			if err != nil {
				t.Fatal(err)
			}
			_ = response.Body.Close()
			// Caddy may add its own catch-all HTTPS upgrade. It must not map an
			// unconfigured host onto a registered storefront or admin domain.
			if location := response.Header.Get("Location"); location != "" && location != "https://unconfigured.example.test"+requestURI {
				t.Fatalf("unknown host entered a configured domain route: %q", location)
			}
		})
	}
}

func reserveAvailableGatewayTestPort(t *testing.T) int {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func newCaddyEdgeIntegrationRequest(t *testing.T, method string, target string, body io.Reader) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(context.Background(), method, target, body)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
