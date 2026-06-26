package service

import (
	"encoding/xml"
	"fmt"
	"strings"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/repository"
	"time"
)

// SitemapService Sitemap 生成服务
type SitemapService struct {
	postRepo *repository.PostRepository
	baseURL  string
}

// NewSitemapService 创建 Sitemap 服务
func NewSitemapService(postRepo *repository.PostRepository, baseURL string) *SitemapService {
	return &SitemapService{
		postRepo: postRepo,
		baseURL:  strings.TrimRight(baseURL, "/"),
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
	// 获取所有已发布的文章
	posts, err := s.postRepo.FindPublished()
	if err != nil {
		return "", fmt.Errorf("failed to fetch published posts: %w", err)
	}

	// 按翻译组分组
	groups := s.groupByTranslation(posts)

	// 生成 URL 列表
	urls := make([]URL, 0)
	for _, group := range groups {
		// 为每个语言版本创建一个 URL 条目
		for _, p := range group {
			url := s.createURL(p, group)
			urls = append(urls, url)
		}
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

// buildPostURL 构建文章 URL
func (s *SitemapService) buildPostURL(p post.Post) string {
	// 根据语言构建 URL 路径
	// 英文: /blog/post-slug
	// 其他语言: /fr/blog/post-slug
	if p.Locale == "en" {
		return fmt.Sprintf("%s/blog/%s", s.baseURL, p.Slug)
	}
	return fmt.Sprintf("%s/%s/blog/%s", s.baseURL, p.Locale, p.Slug)
}

// GenerateSimpleSitemap 生成简单的 Sitemap（不包含 Hreflang）
func (s *SitemapService) GenerateSimpleSitemap(locale string) (string, error) {
	// 获取指定语言的已发布文章
	posts, err := s.postRepo.FindPublishedByLocale(locale)
	if err != nil {
		return "", fmt.Errorf("failed to fetch published posts: %w", err)
	}

	// 生成 URL 列表
	urls := make([]URL, 0)
	for _, p := range posts {
		url := URL{
			Loc:        s.buildPostURL(p),
			LastMod:    p.UpdatedAt.Format(time.RFC3339),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		}
		urls = append(urls, url)
	}

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
