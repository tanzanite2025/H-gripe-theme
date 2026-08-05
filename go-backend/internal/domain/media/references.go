package media

type ReferenceCategory string

const (
	ReferenceCategoryCatalog   ReferenceCategory = "catalog"
	ReferenceCategoryContent   ReferenceCategory = "content"
	ReferenceCategoryCommunity ReferenceCategory = "community"
	ReferenceCategoryCustomer  ReferenceCategory = "customer"
	ReferenceCategorySupport   ReferenceCategory = "support"
)

// AssetReference identifies one persisted record that currently uses a media
// asset. Category, ResourceType, and Field are stable machine-readable values
// used by admin views to group, preview, and later navigate to the source.
type AssetReference struct {
	Category         ReferenceCategory `json:"category"`
	ResourceType     string            `json:"resource_type"`
	ResourceID       uint              `json:"resource_id"`
	ParentResourceID uint              `json:"parent_resource_id,omitempty"`
	Label            string            `json:"label"`
	Field            string            `json:"field"`
}

type AssetReferenceReport struct {
	Total      int              `json:"total"`
	References []AssetReference `json:"references"`
}
