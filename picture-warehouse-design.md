# Picture warehouse 功能设计草案（草稿）

> 目标：在 /company/picture-warehouse 下实现一个两栏照片墙页面，一栏展示用户上传的照片，一栏展示品牌/官方上传的商品图；支持用户上传但必须经过后台审核，通过后才展示。后台逻辑放在一个新的独立插件中，而不是继续堆在 tanzanite-setting 里。

---

## 1. 页面信息架构（/company/picture-warehouse）

### 1.1 顶部结构
- 复用 `products` 布局：
  - 顶部 Company 导航（三项：Our Story / Membership and Points / Picture warehouse）。
  - 下方内容区域专用于图片展示与说明。

- 页面标题 & 简要说明：
  - 标题：`Picture warehouse`
  - 简介：一两句话解释：
    - 左栏是用户实拍图
    - 右栏是官方/商品图

### 1.2 主内容区：两栏照片墙

- **布局原则**：
  - 桌面端：左右两列并排
    - 左列：用户照片
    - 右列：品牌/商品照片
  - 平板/小屏：
    - 上下堆叠：先用户照片，再品牌照片

- 每列上方有一个简短的标题+说明：
  - 左：`Riders photos` / `User photos`
  - 右：`Tanzanite photos` / `Product photos`

- **每列展示数量控制**：
  - 默认仅展示每列最新的最多 6 张照片（brand / user 各自独立计数）；
  - 如对应列的数据量大于 6，则在网格下方提供一个小按钮 `Show more photos` / `Show fewer photos`：
    - 点击展开时显示该列全部照片；
    - 收起时恢复为仅前 6 张；
  - Lightbox 打开与左右切换仍基于完整列表，不受 6 张截断影响。

- 每张图片卡片包含信息（先作为占位字段）：
  - 缩略图
  - 用户地区（必有，例如 `Germany · Berlin` 或 简写 `DE`）
  - 可选：
    - 用户名 / 昵称
    - 标题（例如 "First gravel ride"）
    - 车型/轮组信息

- 交互：
  - 初期只做 **点击放大**（可选，未来再考虑灯箱/滑动）。
  - 初期不做点赞/评论（避免复杂度），只预留数据字段可能性。

### 1.3 单张图片详情弹窗（Lightbox）

> 参考示意：点击任一照片墙中的图片后，进入一个居中的大图弹窗视图，包含评论、分享和推荐模块。

- **总体布局**：
  - 居中大图弹窗，覆盖当前页面（带半透明暗色遮罩）。
  - 上方居中显示该照片的标题（例如一条 build 说明），可选副标题作为简短描述。
  - 左右两侧有上一张 / 下一张的箭头按钮，用于在当前列表内切换照片。
  - 右上角有关闭按钮（`X`）。

- **下方主要内容区域：左右两列**：
  - 左侧：评论区和操作区。
  - 右侧：推荐模块（例如“Like this? Get the same build.”）。

- **左侧：操作按钮栏 + 评论列表**：
  - 顶部有一排操作按钮（示意）：
    - `Message`：切换到留言视图，展示评论列表，并可在将来接入评论表单。
    - `Upload`：入口，用于引导用户上传自己的照片（将来可跳转到上传表单区域）。
    - `Share`：点击后在按钮上方弹出分享面板。
    - `Like`：点赞按钮，带计数（例如 `Like 1`）。
  - 按钮排布风格：扁平矩形按钮，当前激活的按钮用主色突出（如 Message 选中状态）。
  - 下方是 **COMMENTS** 区域：
    - 每条评论包含：头像或首字母圆形标记、昵称、日期、正文内容。
    - 预留 `Reply` 区域，方便后期接入回复功能。

- **Share 弹层（分享与复制链接）**：
  - 点击 `Share` 后，在 Lightbox 下方左侧展示一组分享按钮：
    - `Share to Facebook`（蓝色背景，仅 UI，占位，`title="Coming soon"`）；
    - `Share to X`（黑色背景，仅 UI，占位）；
    - `Share to Reddit`（橙色背景，仅 UI，占位）；
    - `Copy link`：唯一有逻辑的按钮，复制当前照片的分享链接到剪贴板，并在下方显示 `Link copied.` 等提示文案。
  - 分享链接形式：在当前 URL 上追加查询参数，如 `?photo={id}&kind={user|brand}`，用于后续 deep-link 行为。
  - Lightbox 中部大图区域高度已略微压缩（约 45–60vh），为底部评论与推荐模块预留空间，避免在移动端必须大量滚动。

