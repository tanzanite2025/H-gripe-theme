# 博客多语言功能 - 快速开始

**状态**: ✅ 实施完成，待测试  
**版本**: 1.4.0  
**日期**: 2026-05-25

---

## 🚀 快速开始（3步）

### 1. 启动开发环境

```powershell
# Windows
.\start-dev.ps1

# 或手动启动
# 终端 1: cd go-backend && go run cmd/server/main.go
# 终端 2: cd nuxt-i18n && npm run dev
```

### 2. 访问测试

- 前端: http://localhost:3000
- 后端: http://localhost:9000
- API 测试: http://localhost:9000/api/v1/i18n/languages

### 3. 查看效果

访问任意博客文章页面，查看底部的翻译链接。

---

## 📁 重要文件

### 文档（按阅读顺序）

1. **INTEGRATION_FINAL_SUMMARY.md** ⭐ - 完整总结（从这里开始）
2. **I18N_QUICK_REFERENCE.md** - 快速参考卡片
3. **DATA_MIGRATION_AND_FRONTEND_INTEGRATION.md** - 数据迁移指南
4. **FRONTEND_INTEGRATION_COMPLETE.md** - 前端集成说明

### 代码

**后端**:
- `go-backend/internal/api/v1/i18n/handler.go` - i18n API
- `go-backend/internal/service/sitemap_service.go` - Sitemap 生成

**前端**:
- `nuxt-i18n/composables/useI18n.ts` - i18n 组合式函数
- `nuxt-i18n/app/components/PostTranslations.vue` - 翻译链接组件
- `nuxt-i18n/app/components/LanguageSwitcher.vue` - 语言切换器

**工具**:
- `go-backend/scripts/wordpress-export/export-blog-translations.php` - WordPress 导出
- `go-backend/cmd/import/blog_translations.go` - Go 导入工具

---

## 🎯 功能概览

### 后端 API（7个端点）

```bash
# 语言管理
GET  /api/v1/i18n/languages          # 获取34种语言列表
GET  /api/v1/i18n/detect             # 检测用户语言
POST /api/v1/i18n/set-language       # 设置用户语言

# 翻译查询
GET  /api/v1/i18n/translations/:id   # 获取文章翻译

# Sitemap
GET  /sitemap.xml                    # Sitemap 索引
GET  /sitemap-hreflang.xml           # Hreflang Sitemap
GET  /sitemap-:locale.xml            # 单语言 Sitemap
```

### 前端组件

```vue
<!-- 使用 i18n 功能 -->
<script setup>
const { locale, getPostTranslations } = useI18n()
</script>

<!-- 显示翻译链接 -->
<PostTranslations :post-id="123" />

<!-- 语言切换器 -->
<LanguageSwitcher display-mode="dropdown" />
```

---

## 📊 完成情况

| 模块 | 状态 | 文件数 | 代码行数 |
|------|------|--------|----------|
| 后端实现 | ✅ | 8 | ~780 |
| 数据迁移工具 | ✅ | 3 | ~600 |
| 前端集成 | ✅ | 5 | ~565 |
| 文档 | ✅ | 12 | ~2,750 |
| **总计** | **✅** | **28** | **~4,695** |

---

## 🧪 测试

### 快速测试

```powershell
# 测试所有 API 端点
.\go-backend\test-i18n-api.ps1
```

### 手动测试

```bash
# 1. 语言列表
curl http://localhost:9000/api/v1/i18n/languages | jq

# 2. 文章翻译
curl http://localhost:9000/api/v1/i18n/translations/1 | jq

# 3. Sitemap
curl http://localhost:9000/sitemap-hreflang.xml
```

---

## 🔄 数据迁移（可选）

如果需要从 WordPress 迁移数据：

```bash
# 1. WordPress 导出
php go-backend/scripts/wordpress-export/export-blog-translations.php

# 2. 数据库迁移
psql -U tanzanite -d tanzanite -f go-backend/migrations/002_add_post_translation_fields.sql

# 3. 试运行导入
go run go-backend/cmd/import/blog_translations.go --dry-run

# 4. 正式导入
go run go-backend/cmd/import/blog_translations.go
```

详细步骤见: **DATA_MIGRATION_AND_FRONTEND_INTEGRATION.md**

---

## 📚 支持的语言（34种）

**欧洲**: en, fr, de, es, it, pt, ru, nl, pl, tr, sv, no, da, fi, cs, hu, ro  
**亚洲**: zh, zh-TW, ja, ko, vi, th, id, ms, hi, bn, ta, te, mr, ur  
**中东**: ar, fa, he

---

## 🐛 故障排除

### API 连接失败

**检查**:
1. Go 后端是否运行？
2. 端口 9000 是否被占用？
3. `.env.local` 配置是否正确？

**解决**:
```bash
# 检查端口
netstat -ano | findstr :9000

# 查看 Go 后端日志
cd go-backend
go run cmd/server/main.go
```

### 翻译链接不显示

**检查**:
1. 文章是否有 `id` 字段？
2. 文章是否有翻译版本？
3. API 是否返回数据？

**调试**:
```vue
<script setup>
const translations = await getPostTranslations(postId)
console.log('Translations:', translations)
</script>
```

---

## 📞 获取帮助

### 文档
- **完整总结**: INTEGRATION_FINAL_SUMMARY.md
- **快速参考**: I18N_QUICK_REFERENCE.md
- **迁移指南**: DATA_MIGRATION_AND_FRONTEND_INTEGRATION.md

### 常见问题
- 查看各文档的"常见问题"部分
- 查看"故障排除"部分

---

## ⏭️ 下一步

### 立即可做
- [ ] 启动开发环境测试
- [ ] 验证 API 端点
- [ ] 查看前端效果

### 可选工作
- [ ] 执行数据迁移
- [ ] 调整组件样式
- [ ] 添加更多翻译

### 后续计划（Week 1, Day 3-5）
- [ ] tanzanite-setting 增强
- [ ] 集成测试
- [ ] 文档完善

---

## 🎉 总结

✅ **28 个文件** - 完整实现  
✅ **~4,695 行** - 代码和文档  
✅ **34 种语言** - 全球化支持  
✅ **7 个 API** - RESTful 设计  
✅ **3 个组件** - 即插即用  

**状态**: 准备就绪，可以开始测试！

---

**快速链接**:
- [完整总结](INTEGRATION_FINAL_SUMMARY.md)
- [快速参考](go-backend/I18N_QUICK_REFERENCE.md)
- [API 文档](go-backend/API.md)
- [前端指南](FRONTEND_INTEGRATION_COMPLETE.md)
