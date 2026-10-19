package service

import "commerce-platform/internal/domain/media"

func (s *MediaService) projectHistoricalTicketAttachmentConversationAccess(asset *media.MediaAsset) error {
	if asset == nil || !media.IsUnclassifiedHistoricalMediaAssetWithUnknownUploadProvenance(asset) {
		return nil
	}
	if s == nil || s.repo == nil {
		asset.RequiresConversationAccess = true
		return ErrMediaStorageUnavailable
	}
	required, err := s.repo.HistoricalUnclassifiedMediaRequiresConversationAccess(asset)
	// A failed reference read must not turn a potentially private attachment
	// into a public file or expose its delivery fields in the library.
	asset.RequiresConversationAccess = required || err != nil
	return err
}
