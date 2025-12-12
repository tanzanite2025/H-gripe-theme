# Warranty Check QR 扫码方案草稿

## 需求背景
1. 当前 `/support/warranty-check` 页面依赖手动输入序列号 / 质保码，移动端用户输入体验不佳。
2. 期望提供“调用摄像头扫描二维码”功能，直接填充序列号并触发质保查询，从而：
   - 降低误输入概率。
   - 缩短保修查询路径，尤其在线下现场或维修场景。
3. 需要先在文档中确认范围与实现方式，后续再落地开发。

## 目标体验
1. 在已登录状态下的查询表单附近新增“扫描二维码”入口（按钮或卡片）。
2. 点击后弹出摄像头预览（全屏或模态），提示用户对准产品上的二维码/条形码。
3. 扫描成功自动回填 `productCode`，可立即点击“Check”按钮或自动发起查询（需二次确认）。
4. 用户可随时关闭扫描，再次回到表单。

## 技术实现概述
| 模块 | 说明 |
| --- | --- |
| 权限获取 | 使用 `navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })`，需 HTTPS。 |
| 视频渲染 | 通过 `<video>` + `<canvas>` 组件，每帧绘制并解析。可封装成 `QrScannerPanel`。 |
| 解码库 | 候选：`jsQR`（轻量纯 JS）、`@zxing/browser`（功能丰富，支持多码制），根据 bundle 体积和精度取舍。 |
| 状态管理 | 在 `WarrantyCheckPanel` 内维护 `isScanning`, `scanError`, `scannedCode`；成功后写入 `productCode`。 |
| UI 框架 | 采用现有 Tailwind + 暗色主题，提供清晰的“开始/停止扫描”按钮与辅助提示。 |

## 兼容性与限制
1. **HTTPS 必须**：浏览器只在安全上下文暴露摄像头。部署与测试需保证 `https://` 或 `localhost`。 
2. **设备支持差异**：
   - iOS Safari 11+ 支持 `getUserMedia`，但权限弹窗严格，应在 UI 上解释用途。
   - 桌面电脑无摄像头时需隐藏入口或提示“设备不支持”。
3. **性能**：扫码时需在 `requestAnimationFrame` 中循环解析，应限制分辨率（例如 640x480）以避免 CPU 占用过高。
4. **隐私与安全**：只在本地解析数据，不上传视频帧；必要时在隐私政策中补充说明。

## 拟定组件结构
```
WarrantyCheckPanel.vue
└── QrScannerModal.vue
    ├── VideoPreview (展示摄像头画面)
    ├── Overlay (绘制取景框/提示)
    └── Footer actions (Start/Stop, 使用手电筒等可选能力)
```
- `QrScannerModal` 通过 `v-model:open` 控制。
- 扫描成功 emit `"scanned"` 事件，父组件将值写入 `productCode`。
- 失败场景（权限拒绝、未检测到二维码）通过 toast 或 banner 告知。

## 交互与文案要点
1. **入口按钮文案**：`Scan QR` / `扫描二维码`，附加说明“自动识别包装上序列号”。
2. **模态提示**：
   - “允许使用摄像头以扫描质保码”。
   - “确保环境明亮，保持二维码在取景框内”。
3. **权限拒绝**：提供“打开系统设置”或“改为手动输入”的 fallback。

## 实施步骤（待确认）
1. 引入二维码解析库（推荐 `jsqr`，体积 ~30KB）。
2. 封装 `useQrScanner` composable：负责启动摄像头、绘制、解析、停止。 
3. 在 `WarrantyCheckPanel` 中集成 `QrScannerModal`，并与原有表单状态打通。
4. 增加基础测试：
   - Chrome Android / Safari iOS 实机扫描
   - 权限拒绝 / 无摄像头场景
5. 文档化新增权限说明，更新 FAQ。

## 待用户确认的问题
1. **二维码内容格式**：是否只包含纯序列号，或包含 URL（例如 `https://.../?code=XXXX`）。
2. **扫码后流程**：自动提交 vs. 仅填入并等待用户确认。
3. **UI 形式**：模态（全屏）还是内嵌小窗口？是否需要扫码历史记录？
4. **多语言需求**：是否需要增加对应 i18n 文案键，方便后续翻译。

> 请在确认上述问题后告知，我会再更新本文件并开始实现。
