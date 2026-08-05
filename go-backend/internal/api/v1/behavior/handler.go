package behavior

import (
	"net/http"
	"time"

	"tanzanite/internal/pkg/apierror"
	attributionpkg "tanzanite/internal/pkg/attribution"
	"tanzanite/internal/pkg/securecookie"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	behaviorEventService *service.BehaviorEventService
	attributionSigner    *attributionpkg.Signer
	cookieOptions        securecookie.Options
}

func NewHandler(behaviorEventService *service.BehaviorEventService) *Handler {
	return &Handler{behaviorEventService: behaviorEventService}
}

func (h *Handler) ConfigureAttribution(signer *attributionpkg.Signer, cookieOptions securecookie.Options) {
	if h == nil {
		return
	}
	h.attributionSigner = signer
	h.cookieOptions = cookieOptions
}

type batchRequest struct {
	Events []eventRequest `json:"events" binding:"required"`
}

type eventRequest struct {
	EventID     string         `json:"event_id"`
	EventType   string         `json:"event_type"`
	AnonymousID string         `json:"anonymous_id"`
	SessionID   string         `json:"session_id"`
	ProductID   *uint          `json:"product_id"`
	CategoryID  *uint          `json:"category_id"`
	Locale      string         `json:"locale"`
	Path        string         `json:"path"`
	Referrer    string         `json:"referrer"`
	Metadata    map[string]any `json:"metadata"`
	OccurredAt  time.Time      `json:"occurred_at"`
}

func (h *Handler) IngestBatch(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)

	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	inputs := make([]service.BehaviorEventInput, 0, len(req.Events))
	for _, item := range req.Events {
		inputs = append(inputs, service.BehaviorEventInput{
			EventID:     item.EventID,
			EventType:   item.EventType,
			AnonymousID: item.AnonymousID,
			SessionID:   item.SessionID,
			ProductID:   item.ProductID,
			CategoryID:  item.CategoryID,
			Locale:      item.Locale,
			Path:        item.Path,
			Referrer:    item.Referrer,
			Metadata:    item.Metadata,
			OccurredAt:  item.OccurredAt,
		})
	}

	var userID *uint
	if value, exists := c.Get("user_id"); exists {
		if id, ok := value.(uint); ok && id > 0 {
			userID = &id
		}
	}

	result, err := h.behaviorEventService.Ingest(userID, inputs)
	if err != nil {
		if service.IsBehaviorEventValidationError(err) {
			apierror.RespondValidationError(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	h.persistAttributionCookie(c, inputs)

	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"message": "Behavior events accepted",
		"data":    result,
	})
}

func (h *Handler) persistAttributionCookie(c *gin.Context, inputs []service.BehaviorEventInput) {
	if h == nil || h.attributionSigner == nil {
		return
	}
	context, ok := latestAttributionContext(inputs)
	if !ok {
		return
	}
	token, err := h.attributionSigner.Encode(context)
	if err != nil {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     attributionpkg.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cookieOptions.Domain,
		MaxAge:   attributionpkg.CookieMaxAge,
		Secure:   h.cookieOptions.Secure,
		HttpOnly: true,
		SameSite: h.cookieOptions.SameSite,
	})
}

func latestAttributionContext(inputs []service.BehaviorEventInput) (attributionpkg.Context, bool) {
	var latest attributionpkg.Context
	var latestAt time.Time
	for _, input := range inputs {
		if input.EventType != "ad_landing" {
			continue
		}
		context, ok := attributionpkg.FromMetadata(input.Metadata)
		if !ok {
			continue
		}
		if latestAt.IsZero() || input.OccurredAt.After(latestAt) {
			latest = context
			latestAt = input.OccurredAt
		}
	}
	return latest, !latestAt.IsZero()
}
