package publicuploadresponse

import (
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PublicUploadAccessVerifier interface {
	CanServePublicUpload(context.Context, string) (bool, error)
}

// NewRegisteredLocalUploadHTTPHandler shares business authorization and cache policies with stopped recovery verification.
func NewRegisteredLocalUploadHTTPHandler(access PublicUploadAccessVerifier, uploadFileServer, siteLogoFileServer http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization failures must never inherit an immutable/public cache.
		c.Header("Cache-Control", "private, no-store")
		key, ok := sanitizePublicUploadKey(c.Param("key"))
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		if storage.IsPrivateObjectKey(key) {
			c.Status(http.StatusNotFound)
			return
		}

		if access == nil {
			c.Status(http.StatusNotFound)
			return
		}
		allowed, err := access.CanServePublicUpload(c.Request.Context(), key)
		if err != nil {
			logger.Warn("failed to authorize public upload", zap.String("key", key), zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !allowed {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "")
		if namespace := storage.DedicatedPublicObjectNamespaceForStorageKey(key); namespace == storage.SuggestionFeedbackAttachmentNamespace || namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
			c.Header("Cache-Control", "private, no-store")
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("Referrer-Policy", "no-referrer")
		}
		if service.IsCustomerServiceAvatarStorageKey(key) {
			// Avatar URLs are immutable because every replacement receives a new
			// object key, so browser and CDN caches can retain them safely.
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		if service.IsMediaDerivativeStorageKey(key) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}

		fileServer := uploadFileServer
		if service.IsSiteLogoStorageKey(key) {
			fileServer = siteLogoFileServer
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		if service.IsSiteFaviconStorageKey(key) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		if fileServer == nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Request.URL.Path = "/" + key
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// NewRestrictedLocalUploadFileServer reads existing roots without creating directories.
func NewRestrictedLocalUploadFileServer(root string) http.Handler {
	return http.FileServer(restrictedLocalUploadDirectory{root: root})
}
func sanitizePublicUploadKey(raw string) (string, bool) { return storage.NormalizeObjectKey(raw) }

type restrictedLocalUploadDirectory struct{ root string }

func (d restrictedLocalUploadDirectory) Open(name string) (http.File, error) {
	return OpenRestrictedLocalUploadFile(d.root, name)
}
func OpenRestrictedLocalUploadFile(root, name string) (http.File, error) {
	key, ok := sanitizePublicUploadKey(name)
	if !ok {
		return nil, os.ErrNotExist
	}
	filePath := filepath.Join(root, filepath.FromSlash(key))
	cleanRoot, err := filepath.Abs(root)
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
