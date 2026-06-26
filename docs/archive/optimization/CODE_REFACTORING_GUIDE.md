# 🔧 Tanzanite 代码重构实施指南

**创建日期**: 2026-06-26  
**状态**: 🟢 P0阶段已完成基础工具  
**下一步**: 开始应用到现有代码

---

## 📦 已创建的基础工具

### 1. 统一错误处理包 ✅

**文件**: `go-backend/internal/pkg/apierror/error.go`

**功能**:
- 标准化错误响应格式
- 预定义常用错误代码
- 快捷错误响应方法

**使用示例**:

```go
import "tanzanite/internal/pkg/apierror"

// ❌ 重构前
c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

// ✅ 重构后
apierror.RespondNotFound(c, "Order")
apierror.RespondBadRequest(c, "Invalid ID")
apierror.RespondInternalError(c, err)
```

### 2. 统一响应格式包 ✅

**文件**: `go-backend/internal/pkg/response/response.go`

**功能**:
- 标准化成功响应格式
- 统一分页响应结构
- 简化响应代码

**使用示例**:

```go
import "tanzanite/internal/pkg/response"

// ❌ 重构前
c.JSON(200, gin.H{"products": products})
c.JSON(200, gin.H{"data": orders, "total": total, "page": page})

// ✅ 重构后
response.Success(c, products)
response.Paged(c, orders, page, pageSize, total)
```

### 3. 统一分页参数包 ✅

**文件**: `go-backend/internal/pkg/pagination/pagination.go`

**功能**:
- 消除硬编码的分页默认值
- 自动限制最大值
- 简化分页参数解析

**使用示例**:

```go
import "tanzanite/internal/pkg/pagination"

// ❌ 重构前
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
if pageSize > 100 {
    pageSize = 100
}

// ✅ 重构后
params := pagination.ParsePagination(c)
// params.Page, params.PageSize, params.Offset()
```

### 4. 前端统一 API Base URL ✅

**文件**: `nuxt-i18n/app/composables/useApiBase.ts`

**功能**:
- 消除重复的 URL 构造代码
- 统一 API Base URL 逻辑

**使用示例**:

```typescript
// ❌ 重构前（在30+个文件中重复）
const config = useRuntimeConfig()
const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')

// ✅ 重构后
const apiBase = useApiBase()
const response = await $fetch(`${apiBase.value}/products`)
```

---

## 🎯 重构实施步骤

### Phase 1: 工具创建 ✅ 已完成

- [x] 创建 `pkg/apierror` 包
- [x] 创建 `pkg/response` 包
- [x] 创建 `pkg/pagination` 包
- [x] 创建 `useApiBase` composable

### Phase 2: 应用到现有代码（下一步）

#### 步骤 2.1: 重构一个示例Handler

选择一个中等复杂度的handler作为示例：

**候选**: `internal/api/v1/product/handler.go`

**重构清单**:
1. 导入新包
2. 替换错误响应为 `apierror.*`
3. 替换成功响应为 `response.*`
4. 替换分页解析为 `pagination.ParsePagination`
5. 测试验证

#### 步骤 2.2: 批量重构其他Handler

使用相同模式重构：
- `ticket/handler.go`
- `subscription/handler.go`
- `shipping/handler.go`
- 其他20+个handler文件

#### 步骤 2.3: 重构前端Composables

替换所有使用API URL构造的文件：
- `composables/usePublicSettings.ts`
- `composables/useProductAttributes.ts`
- `composables/useShopCategories.ts`
- 其他25+个文件

### Phase 3: 拆分超长文件（稍后）

详见 [CODE_QUALITY_AUDIT_REPORT.md](CODE_QUALITY_AUDIT_REPORT.md) 中的问题3

---

## 📋 重构Checklist模板

### Handler 重构检查清单

