# Tanzanite 架构演进与开发规范 (Archived PHP -> Go/Vue3)

> **⚠️ 核心警告 (CRITICAL WARNING)**
> 本项目正在进行（或已完成）从传统 WordPress/PHP 架构向现代微服务架构的**全面迁移**。
> **绝对禁止**在 PHP 端（包括 `functions.php` 和 `wp-plugin` 目录下的任何插件）中添加、修改或回调任何核心业务逻辑。
> 所有的业务逻辑开发必须且只能在 **Go 后端**中进行！

## 🏛️ 系统架构全景图 (System Architecture)

本项目目前由三个独立解耦的核心子系统组成：

### 1. 客户前端 (Client Frontend)
- **技术栈**: Nuxt 3 (Vue 3 SSR)
- **职责**: 面向 C 端客户/访客的商城浏览、下单、会员中心等交互。
- **数据流向**: 通过 RESTful API 直接与 Go 后端通信。

### 2. 管理后台 (ERP Admin Panel)
- **目录**: `/go-backend/admin-panel`
- **技术栈**: Vue 3 (SPA) + 原生 CSS (遵循 ERP 工业视觉系统规范 v1.0)
- **职责**: 面向 B 端内部运营团队，替代原有臃肿的 WordPress wp-admin 后台。
- **数据流向**: 通过 `/api/v1/admin/*` 接口与 Go 后端通信，受 Admin 鉴权保护。

### 3. 核心接口与数据中枢 (Go Backend)
- **目录**: `/go-backend`
- **技术栈**: Go + Gin + GORM
- **职责**: 承载全站所有核心计算、业务逻辑校验（如积分、优惠券、订单流转、支付）、数据库读写操作。
- **重要说明**: 这是当前系统的**唯一真相来源 (Single Source of Truth)**。

---

## 🚫 PHP 历史包袱排雷指南 (Legacy PHP Deprecation)

在早期的架构中，系统高度依赖 WordPress 的 `functions.php` 以及自定义插件（如 `tanzanite-setting`）。
**当前规范要求：**
- **冻结开发**：现存的 PHP 代码仅供**查阅逻辑参考**或**过渡期间的临时兼容**，禁止任何形式的新功能开发。
- **绞杀者模式 (Strangler Fig Pattern)**：当我们用 Go 后端完成某个模块（例如“积分系统”、“优惠券系统”）的重构并在前端对齐 API 后，必须**物理注释或删除**相应的 PHP 代码（例如 `class-plugin.php` 中引入的 `class-rewards-admin.php` 和旧控制器），以彻底切断 PHP 端的执行路径。
- **防回退机制**：一旦在 Go 端接管的模块，决不允许因为“临时排错方便”等原因重新切回 PHP 插件，以免造成数据踩踏和幽灵 Bug。

### 📝 已完成割接的模块清单
*(更新于 2026-06)*
- [x] **会员积分系统 (Loyalty & Points)**：已在 Go 端 `marketing_service` 实现，PHP 端 `tanzanite-setting` 插件中对应接口及面板已被物理屏蔽。
- [x] **优惠券系统 (Gift Cards & Coupons)**：已在 Go 端及 Vue3 管理面板完成重构对接，PHP 端的旧接口和菜单已废除。

---

> **致开发者**：
> 当您接手本系统的任何新需求时，请首先确保您了解上述三端分离架构。
> 如果遇到遗留的 WordPress/WooCommerce 功能，请思考如何将其抽取为 Go API 并在 Vue3 Admin 面板中提供管理能力，而不是在旧 PHP 泥潭中继续挣扎。
