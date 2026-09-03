package upload

import (
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
)

// SpecCode identifies the intended use of an uploaded asset. The code is
// shared by the API contract, admin upload controls, and validation rules.
type SpecCode string

const (
	SpecProductImage                 SpecCode = "product_image"
	SpecProductDescriptionImage      SpecCode = "product_description_image"
	SpecProductVariantSwatch         SpecCode = "product_variant_swatch"
	SpecMediaLibraryImage            SpecCode = "media_library_image"
	SpecFAQAnswerImage               SpecCode = "faq_answer_image"
	SpecVisualShowcaseHomeCategories SpecCode = "visual_showcase_home_categories"
	SpecVisualShowcaseEditorial      SpecCode = "visual_showcase_editorial"
	SpecSiteLogo                     SpecCode = "site_logo"
	SpecSiteFavicon                  SpecCode = "site_favicon"
	SpecCustomerServiceAvatar        SpecCode = "customer_service_avatar"
	SpecWebsiteProfileAvatar         SpecCode = "website_profile_avatar"
	SpecWebsiteProfileImage          SpecCode = "website_profile_image"
	SpecRefundCancellationImage      SpecCode = "refund_cancellation_image"
	SpecWarrantyEvidence             SpecCode = "warranty_evidence"
	SpecAfterSalesEvidence           SpecCode = "after_sales_evidence"
	SpecSuggestionAttachment         SpecCode = "suggestion_attachment"
	SpecCustomerServiceAttachment    SpecCode = "customer_service_attachment"
	SpecUserShowcaseImage            SpecCode = "user_showcase_image"
)

type UploadSpec struct {
	Code                 string   `json:"code"`
	Kind                 string   `json:"kind"`
	Label                string   `json:"label"`
	Description          string   `json:"description"`
	AcceptedExtensions   []string `json:"accepted_extensions"`
	AcceptedContentTypes []string `json:"accepted_content_types"`
	MaxFileSizeBytes     int64    `json:"max_file_size_bytes"`
	MaxFiles             int      `json:"max_files,omitempty"`
	MaxTotalSizeBytes    int64    `json:"max_total_size_bytes,omitempty"`
	ExactWidth           int      `json:"exact_width,omitempty"`
	ExactHeight          int      `json:"exact_height,omitempty"`
	RecommendedWidth     int      `json:"recommended_width,omitempty"`
	RecommendedHeight    int      `json:"recommended_height,omitempty"`
	RecommendedLongEdge  int      `json:"recommended_long_edge,omitempty"`
	MaxWidth             int      `json:"max_width,omitempty"`
	MaxHeight            int      `json:"max_height,omitempty"`
	MaxPixels            int64    `json:"max_pixels,omitempty"`
	AspectRatioWidth     int      `json:"aspect_ratio_width,omitempty"`
	AspectRatioHeight    int      `json:"aspect_ratio_height,omitempty"`
	AspectRatioLabel     string   `json:"aspect_ratio_label,omitempty"`
	QualityNote          string   `json:"quality_note,omitempty"`
}

type uploadSpecDefinition struct {
	UploadSpec
	FileRule     FileRule
	FilesRule    FilesRule
	HasFilesRule bool
}

