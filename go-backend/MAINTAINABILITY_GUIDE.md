# 博客多语言模块 - 可维护性指南

**创建时间**: 2026-05-25  
**目标**: 确保代码长期可维护，避免混乱

---

## 📋 当前架构分析

### ✅ 做得好的地方

#### 1. 清晰的分层架构
```
Domain (模型) → Repository (数据访问) → Service (业务逻辑) → Handler (API)
```

**优点**:
- 职责分离明确
- 易于测试
- 易于替换实现

#### 2. 统一的命名规范
```go
// Repository 层
FindByTranslationGroup()
FindPublished()

// Service 层
GetTranslations()
GetPublishedPosts()

// Handler 层
GetLanguages()
GetPostTranslations()
```

**优点**:
- 命名一致
- 易于理解
- 易于查找

#### 3. 完整的文档
- API 文档
- 使用指南
- 故障排除
- 代码注释

---

## ⚠️ 潜在的可维护性问题

### 问题 1: 命名冲突风险

**当前情况**:
```typescript
// Nuxt 项目中
import { useI18n } from '#imports'  // @nuxtjs/i18n
import { useI18n } from '~/composables/useI18n'  // 我们的
```

**风险**: 容易混淆，导致使用错误的函数

**解决方案**: 重命名我们的组合式函数

```typescript
// 方案 A: 重命名文件
composables/useBlogI18n.ts

// 方案 B: 重命名导出
export const useBlogI18n = () => { ... }

// 方案 C: 使用命名空间
export const BlogI18n = {
  useI18n: () => { ... }
}
```

**推荐**: 方案 A - 重命名文件为 `useBlogI18n.ts`

---

### 问题 2: 硬编码的语言列表

**当前情况**:
```typescript
// handler.go 中
var SupportedLanguages = []Language{
    {Code: "en", Name: "English", ...},
    {Code: "zh", Name: "Chinese", ...},
    // ... 32 more
}
```

**风险**:
- 添加新语言需要修改代码
- 语言列表分散在多个地方
- 难以动态管理

**解决方案**: 配置化语言列表

```go
// 1. 创建配置文件
// config/languages.yaml
languages:
  - code: en
    name: English
    native_name: English
    enabled: true
  - code: zh
    name: Chinese (Simplified)
    native_name: 简体中文
    enabled: true

// 2. 加载配置
type LanguageConfig struct {
    Languages []Language `yaml:"languages"`
}

func LoadLanguages() ([]Language, error) {
    data, _ := ioutil.ReadFile("config/languages.yaml")
    var config LanguageConfig
    yaml.Unmarshal(data, &config)
    return config.Languages, nil
}

// 3. 使用配置
type Handler struct {
    languages []Language
}

func NewHandler(...) *Handler {
    languages, _ := LoadLanguages()
    return &Handler{languages: languages}
}
```

---

### 问题 3: API 响应格式不统一

**当前情况**:
```go
// 有些返回
c.JSON(200, gin.H{"languages": languages})

// 有些返回
c.JSON(200, gin.H{"data": data, "total": total})

// 有些返回
c.JSON(200, gin.H{"message": "success"})
```

**风险**: 前端需要处理多种响应格式

**解决方案**: 统一响应格式

```go
// 1. 定义标准响应结构
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total      int64 `json:"total,omitempty"`
    Page       int   `json:"page,omitempty"`
    PageSize   int   `json:"page_size,omitempty"`
    TotalPages int   `json:"total_pages,omitempty"`
}

// 2. 创建辅助函数
func SuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(200, Response{
        Success: true,
        Data:    data,
    })
}

func ErrorResponse(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Success: false,
        Error:   message,
    })
}

// 3. 使用统一格式
func (h *Handler) GetLanguages(c *gin.Context) {
    languages := h.languages
    SuccessResponse(c, gin.H{
        "languages": languages,
        "total":     len(languages),
    })
}
```

---

### 问题 4: 缺少错误处理中间件

**当前情况**:
```go
func (h *Handler) GetPostTranslations(c *gin.Context) {
    postID, err := strconv.ParseUint(postIDStr, 10, 32)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid post ID"})
        return
    }
    // ... 更多错误处理
}
```

**风险**: 错误处理代码重复，不一致

**解决方案**: 统一错误处理

```go
// 1. 定义错误类型
type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

// 2. 创建错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            if appErr, ok := err.(*AppError); ok {
                ErrorResponse(c, appErr.Code, appErr.Message)
                return
            }
            
            ErrorResponse(c, 500, "Internal server error")
        }
    }
}

// 3. 使用错误处理
func (h *Handler) GetPostTranslations(c *gin.Context) {
    postID, err := strconv.ParseUint(postIDStr, 10, 32)
    if err != nil {
        c.Error(&AppError{
            Code:    400,
            Message: "Invalid post ID",
            Err:     err,
        })
        return
    }
    // ...
}
```

