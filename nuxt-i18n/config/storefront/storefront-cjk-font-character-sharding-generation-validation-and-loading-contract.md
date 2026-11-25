# CJK 与扩展字符字库分片、生成和加载契约

更新日期：2026-11-19。

前台保留原内置 Maple UI 字体、字体族、字形、字宽和 `font-display: block`。本次只细化 `MapleUICJK` 的文件粒度，不采用系统字体替换，不增加页面等待字体的遮罩，也不修改主导航或底部菜单的加载时序。

## 原来的分片为什么仍下载 6.4 MB

此前已将拉丁、CJK 和阿拉伯文等覆盖字体分开，但 CJK、日韩和部分希腊／西里尔／扩展字符仍共用一个 6,403,884 B 的 WOFF2 文件。CSS `unicode-range` 决定是否下载整个文件，不能让浏览器在一个 WOFF2 文件内按需取字。命中任一相关字符，就会请求整个合并文件。此前分片又沿用了旧 CSS 手写范围，因而虽然覆盖这些范围里的 40,001 个字形，却遗漏了原始 Maple UI cmap 中不在这些范围内的字符。

## 现在的文件与 CSS

同一个 `MapleUICJK` 字体族由多个 `@font-face` 声明组成，每个声明指向独立、带内容哈希的 WOFF2 文件，并声明互不重叠的准确字符范围。分组覆盖来源改为原始字库完整 cmap，不再使用手写 Unicode 大范围作为完整性基准；只排除已经由 `MapleUILatin` 提供且通过原字形校验的 728 个码点。原始 Maple UI cmap 有 42,172 个码点，因此 CJK／扩展分片必须覆盖剩余全部 41,444 个码点，包括框线、部首、数学字母、私用区和全角标点等此前遗漏项。汉字、Hangul、假名／CJK 标点、希腊／西里尔和其余扩展字符分别分组。现有导航、首页、页脚和 FAQ 等命名空间中的常见界面字符优先集中；其余字符按 Unicode 顺序分组，使范围声明保持紧凑。

每片最多 512 个对外字符，压缩后最多 128 KiB；超过体积上限会继续拆分。当前清单共 104 个 CJK／扩展分片，完整覆盖原字库 cmap 中未分配给 Latin 分片的 41,444 个码点。manifest 记录原字库 SHA-256、字节数、码点总数、Latin 排除数和 CJK 分片码点数；检查从原始 WOFF2 重新读取 cmap，并核对每个字体文件的实际 cmap 与声明范围一致。检查还会扫描 `header`、`common`、`dockMenu`、`cookieConsent`、`home`、`footer`、`nav`、`menu`、`shopCategoryMenu`、`productCategoryNavigationCards`、`storefrontDataNotice` 和 `faq` 命名空间：这些当前界面文案使用到的原 Maple UI 字符必须落在常用分片。以后新增翻译字符若只落入剩余覆盖分片，字体门禁会要求重新生成，避免常用页面静默多请求一个低频分片。这样原始字库增加字符、界面文案变化或生成清单错误时，检查不会因为旧 CSS 范围本身过时而错误通过。

浏览器只下载页面实际排版命中的分片。页面已参与排版的屏外内容也可能触发字体请求，不能把字体按字符下载解释为“只下载视口内的字”。后台内容新增生僻字时，对应剩余覆盖分片依然可用；不需要等下一次构建补字。

`app/assets/css/tailwind.css` 的生成区和 `public/fonts/maple-ui.css` 的字体声明保持一致，后者继续供 Stripe 使用。保留 `MapleUILatin, MapleUICJK` 基础字体栈和各语言的内置覆盖字体，其他语言的字体加载入口不变。不会预加载全部 CJK 分片。

首页首屏另外按当前 locale 预加载实际命中的常用界面字符分片：简中加载 2 个 Han 分片和 1 个 CJK 标点分片，日文加载 1 个 Han 分片和 1 个 CJK 标点分片，韩文加载 1 个 Hangul 分片，俄文加载 1 个 Greek/Cyrillic 分片。其余完整覆盖分片不预加载；页面中已参与排版的屏外文本仍会按实际字符请求所需字体。

原 6.4 MB 文件移动到 `scripts/font-source-assets/`，仅用于生成与完整性校验，不复制到 `.output/public/fonts/`，也不作为线上回退文件。后台管理端自己的完整字体资源不属于此次前台修改。

## 生成方式

生成脚本为 `scripts/generate-storefront-cjk-and-extended-font-character-shards.py`，要求 Python 3.12 或更新版本。`scripts/requirements-storefront-font-shard-generation.txt` 固定 `fonttools==4.68.0` 和 `brotli==1.2.1`；生成器运行时会核验安装版本，并拒绝用其他版本写出带错误版本记录的产物。在 `nuxt-i18n` 目录运行：

```powershell
py -3.12 -m pip install -r scripts/requirements-storefront-font-shard-generation.txt
py -3.12 scripts/generate-storefront-cjk-and-extended-font-character-shards.py
```

可用 `--decoded-source-ttf /absolute/path/to/original-decoded.ttf` 复用原字库预解码结果以加快生成；正式检查仍会将每个分片与仓库保留的原 WOFF2 比较。生成时保留 hinting 与布局特性，不转换字体外观。