var uploadSpecDefinitions = map[SpecCode]uploadSpecDefinition{
	SpecProductImage: makeImageSpec(
		SpecProductImage,
		"Product image",
		"Product primary and gallery images shown in cards and product detail.",
		ProductImageRule,
		1600,
		1600,
		0,
		"Upload at least 1600x1600 px when possible; keep the original product framing.",
	),
	SpecProductDescriptionImage: makeImageSpec(
		SpecProductDescriptionImage,
		"Product description image",
		"Images inserted into long-form product descriptions.",
		ProductDescriptionImageRule,
		0,
		0,
		1600,
		"Use an image with a 1600 px or larger long edge so zoomed desktop content stays clear.",
	),
	SpecProductVariantSwatch: makeImageSpec(
		SpecProductVariantSwatch,
		"Variant swatch image",
		"Small color or finish image used as a SKU option swatch.",
		ProductVariantSwatchRule,
		512,
		512,
		0,
		"Use a square 512x512 px image with the subject centered.",
	),
	SpecMediaLibraryImage: makeImageSpec(
		SpecMediaLibraryImage,
		"Media library image",
		"Reusable image uploaded to the general media library.",
		ProductImageRule,
		0,
		0,
		1600,
		"Use at least 1600 px on the long edge for reusable storefront media.",
	),
	SpecFAQAnswerImage: makeImageSpec(
		SpecFAQAnswerImage,
		"FAQ answer image",
		"Single image attached to an FAQ answer.",
		FAQAnswerImageRule,
		800,
		800,
		0,
		"Must be exactly 800x800 px WebP.",
	),
	SpecVisualShowcaseHomeCategories: makeAspectImageSpec(
		SpecVisualShowcaseHomeCategories,
		"Home visual showcase image",
		"Image used by the home product-category visual ugcshowcase.",
		ProductImageRule,
		1920,
		1080,
		16,
		9,
	),
	SpecVisualShowcaseEditorial: makeAspectImageSpec(
		SpecVisualShowcaseEditorial,
		"Editorial visual showcase image",
		"Portrait image used by the editorial visual ugcshowcase.",
		ProductImageRule,
		1200,
		1600,
		3,
		4,
	),
	SpecSiteLogo: {
		FileRule: SiteLogoImageRule,
		UploadSpec: UploadSpec{
			Code:                 string(SpecSiteLogo),
			Kind:                 "image",
			Label:                "Site logo",
			Description:          "The small site mark used by the storefront header and browser metadata.",
			AcceptedExtensions:   []string{".webp"},
			AcceptedContentTypes: []string{SiteLogoImageContentType},
			MaxFileSizeBytes:     SiteLogoImageRule.MaxSize,
			ExactWidth:           SiteLogoImageRule.ExactWidth,
			ExactHeight:          SiteLogoImageRule.ExactHeight,
			RecommendedWidth:     SiteLogoImageRule.ExactWidth,
			RecommendedHeight:    SiteLogoImageRule.ExactHeight,
			MaxWidth:             SiteLogoImageRule.MaxWidth,
			MaxHeight:            SiteLogoImageRule.MaxHeight,
			MaxPixels:            SiteLogoImageRule.MaxPixels,
			AspectRatioWidth:     SiteLogoImageRule.AspectRatioWidth,
			AspectRatioHeight:    SiteLogoImageRule.AspectRatioHeight,
			QualityNote:          "Use a 512x512 px square WebP; the storefront scales it to the header slot.",
		},
	},
	SpecSiteFavicon: makeImageSpec(
		SpecSiteFavicon,
		"Site favicon",
		"Browser tab, bookmark, and PWA icon.",
		FaviconImageRule,
		512,
		512,
		0,
		"Use a square 512x512 px PNG or WebP with transparent padding when needed.",
	),
	SpecCustomerServiceAvatar: makeImageSpec(
		SpecCustomerServiceAvatar,
		"Customer service avatar",
		"Avatar displayed beside a public customer-service agent.",
		CustomerServiceAvatarRule,
		512,
		512,
		0,
		"Use a square 512x512 px portrait with the face centered.",
	),
	SpecWebsiteProfileAvatar: makeImageSpec(
		SpecWebsiteProfileAvatar,
		"Website profile avatar",
		"Avatar used by the website profile content.",
		CustomerServiceAvatarRule,
		512,
		512,
		0,
		"Use a square 512x512 px portrait with the face centered.",
	),
	SpecWebsiteProfileImage: makeImageSpec(
		SpecWebsiteProfileImage,
		"Website profile image",
		"Factory, profile, or company image used by the website profile content.",
		ProductDescriptionImageRule,
		0,
		0,
		1600,
		"Use at least 1600 px on the long edge.",
	),
	SpecRefundCancellationImage: makeImageSpec(
		SpecRefundCancellationImage,
		"Refund & Cancellation Policy image",
		"Illustration or evidence image shown inside the Refund & Cancellation Policy.",
		ProductDescriptionImageRule,
		0,
		0,
		1600,
		"Use at least 1600 px on the long edge.",
	),
	SpecWarrantyEvidence: makeFilesImageSpec(
		SpecWarrantyEvidence,
		"Warranty evidence image",
		"Image evidence attached to a warranty or shipment record.",
		WarrantyImageRule,
		0,
		0,
		1200,
		"Use at least 1200 px on the long edge; preserve readable labels and damage details.",
	),
	SpecAfterSalesEvidence: makeFilesImageSpec(
		SpecAfterSalesEvidence,
		"After-sales evidence image",
		"Image evidence attached to a customer after-sales request.",
		WarrantyImageRule,
		0,
		0,
		1200,
		"Use at least 1200 px on the long edge; preserve readable labels and damage details.",
	),
	SpecSuggestionAttachment: makeImageSpec(
		SpecSuggestionAttachment,
		"Suggestion attachment",
		"Image attached to authenticated customer feedback.",
		SuggestionImageRule,
		0,
		0,
		1200,
		"Use at least 1200 px on the long edge when the image contains readable detail.",
	),
	SpecCustomerServiceAttachment: makeImageSpec(
		SpecCustomerServiceAttachment,
		"Customer service attachment",
		"Image attached to a customer-service conversation.",
		SuggestionImageRule,
		0,
		0,
		1200,
		"Use at least 1200 px on the long edge when the image contains readable detail.",
	),
	SpecUserShowcaseImage: makeFilesImageSpec(
		SpecUserShowcaseImage,
		"User showcase image",
		"WebP image submitted to the public picture warehouse.",
		ShowcaseImageRule,
		0,
		0,
		1600,
		"Use at least 1600 px on the long edge; only WebP is accepted.",
	),
}

