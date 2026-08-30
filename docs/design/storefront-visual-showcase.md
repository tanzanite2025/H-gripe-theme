# Storefront Visual Showcase 实施文档

## 目标

把首页首屏右侧从单张大图升级为可配置的视觉目录：

- 桌面端占首页右侧约 70% 宽度，使用 9 张 3:4 图片组成自由 collage，不固定三列。
- 移动端使用 8 张图片，4 组双图切换，每次切换两张，顶部保留分页点。
- 每张图片底部显示白色文字条，文字使用真实的 `figcaption`，同时维护 `altText`。
- 图片、标题、备注和排序不再硬编码在 `HomeHero.vue`；当前 storefront 不再消费单独的布局/分组编辑值。
- 后台提供独立的 `Visual Showcase` 管理入口，直接上传到专用对象目录并整组保存。
- 展示图不创建媒体库记录；成功替换并保存后，不再引用的旧对象会被直接删除。
- 后端接口无数据或不可用时，首页使用前端 fallback，不让首屏空白。

## 非目标

本次不做以下内容：

- 不把文字烧录进图片作为 SEO 方案。
- 不在首页组件中直接写后台请求、数据模型和图片数组。
- 不做拖拽排序、定时发布和复杂版本审批。
- 不修改现有通用品牌图库的业务含义。

## 文件边界

### Storefront

| 文件 | 责任 |
| --- | --- |
| `nuxt-i18n/app/types/homeHeroVisualShowcase.ts` | 首页视觉目录类型和 API envelope |
| `nuxt-i18n/app/data/homeHeroVisualShowcaseFallback.ts` | 无后台数据时的 9 张 fallback 配置 |
| `nuxt-i18n/app/composables/useHomeHeroVisualShowcase.ts` | SSR 可用的数据读取、locale 处理和 fallback |
| `nuxt-i18n/app/components/home/HomeHeroVisualShowcaseFigure.vue` | 单图、alt、figcaption 和图片加载 |
| `nuxt-i18n/app/components/home/HomeHeroVisualShowcaseDesktopCollage.vue` | 桌面自由 collage |
| `nuxt-i18n/app/components/home/HomeHeroVisualShowcaseMobilePairGallery.vue` | 移动端 8 图、4 组双图和分页点 |
| `nuxt-i18n/app/components/home/HomeHero.vue` | 只负责首屏结构、文案和 CTA 组合 |

### Backend

| 文件 | 责任 |
| --- | --- |
| `go-backend/internal/domain/visualshowcase/visual_showcase.go` | 独立的视觉目录条目模型 |
| `go-backend/internal/repository/visual_showcase_repository.go` | 查询和整组替换 |
| `go-backend/internal/service/visual_showcase_service.go` | 校验、专用对象生命周期和发布过滤 |
| `go-backend/internal/api/v1/visualshowcase/handler.go` | 公共首页读取接口 |
| `go-backend/internal/api/v1/visualshowcase/public_response.go` | 公共响应映射 |
| `go-backend/internal/api/admin/visual_showcase_handler.go` | 后台读取和保存接口 |
| `go-backend/migrations/185_visual_showcase_items.up.sql` | 生产环境表结构 |
| `go-backend/migrations/185_visual_showcase_items.down.sql` | 回滚表结构 |

### Admin

| 文件 | 责任 |
| --- | --- |
| `go-backend/web/admin/src/api/visualShowcase.ts` | 后台 API client |
| `go-backend/web/admin/src/components/admin/visual-showcase/visualShowcaseTypes.ts` | 后台表单类型 |
| `go-backend/web/admin/src/components/admin/visual-showcase/visualShowcaseFormState.ts` | 9 图默认行、API 映射、上传映射和保存校验 |
| `go-backend/web/admin/src/components/admin/visual-showcase/VisualShowcaseItemEditorRow.vue` | 单条图片编辑行 |
| `go-backend/web/admin/src/views/VisualShowcase.vue` | Visual Showcase 页面、专用上传和整组保存 |

## 数据契约

每条视觉目录 item 包含：

```text
showcase_key       固定为 home-hero
locale             en、zh_cn 等站点语种
storage_key        visual-showcase 专用对象键
title              白色条主标题
caption            白色条补充说明
alt_text           图片 alt 属性
desktop_order      桌面排序
mobile_pair_index  兼容字段，当前 storefront 不再用于编辑控制
target_url         可选跳转地址
target_label       可选跳转文本
layout_variant     兼容字段，当前 storefront 使用固定样式
is_published       是否发布
published_from     可选开始时间
published_until    可选结束时间
```

## 前台行为

1. `useHomeHeroVisualShowcase` 在 SSR 阶段读取公开接口。
2. 返回有效条目时，按照 `desktop_order` 渲染桌面端，前 8 条按顺序两两渲染移动端。
3. 接口失败、没有条目或条目不完整时，使用 `homeHeroVisualShowcaseFallback.ts`。
4. 所有图片使用 `HomeHeroVisualShowcaseFigure`，图片比例固定为 `3 / 4`。
5. `figcaption` 使用可见的标题和备注；`alt` 使用 `alt_text`。

## 后台行为

1. 进入 `网站内容 -> 首页视觉目录`。
2. 选择语种，加载 `home-hero` 当前配置。
3. 为每条 item 直接上传 3:4 图片；上传结果只写入 `visual-showcase/<showcase>/<locale>/`。
4. 编辑标题、备注、alt、桌面顺序、发布状态和跳转文案。
5. 点击保存时使用整组替换接口，避免前台读到半组新旧混合数据。
6. 保存成功后，数据库中不再引用的旧 `visual-showcase` 对象会被删除；该流程不会写入媒体库。

## 验收标准

- [x] 桌面右侧 gallery 不出现 `max-width` 或 `margin-left: auto` 导致的左侧空洞。
- [x] 桌面图片宽高比保持 3:4，列数不是固定三列。
- [x] 移动端一次显示两张图片，分页点切换双图。
- [x] 图片底部存在可见的白色文案条。
- [x] 首屏 HTML 中存在 `figcaption` 和非空 `alt`。
- [x] 后台无数据时首页仍显示 fallback。
- [x] 后端 `go test ./...`、admin `npm run typecheck`、storefront `npm run build` 通过。
