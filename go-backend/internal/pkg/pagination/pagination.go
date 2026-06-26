package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 全局分页配置常量
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	DefaultLimit    = 10
	MaxPageSize     = 100
	MaxLimit        = 500
)

// Params 分页参数
type Params struct {
	Page     int
	PageSize int
}

// ParsePagination 从 gin.Context 解析分页参数
// 自动限制最大值和默认值
func ParsePagination(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))

	// 限制有效范围
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// ParseLimit 解析 limit 参数（用于简单列表，不需要分页）
func ParseLimit(c *gin.Context) int {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))

	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return limit
}

// Offset 计算数据库查询的 offset
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 返回 pageSize（用于数据库查询）
func (p Params) Limit() int {
	return p.PageSize
}
