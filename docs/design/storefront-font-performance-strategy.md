# Storefront Font Performance Strategy

Last updated: 2026-09-04

本文档专门说明前台如何在保留 Maple UI 字体美感的同时，让首屏字体尽早进入网络队列。
后续如果 FCP、LCP、字体显示或多语言字体分片跑偏，先按这里排查。

## 目标

前台首屏不得为了速度退回系统字体。当前批准基线是：

```text
MapleUILatin, MapleUICJK
font-display: block
项目自托管 WOFF2 only
```

性能目标是：小体积 Latin 首屏分片必须在第一份 HTML 里被浏览器发现；6.4 MB CJK
大分片必须留在首屏关键路径之外，除非当前渲染文本真的命中它的字符范围。

## 当前实现

1. `nuxt-i18n/public/fonts/MapleUI-Latin.00af3fec5b34.woff2` 是首屏分片，大小
   48,272 bytes。默认英文前台只应该 preload 这一份字体。
2. `nuxt-i18n/nuxt.config.ts` 在 `app.head` 静态声明 Latin preload。
3. `nuxt-i18n/server/plugins/05-font-preload-priority.server.ts` 处理 SSR HTML
   响应，把 Latin preload 去重后移动到 `<head>` 起始位置，排在 Nuxt 生成的 CSS
   link 前面。这样字体不会等 CSS 解析完才被发现。
4. `nuxt-i18n/app/app.vue` 只允许通过
   `storefrontAdditionalFontPreloadLinkForLocale` 追加语言覆盖分片，例如 Latin
   accents、Arabic、Devanagari、Thai。
5. `MapleUI-CJK.f8ce6d72e8cb.woff2` 永远不做 preload。它只通过
   `unicode-range` 被中文、日文、韩文或扩展字符按需触发。
6. 首页 hero/LCP 图片由 `StorefrontImage` 控制，当前激活的首屏图片应带
   `preload`、`fetchpriority="high"` 和桌面/移动端 `media` 条件。

## 禁止事项

不要把 Maple UI 替换成系统字体、通用字体族、平台 UI 字体或 CDN 字体。这里的原则是
“Maple UI 秒开”，不是“先丑后美”。

不要把 Maple UI 的 `font-display` 改成 `swap`、`fallback` 或 `optional`。本项目故意
避免首屏先闪一下系统字体。

不要重新引入 `@nuxt/fonts`、第三方字体 CDN、`local()`、`data:font`，或已经退役的
`StorefrontSystem*` 字体名。

不要 preload `MapleUI-CJK.f8ce6d72e8cb.woff2`。英文首屏里只要混入一个中文标点、
全角空格、日文、韩文或其他命中 CJK unicode-range 的字符，就可能把 6.4 MB 字体拉进
首屏网络队列。

不要删除 Nitro preload-priority 插件，除非 Nuxt 自身已经能保证 Latin preload 排在所有
生成 CSS link 前面，并且测试已经更新来证明这个行为。

## 原理

浏览器默认通常是先下载 HTML，再解析 CSS，等 CSS 里的 `@font-face` 被发现后才请求字体。
慢速移动网络下，这会白白浪费往返时间。

当前做法是把 48 KB Latin 分片放进初始 HTML，并由 SSR 插件挪到 generated CSS 之前。
浏览器拿到 HTML 后就能并发请求字体；同时继续保留 `font-display: block`，避免系统字体
闪现。

CJK 大分片不进默认关键路径，因为它的 `@font-face` 有 `unicode-range`。英文页面正常
只会请求 Latin 分片。

## 排障

如果 FCP 慢，或 Lighthouse 显示字体等待时间长：

1. 任何 server plugin 或 head 修改后，先重启 Nuxt。
2. 查看 SSR HTML 的 `<head>`。
3. 确认第一个字体 preload 是：

```html
<link rel="preload" href="/fonts/MapleUI-Latin.00af3fec5b34.woff2" as="font" type="font/woff2" crossorigin="anonymous" data-hid="storefront-font-preload-latin">
```

4. 确认它出现在 Nuxt 生成的 stylesheet link 之前。
5. 确认没有 CJK 字体 preload。

如果英文首页下载了 CJK 分片：

1. 检查渲染后的首屏文本是否包含 CJK/extended code point。
2. 特别查全角空格、中文标点、日文/韩文字符、从 CMS 复制来的隐藏字符。
3. 把这些字符移出英文首屏，或移到非关键路径内容里。
4. 不要为了修复误下载而删除 `unicode-range`。

如果 LCP 慢：

1. 先确认 Lighthouse 指认的 LCP 元素。
2. 如果是首页 hero 图片，确认当前激活的 `StorefrontImage` 生成了 image preload，并带
   `fetchpriority="high"` 和正确的 `media`。
3. 确认 preload 的 URL 是当前 viewport 实际使用的移动端/桌面端衍生图，不是超大原图。

## 必跑检查

上线字体加载相关改动前，至少跑：

```bash
cd nuxt-i18n
npm run test:font-policy
npm run check:font-performance
npm run check:font-coverage
npm run check:font-preflight
npm run scripts:typecheck
npx nuxi typecheck
```

如果 dev server 正在运行且 Playwright 浏览器可用，再跑：

```bash
cd nuxt-i18n
npm run test:font-contract
```

`npm run check:font-preflight` 会重新生成
`nuxt-i18n/public/_internal/font-preflight.json`。字体合同变更必须带上这份 manifest。

## 更新字体文件

如果字体文件名变了，所有内容寻址引用必须一起改：

```text
nuxt-i18n/app/assets/css/tailwind.css
nuxt-i18n/public/fonts/maple-ui.css
nuxt-i18n/app/utils/storefrontFonts.ts
nuxt-i18n/nuxt.config.ts
nuxt-i18n/server/utils/fontPreloadPriority.ts
nuxt-i18n/scripts/check-font-performance.ts
nuxt-i18n/scripts/generate-font-preflight-manifest.ts
nuxt-i18n/scripts/storefront/font-artifact-contract.ts
nuxt-i18n/tests/storefront-font-contract.spec.ts
docs/design/maple-ui-fonts.md
docs/design/storefront-font-performance-strategy.md
```

然后在生产 build 前跑 `npm run clean`，避免 `.nuxt`、`.output` 或 `dist` 里的旧产物继续
引用退役字体。
