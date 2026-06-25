package spoke

import (
	"math"
	"net/http"
	"strconv"
	domainspoke "tanzanite/internal/domain/spoke"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	spokeRepo *repository.SpokeRepository
}

func NewHandler(spokeRepo *repository.SpokeRepository) *Handler {
	return &Handler{spokeRepo: spokeRepo}
}

func (h *Handler) GetExport(c *gin.Context) {
	c.JSON(http.StatusOK, domainspoke.DefaultExport())
}

func (h *Handler) ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 5
	}

	items, total, err := h.spokeRepo.ListHistory(c.Query("search"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_history_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"meta": gin.H{
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"page":        page,
			"per_page":    pageSize,
		},
	})
}

type CalcRequest struct {
	RimID         string `json:"rimId" binding:"required"`
	HubID         string `json:"hubId" binding:"required"`
	WheelPosition string `json:"wheelPosition" binding:"required"`
	SpokeCount    int    `json:"spokeCount" binding:"required"`
	Crossing      int    `json:"crossing"`
}

func (h *Handler) Calculate(c *gin.Context) {
	var req CalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	export := domainspoke.DefaultExport()

	// 查找 Rim
	var rim *domainspoke.RimModel
	for _, brand := range export.Rims {
		for _, r := range brand.Items {
			if r.ID == req.RimID {
				temp := r
				rim = &temp
				break
			}
		}
		if rim != nil {
			break
		}
	}

	// 查找 Hub
	var hub *domainspoke.HubModel
	for _, brand := range export.Hubs {
		for _, hb := range brand.Items {
			if hb.ID == req.HubID {
				temp := hb
				hub = &temp
				break
			}
		}
		if hub != nil {
			break
		}
	}

	if rim == nil || hub == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Unknown rim or hub geometry"})
		return
	}

	var hubGeo *domainspoke.HubGeometry
	if req.WheelPosition == "front" {
		hubGeo = hub.Front
	} else {
		hubGeo = hub.Rear
	}

	if hubGeo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Hub geometry not available for requested position"})
		return
	}

	// 核心公式: L = sqrt( (ERD/2)^2 + (FlangePCD/2)^2 + FlangeToCenter^2 - 2 * (ERD/2) * (FlangePCD/2) * cos(Angle) )
	r := rim.ERD / 2.0
	fl := hubGeo.LeftFlangePCD / 2.0
	fr := hubGeo.RightFlangePCD / 2.0
	wl := hubGeo.LeftFlange
	wr := hubGeo.RightFlange

	// 将角度转换为弧度
	angleRad := (720.0 * float64(req.Crossing) / float64(req.SpokeCount)) * math.Pi / 180.0

	left := math.Sqrt(r*r + fl*fl + wl*wl - 2*r*fl*math.Cos(angleRad))
	right := math.Sqrt(r*r + fr*fr + wr*wr - 2*r*fr*math.Cos(angleRad))

	// 保留一位小数
	leftRounded := math.Round(left*10) / 10
	rightRounded := math.Round(right*10) / 10

	c.JSON(http.StatusOK, gin.H{
		"leftLengthMm":  leftRounded,
		"rightLengthMm": rightRounded,
		"debug": gin.H{
			"rim": rim,
			"hub": hubGeo,
			"formulaVersion": "v1.0-go-backend",
		},
	})
}
