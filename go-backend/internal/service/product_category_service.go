package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"commerce-platform/internal/domain/product"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const ProductCategoryMaxDepth = 5
const SystemProductCategoryWheelsetSlug = "wheelset"

var productCategorySlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

var (
	ErrProductCategoryNotFound           = errors.New("product category not found")
	ErrProductCategoryInvalid            = errors.New("product category invalid")
	ErrProductCategorySlugExists         = errors.New("product category slug already exists")
	ErrProductCategoryHasChildren        = errors.New("product category has child categories")
	ErrProductCategoryImageInvalid       = errors.New("product category image invalid")
	ErrProductCategoryTranslationInvalid = errors.New("product category translation invalid")
	ErrProductCategorySystemProtected    = errors.New("system product category cannot be changed")
	ErrProductCategorySEOInvalid         = errors.New("product category SEO invalid")
)

type ProductCategoryInput struct {
	ParentID          *uint
	Name              string
	Slug              string
	Description       string
	ImageMediaAssetID *uint
	IsEnabled         bool
	SortOrder         int
}

type ProductCategoryView struct {
	ID                        uint                  `json:"id"`
	ParentID                  *uint                 `json:"parent_id"`
	Name                      string                `json:"name"`
	Slug                      string                `json:"slug"`
	Description               string                `json:"description"`
	MetaTitle                 string                `json:"meta_title"`
	MetaDescription           string                `json:"meta_description"`
	Intro                     string                `json:"intro"`
	RoutePath                 string                `json:"route_path"`
	ImageMediaAssetID         *uint                 `json:"image_media_asset_id,omitempty"`
	ImageURL                  string                `json:"image_url,omitempty"`
	Depth                     int                   `json:"depth"`
	SortOrder                 int                   `json:"sort_order"`
	IsEnabled                 bool                  `json:"is_enabled"`
	CreatedAt                 time.Time             `json:"created_at"`
	UpdatedAt                 time.Time             `json:"updated_at"`
	Children                  []ProductCategoryView `json:"children"`
	TranslationCompleted      int                   `json:"translation_completed,omitempty"`
	TranslationTotal          int                   `json:"translation_total,omitempty"`
	TranslationMissingLocales []string              `json:"translation_missing_locales,omitempty"`
}

type ProductCategoryTranslationInput struct {
	ID          uint
	Locale      string
	Name        string
	Description string
}

