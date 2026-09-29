package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SchwalbeCatalogRefreshSnapshotRow struct {
	ArticleNo       string   `json:"article_no" gorm:"column:article_no"`
	EAN             *string  `json:"ean" gorm:"column:ean"`
	ModelName       string   `json:"model_name" gorm:"column:model_name"`
	ETRTO           string   `json:"etrto" gorm:"column:etrto"`
	InchDesignation *string  `json:"inch_designation" gorm:"column:inch_designation"`
	WeightG         *float64 `json:"weight_g" gorm:"column:weight_g"`
	VersionLabel    *string  `json:"version_label" gorm:"column:version_label"`
	Compound        *string  `json:"compound" gorm:"column:compound"`
	Color           *string  `json:"color" gorm:"column:color"`
	Bead            *string  `json:"bead" gorm:"column:bead"`
	EBikeRating     *string  `json:"e_bike_rating" gorm:"column:e_bike_rating"`
	EPI             *float64 `json:"epi" gorm:"column:epi"`
	LoadKG          *float64 `json:"load_kg" gorm:"column:load_kg"`
	Seal            *string  `json:"seal" gorm:"column:seal"`
	Tread           *string  `json:"tread" gorm:"column:tread"`
	MinPressureBar  *float64 `json:"min_pressure_bar" gorm:"column:min_pressure_bar"`
	MaxPressureBar  *float64 `json:"max_pressure_bar" gorm:"column:max_pressure_bar"`
	MinPressurePSI  *float64 `json:"min_pressure_psi" gorm:"column:min_pressure_psi"`
	MaxPressurePSI  *float64 `json:"max_pressure_psi" gorm:"column:max_pressure_psi"`
	SourceURL       *string  `json:"source_url" gorm:"column:source_url"`
	SourceCheckedAt string   `json:"source_checked_at" gorm:"column:source_checked_at"`
}

type SchwalbeCatalogRefreshAuditService struct {
	db *gorm.DB
}

func NewSchwalbeCatalogRefreshAuditService(db *gorm.DB) *SchwalbeCatalogRefreshAuditService {
	return &SchwalbeCatalogRefreshAuditService{db: db}
}

type SchwalbeCatalogRefreshAuditReport struct {
	GeneratedAt  time.Time                     `json:"generated_at"`
	IncomingRows int                           `json:"incoming_rows"`
	CurrentRows  int                           `json:"current_rows"`
	Catalog      SchwalbeCatalogRefreshCatalog `json:"catalog"`
	Sales        SchwalbeCatalogRefreshSales   `json:"sales"`
	SalesScope   string                        `json:"sales_scope"`
}

type SchwalbeCatalogRefreshCatalog struct {
	Added                 []string                             `json:"added"`
	NotInIncomingSnapshot []string                             `json:"not_in_incoming_snapshot"`
	Changed               []SchwalbeCatalogArticleChange       `json:"changed"`
	UnchangedCount        int                                  `json:"unchanged_count"`
	SourceMetadata        SchwalbeCatalogSourceMetadataSummary `json:"source_metadata"`
}

type SchwalbeCatalogArticleChange struct {
	ArticleNo string                       `json:"article_no"`
	Fields    []SchwalbeCatalogFieldChange `json:"fields"`
}

type SchwalbeCatalogFieldChange struct {
	Field    string `json:"field"`
	Current  any    `json:"current"`
	Incoming any    `json:"incoming"`
}

type SchwalbeCatalogSourceMetadataSummary struct {
	SourceURLChanges         []SchwalbeSourceURLChange       `json:"source_url_changes"`
	CheckedAtChangeGroups    []SchwalbeSourceCheckedAtChange `json:"checked_at_change_groups"`
	CheckedAtChangedArticles int                             `json:"checked_at_changed_articles"`
}

type SchwalbeSourceURLChange struct {
	ArticleNo string  `json:"article_no"`
	Current   *string `json:"current"`
	Incoming  *string `json:"incoming"`
}

type SchwalbeSourceCheckedAtChange struct {
	CurrentArticlesCheckedAt string `json:"current"`
	IncomingCheckedAt        string `json:"incoming"`
	ArticleCount             int    `json:"article_count"`
}

