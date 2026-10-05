package service

// YanwenPublishedCollectionReferenceReader is the only Yanwen dependency
// generic shipping operations may hold. It exposes production collection
// references for validation and public projection without exposing Yanwen
// management CRUD, catalog entities, credentials, or waybill operations.
type YanwenPublishedCollectionReferenceReader interface {
	ListProductionYanwenCollectionReferences() ([]YanwenPublishedCollectionReference, error)
	ListProductionYanwenCollectionReferencesIncludingDisabled() ([]YanwenPublishedCollectionReference, error)
}

var _ YanwenPublishedCollectionReferenceReader = (*YanwenPublishedCollectionService)(nil)