- **右侧推荐模块：Like this? Get the same build.**
  - 标题示例：`Like this? Get the same build.`
  - 下面按类别列出与当前照片关联的产品链接，例如：
    - `Rim`：关联到某条轮圈产品详情（如 `WG55 disc >>`）。
    - `Wheel(s)`：关联到某条整轮或轮组产品详情（如 `Wheelset WG55 disc >>`）。
    - `Hub`：关联到花鼓产品详情；
    - `Tire`：关联到外胎产品详情。
  - 前端会将 `tanz_photo_product_refs` 中对应字段渲染为可点击链接（`<a>`），在新标签页中打开产品 URL；
  - 若对应字段在 `tanz_photo_product_refs` 中缺失，则前端直接不显示该行链接，避免出现空链接或报错。

- **阶段性实现建议**：
  - Phase 1/2 中，Lightbox 可以只实现：大图 + 基本标题 + 左右切换 + 关闭，不做评论/分享/推荐逻辑，只预留 UI 结构。
  - 评论列表、分享弹层、右侧推荐模块可以在 Phase 3 之后，随用户上传与产品数据打通一起实现。

---

## 2. 数据模型与来源（规划）


### 2.1 WordPress 侧数据结构

- **自定义 Post Type：`tanz_photo`**
  - 用途：统一存储用户与品牌的所有照片条目。

- 建议字段（meta，带 key 名）：
  - `tanz_photo_type` (string)
    - 取值：`user` | `brand`。
    - 用途：区分用户上传 vs 品牌/官方上传，用于前端左右栏分组和后台筛选。
  - `tanz_photo_image_id` (int)
    - 关联 WordPress 媒体库 attachment ID。
    - 前端根据 attachment 生成缩略图 URL / 大图 URL。
  - `tanz_photo_region` (string)
    - 用户地区，例如：`Germany`, `USA`, `China - Xiamen`；
    - 对品牌图，可用 `Studio`, `Product shoot` 等简单标记。
  - `tanz_photo_location` (string，可选)
    - 更细粒度的位置，如 `Berlin`, `Seattle`；前端可与 region 组合显示。
  - `tanz_photo_nickname` (string，可选)
    - 上传者昵称或 rider 名称，用于卡片和评论区显示。
  - `tanz_photo_bike_model` (string，可选)
    - 车架 / 轮组 / 车型名称，供后续展示或过滤。
  - `tanz_photo_notes` (string，可选)
    - 额外备注，如拍摄说明、构建细节，主要用于后台参考。
  - `tanz_photo_status` (string)
    - 审核状态：`pending` | `approved` | `rejected`。
    - 仅 `approved` 可通过 gallery 接口返回给前端。
  - `tanz_photo_submitted_at` (datetime，可选)
    - 记录提交时间，用于排序（最新上传在前）。
  - `tanz_photo_approved_at` (datetime，可选)
    - 审核通过时间，用于按发布时间排序。
  - `tanz_photo_rejected_reason` (string，可选)
    - 如被拒绝，管理员可填写原因，暂仅后台可见。
  - `tanz_photo_source` (string，可选)
    - 来源：`web-form` / `manual-import` / `admin-upload` 等。
  - `tanz_photo_product_refs` (string，可选，JSON 或逗号分隔)
    - 存储与照片相关的产品/轮组链接，供右侧 "Like this? Get the same build." 模块使用；
    - 建议采用 JSON 字符串形式，示例：

      ```json
      {
        "rim": "https://example.com/products/wg55-disc-rim",
        "wheel": "https://example.com/products/wg55-disc-wheelset",
        "hub": "https://example.com/products/tanz-hub-01",
        "tire": "https://example.com/products/tire-32c-allroad"
      }
      ```

    - 其中各字段均为 **可选**：
      - `rim`：关联轮圈产品链接（建议使用完整 URL，如产品详情页）；
      - `wheel`：关联整轮/轮组产品链接；
      - `hub`：关联花鼓产品链接；
      - `tire`：关联外胎产品链接。
    - 前端在解析 `product_refs` 时必须做好防护：
      - 将整个字段视为可选；
      - 每个 key 单独判断存在与否，有则渲染对应一行链接，无则跳过该行，避免因字段缺失导致报错。

