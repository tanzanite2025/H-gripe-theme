package middleware

import (
	"net/http"
	"strings"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const commercialCrawlerBlockStatus = http.StatusForbidden

type CommercialCrawlerRule struct {
	Provider  string `json:"provider"`
	UserAgent string `json:"user_agent"`
}

type CommercialIntelligenceSeed struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	Aliases          []string `json:"aliases"`
	Identification   string   `json:"identification"`
	DetectionSignals []string `json:"detection_signals"`
	Threshold        string   `json:"threshold,omitempty"`
	Enforcement      string   `json:"enforcement"`
	Action           string   `json:"action"`
}

var commercialCrawlerRules = []CommercialCrawlerRule{
	{Provider: "Ahrefs", UserAgent: "AhrefsBot"},
	{Provider: "Semrush", UserAgent: "SemrushBot"},
	{Provider: "Similarweb", UserAgent: "SimilarwebBot"},
	{Provider: "BuiltWith", UserAgent: "BuiltWith"},
}

// commercialIntelligenceSeeds is the starting threat-intelligence catalog.
var commercialIntelligenceSeeds = []CommercialIntelligenceSeed{
	{
		ID:             "known-commercial-crawlers",
		Name:           "BuiltWith / Similarweb / Ahrefs / Semrush",
		Category:       "known_crawler",
		Aliases:        []string{"BuiltWith", "SimilarwebBot", "AhrefsBot", "SemrushBot"},
		Identification: "命中已知 User-Agent 标识。",
		DetectionSignals: []string{
			"请求 User-Agent 包含 AhrefsBot、SemrushBot、SimilarwebBot 或 BuiltWith。",
		},
		Enforcement: "enforced",
		Action:      "block_403",
	},
	{
		ID:             "browser-commerce-extensions",
		Name:           "Koala / Commerce Inspector",
		Category:       "browser_extension",
		Aliases:        []string{"Koala Inspector", "Commerce Inspector", "Shine Commerce"},
		Identification: "浏览器插件通常复用普通浏览器请求，不能依赖固定 User-Agent 识别；其标准平台探测入口已在公网边缘返回 404。",
		DetectionSignals: []string{
			"短时间跨大量商品页或商品接口读取。",
			"同一会话连续抓取商品、价格、变体等公开信息。",
			"请求 Shopify 或 WooCommerce 的标准兼容路径。",
		},
		Threshold:   "标准 Shopify/WooCommerce 兼容路径直接返回 HTTP 404。",
		Enforcement: "enforced",
		Action:      "not_found_404",
	},
	{
		ID:             "inventory-crawlers",
		Name:           "Inventory Crawlers",
		Category:       "inventory_probe",
		Aliases:        []string{"库存监听爬虫", "Stock monitor"},
		Identification: "通过重复读取商品、变体或库存可用性推算出货量。",
		DetectionSignals: []string{
			"同一商品或变体在短时间内被重复读取。",
			"同一来源连续遍历大量 SKU 或变体。",
			"短窗口内出现高频目录或库存探测请求。",
		},
		Threshold:   "一分钟内超过 120 次商品读取，或读取超过 40 个不同商品目标。",
		Enforcement: "enforced",
		Action:      "rate_limit_429",
	},
	{
		ID:             "order-scrapers",
		Name:           "Order Scrapers",
		Category:       "order_enumeration",
		Aliases:        []string{"订单号顺序推算", "Order enumeration"},
		Identification: "通过连续或相邻订单号请求推算订单量与 GMV。",
		DetectionSignals: []string{
			"同一会话连续查询相邻或递增的订单标识。",
			"短时间内大量订单查询返回不存在或未授权。",
			"单一来源的订单查询频率异常升高。",
		},
		Threshold:   "一分钟内超过 20 次订单详情读取，或连续命中 5 个相邻订单号。",
		Enforcement: "enforced",
		Action:      "rate_limit_429",
	},
}

// CommercialCrawlerBlocker rejects known commercial intelligence crawlers and,
// when a block service is supplied, persists a short-lived global IP rule for
// the public client address. The variadic form keeps lightweight callers and
// unit tests compatible while production supplies the shared block service.
// The matching is deliberately case-insensitive because user-agent tokens vary
// between product versions while retaining the provider identifier.
func CommercialCrawlerBlocker(blockServices ...*service.GlobalIPBlockService) gin.HandlerFunc {
	var blockService *service.GlobalIPBlockService
	if len(blockServices) > 0 {
		blockService = blockServices[0]
	}

	return func(c *gin.Context) {
		rule := commercialCrawlerRuleForUserAgent(c.GetHeader("User-Agent"))
		if rule == nil {
			c.Next()
			return
		}

		if blockService != nil {
			persistCommercialCrawlerIPBlock(c, blockService, *rule)
		}
		c.Header("Cache-Control", "no-store, max-age=0")
		c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
		c.Header("X-Access-Block", "commercial-crawler")
		c.AbortWithStatus(commercialCrawlerBlockStatus)
	}
}

func CommercialCrawlerProtectionSnapshot() gin.H {
	rules := CommercialCrawlerRules()
	return gin.H{
		"enabled":            true,
		"response_status":    commercialCrawlerBlockStatus,
		"rules":              rules,
		"intelligence_seeds": CommercialIntelligenceSeeds(),
		"robots_txt":         CommercialCrawlerRobotsPolicy(),
		"automatic_ip_block": CommercialCrawlerAutoBlockPolicy(),
		"enforcement": []gin.H{
			{"layer": "Go API middleware", "status": "enabled"},
			{"layer": "Public Nginx edge", "status": "enabled"},
			{"layer": "Platform probe filter", "status": "enabled"},
			{"layer": "robots.txt", "status": "enabled"},
		},
	}
}

func CommercialCrawlerRobotsPolicy() gin.H {
	rules := CommercialCrawlerRules()
	userAgents := make([]string, 0, len(rules))
	for _, rule := range rules {
		userAgents = append(userAgents, rule.UserAgent)
	}

	return gin.H{
		"path":                "/robots.txt",
		"source":              "nuxt-i18n/public/robots.txt",
		"wildcard_user_agent": "*",
		"disallow":            "/",
		"blocked_user_agents": userAgents,
		"block_status":        commercialCrawlerBlockStatus,
	}
}

func CommercialCrawlerRules() []CommercialCrawlerRule {
	rules := make([]CommercialCrawlerRule, len(commercialCrawlerRules))
	copy(rules, commercialCrawlerRules)
	return rules
}

func CommercialIntelligenceSeeds() []CommercialIntelligenceSeed {
	seeds := make([]CommercialIntelligenceSeed, len(commercialIntelligenceSeeds))
	for index, seed := range commercialIntelligenceSeeds {
		seed.Aliases = append([]string(nil), seed.Aliases...)
		seed.DetectionSignals = append([]string(nil), seed.DetectionSignals...)
		seeds[index] = seed
	}
	return seeds
}

func commercialCrawlerRuleForUserAgent(userAgent string) *CommercialCrawlerRule {
	normalizedUserAgent := strings.ToLower(strings.TrimSpace(userAgent))
	if normalizedUserAgent == "" {
		return nil
	}

	for index := range commercialCrawlerRules {
		rule := &commercialCrawlerRules[index]
		if strings.Contains(normalizedUserAgent, strings.ToLower(rule.UserAgent)) {
			return rule
		}
	}

	return nil
}

// SecurityHeaders 设置安全相关的 HTTP 响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "frame-ancestors 'none'; object-src 'none'; base-uri 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(self)")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Next()
	}
}
