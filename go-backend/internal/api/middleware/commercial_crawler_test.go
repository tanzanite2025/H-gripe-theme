package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCommercialCrawlerBlockerRejectsKnownCommercialCrawlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, testCase := range []struct {
		name      string
		userAgent string
	}{
		{name: "Ahrefs", userAgent: "Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)"},
		{name: "Semrush", userAgent: "SemrushBot/7~bl"},
		{name: "Similarweb", userAgent: "SimilarwebBot/1.0"},
		{name: "BuiltWith", userAgent: "BuiltWith/1.0"},
		{name: "case insensitive", userAgent: "mozilla/5.0 (compatible; ahrefsbot/7.0)"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CommercialCrawlerBlocker())
			router.GET("/products", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/products", nil)
			request.Header.Set("User-Agent", testCase.userAgent)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != commercialCrawlerBlockStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, commercialCrawlerBlockStatus)
			}
			if got := recorder.Header().Get("X-Robots-Tag"); got == "" {
				t.Fatal("expected X-Robots-Tag on blocked response")
			}
		})
	}
}

func TestCommercialCrawlerBlockerAllowsRegularVisitors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialCrawlerBlocker())
	router.GET("/products", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, userAgent := range []string{
		"",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0",
		"Googlebot/2.1 (+http://www.google.com/bot.html)",
	} {
		request := httptest.NewRequest(http.MethodGet, "/products", nil)
		if userAgent != "" {
			request.Header.Set("User-Agent", userAgent)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("user-agent %q status = %d, want %d", userAgent, recorder.Code, http.StatusOK)
		}
	}
}

func TestCommercialCrawlerBlockerPersistsShortLivedGlobalIPRule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))

	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	auditRecorder := &recordingCommercialCrawlerAuditRecorder{}
	blockService.ConfigureAuditRecorderFactory(func(_ *gorm.DB) service.IPBlockAuditRecorder {
		return auditRecorder
	})

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(GlobalIPBlocker(blockService))
	router.Use(CommercialCrawlerBlocker(blockService))
	router.GET("/products", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	firstRequest := httptest.NewRequest(http.MethodGet, "/products", nil)
	firstRequest.RemoteAddr = "203.0.113.41:1234"
	firstRequest.Header.Set("User-Agent", "AhrefsBot/7.0")
	firstResponse := httptest.NewRecorder()
	router.ServeHTTP(firstResponse, firstRequest)

	require.Equal(t, http.StatusForbidden, firstResponse.Code)
	require.Equal(t, "commercial-crawler", firstResponse.Header().Get("X-Access-Block"))
	require.Len(t, auditRecorder.logs, 1)

	var rule security.IPBlockRule
	require.NoError(t, db.Where("source = ? AND cidr = ?", security.IPBlockRuleSourceCommercialBot, "203.0.113.41/32").First(&rule).Error)
	require.Equal(t, commercialCrawlerAutoBlockSourceReference("Ahrefs"), rule.SourceReference)
	require.True(t, rule.Enabled)
	require.NotNil(t, rule.ExpiresAt)
	require.True(t, rule.ExpiresAt.After(time.Now().UTC()))

	secondRequest := httptest.NewRequest(http.MethodGet, "/products", nil)
	secondRequest.RemoteAddr = "203.0.113.41:1234"
	secondRequest.Header.Set("User-Agent", "Mozilla/5.0")
	secondResponse := httptest.NewRecorder()
	router.ServeHTTP(secondResponse, secondRequest)

	require.Equal(t, http.StatusForbidden, secondResponse.Code)
	require.Contains(t, secondResponse.Body.String(), "ip_blocked")
	require.Len(t, auditRecorder.logs, 1)

	blocked, match, err := blockService.IsBlocked(context.Background(), "203.0.113.41", time.Now().UTC())
	require.NoError(t, err)
	require.True(t, blocked)
	require.NotNil(t, match)
	require.Equal(t, rule.ID, match.ID)
}

