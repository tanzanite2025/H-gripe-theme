package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerSiteQualityRoutes(group *gin.RouterGroup, handler *SiteQualityHandler, managePermission auth.Permission) {
	group.GET("", handler.ListSiteQualityRuns)
	group.GET("/targets", handler.ListSiteQualityTargets)
	group.POST("/jobs", middleware.RequirePermission(managePermission), handler.CreateSiteQualityJob)
	group.POST("/jobs/cleanup", middleware.RequirePermission(managePermission), handler.CleanupSiteQualityJobs)
	group.GET("/jobs/:id", handler.GetSiteQualityJob)
	group.GET("/findings", handler.ListSiteQualityFindings)
	group.GET("/findings/:id", handler.GetSiteQualityFinding)
	group.GET("/findings/:id/events", handler.ListSiteQualityFindingEvents)
	group.POST("/findings/:id/acknowledge", middleware.RequirePermission(managePermission), handler.AcknowledgeSiteQualityFinding)
	group.POST("/findings/:id/resolve", middleware.RequirePermission(managePermission), handler.ResolveSiteQualityFinding)
	group.POST("/findings/:id/recheck", middleware.RequirePermission(managePermission), handler.RecheckSiteQualityFinding)
}

func registerMediaImageDimensionRoutes(group *gin.RouterGroup, handler *MediaImageDimensionsHandler) {
	group.GET("", handler.List)
	group.POST("/:id/reconcile", middleware.RequirePermission(auth.PermMediaEdit), handler.Reconcile)
}

func registerPreflightContentLinkRoutes(group *gin.RouterGroup, handler *ContentLinkPreflightHandler, managePermission auth.Permission) {
	group.GET("/targets", handler.ListTargets)
	group.POST("/runs", middleware.RequirePermission(managePermission), handler.Run)
	group.GET("/issues", handler.ListIssues)
	group.GET("/issues/:id", handler.GetIssue)
	group.GET("/issues/:id/events", handler.ListIssueEvents)
	group.GET("/stats", handler.Stats)
	group.POST("/issues/:id/apply", middleware.RequirePermission(managePermission), handler.ApplySuggestion)
	group.POST("/issues/:id/resolve", middleware.RequirePermission(managePermission), handler.Resolve)
	group.POST("/issues/:id/recheck", middleware.RequirePermission(managePermission), handler.Recheck)
}