---

### 问题 5: 缺少版本控制

**当前情况**:
```
/api/v1/i18n/languages
```

**风险**: API 变更时难以向后兼容

**解决方案**: 明确版本策略

```go
// 1. 在响应中包含版本信息
type Response struct {
    Version string      `json:"version"`
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
}

// 2. 支持多版本共存
// /api/v1/i18n/languages - 旧版本
// /api/v2/i18n/languages - 新版本

// 3. 使用 API 版本中间件
func APIVersion(version string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("api_version", version)
        c.Next()
    }
}
```

---

### 问题 6: 缺少日志记录

**当前情况**: 没有系统的日志记录

**风险**: 难以调试和监控

**解决方案**: 添加结构化日志

```go
// 1. 使用日志库（如 zap）
import "go.uber.org/zap"

type Handler struct {
    logger *zap.Logger
}

// 2. 记录关键操作
func (h *Handler) GetPostTranslations(c *gin.Context) {
    h.logger.Info("Getting post translations",
        zap.String("post_id", postIDStr),
        zap.String("user_ip", c.ClientIP()),
    )
    
    translations, err := h.postService.GetTranslations(postID)
    if err != nil {
        h.logger.Error("Failed to get translations",
            zap.Error(err),
            zap.Uint("post_id", postID),
        )
        return
    }
    
    h.logger.Info("Successfully retrieved translations",
        zap.Int("count", len(translations)),
    )
}
```

---

### 问题 7: 缺少性能监控

**当前情况**: 没有性能指标收集

**风险**: 难以发现性能瓶颈

**解决方案**: 添加性能监控

```go
// 1. 添加性能监控中间件
func PerformanceMonitor() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // 记录慢请求
        if duration > 100*time.Millisecond {
            log.Printf("Slow request: %s %s took %v",
                c.Request.Method,
                c.Request.URL.Path,
                duration,
            )
        }
    }
}

// 2. 添加 Prometheus 指标
import "github.com/prometheus/client_golang/prometheus"

var (
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests.",
        },
        []string{"path", "method"},
    )
)
```

---

## 🔧 改进建议

### 优先级 1: 立即改进（高优先级）

#### 1.1 重命名 useI18n 避免冲突

```bash
# 重命名文件
mv nuxt-i18n/composables/useI18n.ts nuxt-i18n/composables/useBlogI18n.ts

# 更新导入
# 在所有使用的地方
import { useBlogI18n } from '~/composables/useBlogI18n'
```

#### 1.2 统一 API 响应格式

```go
// 创建 pkg/response/response.go
package response

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

// 在所有 Handler 中使用
```

#### 1.3 添加错误处理中间件

```go
// 在 router.go 中添加
r.Use(middleware.ErrorHandler())
```

---

### 优先级 2: 短期改进（中优先级）

#### 2.1 配置化语言列表

```yaml
# config/languages.yaml
languages:
  - code: en
    name: English
    native_name: English
    enabled: true
    flag_emoji: 🇬🇧
```

#### 2.2 添加日志记录

```go
// 使用 zap 或 logrus
logger, _ := zap.NewProduction()
defer logger.Sync()
```

#### 2.3 添加缓存层

```go
// 缓存 Sitemap
func (s *SitemapService) GenerateHreflangSitemap() (string, error) {
    cacheKey := "sitemap:hreflang"
    
    // 尝试从缓存获取
    if cached, err := s.cache.Get(cacheKey); err == nil {
        return cached, nil
    }
    
    // 生成 Sitemap
    sitemap := s.generate()
    
    // 缓存 1 小时
    s.cache.Set(cacheKey, sitemap, 1*time.Hour)
    
    return sitemap, nil
}
```

---

### 优先级 3: 长期改进（低优先级）

#### 3.1 添加单元测试

```go
// internal/service/sitemap_service_test.go
func TestGenerateHreflangSitemap(t *testing.T) {
    // 创建测试数据
    posts := []post.Post{
        {ID: 1, Locale: "en", TranslationGroupID: ptrUint(1)},
        {ID: 2, Locale: "zh", TranslationGroupID: ptrUint(1)},
    }
    
    // 创建 mock repository
    mockRepo := &MockPostRepository{posts: posts}
    
    // 创建 service
    service := NewSitemapService(mockRepo, "https://test.com")
    
    // 测试
    sitemap, err := service.GenerateHreflangSitemap()
    
    assert.NoError(t, err)
    assert.Contains(t, sitemap, "hreflang=\"en\"")
    assert.Contains(t, sitemap, "hreflang=\"zh\"")
}
```

