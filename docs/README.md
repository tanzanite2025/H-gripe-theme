# Tanzanite 文档入口

本目录是当前迁移工作的有效入口。根目录、`go-backend/`、`nuxt-i18n/`、`wp-plugin/` 下仍保留大量历史文档；未在本文件列为“当前有效”的文档，默认只作为历史参考。

## 当前有效文档

- `README.md`：仓库级架构和 PHP 冻结规则。
- `docs/PHP_TO_GO_MIGRATION_WORKFLOW.md`：PHP → Go 迁移执行顺序、单模块 PR 规则、下一步路线。
- `docs/PHP_TO_GO_API_MATRIX.md`：PHP/WP API → Go API 迁移矩阵与下一批单模块 PR 建议。
- `go-backend/README.md`：Go API 服务说明。
- `go-backend/QUICK_START.md`：Go API 本地/Docker 启动说明。
- `go-backend/DEPLOYMENT.md`：部署参考；实际发布前仍需按目标环境复核。
- `go-backend/web/admin/README.md`：当前管理后台说明。

## 历史参考文档

以下文档可能包含仍有价值的业务说明，但不再作为实现依据：

- `wp-plugin/tanzanite-setting/**`：WordPress 插件时代的管理后台/API 文档，只能用于理解旧行为和迁移映射。
- `go-backend/admin-panel/**`：早期 Vue/Vite demo 后台，当前主线后台是 `go-backend/web/admin`。
- `go-backend/PROJECT_COMPLETE.md`、`go-backend/FRONTEND_PAGES_COMPLETE.md`、`go-backend/CODE_QUALITY_REPORT.md` 等完成度报告：只能作为当时状态记录，不能替代当前代码盘点。
- `nuxt-i18n/*-PLAN.md`、`nuxt-i18n/*-ANALYSIS.md`：按模块参考，实施前必须重新核对实际代码和 API。

## 文档维护规则

1. 任何迁移 PR 只能覆盖一个明确模块。
2. 每个模块 PR 必须在描述中声明范围、非范围、验证命令、是否触碰 PHP。
3. 不允许因为顺手修复而把多个业务模块塞进一个 PR。
4. 如果发现文档与代码冲突，以代码为准，并在当前文档入口记录冲突和下一步。
