package v1

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/api/v1/i18n"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type wpCompatFeaturedImage struct {
	URL    string `json:"url"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
	Alt    string `json:"alt,omitempty"`
}

type wpCompatTranslation struct {
	ID   uint   `json:"id"`
	Slug string `json:"slug"`
}

type wpCompatPostSummary struct {
	ID            uint                           `json:"id"`
	Lang          string                         `json:"lang"`
	Group         string                         `json:"group"`
	Slug          string                         `json:"slug"`
	Title         string                         `json:"title"`
	Excerpt       string                         `json:"excerpt"`
	Date          string                         `json:"date"`
	FeaturedImage *wpCompatFeaturedImage         `json:"featuredImage"`
	Categories    []string                       `json:"categories"`
	Translations  map[string]wpCompatTranslation `json:"translations"`
}

type wpCompatPostDetail struct {
	wpCompatPostSummary
	ContentHTML  string `json:"contentHtml"`
	CanonicalURL string `json:"canonicalUrl"`
}

func registerWordPressCompatRoutes(
	r *gin.Engine,
	postService *service.PostService,
	settingService *service.SettingService,
	loyaltyRepo *repository.LoyaltyRepository,
	marketingService *service.MarketingService,
	authService *service.AuthService,
) {
	// Only expose compatibility routes for modules that have actually moved to Go.
	// Other WordPress plugin routes should continue to be served by WordPress until migrated.
	wp := r.Group("/wp-json/tanzanite/v1")
	{
		wp.GET("/posts", listCompatBlogPosts(postService))
		wp.GET("/post", getCompatBlogPost(postService))
		wp.GET("/translations", getCompatBlogTranslations(postService))

		// 语言列表（公开）
		wp.GET("/languages", getCompatLanguages())

		// 积分兑换配置（公开）
		wp.GET("/redeem/config", getCompatRedeemConfig(settingService))

		// 查询用户积分余额（公开，挂载 OptionalAuthMiddleware，未登录返回0）
		wp.GET("/loyalty/points", middleware.OptionalAuthMiddleware(authService), getCompatUserPoints(loyaltyRepo, settingService))

		// 积分兑换礼品卡（需要登录授权）—— 调用 MarketingService 复用核心事务逻辑
		wp.POST("/giftcards/redeem", middleware.AuthMiddleware(authService), redeemPointsToGiftCard(marketingService, settingService))
		wp.POST("/redeem/exchange", middleware.AuthMiddleware(authService), redeemPointsToGiftCard(marketingService, settingService))
	}
}

func listCompatBlogPosts(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := compatLocale(c)
		status := c.DefaultQuery("status", "published")
		category := strings.ToLower(strings.TrimSpace(c.Query("category")))
		page := parseCompatInt(c, "page", 1)
		perPage := parseCompatInt(c, "per_page", parseCompatInt(c, "page_size", 5))

		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 5
		}

		fetchPage := page
		fetchSize := perPage
		if category != "" {
			fetchPage = 1
			fetchSize = 500
		}

		posts, total, err := postService.List(locale, status, fetchPage, fetchSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if category != "" {
			posts = filterCompatPostsByCategory(posts, category)
			total = int64(len(posts))
			posts = sliceCompatPosts(posts, page, perPage)
		}

		items := make([]wpCompatPostSummary, 0, len(posts))
		for _, item := range posts {
			items = append(items, makeCompatPostSummary(item, category, nil))
		}

		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
			"items":    items,
		})
	}
}

func getCompatBlogPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := compatLocale(c)
		slug := strings.TrimSpace(c.Query("slug"))
		if slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
			return
		}

		item, err := postService.GetBySlug(slug, locale)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		translations := loadCompatTranslations(postService, item)
		c.JSON(http.StatusOK, makeCompatPostDetail(*item, c.Query("category"), translations))
	}
}

func getCompatBlogTranslations(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		group := strings.TrimSpace(c.Query("group"))
		if group == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing group"})
			return
		}

		groupID, err := parseCompatGroupID(group)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"group":        group,
				"translations": map[string]wpCompatTranslation{},
			})
			return
		}

		posts, err := postService.GetTranslationsByGroup(groupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"group":        group,
			"translations": makeCompatTranslationMap(posts),
		})
	}
}

func compatLocale(c *gin.Context) string {
	if lang := strings.TrimSpace(c.Query("lang")); lang != "" {
		return lang
	}
	return middleware.GetLocale(c)
}

func parseCompatInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func parseCompatGroupID(group string) (uint, error) {
	group = strings.TrimPrefix(group, "post-")
	group = strings.TrimPrefix(group, "grp-")
	value, err := strconv.ParseUint(group, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func sliceCompatPosts(posts []post.Post, page, perPage int) []post.Post {
	start := (page - 1) * perPage
	if start >= len(posts) {
		return []post.Post{}
	}
	end := start + perPage
	if end > len(posts) {
		end = len(posts)
	}
	return posts[start:end]
}

func filterCompatPostsByCategory(posts []post.Post, category string) []post.Post {
	filtered := make([]post.Post, 0, len(posts))
	for _, item := range posts {
		if hasCompatCategory(item, category) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func hasCompatCategory(item post.Post, category string) bool {
	if category == "" {
		return true
	}
	for _, itemCategory := range compatCategories(item, "") {
		if itemCategory == category {
			return true
		}
	}
	return false
}

func makeCompatPostSummary(
	item post.Post,
	fallbackCategory string,
	translations map[string]wpCompatTranslation,
) wpCompatPostSummary {
	if translations == nil {
		translations = map[string]wpCompatTranslation{
			item.Locale: {ID: item.ID, Slug: item.Slug},
		}
	}

	return wpCompatPostSummary{
		ID:            item.ID,
		Lang:          item.Locale,
		Group:         compatGroup(item),
		Slug:          item.Slug,
		Title:         item.Title,
		Excerpt:       item.Excerpt,
		Date:          compatPostDate(item),
		FeaturedImage: compatFeaturedImage(item),
		Categories:    compatCategories(item, fallbackCategory),
		Translations:  translations,
	}
}

func makeCompatPostDetail(
	item post.Post,
	fallbackCategory string,
	translations map[string]wpCompatTranslation,
) wpCompatPostDetail {
	return wpCompatPostDetail{
		wpCompatPostSummary: makeCompatPostSummary(item, fallbackCategory, translations),
		ContentHTML:         item.Content,
		CanonicalURL:        item.CanonicalURL,
	}
}

func compatGroup(item post.Post) string {
	if item.TranslationGroupID != nil {
		return fmt.Sprintf("%d", *item.TranslationGroupID)
	}
	return fmt.Sprintf("post-%d", item.ID)
}

func compatPostDate(item post.Post) string {
	if item.PublishedAt != nil {
		return item.PublishedAt.Format(time.RFC3339)
	}
	return item.CreatedAt.Format(time.RFC3339)
}

func compatFeaturedImage(item post.Post) *wpCompatFeaturedImage {
	if strings.TrimSpace(item.FeaturedImg) == "" {
		return nil
	}
	return &wpCompatFeaturedImage{URL: item.FeaturedImg}
}

func compatCategories(item post.Post, fallback string) []string {
	tags := strings.ToLower(item.Tags)
	categories := make([]string, 0, 2)
	for _, candidate := range []string{"news", "wheelsbuild"} {
		if fallback == candidate || strings.Contains(tags, candidate) {
			categories = append(categories, candidate)
		}
	}
	if len(categories) == 0 && fallback != "" {
		categories = append(categories, fallback)
	}
	return categories
}

func loadCompatTranslations(
	postService *service.PostService,
	item *post.Post,
) map[string]wpCompatTranslation {
	posts, err := postService.GetTranslations(item.ID)
	if err != nil || len(posts) == 0 {
		return map[string]wpCompatTranslation{
			item.Locale: {ID: item.ID, Slug: item.Slug},
		}
	}
	return makeCompatTranslationMap(posts)
}

func makeCompatTranslationMap(posts []post.Post) map[string]wpCompatTranslation {
	translations := make(map[string]wpCompatTranslation, len(posts))
	for _, item := range posts {
		translations[item.Locale] = wpCompatTranslation{
			ID:   item.ID,
			Slug: item.Slug,
		}
	}
	return translations
}

// getCompatRedeemConfig 获取积分兑换配置
func getCompatRedeemConfig(settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := middleware.GetLocale(c)
		config, err := settingService.GetRedeemSettings(locale)
		if err != nil {
			// Fail Loudly: 直接返回 500 并附带 Critical 报错日志
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem settings: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"enabled":            config.Enabled,
			"exchange_rate":      config.ExchangeRate,
			"min_points":         config.MinPoints,
			"max_value_per_day":  config.MaxValuePerDay,
			"card_expiry_days":   config.CardExpiryDays,
			"preset_values":      config.PresetValues,
		})
	}
}

// getCompatUserPoints 获取用户积分余额
func getCompatUserPoints(loyaltyRepo *repository.LoyaltyRepository, settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			// 未登录时返回默认值
			c.JSON(http.StatusOK, gin.H{
				"user_id":              0,
				"points":               0,
				"total":                0,
				"available":            0,
				"tier":                 "Ordinary",
				"can_redeem":           false,
				"max_redeemable_value": 0,
				"today_redeemed":       0,
			})
			return
		}
		userID := userIDVal.(uint)

		// 1. 获取兑换配置
		locale := middleware.GetLocale(c)
		config, err := settingService.GetRedeemSettings(locale)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem settings: %v", err)})
			return
		}

		// 2. 获取用户积分汇总
		userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
		if err != nil {
			// 若找不到但用户已登录，说明可能还未产生积分记录，默认初始化返回 0
			c.JSON(http.StatusOK, gin.H{
				"user_id":              userID,
				"points":               0,
				"total":                0,
				"available":            0,
				"tier":                 "Ordinary",
				"can_redeem":           false,
				"max_redeemable_value": 0,
				"today_redeemed":       0,
			})
			return
		}

		// 3. 计算今日已兑换金额
		var sumPoints int
		todayStart := time.Now().Truncate(24 * time.Hour)
		todayEnd := todayStart.Add(24 * time.Hour)
		
		err = loyaltyRepo.GetDB().Model(&loyalty.LoyaltyTransaction{}).
			Where("user_id = ? AND type = ? AND source = ? AND created_at BETWEEN ? AND ?", 
				userID, "spend", "giftcard", todayStart, todayEnd).
			Select("COALESCE(SUM(points), 0)").
			Scan(&sumPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to calculate daily redeemed points: %v", err)})
			return
		}

		// points 扣除是负数，所以取绝对值
		todayRedeemedValue := math.Abs(float64(sumPoints)) / float64(config.ExchangeRate)

		// 4. 计算最大可兑换金额
		maxRedeemableByPoints := math.Floor(float64(userLoyalty.AvailablePoints) / float64(config.ExchangeRate))
		maxRedeemableByLimit := math.Max(0, config.MaxValuePerDay-todayRedeemedValue)
		if config.MaxValuePerDay <= 0 {
			maxRedeemableByLimit = math.MaxFloat64
		}
		maxRedeemableValue := math.Min(maxRedeemableByPoints, maxRedeemableByLimit)

		// 获取级别名称
		tierName := "Ordinary"
		if userLoyalty.MemberLevelID > 0 {
			level, err := loyaltyRepo.FindMemberLevelByID(userLoyalty.MemberLevelID)
			if err == nil && level != nil {
				tierName = level.Name
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":              userID,
			"points":               userLoyalty.AvailablePoints,
			"total":                userLoyalty.TotalPoints,
			"available":            userLoyalty.AvailablePoints,
			"tier":                 tierName,
			"can_redeem":           config.Enabled && userLoyalty.AvailablePoints >= config.MinPoints && maxRedeemableValue > 0,
			"max_redeemable_value": maxRedeemableValue,
			"today_redeemed":       todayRedeemedValue,
		})
	}
}

// redeemPointsToGiftCard 积分兑换礼品卡（兼容路由 - 委托 MarketingService）
func redeemPointsToGiftCard(
	marketingService *service.MarketingService,
	settingService *service.SettingService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] Unauthorized access"})
			return
		}
		userID := userIDVal.(uint)

		// 双字段容错：接受 points 和 points_to_spend
		var req struct {
			Points        int     `json:"points"`
			PointsToSpend int     `json:"points_to_spend"`
			GiftCardValue float64 `json:"giftcard_value" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("[CRITICAL] Invalid request arguments: %v", err)})
			return
		}

		// 容错：优先 points, 其次 points_to_spend
		pointsToSpend := req.Points
		if pointsToSpend <= 0 {
			pointsToSpend = req.PointsToSpend
		}
		if pointsToSpend <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] 'points' or 'points_to_spend' must be > 0"})
			return
		}

		// 获取兑换配置
		locale := middleware.GetLocale(c)
		config, err := settingService.GetRedeemSettings(locale)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem settings: %v", err)})
			return
		}

		// 委托 MarketingService 核心事务方法
		result, err := marketingService.RedeemPointsForGiftCard(userID, pointsToSpend, req.GiftCardValue, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"giftcard_id":      result.GiftCardID,
			"card_code":        result.CardCode,
			"balance":          result.Balance,
			"points_spent":     result.PointsSpent,
			"points_remaining": result.PointsRemaining,
			"expires_at":       result.ExpiresAt,
			"message":          "兑换成功",
		})
	}
}

// getCompatLanguages 获取语言列表（兼容旧 WP API 格式）
func getCompatLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		type compatLangItem struct {
			Code string `json:"code"`
			Name string `json:"name"`
		}

		items := make([]compatLangItem, 0, len(i18n.SupportedLanguages))
		for _, lang := range i18n.SupportedLanguages {
			if lang.Enabled {
				items = append(items, compatLangItem{
					Code: lang.Code,
					Name: lang.NativeName, // 旧 WP 的 name 指向本地化名称，如 "简体中文"
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"locales": items,
		})
	}
}