```markdown
- [ ] 导入必要的包
  ```go
  import (
      "tanzanite/internal/pkg/apierror"
      "tanzanite/internal/pkg/response"
      "tanzanite/internal/pkg/pagination"
  )
  ```

- [ ] 替换所有错误响应
  - [ ] 400 Bad Request → `apierror.RespondBadRequest`
  - [ ] 401 Unauthorized → `apierror.RespondUnauthorized`
  - [ ] 404 Not Found → `apierror.RespondNotFound`
  - [ ] 500 Internal Error → `apierror.RespondInternalError`

- [ ] 替换所有成功响应
  - [ ] 单一数据 → `response.Success`
  - [ ] 分页数据 → `response.Paged`
  - [ ] 创建成功 → `response.Created`
  - [ ] 删除成功 → `response.NoContent`

- [ ] 替换分页参数解析
  - [ ] `page, pageSize` → `pagination.ParsePagination(c)`
  - [ ] `limit` → `pagination.ParseLimit(c)`

- [ ] 测试验证
  - [ ] 编译通过
  - [ ] API响应格式正确
  - [ ] 错误码和消息正确
```

### Composable 重构检查清单

```markdown
- [ ] 替换 API Base URL 构造
  ```typescript
  // 删除重复代码
  - const config = useRuntimeConfig()
  - const base = ...

  // 添加新导入
  + const apiBase = useApiBase()
  
  // 使用
  - `${base}/endpoint`
  + `${apiBase.value}/endpoint`
  ```

- [ ] 验证功能正常
  - [ ] API调用成功
  - [ ] URL格式正确
```

---

## 🧪 测试策略

### 单元测试（推荐）

```go
// internal/pkg/apierror/error_test.go
package apierror_test

import (
    "net/http/httptest"
    "testing"
    "tanzanite/internal/pkg/apierror"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestRespondNotFound(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    apierror.RespondNotFound(c, "Order")
    
    assert.Equal(t, 404, w.Code)
    // 验证JSON响应...
}
```

### 集成测试

使用现有API测试脚本验证：

```bash
# 测试产品API
cd go-backend
./test-product-api.ps1

# 测试订单API
./test-order-api.ps1
```

---

## 📊 预期成果

### 代码减少量

| 重构项 | 预计减少行数 |
|--------|-------------|
| 错误处理重复 | -500行 |
| 分页参数解析 | -200行 |
| API URL构造 | -200行 |
| **总计** | **-900行** |

### 一致性提升

| 指标 | 重构前 | 重构后 | 提升 |
|------|--------|--------|------|
| 错误响应格式统一 | 40% | 100% | +150% |
| 分页行为统一 | 60% | 100% | +67% |
| API响应格式统一 | 50% | 100% | +100% |

---

## 🚨 注意事项

### 1. 向后兼容

如果有外部系统依赖旧的API响应格式：

```go
// 可以添加兼容层
func LegacySuccess(c *gin.Context, data interface{}) {
    // 保持旧格式
    c.JSON(200, gin.H{"data": data})
}
```

### 2. 渐进式重构

**不要**一次性重构所有文件，建议：

1. 先重构1-2个handler作为试点
2. 验证没有问题后再批量重构
3. 每次提交保持功能可用

### 3. 团队协作

- 通知团队成员新的编码规范
- 更新开发文档
- 在PR中强制使用新工具

---

## 📚 相关文档

- [CODE_QUALITY_AUDIT_REPORT.md](CODE_QUALITY_AUDIT_REPORT.md) - 完整审计报告
- [DATA_SYNC_AUDIT_REPORT.md](DATA_SYNC_AUDIT_REPORT.md) - 数据同步审计
- [BUGFIX_COMPLETE_SUMMARY.md](BUGFIX_COMPLETE_SUMMARY.md) - BUG修复总结

---

## ✅ 下一步行动

### 立即执行

1. **重构示例Handler**: 选择 `product/handler.go` 作为试点
2. **验证功能**: 运行测试确保没有破坏现有功能
3. **提交代码**: 创建 PR 供团队审查

### 本周计划

- 重构5-10个handler文件
- 重构10-15个前端composable
- 更新开发文档

### 下周计划

- 完成所有handler重构
- 完成所有composable重构
- 开始拆分超长文件

---

**创建人**: Kiro AI  
**状态**: 🟢 工具已就绪，等待应用  
**预计完成**: P0重构 2-3周
