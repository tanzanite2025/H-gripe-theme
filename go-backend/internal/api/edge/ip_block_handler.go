package edge

import (
	"net/http"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ipBlockCheckPath = "/_internal/security/ip-block-check"

type IPBlockHandler struct {
	blockService *service.GlobalIPBlockService
}

func NewIPBlockHandler(blockService *service.GlobalIPBlockService) *IPBlockHandler {
	return &IPBlockHandler{blockService: blockService}
}

func RegisterRoutes(r *gin.Engine, blockService *service.GlobalIPBlockService) {
	if r == nil {
		return
	}
	NewIPBlockHandler(blockService).register(r)
}

func (h *IPBlockHandler) register(r *gin.Engine) {
	r.GET(ipBlockCheckPath, h.Check)
}

// Check is consumed by the internal Nginx auth_request subrequest. It keeps
// the response body empty because Nginx only needs the status code.
func (h *IPBlockHandler) Check(c *gin.Context) {
	if h == nil || h.blockService == nil {
		appLogger.Error("edge IP block check service is not configured")
		c.Status(http.StatusServiceUnavailable)
		return
	}

	blocked, _, err := h.blockService.IsBlocked(
		c.Request.Context(),
		c.ClientIP(),
		time.Now().UTC(),
	)
	if err != nil {
		appLogger.Error(
			"edge IP block check failed",
			zap.Error(err),
		)
		c.Status(http.StatusServiceUnavailable)
		return
	}
	if blocked {
		c.Header("Cache-Control", "no-store, max-age=0")
		c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
		c.Header("X-Access-Block", "ip")
		c.Status(http.StatusForbidden)
		return
	}

	c.Status(http.StatusNoContent)
}