#### 3.2 添加集成测试

```go
// tests/integration/i18n_test.go
func TestI18nAPI(t *testing.T) {
    // 启动测试服务器
    router := setupRouter()
    
    // 测试语言列表
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/i18n/languages", nil)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotNil(t, response["languages"])
}
```

#### 3.3 添加性能测试

```go
// tests/benchmark/sitemap_bench_test.go
func BenchmarkGenerateHreflangSitemap(b *testing.B) {
    service := setupSitemapService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.GenerateHreflangSitemap()
    }
}
```

---

## 📁 推荐的目录结构

### 当前结构
```
go-backend/
├── internal/
│   ├── domain/post/
│   ├── repository/
│   ├── service/
│   └── api/v1/i18n/
```

### 改进后的结构
```
go-backend/
├── internal/
│   ├── domain/
│   │   └── post/
│   │       ├── model.go
│   │       └── repository.go (接口定义)
│   ├── repository/
│   │   └── post_repository.go (实现)
│   ├── service/
│   │   ├── post_service.go
│   │   └── sitemap_service.go
│   ├── api/
│   │   └── v1/
│   │       └── i18n/
│   │           ├── handler.go
│   │           ├── request.go (请求结构)
│   │           └── response.go (响应结构)
│   └── pkg/
│       ├── response/ (统一响应)
│       ├── errors/ (错误处理)
│       └── logger/ (日志)
├── config/
│   ├── config.yaml
│   └── languages.yaml (新增)
└── tests/
    ├── unit/
    ├── integration/
    └── benchmark/
```

---

## 📝 代码规范

### 1. 命名规范

```go
// ✅ 好的命名
func GetPostTranslations(postID uint) ([]Post, error)
func FindByTranslationGroup(groupID uint) ([]Post, error)

// ❌ 不好的命名
func GetTrans(id uint) ([]Post, error)
func Find(id uint) ([]Post, error)
```

### 2. 注释规范

```go
// ✅ 好的注释
// GetPostTranslations 获取文章的所有翻译版本
// 参数:
//   - postID: 文章ID
// 返回:
//   - []Post: 翻译版本列表
//   - error: 错误信息
func GetPostTranslations(postID uint) ([]Post, error)

// ❌ 不好的注释
// get translations
func GetPostTranslations(postID uint) ([]Post, error)
```

### 3. 错误处理规范

```go
// ✅ 好的错误处理
translations, err := h.postService.GetTranslations(postID)
if err != nil {
    h.logger.Error("Failed to get translations", zap.Error(err))
    return nil, fmt.Errorf("get translations: %w", err)
}

// ❌ 不好的错误处理
translations, _ := h.postService.GetTranslations(postID)
```

---

## 🔍 代码审查清单

### 新增代码审查

- [ ] 命名是否清晰？
- [ ] 是否有足够的注释？
- [ ] 是否有错误处理？
- [ ] 是否有日志记录？
- [ ] 是否有单元测试？
- [ ] 是否符合现有架构？
- [ ] 是否有性能问题？
- [ ] 是否有安全问题？

### 修改代码审查

- [ ] 是否影响现有功能？
- [ ] 是否需要更新文档？
- [ ] 是否需要更新测试？
- [ ] 是否向后兼容？

---

## 📚 文档维护

### 文档结构

```
docs/
├── README.md (总览)
├── ARCHITECTURE.md (架构设计)
├── API.md (API 文档)
├── DEVELOPMENT.md (开发指南)
├── DEPLOYMENT.md (部署指南)
└── MAINTAINABILITY.md (本文档)
```

### 文档更新规则

1. **新增功能**: 更新 API.md 和 README.md
2. **架构变更**: 更新 ARCHITECTURE.md
3. **部署变更**: 更新 DEPLOYMENT.md
4. **开发流程变更**: 更新 DEVELOPMENT.md

---

## 🎯 总结

### 立即执行（本周）

1. ✅ 重命名 `useI18n` 为 `useBlogI18n`
2. ✅ 统一 API 响应格式
3. ✅ 添加错误处理中间件

### 短期执行（本月）

4. ⏳ 配置化语言列表
5. ⏳ 添加日志记录
6. ⏳ 添加缓存层

### 长期执行（下季度）

7. ⏳ 添加单元测试
8. ⏳ 添加集成测试
9. ⏳ 添加性能监控

---

**关键原则**:
1. **保持简单** - 不要过度设计
2. **保持一致** - 遵循现有模式
3. **保持文档** - 及时更新文档
4. **保持测试** - 确保代码质量

---

**文档版本**: v1.0  
**创建日期**: 2026-05-25  
**下次审查**: 2026-06-25