type SchwalbeCatalogRefreshSales struct {
	ActiveProductsScanned    int                              `json:"active_products_scanned"`
	MissingArticleNo         []SchwalbeSalesProductReference  `json:"missing_article_no"`
	NotInCurrentCatalog      []SchwalbeSalesProductReference  `json:"not_in_current_catalog"`
	NotInIncomingSnapshot    []SchwalbeSalesProductReference  `json:"not_in_incoming_snapshot"`
	DuplicateArticleNos      []SchwalbeDuplicateSalesArticle  `json:"duplicate_article_nos"`
	SpecificationDifferences []SchwalbeSalesProductDifference `json:"specification_differences"`
}

type SchwalbeSalesProductReference struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	ArticleNo string `json:"article_no,omitempty"`
}

type SchwalbeDuplicateSalesArticle struct {
	ArticleNo string                          `json:"article_no"`
	Products  []SchwalbeSalesProductReference `json:"products"`
}

type SchwalbeSalesProductDifference struct {
	ProductID                      uint                          `json:"product_id"`
	Name                           string                        `json:"name"`
	ArticleNo                      string                        `json:"article_no"`
	IncomingSnapshotMissing        bool                          `json:"incoming_snapshot_missing"`
	DiffersFromCurrentCatalog      []SchwalbeSpecDifference      `json:"differs_from_current_catalog"`
	DiffersFromIncomingSnapshot    []SchwalbeSpecDifference      `json:"differs_from_incoming_snapshot"`
	OfficialFieldsChangedByRefresh []SchwalbeOfficialFieldChange `json:"official_fields_changed_by_refresh"`
	RequiresSalesProductReview     bool                          `json:"requires_sales_product_review"`
}

type SchwalbeOfficialFieldChange struct {
	Field    string `json:"field"`
	Current  any    `json:"current"`
	Incoming any    `json:"incoming"`
	Product  any    `json:"product"`
}

type SchwalbeSpecDifference struct {
	Field    string `json:"field"`
	Official any    `json:"official"`
	Product  any    `json:"product"`
}

type schwalbeAuditField struct {
	slug    string
	numeric bool
}

var schwalbeAuditFields = []schwalbeAuditField{
	{slug: "ean"},
	{slug: "model_name"},
	{slug: "etrto"},
	{slug: "inch_designation"},
	{slug: "weight_g", numeric: true},
	{slug: "version_label"},
	{slug: "compound"},
	{slug: "color"},
	{slug: "bead"},
	{slug: "e_bike_rating"},
	{slug: "epi", numeric: true},
	{slug: "load_kg", numeric: true},
	{slug: "seal"},
	{slug: "tread"},
	{slug: "min_pressure_bar", numeric: true},
	{slug: "max_pressure_bar", numeric: true},
	{slug: "min_pressure_psi", numeric: true},
	{slug: "max_pressure_psi", numeric: true},
}

func (s *SchwalbeCatalogRefreshAuditService) Review(ctx context.Context, incoming []SchwalbeCatalogRefreshSnapshotRow) (*SchwalbeCatalogRefreshAuditReport, error) {
	incomingByArticle, err := indexSchwalbeRefreshSnapshot(incoming, true)
	if err != nil {
		return nil, err
	}

	var current []SchwalbeCatalogRefreshSnapshotRow
	var products []schwalbeSalesProduct
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := loadSchwalbeCatalogRefreshCatalog(tx, &current); err != nil {
			return err
		}
		if err := loadSchwalbeActiveSalesProducts(tx, &products); err != nil {
			return err
		}
		return nil
	}, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, fmt.Errorf("read Schwalbe catalog and sales snapshots: %w", err)
	}

	currentByArticle, err := indexSchwalbeRefreshSnapshot(current, false)
	if err != nil {
		return nil, fmt.Errorf("index current Schwalbe catalog: %w", err)
	}
	report := buildSchwalbeCatalogRefreshAuditReport(currentByArticle, incomingByArticle, products)
	return report, nil
}

