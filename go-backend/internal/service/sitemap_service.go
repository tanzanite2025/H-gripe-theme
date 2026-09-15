package service

import (
	"commerce-platform/internal/domain/post"
	productdomain "commerce-platform/internal/domain/product"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

// SitemapService Sitemap 生成服务
type SitemapService struct {
	postRepo            *repository.PostRepository
	productRepo         *repository.ProductRepository
	productCategoryRepo *repository.ProductCategoryRepository
	baseURL             string
}

// NewSitemapService creates the backend sitemap service.  The optional
// catalog repositories keep the constructor backwards compatible for callers
// that only need the legacy blog sitemap while allowing the application to
// provide product/category data for the public fallback routes.
func NewSitemapService(postRepo *repository.PostRepository, baseURL string, catalogRepositories ...interface{}) *SitemapService {
	service := &SitemapService{
		postRepo: postRepo,
		baseURL:  strings.TrimRight(baseURL, "/"),
	}
	service.ConfigureCatalogRepositories(catalogRepositories...)
	return service
}

// NewSitemapServiceWithCatalog is the typed constructor used by application
// wiring.  NewSitemapService remains available for legacy integrations.
func NewSitemapServiceWithCatalog(
	postRepo *repository.PostRepository,
	productRepo *repository.ProductRepository,
	productCategoryRepo *repository.ProductCategoryRepository,
	baseURL string,
) *SitemapService {
	return NewSitemapService(postRepo, baseURL, productRepo, productCategoryRepo)
}

// ConfigureCatalogRepositories attaches the public catalog sources used by
// sitemap generation. Nil repositories are accepted so deployments can
// continue serving the blog-only sitemap during a staged migration.
func (s *SitemapService) ConfigureCatalogRepositories(catalogRepositories ...interface{}) {
	if s == nil {
		return
	}
	for _, candidate := range catalogRepositories {
		switch repo := candidate.(type) {
		case *repository.ProductRepository:
			s.productRepo = repo
		case *repository.ProductCategoryRepository:
			s.productCategoryRepo = repo
		}
	}
}

// URLSet Sitemap XML 根元素
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	XHTMLns string   `xml:"xmlns:xhtml,attr"`
	URLs    []URL    `xml:"url"`
}

// URL Sitemap URL 元素
type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
	Links      []Link `xml:"xhtml:link"`
}

// Link Hreflang 链接元素
type Link struct {
	Rel      string `xml:"rel,attr"`
	Hreflang string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
}

// GenerateHreflangSitemap 生成包含 Hreflang 标签的 Sitemap
func (s *SitemapService) GenerateHreflangSitemap() (string, error) {
	urls, err := s.hreflangURLs()
	if err != nil {
		return "", err
	}

	// 创建 URLSet
	urlSet := URLSet{
		XMLNS:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		XHTMLns: "http://www.w3.org/1999/xhtml",
		URLs:    urls,
	}

	// 生成 XML
	output, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap XML: %w", err)
	}

	return xml.Header + string(output), nil
}

func (s *SitemapService) hreflangURLs() ([]URL, error) {
	urls := make([]URL, 0)

	if s == nil {
		return nil, fmt.Errorf("sitemap service is unavailable")
	}
	if s.postRepo != nil {
		posts, err := s.postRepo.FindPublished()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch published posts: %w", err)
		}
		postGroups := s.groupByTranslation(posts)
		postGroupIDs := make([]uint, 0, len(postGroups))
		for groupID := range postGroups {
			postGroupIDs = append(postGroupIDs, groupID)
		}
		sort.Slice(postGroupIDs, func(i, j int) bool { return postGroupIDs[i] < postGroupIDs[j] })
		for _, groupID := range postGroupIDs {
			group := postGroups[groupID]
			sort.Slice(group, func(i, j int) bool {
				if group[i].Locale != group[j].Locale {
					return group[i].Locale < group[j].Locale
				}
				return group[i].ID < group[j].ID
			})
			for _, p := range group {
				urls = append(urls, s.createURL(p, group))
			}
		}
	}

	// Products and categories are part of the backend fallback sitemap as well
	// as the Nuxt dynamic source. Keeping this path live prevents a proxy that
	// serves Go's /sitemap*.xml routes from silently dropping the catalog.
	if s.productRepo != nil {
		products, err := s.productRepo.FindPublished()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch published products: %w", err)
		}
		productGroups := groupProductsByTranslation(products)
		groupIDs := make([]uint, 0, len(productGroups))
		for groupID := range productGroups {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(i, j int) bool { return groupIDs[i] < groupIDs[j] })
		for _, groupID := range groupIDs {
			group := productGroups[groupID]
			sort.Slice(group, func(i, j int) bool {
				if group[i].Locale != group[j].Locale {
					return group[i].Locale < group[j].Locale
				}
				return group[i].ID < group[j].ID
			})
			for _, item := range group {
				urls = append(urls, s.createProductURL(item, group))
			}
		}
	}

	if s.productCategoryRepo != nil {
		categories, err := s.productCategoryRepo.FindPublished()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch published product categories: %w", err)
		}
		byID := productCategoriesByID(categories)
		localesForSitemap := sitemapLocales()
		sort.Slice(categories, func(i, j int) bool { return categories[i].ID < categories[j].ID })
		for _, category := range categories {
			paths := make(map[string]string, len(localesForSitemap))
			for _, locale := range localesForSitemap {
				if path := productCategorySitemapPath(category.ID, locale, byID); path != "" {
					paths[locale] = path
				}
			}
			for _, locale := range localesForSitemap {
				path := paths[locale]
				if path == "" {
					continue
				}
				urls = append(urls, s.createCategoryURL(category, locale, paths))
			}
		}
	}

	sort.Slice(urls, func(i, j int) bool { return urls[i].Loc < urls[j].Loc })
	return urls, nil
}

