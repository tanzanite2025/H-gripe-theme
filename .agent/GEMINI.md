# 🛡️ Global Antigravity Rules

> 🛑 **CRITICAL LANGUAGE RULE (关键原则)**:
> 所有生成的 Artifacts文档 (.md) 、Commit Message 以及用户对话 **必须强制使用中文**。
> 严禁生成英文文档（代码片段除外）。
> Before saving any .md file, ASK YOURSELF: "Is this written in Chinese?"

> 此文件定义 Gemini Code Assist 的全局行为约束。AI 必须严格对齐 Antigravity 插件的防护逻辑。

## 🔒 双层安全保障体系

Antigravity 插件已在底层部署了严格的拦截策略。AI 必须完全理解并遵守同一套规则，形成双层冗余保护。

- **第一层：GEMINI.md** 是 AI 行为约束，在 AI 决策层面约束行为
- **第二层：Antigravity 插件** 是系统级拦截，在命令执行层面拦截危险操作

**如何协同**：
AI 必须主动遵守本文件规则，**不得依赖**插件的拦截功能作为唯一的安全防线。

---

## 🎨 UI/UX 规范 (Frontend Standards)

> 任何涉及 UI 组件的新增或修改，必须优先遵循以下规范，禁止重复造轮子。

### 1. 导航与 Tabs (Navigation)

所有水平胶囊式选项卡 (Pill Tabs) **必须** 使用全局 CSS 组件，严禁编写局部 Scoped CSS：

- **File**: `app/assets/css/components/nav.css` (已在 `nuxt.config.ts` 全局注册)
- **Container Class**: `.nav-pill-tabs`
- **Item Class**: `.nav-pill-item`
- **Active Class**: `.nav-pill-item--active` (白色背景，黑色文字，无渐变)

**❌ 禁止**: 使用 `linear-gradient` 自定义 Active 状态（除非设计稿明确要求打破规范）。

### 2. 卡片与阴影样式 (Cards & Shadows) (Global)

所有嵌套在深色背景中的子卡片 (Inner Cards)，必须遵循 "无边框 + 深阴影" 的设计语言，严禁使用边框描边：

- **Border**: `border-none` (移除所有边框)
- **Shadow**: `shadow-[0_4px_16px_rgba(0,0,0,0.5)]` (纯黑弥散阴影)
- **Background**: 保持各材质原本的背景色 (e.g. `bg-slate-800/40`, `bg-emerald-500/5`)

**❌ 禁止**: 使用 `border` 或 `border-*` 描边来区分层级（除非是特定的高亮交互）。

### 3. 按钮与胶囊样式 (Buttons & Pills) (Global)

所有 "Read More", "View All", "Shop Now" 等操作按钮，以及 Tab 切换器，**必须** 使用全局类 `.premium-button`。

- **Class**: `premium-button` (Defined in `tailwind.css`)
- **Style**: Pill Shape, No Border, Deep Slate Background (50% opacity), Soft Diffuse Shadow.
- **Active State**: `premium-button--active` (White Background, Dark Text).

**示例**:

```html
<NuxtLink to="/shop" class="premium-button">View More</NuxtLink>
```

---

## 📋 核心工作流规范 (Mandatory Workflow)

**所有任务必须严格遵循以下 "3文件 + 1确认" 流程，严禁跳过：**

### 1. 阶段一：规划与任务拆解 (Planning)

在编写任何业务代码前，**必须**先创建/更新以下文件（**注意：所有 Artifacts 文档必须使用中文编写**）：

- **`task.md`**：
  - 将大任务拆解为细粒度、可追踪的 check-list item。
  - 明确每一步的预期结果。
- **`implementation_plan.md`**：
  - 详细描述技术方案、涉及修改的文件列表。
  - 列出潜在风险和破坏性变更。

> **🛑 暂停点**：完成上述文件后，**必须**使用 `notify_user` 请求用户审查规划。
> **❌ 严禁**在用户明确批准计划前开始修改业务代码。

### 2. 阶段二：执行 (Execution)

- **仅在用户批准计划后**，方可开始编写代码。
- 严格按照 `task.md` 的顺序逐项执行。
- 遇到意外复杂情况（如需修改原计划），必须返回阶段一更新 `implementation_plan.md` 并再次请求确认。
- 涉及危险操作（如批量删除、重构核心）前，必须建议创建 Checkpoint。

### 3. 阶段三：验证与总结 (Verification)

- 执行验证测试。
- **创建/更新 `walkthrough.md`**：
  - 记录已完成的工作。
  - 提供测试结果截图/录屏。
  - 总结本次任务的变更点。

---

## 🚨 绝对禁止的操作

1. **禁止删除核心文件**
   - 不得删除 `package.json`, `tsconfig.json`, `go.mod`, `Cargo.toml` 等项目配置文件
   - 不得删除 `.gitignore`, `.env`, `.env.local` 等环境配置
   - 不得删除 `src/extension.ts`, `src/main.go` 等入口文件

2. **禁止执行高危终端命令**（除非用户明确要求且确认）
   - `rm -rf` / `del /s /q` / `rmdir /s`
   - `git reset --hard` / `git clean -fdx`
   - `drop database` / `truncate table`
   - `curl ... | bash` / `wget ... | sh`

3. **禁止修改版本控制历史**
   - 不得执行 `git push --force`
   - 不得执行 `git rebase` 涉及已推送的提交

---

---

> **提示**：特定项目的架构约束和已知陷阱，请在各项目根目录的 `.cursorrules` 或项目级 `GEMINI.md` 中定义。