- 文章本身的 `post_title` / `post_content`：
  - 可以用于存储更长的描述或标题。

### 2.2 REST API 设计（高层）

> 具体路由与权限细节可以在后续迭代时再精细化，这里先定义方向。

- `GET /wp-json/tanz-photo/v1/gallery`
  - 参数：
    - `type` = `user` | `brand` | `all`
    - `status` 默认只返回 `approved`
    - 分页参数 `page` / `per_page`
  - 返回：分页后的照片列表（缩略图 URL + meta 信息）。

- `POST /wp-json/tanz-photo/v1/upload`
  - 用于前端表单上传 **用户照片**，只接受登录用户。
  - 请求格式：`multipart/form-data`。
    - 字段草案：
      - `file`：图片文件，仅允许 `image/webp`（扩展名 `.webp`）；
      - `region`：必填，国家或地区名；
      - `location`：可选，更具体的位置（城市等）；
      - `nickname`：可选，上传者昵称；
      - `bike_model`：可选，车款/轮组信息；
      - `notes`：可选，简短说明；
      - （内部）`source`：服务器端标记为 `web-form`。
  - 行为（高层）：
    - **权限与登录**：
      - `permission_callback` 要求 `is_user_logged_in() && current_user_can('upload_files')`；
      - 未登录或无权限时返回 `401/403`，前端显示「请登录后上传」。
    - **文件校验**：
      - 前端：仅通过 `accept="image/webp"` 和文案提示引导用户上传 WEBP，**不做前端裁剪/压缩**；
      - 后端：只接受 `image/webp` MIME 类型；
      - 后端：限制文件大小（例如 ≤ 5MB）与像素尺寸（最长边 ≤ 800px）：
        - 使用 `wp_get_image_editor` 检查源图尺寸；
        - 若最长边大于 800px，则在保存前用编辑器缩放到最长边 800px 以内；
      - 非法类型/超限大小 → 返回 400 错误，附带错误代码（如 `invalid_type` / `file_too_large`）。
    - **防 spam（基础）**：
      - 按用户 ID 做简单速率限制，例如：
        - 每个用户每天最多 N 次上传（通过 user meta 或 transient 记录计数）；
        - 同一 IP/用户在短时间内重复请求直接拒绝；
      - 所有上传初始状态一律为 `pending`，不会直接在前端展示。
    - **保存逻辑**：
      - 上传到 WP Media Library（使用 `wp_handle_upload` 或等价函数），返回 attachment ID；
      - 创建 `tanz_photo`：
        - `post_type = tanz_photo`，`post_status = publish`；
        - `tanz_photo_type = user`；
        - `tanz_photo_status = pending`；
        - 写入 meta：`image_id/region/location/nickname/bike_model/notes/source`；
        - 记录 `tanz_photo_submitted_at`（当前时间）。
  - 响应：
    - 成功时返回简要对象，如：`{ id, status: 'pending' }`，供前端提示「已提交审核」。

- `POST /wp-json/tanz-photo/v1/review`
  - 用于管理员在后台审核照片：设置 `status = approved` 或 `rejected`。

---

### 2.3 评论系统数据结构与接口

> 不单独建新表，而是复用 WordPress 原生 `wp_comments` / `wp_commentmeta`：
> - 每条评论通过 `comment_post_ID` 关联到对应的 `tanz_photo`；
> - 使用 `comment_type = 'tanz_photo_comment'` 进行区分；
> - 只在前端展示 `comment_approved = 1` 的评论。

- **评论字段（核心）**：
  - `comment_post_ID`：关联 `tanz_photo` 的 post ID；
  - `comment_author`：评论显示名（默认取当前登录用户的 `display_name`）；
  - `comment_author_email`：当前用户邮箱，仅后台可见；
  - `comment_content`：评论正文（纯文本或有限 HTML）；
  - `comment_date_gmt`：创建时间；
  - `comment_approved`：`0`=待审核，`1`=通过，`spam`/`trash` 等继承 WP 规则；
  - `comment_type`：固定为 `tanz_photo_comment`，便于筛选。

- **扩展字段（commentmeta，可选）**：
  - `tanz_comment_location`：评论者位置（如国家/城市），可由前端提供；
  - 其他与构建相关的简短标签，后续再定。