func loadSchwalbeCatalogRefreshCatalog(tx *gorm.DB, rows *[]SchwalbeCatalogRefreshSnapshotRow) error {
	var definitionSlugs []string
	if err := tx.Table("product_spec_definitions AS definition").
		Select("definition.slug").
		Joins("JOIN product_specification_templates AS template ON template.id = definition.product_specification_template_id").
		Where("template.slug = ?", "schwalbe_tire").
		Order("definition.slug ASC").
		Scan(&definitionSlugs).Error; err != nil {
		return fmt.Errorf("load Schwalbe template definitions: %w", err)
	}
	if err := validateSchwalbeDefinitionSlugs(definitionSlugs); err != nil {
		return err
	}
	query := `
		SELECT article_no, ean, model_name, etrto, inch_designation, weight_g,
		       version_label, compound, color, bead, e_bike_rating, epi, load_kg,
		       seal, tread, min_pressure_bar, max_pressure_bar,
		       min_pressure_psi, max_pressure_psi, source_url,
		       source_checked_at::text AS source_checked_at
		FROM schwalbe_tire_specifications
		ORDER BY article_no ASC`
	if err := tx.Raw(query).Scan(rows).Error; err != nil {
		return fmt.Errorf("load current Schwalbe catalog: %w", err)
	}
	return nil
}