// groupByTranslation 按翻译组分组文章
func (s *SitemapService) groupByTranslation(posts []post.Post) map[uint][]post.Post {
	groups := make(map[uint][]post.Post)

	for _, p := range posts {
		// 如果有翻译组ID，使用翻译组ID分组
		if p.TranslationGroupID != nil {
			groups[*p.TranslationGroupID] = append(groups[*p.TranslationGroupID], p)
		} else {
			// 如果没有翻译组ID，使用文章ID作为组ID（单独一组）
			groups[p.ID] = append(groups[p.ID], p)
		}
	}

	return groups
}

// createURL 创建 URL 条目
func (s *SitemapService) createURL(p post.Post, group []post.Post) URL {
	// 构建文章 URL
	loc := s.buildPostURL(p)

	// 格式化最后修改时间
	lastMod := p.UpdatedAt.Format(time.RFC3339)

	// 创建 Hreflang 链接
	links := make([]Link, 0)
	for _, translation := range group {
		link := Link{
			Rel:      "alternate",
			Hreflang: translation.Locale,
			Href:     s.buildPostURL(translation),
		}
		links = append(links, link)
	}

	// 添加 x-default 链接（通常指向英文版本）
	hasEnglish := false
	for _, translation := range group {
		if translation.Locale == "en" {
			links = append(links, Link{
				Rel:      "alternate",
				Hreflang: "x-default",
				Href:     s.buildPostURL(translation),
			})
			hasEnglish = true
			break
		}
	}

	// 如果没有英文版本，使用第一个版本作为默认
	if !hasEnglish && len(group) > 0 {
		links = append(links, Link{
			Rel:      "alternate",
			Hreflang: "x-default",
			Href:     s.buildPostURL(group[0]),
		})
	}

	return URL{
		Loc:        loc,
		LastMod:    lastMod,
		ChangeFreq: "weekly",
		Priority:   "0.8",
		Links:      links,
	}
}

func (s *SitemapService) createProductURL(item productdomain.Product, group []productdomain.Product) URL {
	links := make([]Link, 0, len(group)+1)
	for _, translation := range group {
		links = append(links, Link{
			Rel:      "alternate",
			Hreflang: locales.Normalize(translation.Locale),
			Href:     s.buildProductURL(translation),
		})
	}
	if len(group) > 0 {
		xDefault := group[0]
		for _, translation := range group {
			if locales.Normalize(translation.Locale) == "en" {
				xDefault = translation
				break
			}
		}
		links = append(links, Link{
			Rel:      "alternate",
			Hreflang: "x-default",
			Href:     s.buildProductURL(xDefault),
		})
	}

	return URL{
		Loc:        s.buildProductURL(item),
		LastMod:    item.UpdatedAt.Format(time.RFC3339),
		ChangeFreq: "weekly",
		Priority:   "0.8",
		Links:      links,
	}
}

func (s *SitemapService) createCategoryURL(category productdomain.ProductCategory, locale string, paths map[string]string) URL {
	links := make([]Link, 0, len(paths)+1)
	orderedLocales := make([]string, 0, len(paths))
	for candidateLocale := range paths {
		orderedLocales = append(orderedLocales, candidateLocale)
	}
	sort.Strings(orderedLocales)
	for _, candidateLocale := range orderedLocales {
		links = append(links, Link{
			Rel:      "alternate",
			Hreflang: candidateLocale,
			Href:     s.baseURL + paths[candidateLocale],
		})
	}
	if defaultPath, ok := paths["en"]; ok {
		links = append(links, Link{Rel: "alternate", Hreflang: "x-default", Href: s.baseURL + defaultPath})
	}
	return URL{
		Loc:        s.baseURL + paths[locale],
		LastMod:    category.UpdatedAt.Format(time.RFC3339),
		ChangeFreq: "monthly",
		Priority:   "0.7",
		Links:      links,
	}
}

// buildPostURL 构建文章 URL
func (s *SitemapService) buildPostURL(p post.Post) string {
	return fmt.Sprintf("%s%s", s.baseURL, seodomain.BuildArticleRoute(p.Locale, p.Slug, p.Tags).Path)
}

