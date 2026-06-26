# 🌑 暗黑主题使用指南

本主题提取自 legacy `admin-panel` 的精美暗黑风格设计，并适配了 Element Plus 组件库。

## 📦 已提取的文件

- ✅ `src/styles/theme-dark.css` - 完整的暗黑主题CSS

## 🎨 设计特点

### 核心风格元素
- 🌑 **深黑背景** - #0a0a0a 主背景，#16171d 次级背景
- 💚 **霓虹绿主色** - #10b981 (Emerald) 作为主要强调色
- ✨ **发光效果** - 按钮和交互元素带有霓虹发光阴影
- 🔲 **虚线边框** - 使用 dashed 边框增加科技感
- 🔤 **大写字母** - 标题和按钮使用全大写 + 加宽字距
- 📟 **等宽字体** - 数字和代码使用 monospace 字体
- 🎭 **斜体强调** - 激活状态使用斜体突出

## 🚀 启用方式

### 方式一：全局启用（推荐开发测试）

在 `src/main.js` 中添加：

```javascript
import './styles/theme-dark.css'

// 设置 body 的 data-theme 属性
document.body.setAttribute('data-theme', 'dark')
```

### 方式二：动态切换（推荐生产环境）

#### 1. 创建主题切换 Store

在 `src/stores/theme.js` 中创建：

```javascript
import { defineStore } from 'pinia'

export const useThemeStore = defineStore('theme', {
  state: () => ({
    theme: localStorage.getItem('admin_theme') || 'light'
  }),
  
  actions: {
    toggleTheme() {
      this.theme = this.theme === 'light' ? 'dark' : 'light'
      this.applyTheme()
    },
    
    setTheme(theme) {
      this.theme = theme
      this.applyTheme()
    },
    
    applyTheme() {
      document.body.setAttribute('data-theme', this.theme)
      localStorage.setItem('admin_theme', this.theme)
    },
    
    initTheme() {
      this.applyTheme()
    }
  }
})
```

#### 2. 在 main.js 中初始化

```javascript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './styles/theme-dark.css'  // 引入暗黑主题

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// 初始化主题
import { useThemeStore } from '@/stores/theme'
const themeStore = useThemeStore()
themeStore.initTheme()

app.mount('#app')
```

#### 3. 创建主题切换组件

在 `src/components/ThemeToggle.vue` 中：

```vue
<template>
  <el-switch
    v-model="isDark"
    inline-prompt
    :active-icon="Moon"
    :inactive-icon="Sunny"
    @change="toggleTheme"
    style="--el-switch-on-color: #10b981; --el-switch-off-color: #dcdfe6"
  />
</template>

<script setup>
import { computed } from 'vue'
import { Moon, Sunny } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

const isDark = computed({
  get: () => themeStore.theme === 'dark',
  set: (value) => themeStore.setTheme(value ? 'dark' : 'light')
})

const toggleTheme = () => {
  themeStore.toggleTheme()
}
</script>
```

#### 4. 在布局中添加切换按钮

在 `src/layouts/MainLayout.vue` 的顶部导航栏添加：

```vue
<template>
  <el-header>
    <!-- 其他内容 -->
    <div class="header-right">
      <ThemeToggle />
      <!-- 其他按钮 -->
    </div>
  </el-header>
</template>

<script setup>
import ThemeToggle from '@/components/ThemeToggle.vue'
</script>
```

## 🎯 样式定制

### 自定义霓虹色

在 `theme-dark.css` 中修改这些变量：

```css
:root[data-theme='dark'] {
  --dark-accent-emerald: #10b981;  /* 主色：霓虹绿 */
  --dark-accent-sky: #0ea5e9;      /* 天蓝 */
  --dark-accent-rose: #f43f5e;     /* 玫瑰红 */
  --dark-accent-amber: #f59e0b;    /* 琥珀色 */
}
```

### 统计卡片霓虹渐变

给统计卡片添加不同的渐变类：

