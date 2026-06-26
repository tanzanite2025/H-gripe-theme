package response

import (
	"math"

	"github.com/gin-gonic/gin"
)

// Response 标准API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedResponse 分页响应结构
type PagedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Success 成功响应（200）
// code=0 表示成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code: 0,
		Data: data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Paged 分页响应
func Paged(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	c.JSON(200, PagedResponse{
		Code: 0,
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Created 创建成功响应（201）
func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Response{
		Code:    0,
		Message: "Created successfully",
		Data:    data,
	})
}

// NoContent 无内容响应（204）
// 常用于删除操作
func NoContent(c *gin.Context) {
	c.Status(204)
}
