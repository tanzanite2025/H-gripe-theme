# Reviews - 商品评论管理

**页面路径**: `admin.php?page=tanzanite-settings-reviews`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/tanzanite/v1/reviews`

---

## 📋 功能概述

商品评论管理 提供完整的管理功能。

---

##  主要功能

### 1. 列表管理

- 查看列表
- 搜索筛选
- 批量操作

### 2. 创建/编辑

- 添加新项
- 编辑现有项
- 删除项

### 3. 数据导出

- 导出数据
- CSV 格式
- 批量下载

---

##  REST API

### 获取列表

```
GET /wp-json/tanzanite/v1/reviews
```

### 创建

```
POST /wp-json/tanzanite/v1/reviews
```

### 更新

```
PUT /wp-json/tanzanite/v1/reviews/{id}
```

### 删除

```
DELETE /wp-json/tanzanite/v1/reviews/{id}
```

---

##  前端集成

```javascript
const { $wpApi } = useNuxtApp()

// 获取列表
const items = await $wpApi('/reviews')

// 创建
const response = await $wpApi('/reviews', {
  method: 'POST',
  body: { /* 数据 */ }
})
```

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