- **GET /wp-json/tanz-photo/v1/comments**
  - 用途：为前端 Lightbox 加载某张照片下的评论列表。
  - 参数：
    - `photo_id`：必填，目标 `tanz_photo` 的 ID；
    - `page` / `per_page`：分页控制，默认 `page=1`，`per_page=20`；
  - 行为：
    - 验证 `photo_id` 是否存在且 `post_type=tanz_photo`；
    - 仅查询 `comment_type='tanz_photo_comment'` 且 `comment_approved=1` 的评论；
    - 按 `comment_date_gmt` 倒序返回；
    - 通过响应头返回总数与总页数（如 `X-WP-Total-Comments`）。
  - 返回示例字段：
    - `id`：comment ID；
    - `author`：显示名；
    - `content`：正文（已做基本清洗）；
    - `date_gmt`：时间；
    - `location`：来自 `tanz_comment_location` 的附加信息（如有）。

- **POST /wp-json/tanz-photo/v1/comments**
  - 用途：登录用户给某张 `tanz_photo` 发表一条评论。
  - 权限：
    - `permission_callback` 要求用户已登录（例如 `is_user_logged_in()`），并具备基础 `read` 能力；
    - 未登录返回 `401/403`，前端提示需登录。
  - 请求字段（JSON 或 `application/x-www-form-urlencoded`）：
    - `photo_id`：必填，目标 `tanz_photo` ID；
    - `content`：必填，评论正文；
    - `location`：可选，展示在评论中的地区说明；
  - 行为（高层）：
    - 验证 `photo_id` 合法，并确认其 `post_type=tanz_photo`；
    - 通过 `wp_get_current_user()` 自动填充 `comment_author` / `comment_author_email`；
    - 利用 `wp_insert_comment()` 创建评论：
      - `comment_post_ID = photo_id`；
      - `comment_type = 'tanz_photo_comment'`；
      - `comment_approved = 0`（默认需审核）；
    - 若提供 `location`，则写入 commentmeta `tanz_comment_location`；
  - 响应：
    - 成功时返回 `{ id, status: 'pending' }`，提示用户“评论已提交，等待审核”；
    - 后续前端只在拉取 `comments` 时展示已审核通过的条目。

## 3. 新插件结构（tanzanite-photo-gallery）

> 目标是让图片管理完全独立于 `tanzanite-setting` 插件，降低耦合。

### 3.1 插件职责
- 注册 `tanz_photo` 自定义 post type。
- 注册相关 meta 字段（使用 `register_post_meta`）。
- 注册 REST API 路由：gallery / upload / review。
- 提供一个简单的后台管理 UI：
  - 列出 `pending` 照片
  - 审核通过/拒绝
  - 简单过滤（按类型/状态）。

### 3.2 插件目录
- `tanzanite-photo-gallery/`
  - `tanzanite-photo-gallery.php` （入口）
  - `includes/`
    - `class-tpg-post-type.php` （注册 CPT 和 meta）
    - `class-tpg-rest.php` （REST 路由）
    - `class-tpg-admin.php` （后台审核页面）
  - `assets/` （后台 UI 样式/图标）

---

### 3.3 审核流程（状态与流转）

- **状态定义（基于 `tanz_photo_status`）**：
  - `pending`：
    - 刚创建的照片（无论是用户上传还是管理员后台添加），尚未审核；
    - 不在前端 gallery 中展示。
  - `approved`：
    - 审核通过的照片；
    - 可通过 `GET /gallery` 接口返回，并出现在 Picture warehouse 页中。
  - `rejected`：
    - 审核未通过；
    - 不在前端展示；可保留记录和 `rejected_reason` 供后台参考。

- **典型流转路径**：
  1. **创建**：
     - 用户上传：接口 `POST /upload` → 创建 `tanz_photo`，`status=pending`；
     - 品牌内部上传：管理员在后台新建 `tanz_photo` 或通过接口导入，建议默认也为 `pending`。
  2. **审核**（后台操作）：
     - 管理员在 "Photos" 列表中，查看 `pending` 项。
     - 点选单条记录或批量勾选 → 选择 `Approve` 或 `Reject`：
       - `Approve`：
         - 设置 `tanz_photo_status=approved`；
         - 记录 `tanz_photo_approved_at`（当前时间）。
       - `Reject`：
         - 设置 `tanz_photo_status=rejected`；
         - 可选填写 `tanz_photo_rejected_reason`。
  3. **更新 / 下架**：
     - 若管理员需要临时从前端移除某张已上线照片，可：
       - 将 `approved` 改回 `pending` 或 `rejected`；
       - 前端通过接口只读取 `approved`，自然不再展示该图片。

