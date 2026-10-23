package main

import (
	"net/http"
	"os"
	"path/filepath"

	"commerce-platform/internal/api/publicuploadresponse"
	"commerce-platform/internal/app"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func registerLocalUploadsRoute(router *gin.Engine, deps *app.Dependencies) {
	storageConfig := storage.LoadConfigFromEnv()
	siteLogoStorageConfig := storage.LoadSiteLogoConfigFromEnv(storageConfig)

	uploadFileServer, uploadFileServerOK := localUploadFileServer(storageConfig, "local upload directory")
	siteLogoFileServer, siteLogoFileServerOK := localUploadFileServer(siteLogoStorageConfig, "local site logo upload directory")
	if !uploadFileServerOK && !siteLogoFileServerOK {
		return
	}

	var access publicuploadresponse.PublicUploadAccessVerifier
	if deps != nil {
		access = deps.Services.PublicUploadAccess
	}
	router.GET("/uploads/*key", publicuploadresponse.NewRegisteredLocalUploadHTTPHandler(access, uploadFileServer, siteLogoFileServer))
}

func localUploadFileServer(storageConfig *storage.Config, label string) (http.Handler, bool) {
	if storageConfig == nil || storageConfig.Type != storage.StorageTypeLocal {
		return nil, false
	}

	uploadRoot, err := filepath.Abs(storageConfig.LocalPath)
	if err != nil {
		logger.Warn("failed to resolve "+label, zap.Error(err))
		return nil, false
	}

	if err := os.MkdirAll(uploadRoot, 0o750); err != nil {
		logger.Warn("failed to create "+label, zap.String("path", uploadRoot), zap.Error(err))
		return nil, false
	}

	return http.FileServer(restrictedUploadDirectory{root: uploadRoot}), true
}

func sanitizePublicUploadKey(raw string) (string, bool) {
	return storage.NormalizeObjectKey(raw)
}

type restrictedUploadDirectory struct {
	root string
}

func (d restrictedUploadDirectory) Open(name string) (http.File, error) {
	return publicuploadresponse.OpenRestrictedLocalUploadFile(d.root, name)
}
