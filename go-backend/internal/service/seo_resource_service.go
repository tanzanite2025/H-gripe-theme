package service

import (
	"errors"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	seodomain "commerce-platform/internal/domain/seo"
)

var ErrInvalidSEOCanonicalURL = errors.New("invalid SEO canonical URL")

type SEOResourceService struct {
	posts            *PostService
	products         *ProductService
	settings         *SettingService
	canonicalBaseURL string
}

func NewSEOResourceService(posts *PostService, products *ProductService, settings *SettingService) *SEOResourceService {
	return &SEOResourceService{
		posts:    posts,
		products: products,
		settings: settings,
	}
}

func (s *SEOResourceService) ConfigureCanonicalBaseURL(baseURL string) {
	if s == nil {
		return
	}
	s.canonicalBaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

func (s *SEOResourceService) ListArticles(page, pageSize int, status, locale, search string) ([]post.Post, int64, error) {
	if s == nil || s.posts == nil {
		return nil, 0, ErrPostNotFound
	}
	return s.posts.ListAdmin(page, pageSize, status, locale, search, "")
}

func (s *SEOResourceService) UpdateArticle(id uint, request seodomain.ArticleResourceUpdateRequest) (*post.Post, error) {
	if s == nil || s.posts == nil {
		return nil, ErrPostNotFound
	}
	if request.CanonicalURL != nil {
		if err := s.validateCanonicalURL(*request.CanonicalURL); err != nil {
			return nil, err
		}
	}
	return s.posts.UpdatePostSEO(id, PostSEOUpdateInput{
		MetaTitle:       request.MetaTitle,
		MetaDescription: request.MetaDescription,
		CanonicalURL:    request.CanonicalURL,
	})
}

func (s *SEOResourceService) validateCanonicalURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidSEOCanonicalURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidSEOCanonicalURL
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return ErrInvalidSEOCanonicalURL
	}

	if s.canonicalBaseURL == "" {
		return nil
	}
	base, err := url.Parse(s.canonicalBaseURL)
	if err != nil || base.Host == "" {
		return nil
	}
	if parsed.Hostname() != base.Hostname() {
		return ErrInvalidSEOCanonicalURL
	}
	if base.Scheme != "" && parsed.Scheme != base.Scheme {
		return ErrInvalidSEOCanonicalURL
	}
	return nil
}

func (s *SEOResourceService) GetArticle(id uint) (*post.Post, error) {
	if s == nil || s.posts == nil {
		return nil, ErrPostNotFound
	}
	return s.posts.GetAdminPost(id)
}

func (s *SEOResourceService) ListProducts(page, pageSize int, status, locale, search string) ([]product.Product, int64, error) {
	if s == nil || s.products == nil {
		return nil, 0, ErrProductNotFound
	}
	return s.products.ListAdmin(page, pageSize, status, locale, search, "")
}

func (s *SEOResourceService) GetProduct(id uint) (*product.Product, error) {
	if s == nil || s.products == nil {
		return nil, ErrProductNotFound
	}
	return s.products.GetAdminProduct(id)
}

func (s *SEOResourceService) ProductDiagnostics(item product.Product) (seodomain.ProductSEOReadiness, error) {
	if s == nil {
		return seodomain.ProductSEOReadiness{}, ErrProductNotFound
	}

	brand := ""
	if item.Brand != nil {
		brand = item.Brand.Name
	}
	if s.settings != nil {
		siteSettings, err := s.settings.GetSiteSettings(item.Locale)
		if err != nil {
			return seodomain.ProductSEOReadiness{}, err
		}
		if siteSettings != nil && brand == "" {
			brand = siteSettings.BrandTitle
		}
	}

	routePath := seodomain.BuildProductRoute(item.Locale, item.Slug).Path
	return seodomain.BuildProductSEOReadiness(item, brand, routePath), nil
}

func (s *SEOResourceService) UpdateProduct(id uint, request seodomain.ProductResourceUpdateRequest) (*product.Product, error) {
	if s == nil || s.products == nil {
		return nil, ErrProductNotFound
	}
	return s.products.UpdateProductSEO(id, ProductSEOUpdateInput{
		MetaTitle:       request.MetaTitle,
		MetaDescription: request.MetaDescription,
	})
}