func makeImageSpec(
	code SpecCode,
	label string,
	description string,
	rule FileRule,
	recommendedWidth int,
	recommendedHeight int,
	recommendedLongEdge int,
	qualityNote string,
) uploadSpecDefinition {
	return uploadSpecDefinition{
		UploadSpec: UploadSpec{
			Code:                 string(code),
			Kind:                 "image",
			Label:                label,
			Description:          description,
			AcceptedExtensions:   cloneStrings(rule.AllowedExtensions),
			AcceptedContentTypes: cloneStrings(rule.AllowedContentTypes),
			MaxFileSizeBytes:     rule.MaxSize,
			ExactWidth:           rule.ExactWidth,
			ExactHeight:          rule.ExactHeight,
			RecommendedWidth:     recommendedWidth,
			RecommendedHeight:    recommendedHeight,
			RecommendedLongEdge:  recommendedLongEdge,
			MaxWidth:             rule.MaxWidth,
			MaxHeight:            rule.MaxHeight,
			MaxPixels:            rule.MaxPixels,
			AspectRatioWidth:     rule.AspectRatioWidth,
			AspectRatioHeight:    rule.AspectRatioHeight,
			QualityNote:          qualityNote,
		},
		FileRule: rule,
	}
}

func makeAspectImageSpec(
	code SpecCode,
	label string,
	description string,
	rule FileRule,
	recommendedWidth int,
	recommendedHeight int,
	aspectRatioWidth int,
	aspectRatioHeight int,
) uploadSpecDefinition {
	definition := makeImageSpec(
		code,
		label,
		description,
		rule,
		recommendedWidth,
		recommendedHeight,
		0,
		"Keep the source aspect ratio at "+formatAspectRatio(aspectRatioWidth, aspectRatioHeight)+".",
	)
	definition.FileRule.AspectRatioWidth = aspectRatioWidth
	definition.FileRule.AspectRatioHeight = aspectRatioHeight
	definition.AspectRatioWidth = aspectRatioWidth
	definition.AspectRatioHeight = aspectRatioHeight
	definition.AspectRatioLabel = formatAspectRatio(aspectRatioWidth, aspectRatioHeight)
	return definition
}

