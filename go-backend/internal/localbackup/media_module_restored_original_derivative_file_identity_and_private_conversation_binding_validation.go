package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

// Runs inside the same transaction as inserts, both in rehearsal and formally.
// This reads current owners; it never changes messages, product facts or roles.
func (cfg Configuration) validateMediaModuleRestoredBusinessBindings(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, manifest *Manifest) error {
	files := map[string]FileIdentity{}
	for _, entry := range manifest.Files {
		files[entry.Path] = entry
	}
	current := map[string][]json.RawMessage{}
	for _, table := range strings.Fields("media_assets media_asset_derivatives media_upload_sessions ticket_message_media_attachments product_media product_variants product_variant_option_values product_spec_option_items") {
		rows, err := readProductCatalogRows(ctx, tx, table, "true")
		if err != nil {
			return err
		}
		if len(rows) > businessModuleSnapshotMaximumRows {
			return errors.New("媒体业务关联检查超过容量限制")
		}
		current[table] = rows
	}
	assets := map[string]json.RawMessage{}
	for _, row := range current["media_assets"] {
		assets[readMediaModuleRestoreRowText(row, "id")] = row
	}
	registeredFiles := map[string]string{}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		for _, row := range current[table] {
			key := readMediaModuleRestoreRowText(row, "storage_key")
			if key == "" {
				_, key = cfg.resolveLocalMediaReference(readMediaModuleRestoreRowText(row, "url"))
			}
			visibility := readMediaModuleRestoreRowText(row, "visibility")
			if table == "media_asset_derivatives" {
				visibility = readMediaModuleRestoreRowText(assets[readMediaModuleRestoreRowText(row, "media_asset_id")], "visibility")
			}
			path := "media/" + registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility) + "/" + key
			if _, restored := files[path]; !restored {
				continue
			}
			if registeredFiles[path] != "" {
				return errors.New("恢复文件存在多个媒体对象登记，不能确认唯一读取身份")
			}
			registeredFiles[path] = table + ":" + readMediaModuleRestoreRowText(row, "id")
		}
	}
	for _, table := range []string{"media_assets", "media_asset_derivatives", "site_logo_assets", "site_favicon_assets"} {
		for _, row := range snapshot.Tables[table].Rows {
			field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
			if field("deleted_at") != "" {
				continue
			}
			root, key := cfg.resolveLocalMediaReference(field("url"))
			if normalized, valid := storage.NormalizeObjectKey(field("storage_key")); !valid || normalized != key || normalized != field("storage_key") || root == "" {
				return errors.New("媒体 URL、存储键和本地文件身份不一致")
			}
			visibility := field("visibility")
			if table == "media_asset_derivatives" {
				parent := assets[field("media_asset_id")]
				if len(parent) == 0 || readMediaModuleRestoreRowText(parent, "deleted_at") != "" || readMediaModuleRestoreRowText(parent, "status") != "active" {
					return errors.New("派生文件缺少有效原件")
				}
				visibility = readMediaModuleRestoreRowText(parent, "visibility")
				if strings.HasPrefix(key, "media-derivatives/") && !strings.HasPrefix(key, "media-derivatives/"+field("media_asset_id")+"/") {
					return errors.New("派生文件命名空间与原件不一致")
				}
				version := field("preset_version")
				if version == "" || version == "0" {
					version = "1"
				}
				if visibility == "private" && !strings.HasPrefix(key, storage.CustomerServiceConversationAttachmentDerivativesPrefix+"/"+field("media_asset_id")+"/"+field("preset")+"/v"+version+"/") {
					return errors.New("私有派生文件不属于原附件命名空间")
				}
			}
			if table == "media_assets" || table == "media_asset_derivatives" {
				// Local storage also uses /uploads/<private namespace> URLs;
				// public delivery rejects these keys. Visibility selects the disk.
				if visibility == "private" && root == "uploads" && storage.IsPrivateObjectKey(key) {
					root = "uploads_private"
				}
				if (visibility != "public" && visibility != "private") || root != registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility) {
					return errors.New("媒体可见性与实际存储卷不一致")
				}
				if storage.IsPrivateObjectKey(key) != (visibility == "private") {
					return errors.New("私有文件命名空间与媒体可见性不一致")
				}
			}
			identity, exists := files["media/"+root+"/"+key]
			if !exists {
				return errors.New("原件或派生文件不在认证清单中")
			}
			if size, err := strconv.ParseInt(field("size"), 10, 64); err == nil && size > 0 && size != identity.Size {
				return errors.New("媒体登记大小与实际文件不一致")
			}
			if digest := field("content_sha256"); digest != "" && digest != identity.SHA256 {
				return errors.New("媒体登记摘要与文件字节不一致")
			}
		}
	}
	// Check all bindings in the resulting database: missing metadata can also
	// change the meaning of a retained target-only business reference.
	sessions := map[string]json.RawMessage{}
	for _, row := range current["media_upload_sessions"] {
		field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
		asset := assets[field("media_asset_id")]
		var values map[string]json.RawMessage
		if json.Unmarshal(row, &values) != nil {
			return errors.New("上传身份无法解析")
		}
		actor, valid := decodeCustomerServiceRecoveryMediaUploadActorIdentity(&moduleRecoveryDependencyRow{Values: values}, "upload_actor_type", "upload_actor_account_id", "upload_actor_visitor_session_hash")
		if !valid || actor.ActorType == media.MediaUploadActorLegacyAccount || field("status") != "bound" || field("business_context_type") != "customer_service" || len(asset) == 0 {
			return errors.New("附件上传收据未完成或身份无效")
		}
		if json.Unmarshal(asset, &values) != nil {
			return errors.New("媒体上传主体无法解析")
		}
		assetActor, valid := decodeCustomerServiceRecoveryMediaUploadActorIdentity(&moduleRecoveryDependencyRow{Values: values}, "uploader_actor_type", "uploader_id", "uploader_visitor_session_hash")
		if !valid || assetActor != actor || field("object_key") != readMediaModuleRestoreRowText(asset, "storage_key") || readMediaModuleRestoreRowText(asset, "visibility") != "private" || !strings.HasPrefix(field("object_key"), storage.CustomerServiceConversationAttachmentOriginalsPrefix+"/") {
			return errors.New("私有原件、上传主体与收据不一致")
		}
		var validOwner bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.tickets t WHERE t.id::text=$1 AND t.category='customer_service' AND (($2='visitor' AND t.visitor_session_hash=$3) OR ($2='customer' AND coalesce(nullif(t.customer_user_id,0),t.user_id)::text=$4) OR ($2='administrator' AND EXISTS(SELECT 1 FROM public.users u WHERE u.id::text=$4 AND u.role IN ('admin','manager','support')))))`, field("business_context_id"), string(actor.ActorType), actor.VisitorSessionHash, strconv.FormatUint(uint64(actor.AccountID), 10)).Scan(&validOwner); err != nil {
			return err
		}
		if !validOwner {
			return errors.New("私有上传主体与当前会话不一致")
		}
		sessions[field("id")] = row
	}
	boundSessions := map[string]bool{}
	for _, row := range current["ticket_message_media_attachments"] {
		field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
		asset := assets[field("media_asset_id")]
		if len(asset) == 0 || readMediaModuleRestoreRowText(asset, "deleted_at") != "" || readMediaModuleRestoreRowText(asset, "status") != "active" {
			return errors.New("消息附件缺少有效媒体原件")
		}
		session := sessions[field("upload_session_id")]
		if readMediaModuleRestoreRowText(asset, "visibility") == "private" && len(session) == 0 {
			return errors.New("私有消息附件缺少原上传收据")
		}
		var validMessage bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.ticket_messages m JOIN public.tickets t ON t.id=m.ticket_id WHERE m.id::text=$1 AND m.attachment_reference_format='media_asset_bindings' AND t.category='customer_service' AND ($2='' OR m.ticket_id::text=$2))`, field("ticket_message_id"), readMediaModuleRestoreRowText(session, "business_context_id")).Scan(&validMessage); err != nil {
			return err
		}
		if !validMessage || len(session) > 0 && readMediaModuleRestoreRowText(session, "media_asset_id") != field("media_asset_id") {
			return errors.New("消息、会话、附件及收据不属于同一业务对象")
		}
		if len(session) > 0 {
			boundSessions[field("upload_session_id")] = true
		}
	}
	for id := range sessions {
		if !boundSessions[id] {
			return errors.New("已绑定上传收据缺少实际消息关联")
		}
	}
	for _, row := range snapshot.Tables["media_assets"].Rows {
		if readMediaModuleRestoreRowText(row, "visibility") != "private" || readMediaModuleRestoreRowText(row, "deleted_at") != "" {
			continue
		}
		found := false
		for _, session := range sessions {
			if readMediaModuleRestoreRowText(session, "media_asset_id") == readMediaModuleRestoreRowText(row, "id") {
				found = true
				break
			}
		}
		if !found {
			return errors.New("私有原件缺少可核对的已绑定上传收据")
		}
	}
	privateOwners := map[string]string{}
	for _, session := range sessions {
		privateOwners[readMediaModuleRestoreRowText(session, "object_key")] = readMediaModuleRestoreRowText(session, "business_context_id")
	}
	legacyMessages, err := readProductCatalogRows(ctx, tx, "ticket_messages", `attachment_reference_format='legacy_urls' AND coalesce(attachments,'') NOT IN ('','[]')`)
	if err != nil {
		return err
	}
	if len(legacyMessages) > businessModuleSnapshotMaximumRows {
		return errors.New("历史消息附件检查超过容量限制")
	}
	for _, message := range legacyMessages {
		var references []string
		if json.Unmarshal([]byte(readMediaModuleRestoreRowText(message, "attachments")), &references) != nil {
			return errors.New("历史消息附件无法解析，不能确认恢复媒体后的访问归属")
		}
		for _, reference := range references {
			_, key := cfg.resolveLocalMediaReference(reference)
			if owner := privateOwners[key]; owner != "" && owner != readMediaModuleRestoreRowText(message, "ticket_id") {
				return errors.New("保留的历史消息引用了其他会话的私有附件")
			}
		}
	}
	for _, row := range current["product_media"] {
		asset := assets[readMediaModuleRestoreRowText(row, "media_asset_id")]
		if len(asset) == 0 {
			continue
		}
		decoded := media.MediaAsset{StorageKey: readMediaModuleRestoreRowText(asset, "storage_key"), URL: readMediaModuleRestoreRowText(asset, "url"), Status: readMediaModuleRestoreRowText(asset, "status"), Visibility: readMediaModuleRestoreRowText(asset, "visibility"), LibraryPartition: media.MediaAssetLibraryPartition(readMediaModuleRestoreRowText(asset, "library_partition")), UploaderActorType: media.MediaUploadActorType(readMediaModuleRestoreRowText(asset, "uploader_actor_type")), UploadOrigin: media.MediaAssetUploadOrigin(readMediaModuleRestoreRowText(asset, "upload_origin")), CustomerServiceMediaPurpose: readMediaModuleRestoreRowText(asset, "customer_service_media_purpose")}
		if readMediaModuleRestoreRowText(asset, "deleted_at") != "" || !media.CanMediaAssetBeReusedByPublicBusinessContent(&decoded) {
			return errors.New("商品图片引用了无法公开复用的媒体")
		}
	}
	placement := productCatalogModuleSnapshot{Tables: map[string]businessModuleCompleteTableSnapshot{}}
	for table, rows := range current {
		placement.Tables[table] = businessModuleCompleteTableSnapshot{Rows: rows}
	}
	if err := cfg.validateProductCatalogCompleteMediaPlacementOwnershipAndAssetBindings(placement); err != nil {
		return errors.New("商品媒体归属、规格或原件 URL 关联不一致")
	}
	for _, row := range snapshot.Tables["business_draft_media_upload_receipts"].Rows {
		field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
		asset := assets[field("asset_id")]
		if len(asset) == 0 || field("status") != "bound" || field("original_object_key") != readMediaModuleRestoreRowText(asset, "storage_key") || field("uploader_id") != readMediaModuleRestoreRowText(asset, "uploader_id") {
			return errors.New("编辑器上传收据与绑定原件身份不一致")
		}
		var fields map[string]json.RawMessage
		var keys []string
		if json.Unmarshal(row, &fields) != nil || json.Unmarshal(fields["object_keys_json"], &keys) != nil {
			return errors.New("编辑器上传文件清单无效")
		}
		for _, key := range keys {
			if _, exists := files["media/uploads/"+key]; !exists {
				return errors.New("编辑器上传收据缺少实际文件")
			}
		}
		identity := files["media/uploads/"+field("original_object_key")]
		if field("content_sha256") != identity.SHA256 || field("size") != strconv.FormatInt(identity.Size, 10) {
			return errors.New("编辑器收据字节摘要与原件不一致")
		}
	}
	return nil
}
