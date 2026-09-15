# Spoke Calculator System / 辐条计算器系统手册

Last updated: 2026-08-29

Status: Active reference. Re-audit when the Go spoke API contract or frontend calculator data model changes.

## 1. 系统架构 (Architecture)

本模块采用 **后端黑盒计算 + 前端展示层** 架构。浏览器只获得品牌/型号标识和标签，所有 CAD 几何、公式与验证数据均留在 Go 服务端。

- **核心数据源**: Go `SpokeService` 与数据库；前端通过 `/api/v1/spoke/catalog/export` 获取脱敏目录。
- **计算接口**: `/api/v1/spoke/calc`，受 IP/用户令牌桶限流并记录 `spoke_histories`。
- **组成部分**:
  - `SpokeCalculatorBlueprint.vue`: 主计算器组件 (蓝图式布局，保留 Brand -> Model 级联选择)。
  - `SpokeSmartSearch.vue`: 智能搜索组件 (关键词模糊匹配)。

## 2. 数据管理与同步 (Data Management & Sync)

目录管理入口位于 Go 管理 API，公网接口只返回脱敏标识。

### 2.1 数据结构

- **RIM/HUB geometry**: 仅后端数据库保存，不进入 Nuxt bundle。
- **PRESET_BUILDS**: 后端管理，公网仅返回搜索所需的名称、关键词和 ID。

### 2.2 管理工作流 (Management Workflow)

管理员通过 `/api/admin/spoke-catalog` 维护目录；变更即时由 API 生效，无需生成或提交前端静态 CAD 数据。

## 3. 智能搜索 (Smart Search)

位于计算器下方的搜索栏组件 (`SpokeSmartSearch.vue`)。

- **逻辑**: 搜索脱敏 API 返回的预设元数据，不接触几何原始数据。

## 4. 相关文件索引

- **数据**: `app/data/spoke-calculator/database.ts`
- **计算逻辑**: `go-backend/internal/service/spoke_service.go`
- **Go 导出器**: `go-backend/internal/api/v1/spoke/handler.go`
