# Nuxt SSG + WordPress API 架构文档

**项目**: Tanzanite Theme  
**更新日期**: 2025-11-17  
**架构**: Nuxt 静态生成 + Tanzanite Setting 插件

---

## 📐 完整架构图

```
┌─────────────────────────────────────────────────────────────┐
│  Nuxt 前端 (nuxt-i18n) - 静态生成 + 客户端导航              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  📄 /shop (商品列表页)                                       │
│     ├─ SSR/CSR：从 WordPress API 获取商品列表               │
│     ├─ 显示商品卡片网格                                     │
│     └─ 点击商品 → 跳转到 /products/[slug]                   │
│                                                              │
│  📄 /products/[slug] (商品详情页 - 动态路由)                │
│     ├─ SSG：构建时预渲染所有商品页面                        │
│     ├─ 每个商品独立 HTML 文件                               │
│     ├─ 包含完整 SEO 元数据                                  │
│     ├─ 超快加载速度（静态文件）                             │
│     └─ URL 示例：/products/diamond-ring                     │
│                                                              │
│  📄 /faq (FAQ 页面 - 已实现)                                │
│     └─ 从 WordPress 加载 FAQ JSON                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                            ↓ 
                    WordPress REST API
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  WordPress 后台 (tanzanite.site/wp-admin)                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  🔧 Tanzanite Setting 插件（自建商品系统）                  │
│  ├─────────────────────────────────────────────────────────┤
│  │  📦 商品管理 (Product Management)                        │
│  │     ├─ 自定义商品 CPT (Custom Post Type)                │
│  │     ├─ 商品字段：标题、描述、价格、库存、图片            │
│  │     ├─ 商品分类：戒指、项链、手镯等                      │
│  │     └─ REST API 端点：/wp-json/tanzanite/v1/products    │
│  │                                                           │
│  │  🔗 URLLink 模块                                         │
│  │     ├─ 管理商品 URL 结构                                 │
│  │     ├─ 目录树：products/rings/, products/necklaces/     │
│  │     ├─ 自定义路径：/products/diamond-ring               │
│  │     ├─ 301 重定向管理                                    │
│  │     └─ SEO 优化                                          │
│  │                                                           │
│  │  🌍 SEO 模块 (MyTheme SEO)                              │
│  │     ├─ 多语言支持（34 种语言）                           │
│  │     ├─ 站点地图生成                                      │
│  │     ├─ Meta 标签管理                                     │
│  │     └─ 语言配置：i18n-languages.json                    │
│  │                                                           │
│  │  ❓ FAQ 模块                                             │
│  │     ├─ FAQ 内容管理                                      │
│  │     ├─ 多语言 FAQ                                        │
│  │     └─ JSON 生成：/wp-content/uploads/faq/{locale}.json │
│  │                                                           │
│  └─────────────────────────────────────────────────────────┤
│                                                              │
│  📊 数据库 (MySQL)                                          │
│     ├─ wp_posts (商品数据)                                  │
│     ├─ wp_postmeta (商品元数据)                             │
│     │   ├─ _product_price                                   │
│     │   ├─ _product_stock                                   │
│     │   ├─ _urllink_path (自定义 URL)                       │
│     │   └─ _seo_meta_* (SEO 数据)                           │
│     └─ wp_options (插件配置)                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 数据流程

### 1. 商品创建流程

```
WordPress 后台
    ↓
Tanzanite Setting → 商品管理
    ↓
创建新商品：
  - 标题：Diamond Ring
  - 价格：$1,299
  - 库存：50
  - 图片：上传
  - 描述：高品质钻石戒指
    ↓
URLLink 模块设置 URL：
  - 目录：products/rings
  - Slug：diamond-ring
  - 最终 URL：/products/rings/diamond-ring
    ↓
保存 → 触发 WordPress REST API 更新
    ↓
Nuxt 构建时获取商品数据
    ↓
生成静态页面：
  .output/public/products/rings/diamond-ring/index.html
```

### 2. 用户访问流程

```
用户访问：tanzanite.site/products/diamond-ring
    ↓
Nginx/Apache 检查静态文件
    ↓
找到：.output/public/products/diamond-ring/index.html
    ↓
直接返回静态 HTML（超快！）
    ↓
包含完整商品信息：
  - 标题、价格、图片
  - SEO meta 标签
  - 结构化数据（JSON-LD）
    ↓
页面加载完成（<100ms）
    ↓
用户点击其他商品
    ↓
Nuxt 客户端导航（无刷新）
    ↓
URL 更新：/products/gold-necklace
    ↓
