package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *SettingsHandler) GetConfigurationBackupAndRecoveryOverview(c *gin.Context) {
	file, scopes, err := h.settingsService.ExportPortableSettingsConfiguration(c.Request.Context())
	if err != nil {
		writeSettingsConfigurationBackupError(c, err)
		return
	}
	configured := true
	for _, name := range []string{"LOCAL_BACKUP_SITE_ID", "LOCAL_BACKUP_ENVIRONMENT", "LOCAL_BACKUP_MAINTENANCE_DIRECTORY", "LOCAL_BACKUP_AUTHENTICATION_KEY_FILE"} {
		if os.Getenv(name) == "" {
			configured = false
		}
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"configuration_export_available": true, "configuration_count": len(file.Settings), "excluded_setting_count": file.ExcludedSettingCount, "scopes": scopes, "max_configuration_file_bytes": service.MaximumSettingsConfigurationFileBytes, "whole_storefront_maintenance_configured": configured})
}

func (h *SettingsHandler) DownloadPortableSettingsConfiguration(c *gin.Context) {
	file, _, err := h.settingsService.ExportPortableSettingsConfiguration(c.Request.Context())
	if err != nil {
		writeSettingsConfigurationBackupError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", `attachment; filename="storefront-configuration-`+time.Now().UTC().Format("20060102-150405")+`.json"`)
	c.JSON(http.StatusOK, file)
}

type settingsConfigurationImportHTTPRequest struct {
	Backup               service.SettingsConfigurationBackupFile `json:"backup"`
	ExpectedTargetSHA256 string                                  `json:"expected_target_sha256"`
	Confirmation         string                                  `json:"confirmation"`
}

func decodeBoundedSettingsConfigurationRequest(c *gin.Context, destination any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaximumSettingsConfigurationFileBytes+(64<<10))
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(destination)
	if err == nil {
		err = decoder.Decode(new(any))
		if err == io.EOF {
			return true
		}
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "配置文件超过 8 MiB 上限"})
		return false
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "配置文件格式无效，或包含未知字段"})
	return false
}

func (h *SettingsHandler) PreviewPortableSettingsConfiguration(c *gin.Context) {
	var request settingsConfigurationImportHTTPRequest
	if !decodeBoundedSettingsConfigurationRequest(c, &request) {
		return
	}
	preview, err := h.settingsService.PreviewPortableSettingsConfigurationImport(c.Request.Context(), request.Backup)
	if err != nil {
		writeSettingsConfigurationBackupError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, preview)
}

func (h *SettingsHandler) ImportPortableSettingsConfiguration(c *gin.Context) {
	var request settingsConfigurationImportHTTPRequest
	if !decodeBoundedSettingsConfigurationRequest(c, &request) {
		return
	}
	count, err := h.settingsService.ImportPortableSettingsConfigurationWithPreview(c.Request.Context(), request.Backup, request.ExpectedTargetSHA256, request.Confirmation)
	if err != nil {
		writeSettingsConfigurationBackupError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"count": count, "message": "配置已导入；文件中未包含的配置保持当前值"})
}

func writeSettingsConfigurationBackupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSettingsConfigurationPreviewChanged):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSettingInvalid), errors.Is(err, service.ErrSettingManagedByDomainService):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置备份操作失败，请稍后重试"})
	}
}
