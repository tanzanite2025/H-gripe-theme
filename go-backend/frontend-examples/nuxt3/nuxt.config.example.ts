/**
 * Nuxt 3 配置示例
 * 
 * 将此配置添加到你的 nuxt.config.ts 文件中
 */

export default defineNuxtConfig({
  // 运行时配置
  runtimeConfig: {
    public: {
      // API 基础 URL
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9000',
      
      // 站点 URL（用于生成完整 URL）
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'https://tanzanite.site',
    }
  },

  // 路由配置
  router: {
    options: {
      // 支持多语言路由
      // 例如: /blog/post 和 /zh/blog/post
    }
  },

  // 自动导入组件
  components: [
    {
      path: '~/components',
      pathPrefix: false,
    }
  ],

  // 模块
  modules: [
    // 如果使用 @nuxtjs/i18n 模块，可以配置如下
    // '@nuxtjs/i18n',
  ],

  // i18n 配置（如果使用 @nuxtjs/i18n 模块）
  // i18n: {
  //   locales: [
  //     { code: 'en', iso: 'en-US', name: 'English' },
  //     { code: 'zh', iso: 'zh-CN', name: '简体中文' },
  //     { code: 'fr', iso: 'fr-FR', name: 'Français' },
  //     // ... 其他语言
  //   ],
  //   defaultLocale: 'en',
  //   strategy: 'prefix_except_default', // 英文不加前缀，其他语言加前缀
  //   detectBrowserLanguage: {
  //     useCookie: true,
  //     cookieKey: 'locale',
  //     redirectOn: 'root',
  //   }
  // },

  // Nitro 配置（用于 SSR/SSG）
  nitro: {
    // 预渲染路由（SSG）
    prerender: {
      crawlLinks: true,
      routes: [
        '/',
        '/sitemap.xml',
        // 可以添加更多需要预渲染的路由
      ]
    }
  },

  // 环境变量
  env: {
    API_BASE: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9000',
  },

  // 开发服务器配置
  devServer: {
    port: 3000,
  },

  // 构建配置
  build: {
    transpile: [],
  },

  // TypeScript 配置
  typescript: {
    strict: true,
    typeCheck: true,
  },

  // 实验性功能
  experimental: {
    payloadExtraction: false,
  },

  // CSS 配置
  css: [
    // 全局样式
  ],

  // Vite 配置
  vite: {
    server: {
      proxy: {
        // 开发环境代理 API 请求
        '/api': {
          target: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9000',
          changeOrigin: true,
        },
        '/sitemap.xml': {
          target: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9000',
          changeOrigin: true,
        },
        '/sitemap-hreflang.xml': {
          target: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9000',
          changeOrigin: true,
        }
      }
    }
  }
})