提交生成的 WOFF2、字符范围清单与两处 CSS。普通生产构建只验证已提交的产物，不执行 Python，也不因翻译或后台内容变化而自动重新分配所有分片、破坏缓存。需要调整常见界面字符分组或更新原字体时，显式重新生成并重新验收。

## 验收约束

`storefront-font-character-shard-contract.ts` 是各项前台字体检查共用的清单入口。完整检查包括：

- 原字库与每片 WOFF2 的 SHA-256、内容寻址文件名、体积上限和字符数量。
- 原始 Maple UI cmap 减去 Latin 字体 cmap 后的全部码点都恰好对应一个分片，无丢字、无范围重叠；每片文件实际 cmap 与声明范围完全一致。
- 当前原字库不含 cmap format 14 Unicode 变体序列；生成器和检查会拒绝未实现逐序列覆盖校验的新 format 14 来源，避免把变体序列误当普通码点完整覆盖。
- 当前主界面翻译所需的原字库字符都在常用分片；更改界面翻译后若分组陈旧，门禁会要求重新生成。
- 逐字比较原字库与分片的轮廓、横／竖字宽，并检查字体度量及原 Latin 核心的一致性。
- 全部 20 种语言的 locale 与 message 文本继续由批准的内置字体栈覆盖。
- 前台与 Stripe 声明一致，保持 `block`，构建压缩后的 CSS 范围仍与清单一致。
- 原 critical CSS 的 16 KiB gzip、阻塞 CSS 的 20 KiB gzip 上限保持不变。
- 生产浏览器验证中文、日文、韩文、西里尔和生僻字只下载对应分片，字体像素与布局与原字库一致。

字体预检清单继续使用原 schema 与六个批准字体族，但 `faces` 列表现在按真实文件列出多个 `MapleUICJK` 字体面。后台校验允许该字体族有不同文件名的字符分片，仍拒绝重复文件、其他字体族重复和缺少 CJK 字符范围；管理端列表按文件名作为唯一键。

每项检查的入口分别是 `npm run check:font-preflight`、`check:font-performance`、`check:font-coverage`、`test:font-policy`、`test:font-locales` 以及现有 `tests/storefront-font-contract.spec.ts`。`npm run build` 和 `npm run check:production-artifacts` 继续执行相关生产门禁。

拆分后全部文件的总量可能增加，因为每片都有字体表与布局信息开销；这不等于单个页面会下载全部文件。验收应统计具体页面的实际字体请求及 CSS 成本，不能仅比较字库目录总量，也不能将本地测量直接换算为线上 PageSpeed 分数。

`font-display: block` 的浏览器隐藏期是有限的。此处保留既有行为，不声称它在任意慢网或下载失败时都会无限等待内置字体。

## 2026-11-19 本地生产验证

完整 `npm run build`、`npm run check:production-artifacts`、脚本类型检查、字体覆盖／原字形检查、后台字体预检单元测试及 5 项现有 Chromium 字体测试通过。浏览器包含中日韩、西里尔、生僻字、Latin 的原字库像素／布局对比，确认 FAQ 的“答”字也只触发常用分片。

当前 CJK 字符分片共 104 个，总计 7,631,964 B，最大片 130,912 B。字库总量有分片表开销，但中文首页没有下载全部分片。`maple-ui.css` 为 48,992 B，gzip 10,171 B；它是独立的 Stripe 字体样式入口。前台入口的完整 CSS 与 inline critical CSS 同样包含字符范围，因此不能只看 WOFF2 的节省。当前完整生产构建的首页 entry 加页头 critical CSS 为 15.86 KiB gzip，阻塞 CSS 为 1.57 KiB gzip，均通过原预算。

本地生产预览连接已有本地后端，390×844、DPR 3、Slow 4G、CPU 4×、禁用 HTTP 缓存、不滚动不点击的单次采样：

| 字体下载响应体 | 原记录 | 本次 |
| --- | ---: | ---: |
| CJK／扩展文件 | 6,403,884 B | 118,440 B，3 片，减少 98.15% |
| 加上 Latin 与 Latin accent | 6,457,116 B，3 个文件 | 171,672 B，5 个文件，减少 97.34% |

本次 trace 的 LCP 为 6,080 ms，TTFB 565 ms，FCP 4,164 ms，CLS 约 0.02，`window.load` 12,558 ms，最晚字体完成于 7,198 ms。所有字体请求成功；没有操作时底部业务菜单未挂载。`block` 的有限隐藏期在该仿真下仍可能结束，保留该产品约束，不将报告中的字体策略提示自动改为 `swap`。

此前 8.745 秒 LCP／46.5 秒 load 的记录使用了旧生产预览；它没有加载已生成的 critical CSS。当前预览重新启动后正确应用 critical CSS，并显式连接本地后端，因此这不是只改变字体的严格 A/B 性能比较，不能把 LCP 差值全部归因于字体分片。本次可直接核实的是同一路径实际字体文件下载量下降，以及原字库一致性检查通过。线上 LCP／PageSpeed 仍需部署后的同条件测量。
