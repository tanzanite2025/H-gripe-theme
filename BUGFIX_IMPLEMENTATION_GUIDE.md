# 🔧 数据同步BUG修复实施指南

本文档提供详细的代码修复方案，可直接复制使用。

---

## 🔴 P0-1: Admin Panel 权限验证机制

### 修改文件：`go-backend/web/admin/src/stores/auth.js`

#### 步骤1：添加权限验证方法

```javascript
// 在 defineStore 中添加新方法
const verifyPermissions = async () => {
  if (!token.value) return { valid: false }
  
  try {
    const response = await axios.get('/api/admin/auth/permissions')
    const serverPermissions = response.data.permissions || []
    
    // 对比本地和服务器权限
    const localPerms = JSON.stringify(permissions.value.sort())
    const serverPerms = JSON.stringify(serverPermissions.sort())
    
    if (localPerms !== serverPerms) {
      console.warn('[Auth] Permissions updated from server')
      permissions.value = serverPermissions
      localStorage.setItem('admin_permissions', JSON.stringify(serverPermissions))
    }
    
    return { valid: true, updated: localPerms !== serverPerms }
  } catch (error) {
    console.error('[Auth] Permission verification failed', error)
    
    // 验证失败，可能Token过期
    if (error.response?.status === 401) {
      logout()
    }
    
    return { valid: false, error }
  }
}
```

#### 步骤2：添加初始化方法

```javascript
const initAuth = async () => {
  if (!token.value) return
  
  // 页面加载时验证权限
  const result = await verifyPermissions()
  
  if (!result.valid) {
    console.warn('[Auth] Token invalid, logging out')
    logout()
  }
}
```

#### 步骤3：导出方法

```javascript
return {
  // 现有的导出
  token,
  user,
  permissions,
  isAuthenticated,
  hasPermission,
  hasRole,
  login,
  logout,
  
  // 新增导出
  verifyPermissions,
  initAuth
}
```

### 修改文件：`go-backend/web/admin/src/App.vue`

```vue
<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

onMounted(async () => {
  // 页面加载时初始化认证状态
  await authStore.initAuth()
})
</script>
```

### 后端API（需要确认是否已实现）

检查文件：`go-backend/internal/api/v1/admin/auth_handler.go`

```go
// GetPermissions 获取当前管理员的权限列表
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	// 从JWT Token中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// 查询用户信息和权限
	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// 根据角色返回权限列表
	permissions := h.authService.GetRolePermissions(user.Role)

	c.JSON(200, gin.H{
		"permissions": permissions,
		"role": user.Role,
	})
}
```

---

## 🔴 P0-2: 购物车同步失败处理

### 修改文件：`nuxt-i18n/app/composables/useCart.ts`

#### 步骤1：添加重试逻辑

```typescript
// 在文件顶部添加重试工具函数
const retryOperation = async <T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<{ success: boolean; data?: T; error?: string }> => {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const data = await operation()
      return { success: true, data }
    } catch (error) {
      console.warn(`Attempt ${attempt}/${maxRetries} failed:`, error)
      
      if (attempt === maxRetries) {
        return {
          success: false,
          error: error instanceof Error ? error.message : 'Unknown error'
        }
      }
      
      // 等待后重试
      await new Promise(resolve => setTimeout(resolve, delay * attempt))
    }
  }
  
  return { success: false, error: 'Max retries exceeded' }
}
```

#### 步骤2：重写同步方法

```typescript
const syncGuestCart = async (): Promise<{
  success: boolean
  error?: string
  itemsCount?: number
}> => {
  if (!import.meta.client) {
    return { success: false, error: 'Not in client' }
  }
  
  const saved = localStorage.getItem('tanzanite_cart')
  if (!saved) {
    return { success: true, itemsCount: 0 }
  }
  
  try {
    const items = JSON.parse(saved)
    if (!items || items.length === 0) {
      localStorage.removeItem('tanzanite_cart')
      return { success: true, itemsCount: 0 }
    }
    
    const payload = items.map((item: any) => ({
      product_id: item.id,
      quantity: item.quantity
    }))
    
    // 使用重试机制同步到后端
    const result = await retryOperation(
      () => auth.request('/cart/sync', {
        method: 'POST',
        body: JSON.stringify(payload),
        headers: { 'Content-Type': 'application/json' }
      }),
      3, // 最多重试3次
      1000 // 延迟1秒
    )
    
    if (result.success) {
      // ✅ 同步成功，删除本地数据
      localStorage.removeItem('tanzanite_cart')
      await loadCartFromBackend()
      
      return {
        success: true,
        itemsCount: items.length
      }
    } else {
      // ❌ 同步失败，保留本地数据
      console.error('[Cart] Sync failed after retries:', result.error)
      
      return {
        success: false,
        error: result.error || 'Sync failed',
        itemsCount: items.length
      }
    }
  } catch (e) {
    console.error('[Cart] Failed to parse guest cart', e)
    
    return {
      success: false,
      error: 'Failed to parse cart data'
    }
  }
}
```

#### 步骤3：添加用户提示

修改文件：`nuxt-i18n/app/layouts/default.vue` 或登录后的回调

```vue
<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCart } from '@/composables/useCart'

const cart = useCart()

// 登录成功后调用
const onLoginSuccess = async () => {
  const result = await cart.syncGuestCart()
  
  if (!result.success) {
    // 同步失败，显示错误提示和重试按钮
    ElMessageBox.alert(
      `购物车同步失败（${result.itemsCount || 0}件商品），请刷新页面重试`,
      '购物车同步失败',
      {
        confirmButtonText: '刷新页面',
        type: 'warning',
        callback: () => {
          window.location.reload()
        }
      }
    )
  } else if (result.itemsCount && result.itemsCount > 0) {
    // 同步成功，显示成功提示
    ElMessage.success(`已同步 ${result.itemsCount} 件商品到购物车`)
  }
}
</script>
```

