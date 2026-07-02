package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Tanzanite E-commerce API
// @version 1.0
// @description 完整的电商平台API，包含用户管理、产品管理、订单处理、支付集成等功能
// @termsOfService https://tanzanite.com/terms

// @contact.name API Support
// @contact.url https://tanzanite.com/support
// @contact.email support@tanzanite.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host api.tanzanite.com
// @BasePath /api/v1

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name auth_token
// @description HttpOnly Cookie authentication. Browser clients must send credentials.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API密钥认证

// @tag.name Auth
// @tag.description 用户认证和授权相关接口

// @tag.name Users
// @tag.description 用户管理相关接口

// @tag.name Products
// @tag.description 产品管理相关接口

// @tag.name Orders
// @tag.description 订单管理相关接口

// @tag.name Cart
// @tag.description 购物车相关接口

// @tag.name Payment
// @tag.description 支付相关接口

// @tag.name Reviews
// @tag.description 评论和评分相关接口

// @tag.name Chat
// @tag.description 实时聊天相关接口

// @tag.name Health
// @tag.description 健康检查和系统状态

// setupSwagger 配置Swagger文档路由
func setupSwagger(r *gin.Engine) {
	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 重定向根路径到Swagger UI
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0",
	Host:        "api.tanzanite.com",
	BasePath:    "/api/v1",
	Schemes:     []string{"https", "http"},
	Title:       "Tanzanite E-commerce API",
	Description: "完整的电商平台API文档",
}
