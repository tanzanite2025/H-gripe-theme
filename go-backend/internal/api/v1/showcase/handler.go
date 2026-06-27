package showcase

import (
	"net/http"
	"strconv"
	"tanzanite/internal/pkg/upload"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const showcaseMaxRequestBytes = 55 << 20

type ShowcaseHandler struct {
	service *service.ShowcaseService
}

func NewShowcaseHandler(s *service.ShowcaseService) *ShowcaseHandler {
	return &ShowcaseHandler{service: s}
}

func (h *ShowcaseHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, showcaseMaxRequestBytes)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max memory
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form", "message": err.Error(), "code": "tpg_invalid_form"})
		return
	}

	form := c.Request.MultipartForm
	files := form.File["file[]"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided", "message": "No files provided", "code": "tpg_missing_files"})
		return
	}
	if err := upload.ValidateFiles(files, upload.ShowcaseImageRule); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{"error": err.Error(), "message": err.Error(), "code": upload.ErrorCode(err)})
		return
	}

	region := c.PostForm("region")
	if region == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region is required", "message": "missing_region: Region is required", "code": "tpg_missing_region"})
		return
	}

	params := map[string]string{
		"region":     region,
		"location":   c.PostForm("location"),
		"nickname":   c.PostForm("nickname"),
		"bike_model": c.PostForm("bike_model"),
		"notes":      c.PostForm("notes"),
	}

	userIDVal, exists := c.Get("user_id")
	var userID uint
	if exists {
		userID = userIDVal.(uint)
	} else {
		// 如果前端没带token，可以抛错或设为0 (Guest)。根据原有 PHP 逻辑，必须登录
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	item, err := h.service.UploadPhotos(c.Request.Context(), userID, files, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ShowcaseHandler) List(c *gin.Context) {
	kind := c.Query("type")
	if kind == "" {
		kind = "user"
	}
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	items, err := h.service.List(kind, status, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ShowcaseHandler) AddComment(c *gin.Context) {
	var req struct {
		PhotoID  uint   `json:"photo_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "tpg_empty_comment"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}
	userID := userIDVal.(uint)

	// Fetch username from context
	author := "User"
	if usernameVal, ok := c.Get("username"); ok {
		author = usernameVal.(string)
	}

	comment, err := h.service.AddComment(req.PhotoID, userID, author, req.Content, req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *ShowcaseHandler) ListComments(c *gin.Context) {
	photoIDStr := c.Query("photo_id")
	photoID, _ := strconv.Atoi(photoIDStr)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	comments, err := h.service.ListComments(uint(photoID), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为前端需要的结构 (PhotoComment)
	type PhotoCommentResp struct {
		ID               uint   `json:"id"`
		Author           string `json:"author"`
		Content          string `json:"content"`
		DateGMT          string `json:"date_gmt"`
		DateGMTFormatted string `json:"dateGmtFormatted"`
		Location         string `json:"location"`
	}

	var res []PhotoCommentResp
	for _, c := range comments {
		dateStr := c.CreatedAt.Format("2006-01-02T15:04:05")
		res = append(res, PhotoCommentResp{
			ID:               c.ID,
			Author:           c.Author,
			Content:          c.Content,
			DateGMT:          dateStr,
			DateGMTFormatted: c.CreatedAt.Format("Jan 2, 2006"),
			Location:         c.Location,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (h *ShowcaseHandler) Approve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.Approve(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Showcase approved"})
}

func (h *ShowcaseHandler) Reject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	if err := h.service.Reject(uint(id), req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Showcase rejected"})
}