func TestCommercialCrawlerBlockerDoesNotPersistPrivateClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))

	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	auditRecorder := &recordingCommercialCrawlerAuditRecorder{}
	blockService.ConfigureAuditRecorderFactory(func(_ *gorm.DB) service.IPBlockAuditRecorder {
		return auditRecorder
	})

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(CommercialCrawlerBlocker(blockService))
	router.GET("/products", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.RemoteAddr = "10.0.0.5:1234"
	request.Header.Set("User-Agent", "SemrushBot/7.0")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.Empty(t, auditRecorder.logs)
	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestCommercialCrawlerProtectionSnapshotIncludesBuiltInIntelligenceSeeds(t *testing.T) {
	snapshot := CommercialCrawlerProtectionSnapshot()

	seeds, ok := snapshot["intelligence_seeds"].([]CommercialIntelligenceSeed)
	if !ok {
		t.Fatalf("intelligence_seeds type = %T, want []CommercialIntelligenceSeed", snapshot["intelligence_seeds"])
	}
	if len(seeds) != 4 {
		t.Fatalf("seed count = %d, want 4", len(seeds))
	}

	expectedEnforcement := map[string]string{
		"known-commercial-crawlers":   "enforced",
		"browser-commerce-extensions": "enforced",
		"inventory-crawlers":          "enforced",
		"order-scrapers":              "enforced",
	}

	for _, seed := range seeds {
		expected, exists := expectedEnforcement[seed.ID]
		if !exists {
			t.Fatalf("unexpected intelligence seed %q", seed.ID)
		}
		if seed.Enforcement != expected {
			t.Errorf("seed %q enforcement = %q, want %q", seed.ID, seed.Enforcement, expected)
		}
		if len(seed.DetectionSignals) == 0 {
			t.Errorf("seed %q has no detection signals", seed.ID)
		}
	}
}

func TestCommercialCrawlerRobotsTxtAndNginxStayInSync(t *testing.T) {
	root := findRepoRoot(t)

	robotsPath := filepath.Join(root, "nuxt-i18n", "public", "robots.txt")
	nginxPath := filepath.Join(root, "deployment", "nginx", "theme-web.conf")

	robotsContent, err := os.ReadFile(robotsPath)
	if err != nil {
		t.Fatalf("read robots.txt: %v", err)
	}
	nginxContent, err := os.ReadFile(nginxPath)
	if err != nil {
		t.Fatalf("read nginx conf: %v", err)
	}

	robotsText := string(robotsContent)
	nginxText := strings.ToLower(string(nginxContent))
	rules := CommercialCrawlerRules()

	for _, rule := range rules {
		if !strings.Contains(robotsText, "User-agent: "+rule.UserAgent) {
			t.Fatalf("robots.txt missing user-agent %q", rule.UserAgent)
		}
		if !strings.Contains(nginxText, strings.ToLower(rule.UserAgent)) {
			t.Fatalf("nginx conf missing crawler token %q", rule.UserAgent)
		}
	}

	var commercialGroup strings.Builder
	for _, rule := range rules {
		commercialGroup.WriteString("User-agent: ")
		commercialGroup.WriteString(rule.UserAgent)
		commercialGroup.WriteByte('\n')
	}
	commercialGroup.WriteString("Disallow: /")
	if !strings.Contains(robotsText, commercialGroup.String()) {
		t.Fatal("robots.txt must disallow the complete commercial crawler group")
	}
	if !strings.Contains(robotsText, "User-Agent: *\nDisallow:") {
		t.Fatal("robots.txt must retain an allow-all wildcard group")
	}
}

func TestCommercialCrawlerProtectionSnapshotIncludesRobotsPolicy(t *testing.T) {
	snapshot := CommercialCrawlerProtectionSnapshot()
	policy, ok := snapshot["robots_txt"].(gin.H)
	if !ok {
		t.Fatalf("robots_txt type = %T, want gin.H", snapshot["robots_txt"])
	}
	if policy["path"] != "/robots.txt" {
		t.Fatalf("robots txt path = %v, want /robots.txt", policy["path"])
	}
	if policy["source"] != "nuxt-i18n/public/robots.txt" {
		t.Fatalf("robots txt source = %v, want nuxt-i18n/public/robots.txt", policy["source"])
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "nuxt-i18n", "public", "robots.txt")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

type recordingCommercialCrawlerAuditRecorder struct {
	logs []audit.AuditLog
}

func (r *recordingCommercialCrawlerAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}