浏览器历史记录更新
```

---

## 📊 URL 管理对比

| 内容类型 | URL 示例 | 管理方式 | URLLink 设置 |
|---------|---------|---------|-------------|
| **首页** | `/` | Nuxt pages | ❌ 不需要 |
| **商店列表** | `/shop` | Nuxt pages | ❌ 不需要 |
| **FAQ** | `/faq` | Nuxt pages | ❌ 不需要 |
| **联系我们** | `/contact` | Nuxt pages | ❌ 不需要 |
| **关于我们** | `/company/about` | Nuxt pages | ❌ 不需要 |
| **商品详情** | `/products/diamond-ring` | Tanzanite Setting + URLLink | ✅ **需要** |
| **商品分类** | `/products/rings/` | URLLink 目录树 | ✅ **需要** |

---

## 🎯 URLLink 插件目录树设置

### 目录结构

```
WordPress 后台 → Tanzanite Setting → URL Management

目录树：
└── products/                    # 商品根目录
    ├── rings/                   # 戒指分类
    │   ├── diamond-ring         # 钻石戒指
    │   ├── gold-ring            # 黄金戒指
    │   └── platinum-ring        # 铂金戒指
    ├── necklaces/               # 项链分类
    │   ├── gold-necklace
    │   └── silver-necklace
    └── bracelets/               # 手镯分类
        └── silver-bracelet
```

### 设置步骤

1. **创建根目录**
   - 目录名称：商品
   - Path Slug：`products`
   - 父目录：无（顶层）

2. **创建子分类目录**
   - 戒指：`rings`（父目录：products）
   - 项链：`necklaces`（父目录：products）
   - 手镯：`bracelets`（父目录：products）

3. **附加商品到目录**
   - 在商品管理中编辑商品
   - 设置 Custom Path：`products/rings/diamond-ring`
   - 或使用批量操作将商品附加到目录

4. **保存并同步**
   - 点击"同步到 WordPress"
   - 重建 URL 映射
   - 刷新站点地图

---

## 🛠️ 实施步骤

### 步骤 1：创建 Nuxt 动态路由页面

**文件路径**: `nuxt-i18n/app/pages/products/[slug].vue`

```vue
<template>
  <div class="min-h-screen bg-black text-white">
    <SiteHeader />
    
    <div class="max-w-7xl mx-auto px-4 py-20">
      <!-- 商品详情 -->
      <div class="grid md:grid-cols-2 gap-12">
        <!-- 商品图片 -->
        <div class="aspect-square rounded-2xl overflow-hidden bg-white/5">
          <img 
            :src="product.image" 
            :alt="product.title"
            class="w-full h-full object-cover"
          />
        </div>
        
        <!-- 商品信息 -->
        <div class="space-y-6">
          <h1 class="text-4xl font-bold bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] bg-clip-text text-transparent">
            {{ product.title }}
          </h1>
          
          <div class="text-3xl font-bold text-white">
            ${{ product.price }}
          </div>
          
          <div class="text-white/70 leading-relaxed">
            {{ product.description }}
          </div>
          
          <div class="flex items-center gap-4">
            <span class="text-white/60">库存：</span>
            <span class="text-[#40ffaa] font-semibold">{{ product.stock }} 件</span>
          </div>
          
          <button 
            @click="addToCart"
            class="w-full py-4 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black font-bold text-lg hover:shadow-[0_0_30px_rgba(107,115,255,0.5)] transition-all"
          >
            加入购物车
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const route = useRoute()
const config = useRuntimeConfig()
const { locale } = useI18n()

// 从 Tanzanite Setting API 获取商品数据
const { data: product } = await useFetch(
  `${config.public.wpApiBase}/tanzanite/v1/products/${route.params.slug}`,
  {
    key: `product-${route.params.slug}-${locale.value}`,
  }
)

// SEO 元数据
useHead({
  title: product.value?.seo?.title || product.value?.title,
  meta: [
    { name: 'description', content: product.value?.seo?.description },
    { property: 'og:title', content: product.value?.title },
    { property: 'og:image', content: product.value?.image },
  ]
})
</script>
```

---

### 步骤 2：配置 Nuxt 预渲染

**文件路径**: `nuxt-i18n/nuxt.config.ts`

```typescript
export default defineNuxtConfig({
  nitro: {
    prerender: {
      crawlLinks: true,
      routes: async () => {
        // 从 Tanzanite Setting API 获取所有商品
        const products = await $fetch('https://tanzanite.site/wp-json/tanzanite/v1/products')
        
        // 生成所有商品页面路由
        return products.map(p => `/products/${p.slug}`)
      }
    }
  },
  
  runtimeConfig: {
    public: {
      wpApiBase: 'https://tanzanite.site/wp-json',
      wordpressUrl: 'https://tanzanite.site'
    }
  }
})
```

---

### 步骤 3：Tanzanite Setting 商品 API

**文件路径**: `wp-plugin/tanzanite-setting/includes/products/class-product-api.php`

```php
<?php
class Tanzanite_Product_API {
    public function register_routes() {
        // 获取所有商品
        register_rest_route('tanzanite/v1', '/products', [
            'methods' => 'GET',
            'callback' => [$this, 'get_products'],
            'permission_callback' => '__return_true'
        ]);
        
        // 根据 slug 获取单个商品
        register_rest_route('tanzanite/v1', '/products/(?P<slug>[a-zA-Z0-9-]+)', [
            'methods' => 'GET',
            'callback' => [$this, 'get_product_by_slug'],
            'permission_callback' => '__return_true'
        ]);
    }
    