type ProductCategoryTranslationView struct {
	ID                uint      `json:"id"`
	ProductCategoryID uint      `json:"product_category_id"`
	Locale            string    `json:"locale"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	MetaTitle         string    `json:"meta_title"`
	MetaDescription   string    `json:"meta_description"`
	Intro             string    `json:"intro"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ProductCategoryListView struct {
	Tree     []ProductCategoryView `json:"tree"`
	Flat     []ProductCategoryView `json:"flat"`
	MaxDepth int                   `json:"max_depth"`
}

type ProductCategorySEOView struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	RoutePath       string `json:"route_path"`
	Locale          string `json:"locale"`
	Status          string `json:"status"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	Intro           string `json:"intro"`
}

type ProductBreadcrumbItemView struct {
	Type string `json:"type"`
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
	Path string `json:"path"`
}

type ProductBreadcrumbView struct {
	Status string                      `json:"status"`
	Reason string                      `json:"reason,omitempty"`
	Items  []ProductBreadcrumbItemView `json:"items"`
}

type ProductCategoryService struct {
	repo      *repository.ProductCategoryRepository
	mediaRepo *repository.MediaRepository
}

func NewProductCategoryService(repo *repository.ProductCategoryRepository, mediaRepos ...*repository.MediaRepository) *ProductCategoryService {
	var mediaRepo *repository.MediaRepository
	if len(mediaRepos) > 0 {
		mediaRepo = mediaRepos[0]
	}
	return &ProductCategoryService{repo: repo, mediaRepo: mediaRepo}
}

func (s *ProductCategoryService) List(includeDisabled bool) (*ProductCategoryListView, error) {
	categories, err := s.repo.List(includeDisabled)
	if err != nil {
		return nil, err
	}
	view := buildProductCategoryListView(categories)
	return &view, nil
}

func (s *ProductCategoryService) ListPublic(locale string) (*ProductCategoryListView, error) {
	requestedLocale := locales.Normalize(locale)
	categories, err := s.repo.ListWithTranslations(false, []string{requestedLocale})
	if err != nil {
		return nil, err
	}
	view := buildLocalizedProductCategoryListView(categories, requestedLocale)
	return &view, nil
}

func (s *ProductCategoryService) GetPublicBySlug(slug, locale string) (*ProductCategoryView, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrProductCategoryNotFound
	}

	view, err := s.ListPublic(locale)
	if err != nil {
		return nil, err
	}
	for index := range view.Flat {
		if view.Flat[index].Slug == slug {
			result := view.Flat[index]
			return &result, nil
		}
	}
	return nil, ErrProductCategoryNotFound
}

func (s *ProductCategoryService) BuildProductBreadcrumb(item product.Product, locale string) (*ProductBreadcrumbView, error) {
	normalizedLocale := locales.Normalize(locale)
	result := fallbackProductBreadcrumb(item, normalizedLocale)
	if s == nil || s.repo == nil {
		result.Status = "unavailable"
		result.Reason = "category_service_unavailable"
		return result, nil
	}
	if item.ProductCategoryID == nil || *item.ProductCategoryID == 0 {
		result.Status = "unavailable"
		result.Reason = "missing_category"
		return result, nil
	}

	categories, err := s.repo.ListWithTranslations(false, []string{normalizedLocale})
	if err != nil {
		return nil, err
	}

	byID := productCategoriesByID(categories)
	categoryID := *item.ProductCategoryID
	chain := make([]product.ProductCategory, 0, ProductCategoryMaxDepth)
	visited := make(map[uint]struct{}, len(byID))
	for currentID := categoryID; currentID > 0; {
		if _, seen := visited[currentID]; seen {
			result.Status = "unavailable"
			result.Reason = "category_cycle"
			return result, nil
		}
		visited[currentID] = struct{}{}

		category, ok := byID[currentID]
		if !ok || !category.IsEnabled || strings.TrimSpace(category.Slug) == "" {
			result.Status = "unavailable"
			result.Reason = "category_path_incomplete"
			return result, nil
		}
		chain = append(chain, category)
		if category.ParentID == nil {
			break
		}
		currentID = *category.ParentID
	}

	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}

	categorySlugs := make([]string, 0, len(chain))
	for _, category := range chain {
		categorySlugs = append(categorySlugs, category.Slug)
	}

	items := make([]ProductBreadcrumbItemView, 0, len(chain)+3)
	items = append(items, result.Items[:2]...)
	for index, category := range chain {
		items = append(items, ProductBreadcrumbItemView{
			Type: "category",
			ID:   category.ID,
			Name: productCategoryToLocalizedView(category, normalizedLocale).Name,
			Slug: category.Slug,
			Path: seodomain.BuildCategoryRoute(normalizedLocale, categorySlugs[:index+1]...).Path,
		})
	}
	items = append(items, ProductBreadcrumbItemView{
		Type: "product",
		ID:   item.ID,
		Name: item.Name,
		Slug: item.Slug,
		Path: seodomain.BuildProductRoute(normalizedLocale, item.Slug).Path,
	})
	result.Items = items
	result.Status = "ready"
	result.Reason = ""
	return result, nil
}

func (s *ProductCategoryService) ListSEO(includeDisabled bool, locale, search string) ([]ProductCategorySEOView, error) {
	categories, err := s.repo.ListWithTranslations(includeDisabled, nil)
	if err != nil {
		return nil, err
	}

	normalizedLocale := strings.TrimSpace(locale)
	if normalizedLocale != "" {
		normalizedLocale, err = requireSupportedLocale(normalizedLocale)
		if err != nil {
			return nil, err
		}
	}

	byID := productCategoriesByID(categories)
	views := make([]ProductCategorySEOView, 0, len(categories))
	for _, category := range categories {
		localesForCategory := []string{"en"}
		if normalizedLocale != "" {
			localesForCategory = []string{normalizedLocale}
		} else {
			seenLocales := map[string]struct{}{"en": {}}
			for _, translation := range category.Translations {
				translationLocale := locales.Normalize(translation.Locale)
				if _, exists := seenLocales[translationLocale]; exists {
					continue
				}
				seenLocales[translationLocale] = struct{}{}
				localesForCategory = append(localesForCategory, translationLocale)
			}
		}

		for _, categoryLocale := range localesForCategory {
			view := productCategorySEOView(category, categoryLocale, byID)
			if !matchesProductCategorySEOSearch(view, search) {
				continue
			}
			views = append(views, view)
		}
	}
	return views, nil
}

func (s *ProductCategoryService) GetSEO(id uint, locale string) (*ProductCategorySEOView, error) {
	normalizedLocale := locales.Normalize(locale)
	if normalizedLocale == "" {
		normalizedLocale = "en"
	}
	views, err := s.ListSEO(true, normalizedLocale, "")
	if err != nil {
		return nil, err
	}
	for index := range views {
		if views[index].ID == id {
			result := views[index]
			return &result, nil
		}
	}
	return nil, ErrProductCategoryNotFound
}

func (s *ProductCategoryService) UpdateSEO(id uint, locale string, metaTitle, metaDescription, intro *string) (*ProductCategorySEOView, error) {
	normalizedLocale := locales.Normalize(locale)
	if normalizedLocale == "" {
		normalizedLocale = "en"
	}
	if _, err := requireSupportedLocale(normalizedLocale); err != nil {
		return nil, err
	}

	current, err := s.GetSEO(id, normalizedLocale)
	if err != nil {
		return nil, err
	}

	nextMetaTitle := current.MetaTitle
	if metaTitle != nil {
		nextMetaTitle = strings.TrimSpace(*metaTitle)
	}
	nextMetaDescription := current.MetaDescription
	if metaDescription != nil {
		nextMetaDescription = strings.TrimSpace(*metaDescription)
	}
	nextIntro := current.Intro
	if intro != nil {
		nextIntro, err = safehtml.Sanitize(*intro)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid intro: %v", ErrProductCategorySEOInvalid, err)
		}
	}

	if len([]rune(nextMetaTitle)) > seoMetaTitleLimit {
		return nil, fmt.Errorf("%w: meta_title exceeds %d characters", ErrInvalidSEOSettings, seoMetaTitleLimit)
	}
	if len([]rune(nextMetaDescription)) > seoMetaDescriptionLimit {
		return nil, fmt.Errorf("%w: meta_description exceeds %d characters", ErrInvalidSEOSettings, seoMetaDescriptionLimit)
	}

	if err := s.repo.UpdateSEO(id, normalizedLocale, nextMetaTitle, nextMetaDescription, nextIntro); err != nil {
		return nil, err
	}
	return s.GetSEO(id, normalizedLocale)
}

func (s *ProductCategoryService) Get(id uint) (*ProductCategoryView, error) {
	category, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	view := productCategoryToView(*category)
	return &view, nil
}

func (s *ProductCategoryService) ListTranslations(id uint) ([]ProductCategoryTranslationView, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	translations, err := s.repo.ListTranslations(id)
	if err != nil {
		return nil, err
	}

	views := make([]ProductCategoryTranslationView, 0, len(translations))
	for _, translation := range translations {
		views = append(views, productCategoryTranslationToView(translation))
	}
	return views, nil
}

func (s *ProductCategoryService) UpdateTranslations(id uint, inputs []ProductCategoryTranslationInput) ([]ProductCategoryTranslationView, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	existingTranslations, err := s.repo.ListTranslations(id)
	if err != nil {
		return nil, err
	}
	existingByLocale := make(map[string]product.ProductCategoryTranslation, len(existingTranslations))
	for _, translation := range existingTranslations {
		existingByLocale[locales.Normalize(translation.Locale)] = translation
	}

	translations := make([]product.ProductCategoryTranslation, 0, len(inputs))
	seenLocales := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		locale, err := requireSupportedLocale(input.Locale)
		if err != nil {
			return nil, err
		}
		if _, exists := seenLocales[locale]; exists {
			return nil, fmt.Errorf("%w: duplicate locale %s", ErrProductCategoryTranslationInvalid, locale)
		}
		seenLocales[locale] = struct{}{}

		name := strings.TrimSpace(input.Name)
		description := strings.TrimSpace(input.Description)
		if name == "" {
			continue
		}
		if len(name) > 120 {
			return nil, fmt.Errorf("%w: translated name is too long for %s", ErrProductCategoryTranslationInvalid, locale)
		}

		translations = append(translations, product.ProductCategoryTranslation{
			ProductCategoryID: id,
			Locale:            locale,
			Name:              name,
			Description:       description,
			MetaTitle:         existingByLocale[locale].MetaTitle,
			MetaDesc:          existingByLocale[locale].MetaDesc,
			SEOIntro:          existingByLocale[locale].SEOIntro,
		})
	}

	if err := s.repo.ReplaceTranslations(id, translations); err != nil {
		return nil, err
	}
	return s.ListTranslations(id)
}

func (s *ProductCategoryService) Create(input ProductCategoryInput) (*ProductCategoryView, error) {
	category, err := normalizeProductCategoryInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.applyCategoryImage(category); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsBySlug(category.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductCategorySlugExists
	}

	all, err := s.repo.List(true)
	if err != nil {
		return nil, err
	}
	byID := productCategoriesByID(all)
	depth, err := nextProductCategoryDepth(category.ParentID, byID)
	if err != nil {
		return nil, err
	}
	category.Depth = depth

	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return s.Get(category.ID)
}

func (s *ProductCategoryService) Update(id uint, input ProductCategoryInput) (*ProductCategoryView, error) {
	existing, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	category, err := normalizeProductCategoryInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.applyCategoryImage(category); err != nil {
		return nil, err
	}
	category.ID = id

	if isSystemProductCategorySlug(existing.Slug) {
		if category.Slug != existing.Slug {
			return nil, fmt.Errorf("%w: slug %s is required", ErrProductCategorySystemProtected, existing.Slug)
		}
		if !category.IsEnabled {
			return nil, fmt.Errorf("%w: %s must remain enabled", ErrProductCategorySystemProtected, existing.Slug)
		}
	}

	exists, err := s.repo.ExistsBySlug(category.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductCategorySlugExists
	}

	all, err := s.repo.List(true)
	if err != nil {
		return nil, err
	}
	byID := productCategoriesByID(all)
	childMap := productCategoryChildrenByParent(all)

	if _, ok := byID[id]; !ok {
		return nil, ErrProductCategoryNotFound
	}
	if category.ParentID != nil {
		if *category.ParentID == id {
			return nil, fmt.Errorf("%w: category cannot be its own parent", ErrProductCategoryInvalid)
		}
		if _, ok := byID[*category.ParentID]; !ok {
			return nil, fmt.Errorf("%w: parent category not found", ErrProductCategoryInvalid)
		}
		if productCategoryIsDescendant(*category.ParentID, id, byID) {
			return nil, fmt.Errorf("%w: parent category cannot be a descendant", ErrProductCategoryInvalid)
		}
	}

	newDepth, err := nextProductCategoryDepth(category.ParentID, byID)
	if err != nil {
		return nil, err
	}
	category.Depth = newDepth

	depthShift := newDepth - existing.Depth
	descendantDepths := make(map[uint]int)
	for _, descendantID := range productCategorySubtreeIDs(id, childMap) {
		item, ok := byID[descendantID]
		if !ok {
			continue
		}
		depth := item.Depth + depthShift
		if descendantID == id {
			depth = newDepth
		}
		if depth < 1 || depth > ProductCategoryMaxDepth {
			return nil, fmt.Errorf("%w: category tree supports at most %d levels", ErrProductCategoryInvalid, ProductCategoryMaxDepth)
		}
		descendantDepths[descendantID] = depth
	}

	if err := s.repo.Update(category, descendantDepths); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ProductCategoryService) Delete(id uint) error {
	existing, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductCategoryNotFound
	}
	if err != nil {
		return err
	}
	if isSystemProductCategorySlug(existing.Slug) {
		return fmt.Errorf("%w: %s", ErrProductCategorySystemProtected, existing.Slug)
	}
	childCount, err := s.repo.CountChildren(id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return fmt.Errorf("%w: delete child categories first", ErrProductCategoryHasChildren)
	}
	return s.repo.Delete(id)
}

func isSystemProductCategorySlug(slug string) bool {
	switch strings.ToLower(strings.TrimSpace(slug)) {
	case SystemProductCategoryWheelsetSlug:
		return true
	default:
		return false
	}
}

func normalizeProductCategoryInput(input ProductCategoryInput) (*product.ProductCategory, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrProductCategoryInvalid)
	}
	if len(name) > 120 || len(slug) > 120 || !productCategorySlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must use lowercase letters, numbers, dashes, or underscores", ErrProductCategoryInvalid)
	}
	if input.ParentID != nil && *input.ParentID == 0 {
		input.ParentID = nil
	}

	return &product.ProductCategory{
		ParentID:          input.ParentID,
		Name:              name,
		Slug:              slug,
		Description:       strings.TrimSpace(input.Description),
		ImageMediaAssetID: input.ImageMediaAssetID,
		SortOrder:         input.SortOrder,
		IsEnabled:         input.IsEnabled,
	}, nil
}

func (s *ProductCategoryService) applyCategoryImage(category *product.ProductCategory) error {
	if category == nil || category.ImageMediaAssetID == nil {
		if category != nil {
			category.ImageURL = ""
		}
		return nil
	}
	if *category.ImageMediaAssetID == 0 || s.mediaRepo == nil {
		return ErrProductCategoryImageInvalid
	}

	asset, err := s.mediaRepo.FindAssetByID(*category.ImageMediaAssetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductCategoryImageInvalid
		}
		return err
	}
	if asset.MediaType != "image" || asset.Status != "active" || asset.Visibility != "public" || strings.TrimSpace(asset.URL) == "" {
		return ErrProductCategoryImageInvalid
	}
	category.ImageURL = strings.TrimSpace(asset.URL)
	return nil
}

func nextProductCategoryDepth(parentID *uint, byID map[uint]product.ProductCategory) (int, error) {
	if parentID == nil {
		return 1, nil
	}
	parent, ok := byID[*parentID]
	if !ok {
		return 0, fmt.Errorf("%w: parent category not found", ErrProductCategoryInvalid)
	}
	depth := parent.Depth + 1
	if depth > ProductCategoryMaxDepth {
		return 0, fmt.Errorf("%w: category tree supports at most %d levels", ErrProductCategoryInvalid, ProductCategoryMaxDepth)
	}
	return depth, nil
}

func productCategoriesByID(categories []product.ProductCategory) map[uint]product.ProductCategory {
	byID := make(map[uint]product.ProductCategory, len(categories))
	for _, category := range categories {
		byID[category.ID] = category
	}
	return byID
}

func productCategoryChildrenByParent(categories []product.ProductCategory) map[uint][]uint {
	children := make(map[uint][]uint)
	for _, category := range categories {
		if category.ParentID == nil {
			continue
		}
		children[*category.ParentID] = append(children[*category.ParentID], category.ID)
	}
	return children
}

func productCategoryIsDescendant(candidateID, ancestorID uint, byID map[uint]product.ProductCategory) bool {
	visited := make(map[uint]struct{}, len(byID))
	currentID := candidateID
	for {
		if currentID == ancestorID {
			return true
		}
		if _, seen := visited[currentID]; seen {
			return true
		}
		visited[currentID] = struct{}{}
		current, ok := byID[currentID]
		if !ok || current.ParentID == nil {
			return false
		}
		currentID = *current.ParentID
	}
}

func productCategorySubtreeIDs(rootID uint, childMap map[uint][]uint) []uint {
	ids := make([]uint, 0)
	visited := make(map[uint]struct{})
	var walk func(uint)
	walk = func(id uint) {
		if _, seen := visited[id]; seen {
			return
		}
		visited[id] = struct{}{}
		ids = append(ids, id)
		for _, childID := range childMap[id] {
			walk(childID)
		}
	}
	walk(rootID)
	return ids
}

func buildProductCategoryListView(categories []product.ProductCategory) ProductCategoryListView {
	flat := make([]ProductCategoryView, 0, len(categories))
	knownIDs := make(map[uint]struct{}, len(categories))
	for _, category := range categories {
		flat = append(flat, productCategoryToAdminView(category))
		knownIDs[category.ID] = struct{}{}
	}

	children := make(map[uint][]product.ProductCategory)
	for _, category := range categories {
		parentID := uint(0)
		if category.ParentID != nil {
			if _, ok := knownIDs[*category.ParentID]; ok {
				parentID = *category.ParentID
			}
		}
		children[parentID] = append(children[parentID], category)
	}

	visited := make(map[uint]struct{}, len(categories))
	var build func(parentID uint, depth int) []ProductCategoryView
	build = func(parentID uint, depth int) []ProductCategoryView {
		if depth > ProductCategoryMaxDepth {
			return nil
		}
		items := make([]ProductCategoryView, 0, len(children[parentID]))
		for _, category := range children[parentID] {
			if _, seen := visited[category.ID]; seen {
				continue
			}
			visited[category.ID] = struct{}{}
			view := productCategoryToAdminView(category)
			view.Children = build(category.ID, depth+1)
			items = append(items, view)
		}
		return items
	}

	tree := build(0, 1)
	view := ProductCategoryListView{Tree: tree, Flat: flat, MaxDepth: ProductCategoryMaxDepth}
	applyProductCategoryRoutePaths(&view, categories, "en")
	return view
}

func productCategoryToAdminView(category product.ProductCategory) ProductCategoryView {
	view := productCategoryToView(category)
	supportedLocales := locales.EnabledLocaleCodes()
	translatedLocales := make(map[string]struct{}, len(category.Translations))

	for _, translation := range category.Translations {
		locale := locales.Normalize(translation.Locale)
		if !locales.IsSupportedCode(locale) || strings.TrimSpace(translation.Name) == "" {
			continue
		}
		translatedLocales[locale] = struct{}{}
	}

	view.TranslationCompleted = len(translatedLocales)
	view.TranslationTotal = len(supportedLocales)
	for _, locale := range supportedLocales {
		if _, ok := translatedLocales[locale]; ok {
			continue
		}
		view.TranslationMissingLocales = append(view.TranslationMissingLocales, locale)
	}
	return view
}

func buildLocalizedProductCategoryListView(categories []product.ProductCategory, locale string) ProductCategoryListView {
	flat := make([]ProductCategoryView, 0, len(categories))
	knownIDs := make(map[uint]struct{}, len(categories))
	for _, category := range categories {
		flat = append(flat, productCategoryToLocalizedView(category, locale))
		knownIDs[category.ID] = struct{}{}
	}

	children := make(map[uint][]product.ProductCategory)
	for _, category := range categories {
		parentID := uint(0)
		if category.ParentID != nil {
			if _, ok := knownIDs[*category.ParentID]; ok {
				parentID = *category.ParentID
			}
		}
		children[parentID] = append(children[parentID], category)
	}

	visited := make(map[uint]struct{}, len(categories))
	var build func(parentID uint, depth int) []ProductCategoryView
	build = func(parentID uint, depth int) []ProductCategoryView {
		if depth > ProductCategoryMaxDepth {
			return nil
		}
		items := make([]ProductCategoryView, 0, len(children[parentID]))
		for _, category := range children[parentID] {
			if _, seen := visited[category.ID]; seen {
				continue
			}
			visited[category.ID] = struct{}{}
			view := productCategoryToLocalizedView(category, locale)
			view.Children = build(category.ID, depth+1)
			items = append(items, view)
		}
		return items
	}

	tree := build(0, 1)
	view := ProductCategoryListView{Tree: tree, Flat: flat, MaxDepth: ProductCategoryMaxDepth}
	applyProductCategoryRoutePaths(&view, categories, locale)
	return view
}

func productCategoryToView(category product.ProductCategory) ProductCategoryView {
	return ProductCategoryView{
		ID:                category.ID,
		ParentID:          category.ParentID,
		Name:              category.Name,
		Slug:              category.Slug,
		Description:       category.Description,
		MetaTitle:         category.MetaTitle,
		MetaDescription:   category.MetaDesc,
		Intro:             firstNonEmptyCategoryValue(category.SEOIntro, category.Description),
		ImageMediaAssetID: category.ImageMediaAssetID,
		ImageURL:          category.ImageURL,
		Depth:             category.Depth,
		SortOrder:         category.SortOrder,
		IsEnabled:         category.IsEnabled,
		CreatedAt:         category.CreatedAt,
		UpdatedAt:         category.UpdatedAt,
		Children:          []ProductCategoryView{},
	}
}

func productCategoryToLocalizedView(category product.ProductCategory, locale string) ProductCategoryView {
	view := productCategoryToView(category)
	for _, translation := range category.Translations {
		if locales.Normalize(translation.Locale) != locale {
			continue
		}
		if name := strings.TrimSpace(translation.Name); name != "" {
			view.Name = name
		}
		if description := strings.TrimSpace(translation.Description); description != "" {
			view.Description = description
		}
		if locale != "en" {
			if metaTitle := strings.TrimSpace(translation.MetaTitle); metaTitle != "" {
				view.MetaTitle = metaTitle
			}
			if metaDescription := strings.TrimSpace(translation.MetaDesc); metaDescription != "" {
				view.MetaDescription = metaDescription
			}
			if intro := strings.TrimSpace(translation.SEOIntro); intro != "" {
				view.Intro = intro
			} else if description := strings.TrimSpace(translation.Description); description != "" {
				view.Intro = description
			}
		}
		break
	}
	return view
}

func productCategorySEOView(
	category product.ProductCategory,
	locale string,
	byID map[uint]product.ProductCategory,
) ProductCategorySEOView {
	localized := productCategoryToLocalizedView(category, locale)
	metaTitle := strings.TrimSpace(category.MetaTitle)
	metaDescription := strings.TrimSpace(category.MetaDesc)
	intro := strings.TrimSpace(category.SEOIntro)
	if locale != "en" {
		metaTitle = ""
		metaDescription = ""
		intro = ""
		for _, translation := range category.Translations {
			if locales.Normalize(translation.Locale) != locale {
				continue
			}
			metaTitle = strings.TrimSpace(translation.MetaTitle)
			metaDescription = strings.TrimSpace(translation.MetaDesc)
			intro = strings.TrimSpace(translation.SEOIntro)
			break
		}
	}
	return ProductCategorySEOView{
		ID:              category.ID,
		Name:            localized.Name,
		Slug:            category.Slug,
		RoutePath:       productCategoryRoutePath(category.ID, locale, byID),
		Locale:          locale,
		Status:          map[bool]string{true: "active", false: "inactive"}[category.IsEnabled],
		MetaTitle:       metaTitle,
		MetaDescription: metaDescription,
		Intro:           intro,
	}
}

func matchesProductCategorySEOSearch(view ProductCategorySEOView, search string) bool {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return true
	}
	for _, value := range []string{view.Name, view.Slug, view.RoutePath} {
		if strings.Contains(strings.ToLower(value), search) {
			return true
		}
	}
	return false
}

func productCategoryTranslationToView(translation product.ProductCategoryTranslation) ProductCategoryTranslationView {
	return ProductCategoryTranslationView{
		ID:                translation.ID,
		ProductCategoryID: translation.ProductCategoryID,
		Locale:            locales.Normalize(translation.Locale),
		Name:              translation.Name,
		Description:       translation.Description,
		MetaTitle:         translation.MetaTitle,
		MetaDescription:   translation.MetaDesc,
		Intro:             translation.SEOIntro,
		CreatedAt:         translation.CreatedAt,
		UpdatedAt:         translation.UpdatedAt,
	}
}

func applyProductCategoryRoutePaths(
	view *ProductCategoryListView,
	categories []product.ProductCategory,
	locale string,
) {
	if view == nil {
		return
	}
	byID := productCategoriesByID(categories)
	paths := make(map[uint]string, len(categories))
	for _, category := range categories {
		paths[category.ID] = productCategoryRoutePath(category.ID, locale, byID)
	}

	for index := range view.Flat {
		view.Flat[index].RoutePath = paths[view.Flat[index].ID]
	}
	var applyTree func([]ProductCategoryView)
	applyTree = func(items []ProductCategoryView) {
		for index := range items {
			items[index].RoutePath = paths[items[index].ID]
			applyTree(items[index].Children)
		}
	}
	applyTree(view.Tree)
}

func productCategoryRoutePath(id uint, locale string, byID map[uint]product.ProductCategory) string {
	segments := make([]string, 0, ProductCategoryMaxDepth)
	visited := make(map[uint]struct{}, len(byID))
	currentID := id
	for currentID > 0 {
		if _, seen := visited[currentID]; seen {
			return ""
		}
		visited[currentID] = struct{}{}

		category, ok := byID[currentID]
		if !ok || strings.TrimSpace(category.Slug) == "" {
			return ""
		}
		segments = append(segments, category.Slug)
		if category.ParentID == nil {
			break
		}
		currentID = *category.ParentID
	}

	for left, right := 0, len(segments)-1; left < right; left, right = left+1, right-1 {
		segments[left], segments[right] = segments[right], segments[left]
	}
	return seodomain.BuildCategoryRoute(locale, segments...).Path
}

func fallbackProductBreadcrumb(item product.Product, locale string) *ProductBreadcrumbView {
	return &ProductBreadcrumbView{
		Status: "ready",
		Items: []ProductBreadcrumbItemView{
			{
				Type: "home",
				Name: "Home",
				Path: seodomain.BuildStaticRoute(locale, "/").Path,
			},
			{
				Type: "shop",
				Name: "Shop",
				Path: seodomain.BuildStaticRoute(locale, "/shop").Path,
			},
			{
				Type: "product",
				ID:   item.ID,
				Name: item.Name,
				Slug: item.Slug,
				Path: seodomain.BuildProductRoute(locale, item.Slug).Path,
			},
		},
	}
}

func firstNonEmptyCategoryValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