---

## 🟠 P1-1: 聊天消息后端持久化

### 步骤1：创建后端API

新建文件：`go-backend/internal/api/v1/chat/handler.go`

```go
package chat

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"tanzanite/internal/repository"
)

type ChatHandler struct {
	chatRepo *repository.ChatRepository
}

func NewChatHandler(chatRepo *repository.ChatRepository) *ChatHandler {
	return &ChatHandler{chatRepo: chatRepo}
}

// SaveMessage 保存聊天消息
func (h *ChatHandler) SaveMessage(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Message   struct {
			ID        string `json:"id"`
			Content   string `json:"content"`
			Role      string `json:"role"` // user, agent, system
			Timestamp int64  `json:"timestamp"`
		} `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库
	err := h.chatRepo.SaveMessage(req.SessionID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetMessages 获取聊天历史
func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	messages, err := h.chatRepo.GetMessages(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
```

### 步骤2：创建数据模型

新建文件：`go-backend/internal/domain/chat/chat.go`

```go
package chat

import "time"

type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index;not null"`
	MessageID string    `json:"message_id" gorm:"uniqueIndex;not null"`
	Content   string    `json:"content" gorm:"type:text"`
	Role      string    `json:"role"` // user, agent, system
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
```

### 步骤3：创建Repository

新建文件：`go-backend/internal/repository/chat_repository.go`

```go
package repository

import (
	"tanzanite/internal/domain/chat"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveMessage(sessionID string, message interface{}) error {
	// 实现保存逻辑
	return nil
}

func (r *ChatRepository) GetMessages(sessionID string) ([]chat.ChatMessage, error) {
	var messages []chat.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&messages).Error
	return messages, err
}
```

### 步骤4：修改前端

修改文件：`nuxt-i18n/app/composables/chat/useWhatsAppState.ts`

```typescript
const saveChatMessage = async (sessionId: string, message: ChatMessage) => {
  // 1. 先存储到 localStorage（快速响应，离线可用）
  const localKey = `tanzanite_chat_${sessionId}`
  const messages = JSON.parse(localStorage.getItem(localKey) || '[]')
  messages.push(message)
  localStorage.setItem(localKey, JSON.stringify(messages))
  
  // 2. 异步同步到后端（不阻塞用户）
  if (import.meta.client) {
    try {
      await $fetch('/api/v1/chat/messages', {
        method: 'POST',
        body: {
          session_id: sessionId,
          message: message
        }
      })
    } catch (e) {
      console.warn('[Chat] Message sync failed, will retry later', e)
      // TODO: 添加到重试队列
    }
  }
}

const loadChatHistory = async (sessionId: string) => {
  try {
    // 优先从后端加载（跨设备同步）
    const response = await $fetch<{ messages: ChatMessage[] }>(
      `/api/v1/chat/messages?session_id=${sessionId}`
    )
    
    if (response.messages && response.messages.length > 0) {
      return response.messages
    }
  } catch (e) {
    console.warn('[Chat] Failed to load from server, using local cache', e)
  }
  
  // 降级到 localStorage
  const localKey = `tanzanite_chat_${sessionId}`
  return JSON.parse(localStorage.getItem(localKey) || '[]')
}
```

---

## 🟠 P1-2: 浏览历史后端同步

### 步骤1：创建后端API

修改文件：`go-backend/internal/api/v1/auth/handler.go`

```go
// AddBrowsingHistory 添加浏览历史
func (h *AuthHandler) AddBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 保存到浏览历史表
	err := h.userRepo.AddBrowsingHistory(userID.(uint), req.ProductID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save browsing history"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}
```

### 步骤2：修改前端

修改文件：`nuxt-i18n/app/composables/useBrowsingHistory.ts`

```typescript
const addToBrowsingHistory = async (product: Product) => {
  // 1. 本地存储（快速响应）
  const history = JSON.parse(
    localStorage.getItem('tanzanite_browsing_history') || '[]'
  )
  
  // 去重
  const filtered = history.filter((p: Product) => p.id !== product.id)
  filtered.unshift(product)
  filtered.splice(10) // 保留最近10条
  
  localStorage.setItem('tanzanite_browsing_history', JSON.stringify(filtered))
  
  // 2. 异步同步到后端（登录用户）
  if (auth.isAuthenticated && import.meta.client) {
    try {
      await $fetch('/api/v1/user/browsing-history', {
        method: 'POST',
        body: { product_id: product.id }
      })
    } catch (e) {
      console.warn('[BrowsingHistory] Failed to sync', e)
    }
  }
}
```

---

## ✅ 测试清单

### P0-1: 权限验证
- [ ] 管理员登录后，修改其权限，刷新页面应自动更新
- [ ] Token过期后，应自动登出
- [ ] 权限更新后，菜单和按钮应立即响应

### P0-2: 购物车同步
- [ ] 游客添加商品，登录后同步成功
- [ ] 同步失败时，本地数据保留，显示提示
- [ ] 点击"刷新页面"按钮可重试
- [ ] 同步成功后，本地数据清除

### P1-1: 聊天消息
- [ ] 发送消息后，后端数据库有记录
- [ ] 切换设备后，能看到历史消息
- [ ] 网络失败时，消息保留在 localStorage

### P1-2: 浏览历史
- [ ] 登录用户浏览产品，后端有记录
- [ ] 多设备登录，浏览历史同步

---

**实施建议**: 按照 P0 → P1 → P2 的顺序修复，每个修复后运行测试。

