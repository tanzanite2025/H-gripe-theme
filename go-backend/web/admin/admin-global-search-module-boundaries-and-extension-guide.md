# 后台独立搜索的模块边界与扩展方式

底部搜索入口与客服入口各自控制弹窗，只共享按钮外观。搜索当前提供页面查询；默认推荐 10 个常用页面，输入后查询全部有权限访问的导航 TAB。

| 职责 | 文件 | 扩展规则 |
| --- | --- | --- |
| 底栏工具组合 | `src/components/admin/AdminBottomUtilityDockToolActions.vue` | 在此组合搜索、客服及后续工具；主布局只传入权限与导航，不接入各工具内部实现 |
| 底栏容器与排列 | `src/components/admin/AdminBottomUtilityDock.vue` | 负责位置、排列与导航语义，不管理工具状态 |
| 按钮尺寸、焦点样式、角标 | `src/components/admin/AdminBottomUtilityDockActionButton.vue` | 新增工具入口复用按钮，保持单层按钮，不加外围胶囊 |
| 搜索入口与打开/关闭 | `src/components/admin/search/AdminGlobalSearchDockAction.vue` | 保留按需加载和关闭后焦点返回 |
| 数据源组合 | `src/components/admin/search/AdminGlobalSearchDialog.vue` | 新数据源在这里加入，不加入布局或展示组件 |
| 输入、结果、加载状态的展示 | `src/components/admin/search/AdminGlobalSearchDialogContent.vue` | 仅接收数据和发出交互事件，不请求业务接口 |
| 查询输入、键盘跳转与关闭 | `src/composables/search/useAdminGlobalSearchController.ts` | 组合请求管理，保留输入法、原生链接、导航失败与过期跳转的保护 |
| 数据源请求、防抖、取消与失败重试 | `src/composables/search/useAdminGlobalSearchSourceRequests.ts` | 各数据源独立加载与失败；仅重试失败源，保留成功结果和其他正在进行的请求；旧请求不得覆盖新查询、权限变化或关闭后的状态 |
| 数据源与结果契约 | `src/modules/search/adminGlobalSearchResultAndSourceContracts.ts` | 数据源 ID 唯一，结果 ID 在所属数据源内唯一；支持同步及异步结果，异步接口传递 AbortSignal |
| 页面目录与匹配 | `src/modules/search/adminPermissionScopedPageSearchCatalog.ts` | 由过滤权限后的原始导航生成；名称、路由、权限继续统一维护在 adminNavigation，不复制到搜索中 |
| 页面描述与搜索别名 | `src/modules/search/adminPageSearchDescriptionAndKeywordMetadata.ts` | 按导航 ID 增加补充信息；没有专用描述时使用本地化默认描述 |
| 常用页面推荐顺序 | `src/modules/search/adminCommonPageRecommendationSelection.ts` | 推荐与可搜索范围分开；只选择当前有权限的推荐项 |
| 页面数据源适配 | `src/modules/search/adminNavigationPageSearchSource.ts` | 将页面转换为统一结果，保留原生路由链接 |

后续接入订单、商品搜索时，各自建立独立的数据源文件和业务接口适配。服务端仍须校验业务权限；数据源只返回当前用户可访问的记录。将返回数据转换为统一结果，使用 `routeTarget` 指向详情，并在组合组件中加入数据源。通用控制器和结果展示无需增加按业务类型判断的分支。

本地页面查询保持即时执行。远程数据源可以在契约中设置 `queryDebounceMs: 250`，仅对非空查询防抖；默认推荐和用户点击的失败重试即时执行。查询或数据源变化时立即清空旧结果、取消等待中的防抖和正在执行的请求；接口适配必须继续将 `signal` 传给实际请求。各分组分别展示加载状态，已有结果在其他数据源加载或重试期间继续可用。

当前不增加收藏/历史持久化、工具插件注册中心或订单/商品后端查询。通用弹窗文案在 `adminGlobalSearch`，页面描述和推荐标签在 `adminPageSearch`。

## 客服入口与通知的模块边界

| 职责 | 文件 | 扩展规则 |
| --- | --- | --- |
| 客服按钮与弹窗展示 | `src/components/admin/customer-service/CustomerServiceDockInboxAction.vue` | 只消费控制器状态，不维护定时器、WebSocket 或请求 |
| 入口状态与深链接 | `src/composables/customerService/useCustomerServiceDockInboxController.ts` | 管理打开/关闭、路由 inbox 参数、通知打开操作；打开工作台时暂停入口连接，关闭时恢复 |
| 未读计数与刷新合并 | `src/composables/customerService/useCustomerServiceUnreadConversationSummary.ts` | 独立处理防抖、取消请求和迟到响应；释放作用域时取消定时器与请求 |
| 实时传输 | `src/composables/customerService/useCustomerServiceRealtime.ts` | 只处理连接、重连、游标、事件分发及生命周期；通知策略交给通知模块 |
| 通知偏好、权限与静默时段 | `src/composables/customerService/useCustomerServiceNotificationSettingsAndQuietHours.ts` | 入口与工作台复用相同设置；权限申请保留为显式用户操作 |
| 客户消息通知策略 | `src/composables/customerService/notifyCustomerServiceIncomingCustomerMessage.ts` | 统一客户消息筛选、静默时段、声音和桌面提醒开关 |
| 跨标签去重 | `src/modules/customer-service/customerServiceCrossTabMessageNotificationClaims.ts` | 复用原有事件 ID、Web Locks 与 BroadcastChannel 机制 |
| 浏览器声音与桌面通知 | `src/modules/customer-service/customerServiceBrowserMessageNotificationEffects.ts` | 封装浏览器效果，保持业务策略与设置管理分离 |

通知设置继续使用原来的本地存储键；跨标签通知的声明键也保持兼容。会话列表 API 的第三个可选参数支持 `AbortSignal`，已有调用方式仍然有效。布局按客服权限挂载入口，入口卸载后不再连接、更新计数或响应旧通知操作。

修改页面目录或权限后运行页面目录测试；新增异步数据源后运行 `src/composables/search` 下的请求管理与控制器测试，验证防抖、取消、独立加载、失败重试和权限变化。界面改动还需运行 `npm run typecheck`、`npm run build`，检查桌面与移动端的按钮尺寸、输入焦点、原生链接和弹窗滚动。
