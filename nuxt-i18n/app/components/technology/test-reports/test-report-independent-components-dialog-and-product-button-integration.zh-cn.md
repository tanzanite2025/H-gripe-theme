# 独立测试组件、通用弹窗与商品详情按钮接入说明

七项轮圈测试各有独立组件，位于 `app/components/technology/test-reports/rim/`。各组件绑定自身测试定义，并通过公共布局自行加载 `testReportRim` 翻译、显示素材及管理移动端切换。报告页只负责组合这些组件；独立使用时无需挂载报告页。

按钮和弹窗只依赖 `testReportBlockComponentRegistry.ts` 中的组件、翻译命名空间及标题键，不读取轮圈素材配置。弹窗通用文案使用独立的 `testReportDialog` 命名空间；组件通过 `presentation="dialog"` 自行调整外观，弹窗不查询业务组件的内部 CSS 类。

| 测试标识 testId | 独立组件 |
| --- | --- |
| `rim-strength-test` | `RimStrengthTestReportBlock.vue` |
| `rim-impact-test` | `RimImpactTestReportBlock.vue` |
| `rim-inflation-test` | `RimInflationTestReportBlock.vue` |
| `rim-radial-load-test` | `RimRadialLoadTestReportBlock.vue` |
| `rim-spoke-hole-tension-test` | `RimSpokeHoleTensionTestReportBlock.vue` |
| `rim-cross-section-inspection` | `RimCrossSectionInspectionReportBlock.vue` |
| `rim-xray-flatness-inspection` | `RimXRayFlatnessInspectionReportBlock.vue` |

## 直接嵌入一个测试块

```vue
<template>
  <RimImpactTestReportBlock heading-tag="h3" />
</template>

<script setup lang="ts">
import RimImpactTestReportBlock from '~/components/technology/test-reports/rim/RimImpactTestReportBlock.vue'
</script>
```

独立组件默认生成唯一实例 ID，允许同一测试同时出现在页面和弹窗中。报告页通过 `use-page-anchor` 保留原有测试锚点；同一页面只为每项测试的一个实例启用此属性。

可以使用 `header-actions` 插槽添加按钮。布局根据组件内容区的可用宽度切换：达到 896px 时显示描述、视频、四宫格三列；较窄时描述独占一行，媒体通过两颗圆点或滑动切换。

## 在商品详情组件中放一个弹窗按钮

```vue
<template>
  <TestReportDialogTriggerButton test-id="rim-impact-test" label="查看冲击测试" />
</template>

<script setup lang="ts">
import TestReportDialogTriggerButton from '~/components/technology/test-reports/TestReportDialogTriggerButton.vue'
</script>
```

按钮与通用弹窗外壳一起就绪，点击后立即显示弹窗，通过 `app/utils/testReportBlockComponentRegistry.ts` 按需加载指定测试组件，展示与报告页一致的内容。`test-id` 可以响应式更新，按钮始终打开当前指定的测试。

`class`、`style`、`disabled`、`data-*` 和 `@click` 等属性及事件会透传给实际的按钮元素。默认插槽可以替换按钮文案。使用 `label` 或文字插槽时，可访问名称跟随实际文案；如果插槽只有图标，请通过 `aria-label` 指定名称。显式传入的 `aria-label` 优先于默认名称。按钮自带弹窗状态，不依赖商品页的标签页、路由查询参数或全局选中测试。

按钮、弹窗及测试内容的 setup 均保持同步，不要求接入位置提供 `Suspense`，可在全局商品弹层中动态挂载。`useIndependentTestReportMessages` 为同一 Nuxt 应用的多个实例共用正文翻译加载，在 SSR 时按请求隔离，并通过 Nuxt 状态把翻译交给客户端，避免首次水合缺失文案。少量弹窗通用文案随外壳提供，中英文以外沿用英文回退。

测试组件或正文翻译加载失败时，弹窗内显示失败提示和重试按钮，外壳仍可正常关闭或返回。恢复网络后可直接重试；关闭后到达的延迟响应不会重新打开弹窗。

弹窗支持关闭按钮、Escape、点击遮罩和浏览器返回；关闭时停止视频、恢复原有页面滚动状态，并把焦点还给打开按钮。移动端支持安全区域和弹窗内部滚动，关闭按钮固定在内容滚动区外。测试弹窗以 push 模式加入已有弹层栈：从商品弹层中打开后，返回一次只关闭测试，返回第二次才关闭商品弹层。页面和弹窗中的同一测试保持各自独立的媒体切换状态。

## 已有页面自行管理弹窗状态

也可以直接导入 `TestReportBlockDialog.vue`，通过 `v-if` 和 `test-id` 控制挂载，监听 `close` 关闭。每次重新挂载会重置该弹窗内的媒体切换状态。如果使用自己的按钮，可通过 `return-focus-element` 传入按钮 DOM 元素；未传时使用弹窗挂载前的焦点元素。

外部将 `v-if` 设为 false 时，也会正常关闭弹层并消耗弹窗产生的同地址历史记录，避免返回键额外停留一次。真正的路由跳转由全局弹层栈统一清理，无需商品详情再清理弹窗状态。

## 后续新增轮组等测试

先实现该测试的独立组件，支持 `TestReportBlockPresentationProps` 中的标题级别、页面锚点及弹窗外观约定。再将测试标识加入 `TestReportBlockId`，并在注册表补充异步组件、翻译命名空间及标题键。轮圈内部配置使用自己的 `RimTestReportBlockId`，新增其他类别不会扩大轮圈配置接受的标识范围。

完成注册后，已有 `TestReportDialogTriggerButton` 和 `TestReportBlockDialog` 无需修改。目前注册的内容为上表七项轮圈测试；轮组和编轮的测试内容尚未在这里注册。

## 素材与后台正文

素材仍在 `app/data/rimTestReportTestDefinitionsAndMediaConfiguration.ts` 配置，按测试项目的 `videos/`、`photos/` 目录存放。修改一份测试素材会同时更新报告页、独立组件和弹窗。

打气测试对应 `id: 'inflation'`、按钮标识 `rim-inflation-test`，素材目录为 `public/testreport/rimtestreport/inflation-test/`。当前保留视频和四张照片占位；具体压力、保压时长及合格阈值按该型号经确认的试验记录补充，不预设数值或认证标准。

商品详情分页目前通过 `SafeRichText` 显示后台 HTML。上述 Vue 组件按钮应放在前端模板或插槽中；如需让后台正文指定按钮位置，需另行接入受控触发标记及对应的渲染处理。

## 可重复运行的浏览器回归

启动 Nuxt 开发服务后，在 `nuxt-i18n` 目录运行：

```powershell
playwright-cli open http://127.0.0.1:10240/zh_cn/technology/wheelset-test-report
playwright-cli run-code --filename=scripts/test-report-independent-components-network-recovery-and-dialog-lifecycle-browser-regression.js
playwright-cli close
```

脚本在独立 Vue 根节点中挂载按钮和弹窗，不新增测试路由；覆盖组件和翻译请求失败后的重试、加载中关闭、20 次快速关闭再打开、打气测试的独立按钮接入、两个原生弹窗的底层卸载，以及离开页面时清理资源。请求被主动阻断时，浏览器的网络错误属于预期；脚本要求没有未捕获的浏览器异常，且场景结束后没有残留弹窗、滚动锁和弹层历史状态。