    public function get_products($request) {
        $args = [
            'post_type' => 'tanzanite_product',
            'posts_per_page' => -1,
            'post_status' => 'publish'
        ];
        
        $products = get_posts($args);
        $data = [];
        
        foreach ($products as $product) {
            $data[] = $this->format_product($product);
        }
        
        return rest_ensure_response($data);
    }
    
    public function get_product_by_slug($request) {
        $slug = $request['slug'];
        
        $args = [
            'post_type' => 'tanzanite_product',
            'name' => $slug,
            'posts_per_page' => 1,
            'post_status' => 'publish'
        ];
        
        $products = get_posts($args);
        
        if (empty($products)) {
            return new WP_Error('not_found', '商品未找到', ['status' => 404]);
        }
        
        return rest_ensure_response($this->format_product($products[0]));
    }
    
    private function format_product($product) {
        return [
            'id' => $product->ID,
            'title' => $product->post_title,
            'slug' => $product->post_name,
            'description' => $product->post_content,
            'price' => get_post_meta($product->ID, '_product_price', true),
            'stock' => get_post_meta($product->ID, '_product_stock', true),
            'image' => get_the_post_thumbnail_url($product->ID, 'large'),
            'url' => $this->get_product_url($product->ID),
            'seo' => [
                'title' => get_post_meta($product->ID, '_seo_title', true),
                'description' => get_post_meta($product->ID, '_seo_description', true)
            ]
        ];
    }
    
    private function get_product_url($post_id) {
        // 从 URLLink 获取自定义路径
        $custom_path = get_post_meta($post_id, '_urllink_path', true);
        return $custom_path ?: get_permalink($post_id);
    }
}
```

---

## ⚡ 性能优势

### 构建时（npm run generate）

```
1. Nuxt 调用 Tanzanite Setting API
2. 获取所有商品数据（例如 100 个商品）
3. 为每个商品生成静态 HTML
4. 生成 100 个独立文件
5. 总耗时：~30 秒
```

### 用户访问时

```
1. 用户访问 /products/diamond-ring
2. Nginx 直接返回静态 HTML
3. 加载时间：<100ms
4. 无需查询 WordPress 数据库
5. 无需 PHP 处理
```

### 性能对比

| 指标 | 传统 WordPress | Nuxt SSG |
|-----|---------------|----------|
| **首次加载** | 500-1000ms | <100ms |
| **服务器负载** | 高（每次查询数据库） | 极低（静态文件） |
| **并发能力** | 100 req/s | 10000+ req/s |
| **SEO** | 好 | 完美 |
| **CDN 友好** | 一般 | 极好 |

---

## 🔄 内容更新流程

### 方式 1：手动重新构建

```bash
# 在 Nuxt 项目目录
cd nuxt-i18n
npm run generate

# 部署到服务器
rsync -avz .output/public/ user@server:/var/www/tanzanite.site/
```

### 方式 2：自动化 Webhook（推荐）

```php
// WordPress 插件中添加 Webhook
add_action('save_post_tanzanite_product', function($post_id) {
    // 触发 Nuxt 重新构建
    wp_remote_post('https://your-ci-cd-service.com/webhook', [
        'body' => [
            'event' => 'product_updated',
            'post_id' => $post_id
        ]
    ]);
});
```

---

## ✅ 架构优势总结

### Tanzanite Setting 负责

- ✅ 商品数据管理
- ✅ 商品 REST API
- ✅ URL 结构管理（URLLink）
- ✅ SEO 元数据
- ✅ 多语言支持

### Nuxt 负责

- ✅ 静态页面生成（SSG）
- ✅ 客户端路由
- ✅ 用户界面渲染
- ✅ 购物车逻辑
- ✅ 性能优化

### 完美结合

- 🚀 **静态速度** + **动态管理**
- 🔗 **完美 URL** + **SEO 优化**
- 📦 **自建系统** + **无 WooCommerce 依赖**
- 🌍 **多语言支持** + **全球 CDN**
- ⚡ **极致性能** + **灵活管理**

---

## 📝 注意事项

1. **商品数量限制**
   - 建议商品数量 < 1000
   - 超过 1000 需要考虑增量构建

2. **构建时间**
   - 100 商品：~30 秒
   - 500 商品：~2 分钟
   - 1000 商品：~5 分钟

3. **内容更新延迟**
   - 手动构建：立即生效
   - 自动构建：5-10 分钟延迟
   - 可接受的延迟范围

4. **缓存策略**
   - 静态文件：长期缓存（1 年）
   - API 请求：短期缓存（5 分钟）
   - 购物车数据：不缓存

---

**最后更新**: 2025-11-17  
**维护者**: Tanzanite Team
