# Spoke Calculator System / 辐条计算器系统手册

> **最后更新**: 2025-12-31
> **状态**: Active

## 1. 系统架构 (Architecture)

本模块已从传统的 "Runtime API 获取" 模式重构为 **"前端静态数据库 (Frontend Static Database)"** 模式。

- **核心数据源**: `app/data/spoke-calculator/database.ts`
- **优势**: 0ms 延迟加载，强类型校验，无后端运行风险。
- **组成部分**:
  - `SpokeCalculatorCore.vue`: 主计算器组件 (Brand -> Model 级联选择)。
  - `SpokeSmartSearch.vue`: 智能搜索组件 (关键词模糊匹配)。
  - `spokeMath.ts`: 公共计算算法库。

## 2. 数据管理与同步 (Data Management & Sync)

虽然前端使用静态文件，但数据的"真理来源 (Source of Truth)" 依然是 **WordPress 后台**。

### 2.1 数据结构

- **RIM_DATABASE**: 轮圈几何数据 (ERD, Name).
- **HUB_DATABASE**: 花鼓几何数据 (Flange PCD, Distances).
- **PRESET_BUILDS**: 官方预设组合 (仅前端维护，用于智能搜索).

### 2.2 同步工作流 (Sync Workflow)

当运营人员在 WordPress 后台添加了新的 Spoke Product (花鼓或轮圈) 后，开发人员需执行以下操作将数据同步到前端：

1. **准备环境**: 确保本地前端项目 (`nuxt-i18n`) 已连通外网。
2. **运行命令**:

    ```bash
    npm run sync-data
    ```

3. **脚本行为 (`scripts/sync-spoke-data.mjs`)**:
    - 连接 WordPress API: `/wp-json/tanzanite/v1/spoke-db-export`
    - 拉取最新的 `rims` 和 `hubs` 数据。
    - 读取本地 `database.ts`，**自动保留** 您手动维护的 `PRESET_BUILDS` 部分。
    - 生成新的 `database.ts` 文件并覆盖。
4. **提交代码**: 检查 `database.ts` 变更无误后，提交 Git。

## 3. 智能搜索 (Smart Search)

位于计算器下方的搜索栏组件 (`SpokeSmartSearch.vue`)。

- **逻辑**: 它**不**搜索原始的花鼓/轮圈数据（因为那是未验证的），而是搜索 `PRESET_BUILDS`。
- **管理**: 若要让某个花鼓组合在搜索中出现，必须在 `database.ts` 底部的 `PRESET_BUILDS` 数组中手动添加配置。

```typescript
// Example Preset
{
  id: 'tz_ar45_dt350',
  name: 'Tanzanite AR 45 + DT Swiss 350', // 搜索显示名
  keywords: ['350', 'dt swiss'],         // 搜索关键词
  // ... 具体的几何引用 ID
}
```

## 4. 相关文件索引

- **数据**: `app/data/spoke-calculator/database.ts`
- **同步脚本**: `scripts/sync-spoke-data.mjs`
- **计算逻辑**: `app/utils/spokeMath.ts`
- **WP 导出器**: `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-spoke-export-controller.php`