func loadSchwalbeActiveSalesProducts(tx *gorm.DB, products *[]schwalbeSalesProduct) error {
	query := `
		SELECT p.id AS product_id, p.name, definition.slug, spec_value.value AS spec_value
		FROM products AS p
		JOIN product_specification_templates AS template
		  ON template.id = p.product_specification_template_id
		 AND template.slug = ?
	JOIN product_spec_definitions AS definition
		  ON definition.product_specification_template_id = template.id
		LEFT JOIN product_spec_values AS spec_value
		  ON spec_value.product_id = p.id
		 AND spec_value.spec_definition_id = definition.id
		WHERE p.status = 'active'
		  AND p.deleted_at IS NULL
		  AND EXISTS (
		      SELECT 1
		      FROM product_variants AS variant
		      WHERE variant.product_id = p.id
		        AND variant.is_active IS TRUE
		        AND variant.deleted_at IS NULL
		  )
		ORDER BY p.id ASC, definition.sort_order ASC, definition.slug ASC`
	var rows []schwalbeSalesSpecRow
	if err := tx.Raw(query, "schwalbe_tire").Scan(&rows).Error; err != nil {
		return fmt.Errorf("load active Schwalbe sales product specifications: %w", err)
	}

	byID := make(map[uint]*schwalbeSalesProduct)
	for _, row := range rows {
		product := byID[row.ProductID]
		if product == nil {
			product = &schwalbeSalesProduct{ID: row.ProductID, Name: row.Name, Values: make(map[string]*string, 19)}
			byID[row.ProductID] = product
		}
		if row.SpecValue.Valid && strings.TrimSpace(row.SpecValue.String) != "" {
			value := row.SpecValue.String
			product.Values[row.Slug] = &value
		} else {
			product.Values[row.Slug] = nil
		}
	}
	ordered := make([]schwalbeSalesProduct, 0, len(byID))
	for _, product := range byID {
		ordered = append(ordered, *product)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	*products = ordered
	return nil
}

type schwalbeSalesSpecRow struct {
	ProductID uint           `gorm:"column:product_id"`
	Name      string         `gorm:"column:name"`
	Slug      string         `gorm:"column:slug"`
	SpecValue sql.NullString `gorm:"column:spec_value"`
}

type schwalbeSalesProduct struct {
	ID     uint
	Name   string
	Values map[string]*string
}

func validateSchwalbeDefinitionSlugs(actual []string) error {
	want := make(map[string]struct{}, len(schwalbeAuditFields)+1)
	want["article_no"] = struct{}{}
	for _, field := range schwalbeAuditFields {
		want[field.slug] = struct{}{}
	}
	got := make(map[string]struct{}, len(actual))
	for _, slug := range actual {
		got[slug] = struct{}{}
	}
	missing := make([]string, 0)
	for slug := range want {
		if _, ok := got[slug]; !ok {
			missing = append(missing, slug)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("Schwalbe template is missing required specification definitions: %s", strings.Join(missing, ", "))
	}
	unexpected := make([]string, 0)
	for slug := range got {
		if _, ok := want[slug]; !ok {
			unexpected = append(unexpected, slug)
		}
	}
	sort.Strings(unexpected)
	if len(unexpected) > 0 {
		return fmt.Errorf("Schwalbe template has specification definitions outside the 19-field contract: %s", strings.Join(unexpected, ", "))
	}
	return nil
}

func indexSchwalbeRefreshSnapshot(rows []SchwalbeCatalogRefreshSnapshotRow, validateIncomingContract bool) (map[string]SchwalbeCatalogRefreshSnapshotRow, error) {
	if validateIncomingContract && len(rows) == 0 {
		return nil, fmt.Errorf("snapshot is empty; refusing to review an empty catalog refresh")
	}
	indexed := make(map[string]SchwalbeCatalogRefreshSnapshotRow, len(rows))
	for index, row := range rows {
		key := normalizeSchwalbeArticleNo(row.ArticleNo)
		if key == "" {
			return nil, fmt.Errorf("snapshot row %d has an empty article_no", index+1)
		}
		if _, exists := indexed[key]; exists {
			return nil, fmt.Errorf("snapshot contains duplicate Article No. %q after normalization", strings.TrimSpace(row.ArticleNo))
		}
		if validateIncomingContract && (strings.TrimSpace(row.ModelName) == "" || strings.TrimSpace(row.ETRTO) == "") {
			return nil, fmt.Errorf("snapshot row %q is missing required model_name or etrto", row.ArticleNo)
		}
		if validateIncomingContract {
			for _, field := range schwalbeAuditFields {
				if !field.numeric {
					continue
				}
				value := snapshotNumericValue(row, field.slug)
				if value != nil && (!isFiniteNumber(*value) || *value <= 0) {
					return nil, fmt.Errorf("snapshot row %q has invalid positive measurement %s=%v", row.ArticleNo, field.slug, *value)
				}
				if field.slug == "epi" && value != nil && math.Trunc(*value) != *value {
					return nil, fmt.Errorf("snapshot row %q has non-integer EPI %v", row.ArticleNo, *value)
				}
			}
			if row.MinPressureBar != nil && row.MaxPressureBar != nil && *row.MinPressureBar > *row.MaxPressureBar {
				return nil, fmt.Errorf("snapshot row %q has min_pressure_bar greater than max_pressure_bar", row.ArticleNo)
			}
			if row.MinPressurePSI != nil && row.MaxPressurePSI != nil && *row.MinPressurePSI > *row.MaxPressurePSI {
				return nil, fmt.Errorf("snapshot row %q has min_pressure_psi greater than max_pressure_psi", row.ArticleNo)
			}
		}
		indexed[key] = row
	}
	return indexed, nil
}

func buildSchwalbeCatalogRefreshAuditReport(current, incoming map[string]SchwalbeCatalogRefreshSnapshotRow, products []schwalbeSalesProduct) *SchwalbeCatalogRefreshAuditReport {
	checkedAtGroups := make(map[string]int)
	report := &SchwalbeCatalogRefreshAuditReport{
		GeneratedAt:  time.Now().UTC(),
		IncomingRows: len(incoming),
		CurrentRows:  len(current),
		SalesScope:   "Products with status=active, not soft-deleted, and at least one active non-deleted variant; stock quantity is ignored.",
		Catalog: SchwalbeCatalogRefreshCatalog{
			Added:                 make([]string, 0),
			NotInIncomingSnapshot: make([]string, 0),
			Changed:               make([]SchwalbeCatalogArticleChange, 0),
			SourceMetadata: SchwalbeCatalogSourceMetadataSummary{
				SourceURLChanges:      make([]SchwalbeSourceURLChange, 0),
				CheckedAtChangeGroups: make([]SchwalbeSourceCheckedAtChange, 0),
			},
		},
		Sales: SchwalbeCatalogRefreshSales{
			MissingArticleNo:         make([]SchwalbeSalesProductReference, 0),
			NotInCurrentCatalog:      make([]SchwalbeSalesProductReference, 0),
			NotInIncomingSnapshot:    make([]SchwalbeSalesProductReference, 0),
			DuplicateArticleNos:      make([]SchwalbeDuplicateSalesArticle, 0),
			SpecificationDifferences: make([]SchwalbeSalesProductDifference, 0),
		},
	}

	for key, newRow := range incoming {
		oldRow, exists := current[key]
		if !exists {
			report.Catalog.Added = append(report.Catalog.Added, newRow.ArticleNo)
			continue
		}
		fieldChanges := compareOfficialCatalogFields(oldRow, newRow)
		if len(fieldChanges) > 0 {
			report.Catalog.Changed = append(report.Catalog.Changed, SchwalbeCatalogArticleChange{ArticleNo: oldRow.ArticleNo, Fields: fieldChanges})
		} else {
			report.Catalog.UnchangedCount++
		}
		if !auditValuesEqual("source_url", oldRow.SourceURL, newRow.SourceURL) {
			report.Catalog.SourceMetadata.SourceURLChanges = append(report.Catalog.SourceMetadata.SourceURLChanges, SchwalbeSourceURLChange{
				ArticleNo: oldRow.ArticleNo, Current: cloneString(oldRow.SourceURL), Incoming: cloneString(newRow.SourceURL),
			})
		}
		oldDate := strings.TrimSpace(oldRow.SourceCheckedAt)
		newDate := strings.TrimSpace(newRow.SourceCheckedAt)
		if oldDate != newDate {
			report.Catalog.SourceMetadata.CheckedAtChangedArticles++
			dateKey := oldDate + "\x00" + newDate
			checkedAtGroups[dateKey]++
		}
	}
	for key, oldRow := range current {
		if _, exists := incoming[key]; !exists {
			report.Catalog.NotInIncomingSnapshot = append(report.Catalog.NotInIncomingSnapshot, oldRow.ArticleNo)
		}
	}
	for dateKey, articleCount := range checkedAtGroups {
		parts := strings.SplitN(dateKey, "\x00", 2)
		report.Catalog.SourceMetadata.CheckedAtChangeGroups = append(report.Catalog.SourceMetadata.CheckedAtChangeGroups, SchwalbeSourceCheckedAtChange{
			CurrentArticlesCheckedAt: parts[0], IncomingCheckedAt: parts[1], ArticleCount: articleCount,
		})
	}
	sort.Slice(report.Catalog.Added, func(i, j int) bool { return report.Catalog.Added[i] < report.Catalog.Added[j] })
	sort.Slice(report.Catalog.NotInIncomingSnapshot, func(i, j int) bool {
		return report.Catalog.NotInIncomingSnapshot[i] < report.Catalog.NotInIncomingSnapshot[j]
	})
	sort.Slice(report.Catalog.Changed, func(i, j int) bool { return report.Catalog.Changed[i].ArticleNo < report.Catalog.Changed[j].ArticleNo })
	sort.Slice(report.Catalog.SourceMetadata.SourceURLChanges, func(i, j int) bool {
		return report.Catalog.SourceMetadata.SourceURLChanges[i].ArticleNo < report.Catalog.SourceMetadata.SourceURLChanges[j].ArticleNo
	})
	sort.Slice(report.Catalog.SourceMetadata.CheckedAtChangeGroups, func(i, j int) bool {
		if report.Catalog.SourceMetadata.CheckedAtChangeGroups[i].CurrentArticlesCheckedAt == report.Catalog.SourceMetadata.CheckedAtChangeGroups[j].CurrentArticlesCheckedAt {
			return report.Catalog.SourceMetadata.CheckedAtChangeGroups[i].IncomingCheckedAt < report.Catalog.SourceMetadata.CheckedAtChangeGroups[j].IncomingCheckedAt
		}
		return report.Catalog.SourceMetadata.CheckedAtChangeGroups[i].CurrentArticlesCheckedAt < report.Catalog.SourceMetadata.CheckedAtChangeGroups[j].CurrentArticlesCheckedAt
	})

	report.Sales.ActiveProductsScanned = len(products)
	salesByArticle := make(map[string][]schwalbeSalesProduct)
	for _, product := range products {
		articleNo := product.Values["article_no"]
		if articleNo == nil || strings.TrimSpace(*articleNo) == "" {
			report.Sales.MissingArticleNo = append(report.Sales.MissingArticleNo, salesProductReference(product, ""))
			continue
		}
		key := normalizeSchwalbeArticleNo(*articleNo)
		salesByArticle[key] = append(salesByArticle[key], product)
		currentRow, inCurrent := current[key]
		incomingRow, inIncoming := incoming[key]
		ref := salesProductReference(product, *articleNo)
		if !inCurrent {
			report.Sales.NotInCurrentCatalog = append(report.Sales.NotInCurrentCatalog, ref)
		}
		if !inIncoming {
			report.Sales.NotInIncomingSnapshot = append(report.Sales.NotInIncomingSnapshot, ref)
		}

		var currentDiffs, incomingDiffs []SchwalbeSpecDifference
		if inCurrent {
			currentDiffs = compareProductWithCatalog(currentRow, product)
		}
		if inIncoming {
			incomingDiffs = compareProductWithCatalog(incomingRow, product)
		}
		var officialChanges []SchwalbeOfficialFieldChange
		if inCurrent && inIncoming {
			officialChanges = compareCurrentAndIncomingForProduct(currentRow, incomingRow, product)
		}
		if !inIncoming || len(currentDiffs) > 0 || len(incomingDiffs) > 0 || len(officialChanges) > 0 {
			requiresReview := !inIncoming || len(incomingDiffs) > 0
			report.Sales.SpecificationDifferences = append(report.Sales.SpecificationDifferences, SchwalbeSalesProductDifference{
				ProductID: product.ID, Name: product.Name, ArticleNo: *articleNo,
				IncomingSnapshotMissing:   !inIncoming,
				DiffersFromCurrentCatalog: currentDiffs, DiffersFromIncomingSnapshot: incomingDiffs,
				OfficialFieldsChangedByRefresh: officialChanges, RequiresSalesProductReview: requiresReview,
			})
		}
	}
	for key, matchedProducts := range salesByArticle {
		if len(matchedProducts) < 2 {
			continue
		}
		refs := make([]SchwalbeSalesProductReference, 0, len(matchedProducts))
		for _, product := range matchedProducts {
			refs = append(refs, salesProductReference(product, *product.Values["article_no"]))
		}
		sort.Slice(refs, func(i, j int) bool { return refs[i].ProductID < refs[j].ProductID })
		report.Sales.DuplicateArticleNos = append(report.Sales.DuplicateArticleNos, SchwalbeDuplicateSalesArticle{ArticleNo: key, Products: refs})
	}
	sort.Slice(report.Sales.MissingArticleNo, func(i, j int) bool {
		return report.Sales.MissingArticleNo[i].ProductID < report.Sales.MissingArticleNo[j].ProductID
	})
	sort.Slice(report.Sales.NotInCurrentCatalog, func(i, j int) bool {
		return report.Sales.NotInCurrentCatalog[i].ProductID < report.Sales.NotInCurrentCatalog[j].ProductID
	})
	sort.Slice(report.Sales.NotInIncomingSnapshot, func(i, j int) bool {
		return report.Sales.NotInIncomingSnapshot[i].ProductID < report.Sales.NotInIncomingSnapshot[j].ProductID
	})
	sort.Slice(report.Sales.DuplicateArticleNos, func(i, j int) bool {
		return report.Sales.DuplicateArticleNos[i].ArticleNo < report.Sales.DuplicateArticleNos[j].ArticleNo
	})
	sort.Slice(report.Sales.SpecificationDifferences, func(i, j int) bool {
		return report.Sales.SpecificationDifferences[i].ProductID < report.Sales.SpecificationDifferences[j].ProductID
	})
	return report
}

func compareOfficialCatalogFields(current, incoming SchwalbeCatalogRefreshSnapshotRow) []SchwalbeCatalogFieldChange {
	differences := make([]SchwalbeCatalogFieldChange, 0)
	currentValues := schwalbeCatalogValues(current)
	incomingValues := schwalbeCatalogValues(incoming)
	for _, field := range schwalbeAuditFields {
		oldValue, newValue := currentValues[field.slug], incomingValues[field.slug]
		if !auditValuesEqual(field.slug, oldValue, newValue) {
			differences = append(differences, SchwalbeCatalogFieldChange{Field: field.slug, Current: displayAuditValue(field.slug, oldValue), Incoming: displayAuditValue(field.slug, newValue)})
		}
	}
	return differences
}

func compareProductWithCatalog(catalog SchwalbeCatalogRefreshSnapshotRow, product schwalbeSalesProduct) []SchwalbeSpecDifference {
	differences := make([]SchwalbeSpecDifference, 0)
	catalogValues := schwalbeCatalogValues(catalog)
	for _, field := range schwalbeAuditFields {
		official, saved := catalogValues[field.slug], product.Values[field.slug]
		if !auditValuesEqual(field.slug, official, saved) {
			differences = append(differences, SchwalbeSpecDifference{Field: field.slug, Official: displayAuditValue(field.slug, official), Product: displayAuditValue(field.slug, saved)})
		}
	}
	return differences
}

func compareCurrentAndIncomingForProduct(current, incoming SchwalbeCatalogRefreshSnapshotRow, product schwalbeSalesProduct) []SchwalbeOfficialFieldChange {
	changes := make([]SchwalbeOfficialFieldChange, 0)
	currentValues, incomingValues := schwalbeCatalogValues(current), schwalbeCatalogValues(incoming)
	for _, field := range schwalbeAuditFields {
		oldValue, newValue := currentValues[field.slug], incomingValues[field.slug]
		if auditValuesEqual(field.slug, oldValue, newValue) {
			continue
		}
		changes = append(changes, SchwalbeOfficialFieldChange{
			Field: field.slug, Current: displayAuditValue(field.slug, oldValue), Incoming: displayAuditValue(field.slug, newValue),
			Product: displayAuditValue(field.slug, product.Values[field.slug]),
		})
	}
	return changes
}

func schwalbeCatalogValues(row SchwalbeCatalogRefreshSnapshotRow) map[string]*string {
	return map[string]*string{
		"ean": row.EAN, "model_name": schwalbeAuditStringPtr(row.ModelName), "etrto": schwalbeAuditStringPtr(row.ETRTO),
		"inch_designation": row.InchDesignation, "weight_g": numberStringValue(row.WeightG),
		"version_label": row.VersionLabel, "compound": row.Compound, "color": row.Color,
		"bead": row.Bead, "e_bike_rating": row.EBikeRating, "epi": numberStringValue(row.EPI),
		"load_kg": numberStringValue(row.LoadKG), "seal": row.Seal, "tread": row.Tread,
		"min_pressure_bar": numberStringValue(row.MinPressureBar), "max_pressure_bar": numberStringValue(row.MaxPressureBar),
		"min_pressure_psi": numberStringValue(row.MinPressurePSI), "max_pressure_psi": numberStringValue(row.MaxPressurePSI),
	}
}

func snapshotNumericValue(row SchwalbeCatalogRefreshSnapshotRow, slug string) *float64 {
	switch slug {
	case "weight_g":
		return row.WeightG
	case "epi":
		return row.EPI
	case "load_kg":
		return row.LoadKG
	case "min_pressure_bar":
		return row.MinPressureBar
	case "max_pressure_bar":
		return row.MaxPressureBar
	case "min_pressure_psi":
		return row.MinPressurePSI
	case "max_pressure_psi":
		return row.MaxPressurePSI
	default:
		return nil
	}
}

func auditValuesEqual(field string, left, right *string) bool {
	leftValue, rightValue := normalizedAuditValue(field, left), normalizedAuditValue(field, right)
	return leftValue == rightValue
}

func normalizedAuditValue(field string, value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "<null>"
	}
	trimmed := strings.TrimSpace(*value)
	for _, specField := range schwalbeAuditFields {
		if specField.slug == field && specField.numeric {
			parsed, err := strconv.ParseFloat(trimmed, 64)
			if err == nil && isFiniteNumber(parsed) {
				return "number:" + strconv.FormatFloat(parsed, 'f', -1, 64)
			}
			return "invalid-number:" + trimmed
		}
	}
	if field == "etrto" {
		trimmed = strings.NewReplacer("–", "-", "—", "-", "−", "-", "‑", "-").Replace(trimmed)
	}
	return "text:" + trimmed
}

func displayAuditValue(field string, value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	for _, specField := range schwalbeAuditFields {
		if specField.slug == field && specField.numeric {
			if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil && isFiniteNumber(parsed) {
				return parsed
			}
		}
	}
	return trimmed
}

func normalizeSchwalbeArticleNo(value string) string {
	value = strings.Map(func(r rune) rune {
		switch r {
		case '\u00ad', '\u200b', '\u200c', '\u200d', '\u2060', '\ufeff':
			return -1
		default:
			return r
		}
	}, value)
	return strings.ToLower(strings.TrimSpace(value))
}

func salesProductReference(product schwalbeSalesProduct, articleNo string) SchwalbeSalesProductReference {
	return SchwalbeSalesProductReference{ProductID: product.ID, Name: product.Name, ArticleNo: articleNo}
}

func schwalbeAuditStringPtr(value string) *string {
	return &value
}

func numberStringValue(value *float64) *string {
	if value == nil {
		return nil
	}
	formatted := strconv.FormatFloat(*value, 'f', -1, 64)
	return &formatted
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func isFiniteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func WriteSchwalbeCatalogRefreshAuditText(w io.Writer, report *SchwalbeCatalogRefreshAuditReport) error {
	if _, err := fmt.Fprintf(w, "Schwalbe refresh audit generated_at=%s current_rows=%d incoming_rows=%d\n", report.GeneratedAt.Format(time.RFC3339), report.CurrentRows, report.IncomingRows); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Catalog: added=%d absent_from_snapshot=%d changed=%d unchanged=%d url_changes=%d checked_at_changed=%d\n",
		len(report.Catalog.Added), len(report.Catalog.NotInIncomingSnapshot), len(report.Catalog.Changed), report.Catalog.UnchangedCount,
		len(report.Catalog.SourceMetadata.SourceURLChanges), report.Catalog.SourceMetadata.CheckedAtChangedArticles); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Sales: active=%d missing_article_no=%d not_in_current_catalog=%d not_in_incoming_snapshot=%d duplicate_article_no=%d products_with_differences=%d\n",
		report.Sales.ActiveProductsScanned, len(report.Sales.MissingArticleNo), len(report.Sales.NotInCurrentCatalog), len(report.Sales.NotInIncomingSnapshot), len(report.Sales.DuplicateArticleNos), len(report.Sales.SpecificationDifferences)); err != nil {
		return err
	}
	for _, change := range report.Catalog.Changed {
		if _, err := fmt.Fprintf(w, "CATALOG changed article_no=%s fields=%s\n", change.ArticleNo, catalogChangeFieldNames(change.Fields)); err != nil {
			return err
		}
	}
	for _, articleNo := range report.Catalog.Added {
		if _, err := fmt.Fprintf(w, "CATALOG added article_no=%s\n", articleNo); err != nil {
			return err
		}
	}
	for _, articleNo := range report.Catalog.NotInIncomingSnapshot {
		if _, err := fmt.Fprintf(w, "CATALOG absent_from_incoming_snapshot article_no=%s (review only; no delete is performed)\n", articleNo); err != nil {
			return err
		}
	}
	for _, difference := range report.Sales.SpecificationDifferences {
		if _, err := fmt.Fprintf(w, "SALES product_id=%d article_no=%s review=%t current_diff=%s incoming_diff=%s refresh_fields=%s\n",
			difference.ProductID, difference.ArticleNo, difference.RequiresSalesProductReview,
			differenceFieldNames(difference.DiffersFromCurrentCatalog), differenceFieldNames(difference.DiffersFromIncomingSnapshot), officialChangeFieldNames(difference.OfficialFieldsChangedByRefresh)); err != nil {
			return err
		}
	}
	return nil
}

func differenceFieldNames(differences []SchwalbeSpecDifference) string {
	fields := make([]string, 0, len(differences))
	for _, difference := range differences {
		fields = append(fields, difference.Field)
	}
	return strings.Join(fields, ",")
}

func catalogChangeFieldNames(changes []SchwalbeCatalogFieldChange) string {
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		fields = append(fields, change.Field)
	}
	return strings.Join(fields, ",")
}

func officialChangeFieldNames(changes []SchwalbeOfficialFieldChange) string {
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		fields = append(fields, change.Field)
	}
	return strings.Join(fields, ",")
}
