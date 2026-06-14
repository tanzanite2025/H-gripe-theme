# Tanzanite 架构演进与开发规范 (PHP -> Go/Vue3)

> **⚠️ 核心警告 (CRITICAL WARNING)**
> 本项目正在进行（或已完成）从传统 WordPress/PHP 架构向现代微服务架构的**全面迁移**。
> **绝对禁止**在 PHP 端（包括 `wp-plugin` 目录下的任何插件）中添加、修改或回调任何核心业务逻辑。
> 所有的业务逻辑开发必须且只能在 **Go 后端**中进行！

## 当前文档入口

先读 `docs/README.md` 和 `docs/PHP_TO_GO_MIGRATION_WORKFLOW.md`。未被文档入口列为“当前有效”的历史文档，只能作为旧行为参考，不能作为新实现依据。

## 🏛️ 系统架构全景图 (System Architecture)

本项目目前由三个独立解耦的核心子系统组成：

### 1. 客户前端 (Client Frontend)
- **技术栈**: Nuxt 3 (Vue 3 SSR)
- **职责**: 面向 C 端客户/访客的商城浏览、下单、会员中心等交互。
- **数据流向**: 通过 RESTful API 直接与 Go 后端通信。

### 2. 管理后台 (ERP Admin Panel)
- **当前主线目录**: `/go-backend/web/admin`
- **历史目录**: `/go-backend/admin-panel` 仅作旧 demo 参考，不再作为迁移目标。
- **技术栈**: Vue 3 + Vite + Element Plus
- **职责**: 面向 B 端内部运营团队，替代原有臃肿的 WordPress wp-admin 后台。
- **数据流向**: 通过 `/api/admin/*` 接口与 Go 后端通信，受 Admin 鉴权保护。

### 3. 核心接口与数据中枢 (Go Backend)
- **目录**: `/go-backend`
- **技术栈**: Go + Gin + GORM
- **职责**: 承载全站所有核心计算、业务逻辑校验（如积分、优惠券、订单流转、支付）、数据库读写操作。
- **重要说明**: 这是当前系统的**唯一真相来源 (Single Source of Truth)**。

---

## 🚫 PHP 历史包袱排雷指南 (Legacy PHP Deprecation)

根目录曾保留一组 WordPress 主题壳文件和主题元数据（`index.php`、`header.php`、`footer.php`、`page.php`、`single.php`、`functions.php`、`style.css`），这是早期虚拟机限制下的历史包袱；当前 C 端前台主线是 `nuxt-i18n`。
在早期的架构中，系统还高度依赖 WordPress 自定义插件（如 `tanzanite-setting`）。
**当前规范要求：**
- **冻结开发**：现存的 PHP 代码仅供**查阅逻辑参考**或**过渡期间的临时兼容**，禁止任何形式的新功能开发。
- **绞杀者模式 (Strangler Fig Pattern)**：当我们用 Go 后端完成某个模块（例如“积分系统”、“优惠券系统”）的重构并在前端对齐 API 后，必须**物理注释或删除**相应的 PHP 代码（例如 `class-plugin.php` 中引入的 `class-rewards-admin.php` 和旧控制器），以彻底切断 PHP 端的执行路径。
- **防回退机制**：一旦在 Go 端接管的模块，决不允许因为“临时排错方便”等原因重新切回 PHP 插件，以免造成数据踩踏和幽灵 Bug。
- **根目录 WordPress 壳已删除**：不要恢复根目录 WordPress 主题入口或 `style.css` 主题元数据；前台页面和布局只应在 `nuxt-i18n` 维护。

### 迁移推进规则

- 做完一个模块就停下来开 PR。
- 一个 PR 只做一个业务模块或一个基础设施模块。
- 不把 API 矩阵、前端切流、数据迁移、PHP 删除混在同一个 PR。
- 模块顺序和验收标准见 `docs/PHP_TO_GO_MIGRATION_WORKFLOW.md`。

---

> **致开发者**：
> 当您接手本系统的任何新需求时，请首先确保您了解上述三端分离架构。
> 如果遇到遗留的 WordPress/WooCommerce 功能，请思考如何将其抽取为 Go API 并在 Vue3 Admin 面板中提供管理能力，而不是在旧 PHP 泥潭中继续挣扎。