```vue
<el-card class="stat-card orders">
  <!-- 订单统计 - 天蓝渐变 -->
</el-card>

<el-card class="stat-card users">
  <!-- 用户统计 - 玫瑰渐变 -->
</el-card>

<el-card class="stat-card revenue">
  <!-- 销售额 - 霓虹绿渐变 -->
</el-card>

<el-card class="stat-card tickets">
  <!-- 工单 - 琥珀渐变 -->
</el-card>
```

### 发光按钮

按钮会自动获得霓虹发光效果：

```vue
<el-button type="primary">
  <!-- 自动带有霓虹绿发光 -->
</el-button>
```

额外添加脉动效果：

```vue
<el-button type="primary" class="glow-effect">
  <!-- 带脉动发光效果 -->
</el-button>
```

## 📝 组件适配清单

已适配的 Element Plus 组件：

- ✅ `el-card` - 卡片（虚线边框 + 渐变）
- ✅ `el-button` - 按钮（霓虹发光）
- ✅ `el-table` - 表格（深色背景）
- ✅ `el-input` - 输入框（深色 + 霓虹边框）
- ✅ `el-menu` - 菜单（深色 + 激活态）
- ✅ `el-tag` - 标签（等宽字体 + 霓虹色）
- ✅ `el-dialog` - 对话框（深色 + 虚线边框）
- ✅ `el-pagination` - 分页（深色 + 霓虹激活态）
- ✅ `el-message` - 消息提示（深色 + 彩色边框）
- ✅ `el-loading` - 加载动画（霓虹绿）
- ✅ `el-empty` - 空状态（等宽字体）

## 🔧 与现有样式的兼容性

### 不冲突的设计

暗黑主题使用 `body[data-theme='dark']` 选择器，只有当 `data-theme="dark"` 时才生效，不会影响默认的亮色主题。

### 优先级

暗黑主题的样式优先级高于默认样式，确保正确覆盖。

## 🎨 设计对比

| 特性 | 默认 Element Plus 主题 | Tanzanite 暗黑主题 |
|-----|---------------------|------------------|
| 背景色 | 白色 #FFFFFF | 深黑 #0a0a0a |
| 主色调 | 蓝色 #409EFF | 霓虹绿 #10b981 |
| 边框 | 实线 solid | 虚线 dashed |
| 按钮 | 扁平风格 | 发光效果 |
| 字体 | 常规 | 大写 + 等宽 |
| 卡片 | 直角 | 圆角 24px |
| 激活态 | 填充色 | 边框 + 斜体 |

## 📸 预览效果

### Dashboard（仪表板）
- 统计卡片：深黑背景 + 霓虹渐变悬停效果
- 图表：适配深色背景
- 快速操作：霓虹绿发光按钮

### 数据表格
- 深黑背景
- 悬停行高亮
- 等宽字体数字

### 侧边栏菜单
- 深黑背景 + 虚线分隔
- 激活项：霓虹绿边框 + 斜体
- 全大写字母

## ⚠️ 注意事项

1. **图表颜色** - 需要在 ECharts 配置中单独设置深色主题
2. **图片资源** - 建议使用 SVG 或透明背景的 PNG
3. **文字对比度** - 确保文字颜色与深色背景对比度足够
4. **测试兼容性** - 在不同浏览器测试发光效果

## 🔄 未来改进

- [ ] 添加自定义颜色选择器
- [ ] 支持更多组件适配
- [ ] 添加主题预设（蓝色科技风、紫色梦幻风等）
- [ ] 优化动画性能
- [ ] 添加色盲友好模式

## 📚 相关资源

- [Element Plus 官方文档](https://element-plus.org/)
- [Tailwind CSS 颜色系统](https://tailwindcss.com/docs/customizing-colors)
- [CSS 发光效果教程](https://css-tricks.com/glow-effects/)

---

**提取自**: legacy `admin-panel` 项目  
**适配日期**: 2026-06-26  
**维护者**: Tanzanite Team
