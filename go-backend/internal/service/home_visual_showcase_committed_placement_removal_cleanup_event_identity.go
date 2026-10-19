package service

import (
	"commerce-platform/internal/domain/outbox"

	"github.com/google/uuid"
)

// Each committed detachment needs its own event: an earlier event may already
// have completed while another slot still retained the same object. Retrying
// delivery of this event keeps its identity; a later detachment gets a new one.
func newVisualShowcaseCommittedPlacementRemovalObjectCleanupEvent(showcaseScope string, objectKeys []string) (*outbox.Event, error) {
	return newObjectStorageCleanupEvent(
		objectCleanupResourceHomeVisualTile,
		showcaseScope+":removal:"+uuid.NewString(),
		outbox.AggregateTypeHomeVisualTileSet,
		showcaseScope,
		objectKeys,
	)
}
