package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerOperationsRoutes(
	authenticated *gin.RouterGroup,
	adminAccountHandler *AdminAccountHandler,
	opsOverviewHandler *OpsOverviewHandler,
	opsNetworkSummaryHandler *OpsNetworkSummaryHandler,
	outboxReconciliationHandler *OutboxReconciliationHandler,
	opsDeploymentPreflightHandler *OpsDeploymentPreflightHandler,
	opsDeploymentWorkflowHandler *OpsDeploymentWorkflowHandler,
	opsDomainBindingHandler *OpsDomainBindingHandler,
	opsConnectorHandler *OpsConnectorHandler,
	opsVPSBindingHandler *OpsVPSBindingHandler,
	opsProjectBindingHandler *OpsProjectBindingHandler,
	auditHandler *AuditHandler,
) {
	// 运维中心：当前维护域名、连接器、VPS 和项目的声明式台账。
	opsGroup := authenticated.Group("/ops")
	opsGroup.Use(middleware.RequirePermission(auth.PermOpsView))
	{
		opsGroup.GET("/overview", opsOverviewHandler.Get)
		opsGroup.GET("/network/summary", opsNetworkSummaryHandler.Get)
		opsGroup.GET("/outbox/unknown", outboxReconciliationHandler.ListUnknown)
		opsGroup.POST("/outbox/unknown/:id/resume", middleware.RequirePermission(auth.PermSystemManage), outboxReconciliationHandler.Resume)
		opsGroup.POST("/outbox/unknown/:id/mark-processed", middleware.RequirePermission(auth.PermSystemManage), outboxReconciliationHandler.MarkProcessed)
		opsGroup.GET("/admin-accounts", middleware.AdminOnly(), adminAccountHandler.List)
		opsGroup.POST("/admin-accounts/ensure", middleware.AdminOnly(), adminAccountHandler.Ensure)
		opsGroup.GET("/deployments/preflight-overview", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetOverview)
		opsGroup.GET("/deployments/preflight", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetProjectReportByQuery)
		opsGroup.GET("/workflows", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentWorkflowHandler.List)
		opsGroup.GET("/workflows/:id", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentWorkflowHandler.Get)
		opsGroup.POST("/workflows", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Create)
		opsGroup.POST("/workflows/:id/validate", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Validate)
		opsGroup.POST("/workflows/:id/approve", middleware.RequirePermission(auth.PermOpsWorkflowApprove), opsDeploymentWorkflowHandler.Approve)
		opsGroup.POST("/workflows/:id/execute", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Execute)
		opsGroup.POST("/workflows/:id/retry", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.RetryFailedStep)
		opsGroup.POST("/workflows/:id/rollback", middleware.RequirePermission(auth.PermOpsDeployRollback), opsDeploymentWorkflowHandler.Rollback)
		opsGroup.POST("/workflows/:id/cancel", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Cancel)

		domainsGroup := opsGroup.Group("/domains")
		domainsGroup.Use(middleware.RequirePermission(auth.PermOpsDomainView))
		{
			domainsGroup.GET("", opsDomainBindingHandler.List)
			domainsGroup.GET("/:id", opsDomainBindingHandler.Get)
			domainsGroup.GET("/:id/diff", opsDomainBindingHandler.Diff)
			domainsGroup.GET("/:id/preview", opsDomainBindingHandler.Preview)
			domainsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsDomainSync), opsDomainBindingHandler.Sync)
			domainsGroup.POST("", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.Create)
			domainsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.Update)
			domainsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.UpdateStatus)
		}

		connectorsGroup := opsGroup.Group("/connectors")
		connectorsGroup.Use(middleware.RequirePermission(auth.PermOpsConnectorView))
		{
			connectorsGroup.GET("", opsConnectorHandler.List)
			connectorsGroup.GET("/:id", opsConnectorHandler.Get)
			connectorsGroup.POST("", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Create)
			connectorsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Update)
			connectorsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.UpdateStatus)
			connectorsGroup.POST("/:id/test", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Test)
			connectorsGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.StartOAuth)
		}

		vpsGroup := opsGroup.Group("/vps")
		vpsGroup.Use(middleware.RequirePermission(auth.PermOpsVPSView))
		{
			vpsGroup.GET("", opsVPSBindingHandler.List)
			vpsGroup.GET("/:id", opsVPSBindingHandler.Get)
			vpsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsVPSSync), opsVPSBindingHandler.Sync)
			vpsGroup.POST("", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.Create)
			vpsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.Update)
			vpsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.UpdateStatus)
		}

		projectsGroup := opsGroup.Group("/projects")
		projectsGroup.Use(middleware.RequirePermission(auth.PermOpsProjectView))
		{
			projectsGroup.GET("", opsProjectBindingHandler.List)
			projectsGroup.GET("/:id/preflight", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetProjectReport)
			projectsGroup.GET("/:id", opsProjectBindingHandler.Get)
			projectsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsProjectSync), opsProjectBindingHandler.Sync)
			projectsGroup.POST("", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.Create)
			projectsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.Update)
			projectsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.UpdateStatus)
		}
	}

	// 审计日志（需要日志查看权限）
	logsGroup := authenticated.Group("/logs")
	logsGroup.Use(middleware.RequirePermission(auth.PermLogsView))
	{
		logsGroup.GET("", auditHandler.ListAuditLogs)
		logsGroup.GET("/stats", auditHandler.GetAuditStats)
		logsGroup.GET("/recent", auditHandler.GetRecentActivities)
		logsGroup.GET("/search", auditHandler.SearchAuditLogs)
		logsGroup.GET("/:id", auditHandler.GetAuditLog)
		logsGroup.GET("/user/:user_id", auditHandler.GetUserAuditLogs)
		logsGroup.POST("/cleanup", middleware.AdminOnly(), auditHandler.DeleteOldLogs)
	}
}