func (s *SitemapService) buildProductURL(item productdomain.Product) string {
	return fmt.Sprintf("%s%s", s.baseURL, seodomain.BuildProductRoute(item.Locale, item.Slug).Path)
}

// GenerateSimpleSitemap 生成简单的 Sitemap（不包含 Hreflang）
func (s *SitemapService) GenerateSimpleSitemap(locale string) (string, error) {
	normalizedLocale := locales.Normalize(locale)
	if s == nil {
		return "", fmt.Errorf("sitemap service is unavailable")
	}

	// 生成 URL 列表
	urls := make([]URL, 0)
	if s.postRepo != nil {
		posts, err := s.postRepo.FindPublishedByLocale(normalizedLocale)
		if err != nil {
			return "", fmt.Errorf("failed to fetch published posts: %w", err)
		}
		for _, p := range posts {
			url := URL{
				Loc:        s.buildPostURL(p),
				LastMod:    p.UpdatedAt.Format(time.RFC3339),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			}
			urls = append(urls, url)
		}
	}

	if s.productRepo != nil {
		products, err := s.productRepo.FindPublishedByLocale(normalizedLocale)
		if err != nil {
			return "", fmt.Errorf("failed to fetch published products: %w", err)
		}
		for _, item := range products {
			urls = append(urls, URL{
				Loc:        s.buildProductURL(item),
				LastMod:    item.UpdatedAt.Format(time.RFC3339),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			})
		}
	}

	if s.productCategoryRepo != nil {
		categories, err := s.productCategoryRepo.FindPublished()
		if err != nil {
			return "", fmt.Errorf("failed to fetch published product categories: %w", err)
		}
		byID := productCategoriesByID(categories)
		for _, category := range categories {
			path := productCategorySitemapPath(category.ID, normalizedLocale, byID)
			if path == "" {
				continue
			}
			urls = append(urls, URL{
				Loc:        s.baseURL + path,
				LastMod:    category.UpdatedAt.Format(time.RFC3339),
				ChangeFreq: "monthly",
				Priority:   "0.7",
			})
		}
	}

	sort.Slice(urls, func(i, j int) bool { return urls[i].Loc < urls[j].Loc })

	// 创建 URLSet
	urlSet := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	// 生成 XML
	output, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap XML: %w", err)
	}

	return xml.Header + string(output), nil
}

func groupProductsByTranslation(products []productdomain.Product) map[uint][]productdomain.Product {
	groups := make(map[uint][]productdomain.Product)
	for _, item := range products {
		rootID := item.ID
		if item.ParentID != nil && *item.ParentID > 0 {
			rootID = *item.ParentID
		}
		groups[rootID] = append(groups[rootID], item)
	}
	return groups
}

func productCategorySitemapPath(id uint, locale string, byID map[uint]productdomain.ProductCategory) string {
	segments := make([]string, 0, 5)
	visited := make(map[uint]struct{}, len(byID))
	for currentID := id; currentID > 0; {
		if _, seen := visited[currentID]; seen {
			return ""
		}
		visited[currentID] = struct{}{}
		category, ok := byID[currentID]
		if !ok || !category.IsEnabled || strings.TrimSpace(category.Slug) == "" {
			return ""
		}
		segments = append(segments, category.Slug)
		if category.ParentID == nil || *category.ParentID == 0 {
			break
		}
		currentID = *category.ParentID
	}
	for left, right := 0, len(segments)-1; left < right; left, right = left+1, right-1 {
		segments[left], segments[right] = segments[right], segments[left]
	}
	return seodomain.BuildCategoryRoute(locale, segments...).Path
}

func sitemapLocales() []string {
	result := locales.EnabledLocaleCodes()
	if len(result) == 0 {
		return []string{"en"}
	}
	return result
}

// SitemapIndex Sitemap 索引
type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	XMLNS    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// Sitemap Sitemap 索引条目
type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// GenerateSitemapIndex 生成 Sitemap 索引
func (s *SitemapService) GenerateSitemapIndex(locales []string) (string, error) {
	sitemaps := make([]Sitemap, 0)

	// 添加 Hreflang Sitemap
	sitemaps = append(sitemaps, Sitemap{
		Loc:     fmt.Sprintf("%s/sitemap-hreflang.xml", s.baseURL),
		LastMod: time.Now().Format(time.RFC3339),
	})

	// 为每个语言添加单独的 Sitemap
	for _, locale := range locales {
		sitemaps = append(sitemaps, Sitemap{
			Loc:     fmt.Sprintf("%s/sitemap-%s.xml", s.baseURL, locale),
			LastMod: time.Now().Format(time.RFC3339),
		})
	}

	// 创建 SitemapIndex
	index := SitemapIndex{
		XMLNS:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: sitemaps,
	}

	// 生成 XML
	output, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap index XML: %w", err)
	}

	return xml.Header + string(output), nil
}
