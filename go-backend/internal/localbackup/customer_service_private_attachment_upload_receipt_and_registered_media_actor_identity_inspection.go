package localbackup

import (
	"encoding/json"
	"strconv"

	"commerce-platform/internal/domain/media"
)

// Use the same identity validity rules as attachment access. Matching paths and
// account numbers alone do not establish the upload actor's namespace.
func decodeCustomerServiceRecoveryMediaUploadActorIdentity(row *moduleRecoveryDependencyRow, actorColumn, accountColumn, visitorColumn string) (media.MediaUploadActorIdentity, bool) {
	var actor, visitor string
	if row == nil || string(row.Values[actorColumn]) == "null" || string(row.Values[visitorColumn]) == "null" ||
		json.Unmarshal(row.Values[actorColumn], &actor) != nil || json.Unmarshal(row.Values[visitorColumn], &visitor) != nil {
		return media.MediaUploadActorIdentity{}, false
	}
	accountText := string(row.Values[accountColumn])
	account, err := strconv.ParseUint(accountText, 10, strconv.IntSize)
	if err != nil || strconv.FormatUint(account, 10) != accountText {
		return media.MediaUploadActorIdentity{}, false
	}
	identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorType(actor), AccountID: uint(account), VisitorSessionHash: visitor}
	return identity, identity.IsValidMediaUploadActorIdentity()
}

func (graph *moduleRecoveryRowDependencyGraph) requireCustomerServiceUploadActorIdentityColumns(table string, columns ...string) bool {
	for _, column := range columns {
		if !graph.Observation.Columns[table][column] {
			graph.markIncomplete("upload_actor_identity_schema_missing", table, "缺少附件上传主体核对所需字段，不能确认恢复后的私有附件可读取。")
			return false
		}
	}
	return true
}

func (graph *moduleRecoveryRowDependencyGraph) inspectCustomerServicePrivateAttachmentUploadReceiptAndRegisteredMediaActorIdentities() {
	if len(graph.Rows["media_upload_sessions"]) > 0 {
		sessionColumnsPresent := graph.requireCustomerServiceUploadActorIdentityColumns("media_upload_sessions", "upload_actor_type", "upload_actor_account_id", "upload_actor_visitor_session_hash", "media_asset_id")
		assetColumnsPresent := graph.requireCustomerServiceUploadActorIdentityColumns("media_assets", "uploader_actor_type", "uploader_id", "uploader_visitor_session_hash")
		for _, session := range graph.Rows["media_upload_sessions"] {
			identity := moduleRecoveryDependencyRowPublicIdentity(session, graph.Observation.PrimaryKeys["media_upload_sessions"])
			actor, valid := decodeCustomerServiceRecoveryMediaUploadActorIdentity(session, "upload_actor_type", "upload_actor_account_id", "upload_actor_visitor_session_hash")
			if sessionColumnsPresent && (!valid || actor.ActorType == media.MediaUploadActorLegacyAccount) {
				graph.Report.recordIssue("upload_actor_identity_invalid", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据主体不符合客户、客服账号或访客的完整身份规则。")
			}
			asset := findCustomerServiceDependencyRowByID(graph, "media_assets", session.Values["media_asset_id"])
			if asset == nil {
				graph.Report.recordIssue("upload_receipt_registered_asset_missing", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据缺少可核对的媒体登记，无法确认私有附件的读取身份。")
				continue
			}
			assetActor, assetValid := decodeCustomerServiceRecoveryMediaUploadActorIdentity(asset, "uploader_actor_type", "uploader_id", "uploader_visitor_session_hash")
			if sessionColumnsPresent && assetColumnsPresent && (!valid || !assetValid || actor != assetActor) {
				graph.Report.recordIssue("upload_receipt_asset_actor_identity_conflict", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据与媒体登记的完整上传主体无效或不一致，恢复后私有附件将无法通过读取校验。")
			}
		}
	}
	if len(graph.Rows["ticket_message_media_attachments"]) == 0 {
		return
	}
	if !graph.requireCustomerServiceUploadActorIdentityColumns("media_assets", "visibility") ||
		!graph.requireCustomerServiceUploadActorIdentityColumns("ticket_message_media_attachments", "upload_session_id") {
		return
	}
	for _, binding := range graph.Rows["ticket_message_media_attachments"] {
		asset := findCustomerServiceDependencyRowByID(graph, "media_assets", binding.Values["media_asset_id"])
		// Public reusable media does not require an upload receipt or a modern
		// upload actor registration. Do not infer or rewrite historical actors.
		if asset == nil || asset.referenceColumnText("visibility") != "private" {
			continue
		}
		if binding.referenceColumnText("upload_session_id") == "" || findCustomerServiceDependencyRowByID(graph, "media_upload_sessions", binding.Values["upload_session_id"]) == nil {
			graph.Report.recordIssue("private_attachment_upload_receipt_missing", "blocking", graph.Side, "ticket_message_media_attachments", moduleRecoveryDependencyRowPublicIdentity(binding, graph.Observation.PrimaryKeys["ticket_message_media_attachments"]), "私有附件绑定缺少上传收据，文件存在也不能通过消息附件读取校验。")
		}
	}
}