---

### 3.4 后台列表视图（管理 UI）

- **列表位置**：
  - 在 WP 管理菜单中新增项 `Tanzanite Photos`（或更短名称），指向 `tanz_photo` 的列表页面。

- **列表列（Columns）**：
  - `Thumbnail`：
    - 显示一张小缩略图，方便快速识别内容。
  - `Title`：
    - 使用文章标题；点击可进入编辑页面。
  - `Type`：
    - 显示 `User` / `Brand`，来源于 `tanz_photo_type`。
  - `Region`：
    - 组合展示 `tanz_photo_region` + `tanz_photo_location`（如存在）。
  - `Status`：
    - `Pending` / `Approved` / `Rejected`，直观呈现当前审核状态。
  - `Submitted`：
    - 显示 `tanz_photo_submitted_at` 或文章创建时间。
  - `Approved at`（可选）：
    - 若 `approved`，展示 `tanz_photo_approved_at`；
  - `Source`（可选）：
    - 展示 `tanz_photo_source`，帮助区分上传入口。

- **过滤与视图（Filters）**：
  - 顶部筛选：
    - 按 `Status`：`All` / `Pending` / `Approved` / `Rejected`；
    - 按 `Type`：`All` / `User` / `Brand`。
  - 搜索框：
    - 支持按标题、昵称、地区等关键字搜索（使用 WP 默认搜索能力）。

- **批量操作与单条操作**：
  - 批量操作（Bulk actions）：
    - `Approve`：将选中项状态改为 `approved`；
    - `Reject`：将选中项改为 `rejected`；
    - `Move to Trash`：使用 WP 默认删除行为（需要与 `status` 状态区分）。
  - 单条记录行内操作：
    - `Edit`：进入普通编辑界面，可修改 meta 字段。
    - `Quick Approve` / `Quick Reject`：无需进入详情页即可变更状态。

- **权限（capabilities）简要说明**：
  - 普通编辑者/店铺运营：
    - 可查看 `tanz_photo` 列表和详情；
    - 可执行 `Approve` / `Reject` / `Edit`；
  - 一般作者或订阅用户：
    - 默认无权访问后台该列表，只能通过前端接口上传（如允许的话）。

### 3.5 实现状态（插件骨架）

- 已在 `wp-plugin/` 下创建独立插件：`tanzanite-photo-gallery`：
  - 入口文件：`tanzanite-photo-gallery.php`；
  - 包含三个主要类：`TPG_Post_Type`、`TPG_REST`、`TPG_Admin`。
- 现有功能（骨架级别）：
  - 注册 CPT `tanz_photo`，并将所有设计中的 meta 字段通过 `register_post_meta` 注册；
  - 注册 REST 路由：`/gallery`、`/upload`、`/review`；
  - `/gallery` 已实现基础查询逻辑：支持按 `type`（user/brand/all）、`status`（pending/approved/rejected/all）、`page`、`per_page` 查询 `tanz_photo` 并返回元数据数组；
  - 在后台侧边栏添加 `Tanzanite Photos` 菜单，并在列表中显示 `Title / Type / Region / Status / Date` 等基本列。

## 4. 开发阶段规划（更新：2025-12-06）

**Phase 1：静态骨架与基础展示 (已实现)**
- 页面布局：两栏结构（Riders / Brand）
- 数据展示：对接 `/gallery` 接口，支持分页（Show more）。
- Lightbox：大图浏览，支持左右切换。

**Phase 2：用户上传 (已实现)**
- 上传表单：支持 Region, Location, Nickname 等字段。
- 接口对接：`POST /upload`。
- 状态处理：上传后进入 Pending 状态。

**Phase 3：评论与互动 (已实现)**
- 评论展示：Lightbox 中加载评论列表。
- 评论发表：支持登录用户发表评论（需审核）。
- 分享功能：支持 Copy Link，社交按钮占位。

**Phase 4：推荐与管理 (进行中)**
- 产品推荐：已实现 "Like This? Get The Same Build" 链接展示（基于 `productRefs`）。
- 后台管理：`tanzanite-photo-gallery` 插件已支持基础审核流程。
