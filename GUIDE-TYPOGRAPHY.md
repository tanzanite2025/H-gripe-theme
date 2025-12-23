# Guides 文本与小蓝点使用规范

本规范用于约束 `/guides/*` 页面（包括 Tire Guides / Technical / Wheelset Buyers 等）中文本标题、段落与小蓝点的使用方式，避免样式与语义混乱。

## 一、核心类一览

- **`.sizecharts-section`**  
  - Section 容器，用在每个 TAB 的 `<section>` 上。

- **`.sizecharts-section__title`**  
  - TAB 级主标题，一般为 `h2`，内容居中，无小蓝点。

- **`.sizecharts-section__subheading`**  
  - Section 内的小标题，一般为 `h3`，内容居中，无小蓝点，文本高亮为蓝色。

- **`.sizecharts-section__intro`**  
  - 带小蓝点的“要点式段落”。  
  - 特征：灰色文字、左侧有小蓝点、整体居中。

- **`.sizecharts-section__list`**  
  - 文本列表（`ul` / `li`），每个 `li` 左侧有小蓝点。

- **`.sizecharts-section__list--numbered`**  
  - 有序列表（`ol`），显示 `1. 2. 3.` 数字，不显示小蓝点。

> 说明：`__intro` / `__list` 同时负责文字颜色、对齐与小蓝点，因此“要不要小蓝点”不是随便换 class，而是要换成对应的结构模式。

---

## 二、什么时候用 `__intro`

**适用场景：**

- 需要突出一两句“要点”，但不想写成 `<ul><li>` 列表。  
- 内容偏短，类似 bullet point，适合一条一行展示。  
- 希望这一行**居中 + 带小蓝点**。

**示例：**

```html
<section class="sizecharts-section">
  <h2 class="sizecharts-section__title">Installation</h2>
  <p class="sizecharts-section__intro">
    Always check the rim bed and tire bead before inflating.
  </p>
  <p class="sizecharts-section__intro">
    Use a calibrated torque wrench for critical fasteners.
  </p>
</section>
```

**不适用场景：**

- 长段落说明性文本（例如一整段故事、背景介绍）。  
- 作为“1. / 2. / 3.” 这种小标题本身。  
- 已经在 `ol/ul` 列表内部的内容（那应使用 `__list`）。

这类场景，不要用 `__intro`，而是用普通 `<p>` 或在局部 section 上单独定义样式。

---

## 三、什么时候用 `__list`（带小蓝点的列表）

**适用场景：**

- 标准多条要点列表，例如「Accepted Components」「Important Notes」。
- 需要每条左侧有一致的小蓝点，且多条之间行距统一。

**基本用法：**

```html
<ul class="sizecharts-section__list">
  <li>Wheelsets (including hubs)</li>
  <li>Other related accessories (please confirm with us in advance)</li>
</ul>
```

**1/2/3/4 标题 + 子项的推荐结构：**

```html
<ol class="sizecharts-section__list sizecharts-section__list--numbered">
  <li>
    <strong>Service Process</strong>
    <ul class="sizecharts-section__list">
      <li><strong>Shipping:</strong> ...</li>
      <li><strong>Confirmation:</strong> ...</li>
      <li><strong>Assembly:</strong> ...</li>
      <li><strong>Testing:</strong> ...</li>
      <li><strong>Delivery:</strong> ...</li>
    </ul>
  </li>
  <li>
    <strong>Accepted Components</strong>
    <ul class="sizecharts-section__list">
      <li>Wheelsets (including hubs)</li>
      <li>Other related accessories ...</li>
    </ul>
  </li>
</ol>
```

- 外层 `ol.sizecharts-section__list--numbered`：负责 `1. / 2. / 3.` **数字**，**不带蓝点**。
- 内层 `ul.sizecharts-section__list`：负责每条子项的小蓝点。

这样可以实现：

- 「1. Service Process」这一行是纯标题，无蓝点。  
- 标题下方每一条子项都有小蓝点。

---

## 四、什么时候用普通 `<p>`（不带小蓝点）