func makeFilesImageSpec(
	code SpecCode,
	label string,
	description string,
	rule FilesRule,
	recommendedWidth int,
	recommendedHeight int,
	recommendedLongEdge int,
	qualityNote string,
) uploadSpecDefinition {
	definition := makeImageSpec(
		code,
		label,
		description,
		rule.FileRule,
		recommendedWidth,
		recommendedHeight,
		recommendedLongEdge,
		qualityNote,
	)
	definition.FilesRule = rule
	definition.HasFilesRule = true
	definition.UploadSpec.MaxFiles = rule.MaxFiles
	definition.UploadSpec.MaxTotalSizeBytes = rule.MaxTotalSize
	return definition
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

func formatAspectRatio(width, height int) string {
	return strconv.Itoa(width) + ":" + strconv.Itoa(height)
}

func normalizeSpecCode(value string) SpecCode {
	return SpecCode(strings.ToLower(strings.TrimSpace(value)))
}

// GetUploadSpec returns a copy of the public contract for one upload purpose.
func GetUploadSpec(code string) (UploadSpec, bool) {
	definition, ok := uploadSpecDefinitions[normalizeSpecCode(code)]
	if !ok {
		return UploadSpec{}, false
	}
	spec := definition.UploadSpec
	spec.AcceptedExtensions = cloneStrings(spec.AcceptedExtensions)
	spec.AcceptedContentTypes = cloneStrings(spec.AcceptedContentTypes)
	return spec, true
}

// ListUploadSpecs returns all upload contracts in stable code order.
func ListUploadSpecs() []UploadSpec {
	codes := make([]string, 0, len(uploadSpecDefinitions))
	for code := range uploadSpecDefinitions {
		codes = append(codes, string(code))
	}
	sort.Strings(codes)

	result := make([]UploadSpec, 0, len(codes))
	for _, code := range codes {
		spec, _ := GetUploadSpec(code)
		result = append(result, spec)
	}
	return result
}

// ValidateSpecFile applies the authoritative backend rule for one purpose.
func ValidateSpecFile(file *multipart.FileHeader, code string) error {
	definition, ok := uploadSpecDefinitions[normalizeSpecCode(code)]
	if !ok {
		return validationError(CodeInvalidType, "invalid_type: unknown upload specification %q", strings.TrimSpace(code))
	}
	return ValidateFile(file, definition.FileRule)
}

// ValidateSpecFiles applies a multi-file rule, including total request limits.
func ValidateSpecFiles(files []*multipart.FileHeader, code string) error {
	definition, ok := uploadSpecDefinitions[normalizeSpecCode(code)]
	if !ok {
		return validationError(CodeInvalidType, "invalid_type: unknown upload specification %q", strings.TrimSpace(code))
	}
	if definition.HasFilesRule {
		return ValidateFiles(files, definition.FilesRule)
	}
	for _, file := range files {
		if err := ValidateSpecFile(file, code); err != nil {
			return err
		}
	}
	return nil
}

// RuleForSpec returns the single-file rule used by an upload purpose.
func RuleForSpec(code string) (FileRule, bool) {
	definition, ok := uploadSpecDefinitions[normalizeSpecCode(code)]
	if !ok || definition.HasFilesRule {
		return FileRule{}, false
	}
	return definition.FileRule, true
}

// FilesRuleForSpec returns the multi-file rule used by an upload purpose.
func FilesRuleForSpec(code string) (FilesRule, bool) {
	definition, ok := uploadSpecDefinitions[normalizeSpecCode(code)]
	if !ok || !definition.HasFilesRule {
		return FilesRule{}, false
	}
	return definition.FilesRule, true
}
