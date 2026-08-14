package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"commerce-platform/internal/app"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func registerLocalUploadsRoute(router *gin.Engine, deps *app.Dependencies) {
	storageConfig := storage.LoadConfigFromEnv()
	if storageConfig.Type != storage.StorageTypeLocal {
		return
	}

	uploadRoot, err := filepath.Abs(storageConfig.LocalPath)
	if err != nil {
		logger.Warn("failed to resolve local upload directory", zap.Error(err))
		return
	}

	if err := os.MkdirAll(uploadRoot, 0o750); err != nil {
		logger.Warn("failed to create local upload directory", zap.String("path", uploadRoot), zap.Error(err))
		return
	}

	fileServer := http.FileServer(restrictedUploadDirectory{root: uploadRoot})
	router.GET("/uploads/*key", func(c *gin.Context) {
		key, ok := sanitizePublicUploadKey(c.Param("key"))
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}

		if deps == nil || deps.Services.PublicUploadAccess == nil {
			c.Status(http.StatusNotFound)
			return
		}
		allowed, err := deps.Services.PublicUploadAccess.CanServePublicUpload(c.Request.Context(), key)
		if err != nil {
			logger.Warn("failed to authorize public upload", zap.String("key", key), zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !allowed {
			c.Status(http.StatusNotFound)
			return
		}

		c.Request.URL.Path = "/" + key
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

func sanitizePublicUploadKey(raw string) (string, bool) {
	return storage.NormalizeObjectKey(raw)
}

type restrictedUploadDirectory struct {
	root string
}

func (d restrictedUploadDirectory) Open(name string) (http.File, error) {
	key, ok := sanitizePublicUploadKey(name)
	if !ok {
		return nil, os.ErrNotExist
	}
	filePath := filepath.Join(d.root, filepath.FromSlash(key))
	cleanRoot, err := filepath.Abs(d.root)
	if err != nil {
		return nil, err
	}
	cleanPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return nil, err
	}
	parentPrefix := ".." + string(os.PathSeparator)
	if rel == ".." || strings.HasPrefix(rel, parentPrefix) || filepath.IsAbs(rel) {
		return nil, os.ErrNotExist
	}

	info, err := os.Lstat(cleanPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, os.ErrNotExist
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("upload object is not a regular file")
	}

	realRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil {
		return nil, err
	}
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return nil, err
	}
	realRel, err := filepath.Rel(realRoot, realPath)
	if err != nil {
		return nil, err
	}
	if realRel == ".." || strings.HasPrefix(realRel, parentPrefix) || filepath.IsAbs(realRel) {
		return nil, os.ErrNotExist
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