**适用场景：**

- 普通说明性文本：背景介绍、服务条款、补充说明等。  
- 不希望出现蓝点，只是简单的正文段落。

**写法：**

```html
<section class="sizecharts-section">
  <h2 class="sizecharts-section__title">Sample assembly</h2>
  <p>
    We provide professional assembly services, supporting customers who wish to send in carbon fiber rims, wheelsets, and bicycle components.
  </p>
  <p>
    Please contact our sales team at support@tanzanite.site to begin the quotation process, or reach us directly via online chat.
  </p>
</section>
```

如果需要这类段落**居中或变色**，推荐做法是：

- 给当前 section 加一个局部类，例如：`class="sizecharts-section my-section"`。  
- 在局部样式中单独处理：

```css
.my-section > p {
  text-align: center;
  color: #f9fafb;
}
```

而不是复用 `sizecharts-section__intro`，以免意外带上小蓝点。

---

## 五、禁止事项

- **不要手动打 `•` 字符**。  
  一律使用 `sizecharts-section__intro` 或 `sizecharts-section__list` 生成蓝点。

- **不要把 `sizecharts-section__intro` 用在 1/2/3/4 标题行上。**  
  1/2/3/4 这类数字标题应由 `ol.sizecharts-section__list--numbered` 负责，保持数字前干净无蓝点。

- **不要在同一条文本上叠加浏览器默认黑点和自定义蓝点。**  
  所有真正要显示点的地方，都用 `sizecharts-section__list`，不要再依赖 `list-style-type: disc`。

---

## 六、常见模式快速模板

1. **单一 TAB 下的多个要点句子（每句一个蓝点，居中）**
   - 标题：`h2.sizecharts-section__title`  
   - 每句：`p.sizecharts-section__intro`

2. **1/2/3/4 + 子项列表（子项有蓝点）**
   - 外层：`ol.sizecharts-section__list sizecharts-section__list--numbered`  
   - 每个编号标题内，用 `ul.sizecharts-section__list` 放多条子项。

3. **纯说明正文（无蓝点，可左对齐或局部居中）**
   - 使用普通 `<p>`，必要时在 section 上加局部类设置对齐和颜色。

---

## 七、CTA 按钮使用规范

### 7.1 独立 CTA 按钮（单行居中）

- 典型场景：
  - Appearance Logo TAB 中的「Explore all weave & finish options」。
  - Technical / Tire Guides 中类似「View full chart」「Go to calculator」等主要操作。
- 推荐结构：

```html
<section class="sizecharts-section">
  <h2 class="sizecharts-section__title">Appearance Logo</h2>
  <p class="sizecharts-section__intro">
    Why we recommend laser-engraved logos for most builds.
  </p>
  <p class="guide-section__cta-wrapper">
    <button class="wheelset-inline-button">Explore all weave &amp; finish options</button>
  </p>
</section>
```

- 推荐样式（示例）：

```css
.guide-section__cta-wrapper {
  text-align: center;
}
```

或直接给按钮添加一个居中容器，总体原则是：**独立 CTA 按钮单独占一行，并在水平方向居中**，在桌面和移动端都方便点击。

### 7.2 句子内部的按钮（保持内联）

- 典型场景：
  - Sample assembly TAB 中句子里的「Click to view documentation on spoke weaving」按钮。
- 推荐结构：

```html
<p>
  If you wish to assemble your own wheels and purchase parts exclusively from TANZANITE, we are happy to assist you.
  <button class="wheelset-inline-button">
    Click to view documentation on spoke weaving
  </button>
  <button class="wheelset-inline-button">
    Check out the rim weaving holes we can customize
  </button>
</p>
```

- 规则：
  - 这类按钮视为句子的一部分，**跟随文本在同一行/同一段落内排版**。
  - 不额外做整体居中，避免把句子“拆成几段漂浮的按钮块”。
  - 在移动端由浏览器自然换行即可。

总结：

- **独立 CTA**：单独一行，居中对齐。
- **内联按钮**：作为句子一部分，跟随文本对齐方式（通常左对齐）。
