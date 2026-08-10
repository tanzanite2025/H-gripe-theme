import { createRouter, createWebHistory } from 'vue-router'
import type {
  LocationQuery,
  LocationQueryRaw,
  LocationQueryValue,
  RouteLocationNormalized,
  RouteLocationRaw,
  RouteRecordRaw
} from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const firstQueryValue = (value: LocationQueryValue | LocationQueryValue[]): LocationQueryValue => Array.isArray(value) ? value[0] : value

const stripLegacyTabQuery = (query: LocationQuery = {}): LocationQueryRaw => {
  const { tab, subtab, ...remaining } = query
  return remaining
}

const domainRedirect = (defaultRouteName: string, tabRoutes: Record<string, string>) => (to: RouteLocationNormalized): RouteLocationRaw => ({
  name: tabRoutes[String(firstQueryValue(to.query.tab) || '')] || defaultRouteName,
  query: stripLegacyTabQuery(to.query),
  hash: to.hash,
})

const marketingRedirect = (to: RouteLocationNormalized): RouteLocationRaw => {
  const tab = String(firstQueryValue(to.query.tab) || '')
  const target = tab === 'loyalty'
    ? (firstQueryValue(to.query.subtab) === 'rules' ? 'MarketingLoyaltyRules' : 'MarketingLoyaltyTransactions')
    : {
        coupons: 'MarketingCoupons',
        giftcards: 'MarketingGiftCards',
        levels: 'MarketingLevels',
      }[tab] || 'MarketingCoupons'

  return {
    name: target,
    query: stripLegacyTabQuery(to.query),
    hash: to.hash,
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表板' }
      },
      {
        path: 'catalog/products',
        name: 'CatalogProducts',
        component: () => import('@/views/Products.vue'),
        meta: { title: '商品管理', permission: 'product:view' }
      },
      {
        path: 'catalog/templates',
        name: 'CatalogProductTemplates',
        component: () => import('@/views/ProductTypes.vue'),
        meta: { title: '产品模板', permission: 'product:view' }
      },
      {
        path: 'catalog/information-templates',
        name: 'CatalogProductInformationTemplates',
        component: () => import('@/views/ProductInformationTemplates.vue'),
        meta: { title: '产品信息模板', permission: 'product:view' }
      },
      {
        path: 'google-merchant',
        name: 'GoogleMerchant',
        component: () => import('@/views/GoogleMerchant.vue'),
        meta: { title: 'Google Merchant', permission: 'merchant:view' }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单管理', permission: 'order:view' }
      },
      {
        path: 'payment/settings',
        name: 'PaymentSettings',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '支付设置', permission: 'settings:view' }
      },
      {
        path: 'payment/currency',
        name: 'PaymentCurrency',
        redirect: { name: 'SettingsCurrency' },
        meta: { title: '价格币种', permission: 'settings:view' }
      },
      {
        path: 'payment/risk/reviews',
        name: 'PaymentRiskReviews',
        component: () => import('@/views/PaymentRisk.vue'),
        meta: { title: '人工复核', permission: 'order:view' }
      },
      {
        path: 'payment/risk/refund-recommendations',
        name: 'PaymentRiskRefundRecommendations',
        component: () => import('@/views/PaymentRisk.vue'),
        meta: { title: '退款建议', permission: 'order:view' }
      },
      {
        path: 'payment/risk/disputes',
        name: 'PaymentRiskDisputes',
        component: () => import('@/views/PaymentRisk.vue'),
        meta: { title: 'Stripe 拒付', permission: 'order:view' }
      },
      {
        path: 'payment/risk/controls',
        name: 'PaymentRiskControls',
        component: () => import('@/views/PaymentRisk.vue'),
        meta: { title: '人工保护', permission: 'order:view' }
      },
      {
        path: 'warranty',
        redirect: domainRedirect('WarrantyRegistrations', {
          registrations: 'WarrantyRegistrations',
          claims: 'WarrantyClaims',
          expiring: 'WarrantyExpiring',
          boundary: 'WarrantyBoundary',
        }),
        meta: { permission: 'product:view' }
      },
      {
        path: 'warranty/registrations',
        name: 'WarrantyRegistrations',
        component: () => import('@/views/Warranty.vue'),
        meta: { title: '注册记录', permission: 'product:view' }
      },
      {
        path: 'warranty/claims',
        name: 'WarrantyClaims',
        component: () => import('@/views/Warranty.vue'),
        meta: { title: '保修申请', permission: 'product:view' }
      },
      {
        path: 'warranty/expiring',
        name: 'WarrantyExpiring',
        component: () => import('@/views/Warranty.vue'),
        meta: { title: '即将到期', permission: 'product:view' }
      },
      {
        path: 'warranty/boundary',
        name: 'WarrantyBoundary',
        component: () => import('@/views/Warranty.vue'),
        meta: { title: '数据边界', permission: 'product:view' }
      },
      {
        path: 'shipping',
        redirect: domainRedirect('ShippingTemplates', {
          overview: 'ShippingTemplates',
          templates: 'ShippingTemplates',
          zones: 'ShippingZones',
          carriers: 'ShippingCarriers',
          services: 'ShippingServices',
          quote: 'ShippingQuote',
          packaging: 'ShippingPackaging',
          tracking: 'ShippingTracking',
          trackingShipments: 'ShippingTrackingShipments',
        }),
        meta: { permission: 'shipping:view' }
      },
      {
        path: 'shipping/templates',
        name: 'ShippingTemplates',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '运费模板', permission: 'shipping:view' }
      },
      {
        path: 'shipping/zones',
        name: 'ShippingZones',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '配送区域', permission: 'shipping:view' }
      },
      {
        path: 'shipping/carriers',
        name: 'ShippingCarriers',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '承运商', permission: 'shipping:view' }
      },
      {
        path: 'shipping/services',
        name: 'ShippingServices',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '线路服务', permission: 'shipping:view' }
      },
      {
        path: 'shipping/quote',
        name: 'ShippingQuote',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '试算器', permission: 'shipping:view' }
      },
      {
        path: 'shipping/packaging',
        name: 'ShippingPackaging',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '包装规则', permission: 'shipping:view' }
      },
      {
        path: 'shipping/tracking',
        name: 'ShippingTracking',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '追踪配置', permission: 'shipping:view' }
      },
      {
        path: 'shipping/trackingshipments',
        name: 'ShippingTrackingShipments',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '追踪任务', permission: 'shipping:view' }
      },
      {
        path: 'access/admin-users',
        name: 'AccessAdminUsers',
        component: () => import('@/views/Users.vue'),
        meta: { title: '后台账号', permission: 'user:view' }
      },
      {
        path: 'access/customers',
        name: 'AccessCustomers',
        component: () => import('@/views/Customers.vue'),
        meta: { title: '客户账户', permission: 'user:view' }
      },
      {
        path: 'content/blog',
        name: 'ContentBlog',
        component: () => import('@/views/Content.vue'),
        meta: { title: '博客内容', permission: 'content:view' }
      },
      {
        path: 'content/faqs',
        name: 'ContentFAQs',
        component: () => import('@/views/FAQs.vue'),
        meta: { title: 'FAQ 内容', permission: 'faq:view' }
      },
      {
        path: 'content/brand-gallery',
        name: 'ContentBrandGallery',
        component: () => import('@/views/Galleries.vue'),
        meta: { title: '品牌图库', permission: 'gallery:view' }
      },
      {
        path: 'content/media-library',
        name: 'ContentMediaLibrary',
        component: () => import('@/views/MediaLibrary.vue'),
        meta: { title: '媒体库', permission: 'media:view' }
      },
      {
        path: 'support/conversations',
        name: 'SupportConversations',
        component: () => import('@/views/CustomerServiceChats.vue'),
        meta: { title: '客服对话', permission: 'ticket:view' }
      },
      {
        path: 'support/auto-replies',
        name: 'SupportAutoReplies',
        component: () => import('@/views/AutoReplyRules.vue'),
        meta: { title: '自动回复', permission: 'ticket:view' }
      },
      {
        path: 'support/public-chat',
        name: 'SupportPublicChat',
        component: () => import('@/views/Settings.vue'),
        meta: { title: 'Public Chat', permission: 'settings:view' }
      },
      {
        path: 'visitor-profiles',
        redirect: domainRedirect('VisitorProfilesProfiles', {
          profiles: 'VisitorProfilesProfiles',
          risk: 'VisitorProfilesRisk',
        }),
        meta: { permission: 'ticket:view' }
      },
      {
        path: 'visitor-profiles/profiles',
        name: 'VisitorProfilesProfiles',
        component: () => import('@/views/VisitorProfiles.vue'),
        meta: { title: '访客画像', permission: 'ticket:view' }
      },
      {
        path: 'visitor-profiles/risk',
        name: 'VisitorProfilesRisk',
        component: () => import('@/views/VisitorProfiles.vue'),
        meta: { title: '风险事实', permission: 'ticket:view' }
      },
      {
        path: 'tickets',
        name: 'Tickets',
        component: () => import('@/views/Tickets.vue'),
        meta: { title: '工单管理', permission: 'ticket:view' }
      },
      {
        path: 'marketing',
        redirect: marketingRedirect,
        meta: { permission: 'marketing:view' }
      },
      {
        path: 'marketing/coupons',
        name: 'MarketingCoupons',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '优惠券', permission: 'marketing:view' }
      },
      {
        path: 'marketing/giftcards',
        name: 'MarketingGiftCards',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '礼品卡', permission: 'marketing:view' }
      },
      {
        path: 'marketing/loyalty/transactions',
        name: 'MarketingLoyaltyTransactions',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '积分流水', permission: 'marketing:view' }
      },
      {
        path: 'marketing/loyalty/rules',
        name: 'MarketingLoyaltyRules',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '积分规则', permission: 'marketing:view' }
      },
      {
        path: 'marketing/levels',
        name: 'MarketingLevels',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '会员等级', permission: 'marketing:view' }
      },
      {
        path: 'marketing/subscriptions',
        name: 'MarketingSubscriptions',
        component: () => import('@/views/Subscriptions.vue'),
        meta: { title: '邮件订阅', permission: 'subscription:view' }
      },
      {
        path: 'seo',
        redirect: { name: 'SEOHome' },
        meta: { permission: 'seo:view' }
      },
      {
        path: 'seo/home',
        name: 'SEOHome',
        component: () => import('@/views/seo/Home.vue'),
        meta: { title: 'SEO / 首页', permission: 'seo:view' }
      },
      {
        path: 'seo/articles',
        name: 'SEOArticles',
        component: () => import('@/views/seo/Articles.vue'),
        meta: { title: 'SEO / 文章', permission: 'seo:view' }
      },
      {
        path: 'seo/products',
        name: 'SEOProducts',
        component: () => import('@/views/seo/Products.vue'),
        meta: { title: 'SEO / 产品', permission: 'seo:view' }
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: () => import('@/views/Analytics.vue'),
        meta: { title: 'Analytics', permission: 'analytics:view' }
      },
      {
        path: 'settings',
        redirect: domainRedirect('SettingsSite', {
          site: 'SettingsSite',
          email: 'SettingsEmail',
          social: 'SettingsSocial',
          currency: 'SettingsCurrency',
          markets: 'SettingsMarkets',
          api: 'SettingsApi',
          commercial_crawler: 'SettingsCommercialCrawler',
        }),
        meta: { permission: 'settings:view' }
      },
      {
        path: 'settings/site',
        name: 'SettingsSite',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '站点', permission: 'settings:view' }
      },
      {
        path: 'settings/email',
        name: 'SettingsEmail',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '邮件', permission: 'settings:view' }
      },
      {
        path: 'settings/social',
        name: 'SettingsSocial',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '社交媒体', permission: 'settings:view' }
      },
      {
        path: 'settings/currency',
        name: 'SettingsCurrency',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '价格币种', permission: 'settings:view' }
      },
      {
        path: 'settings/markets',
        name: 'SettingsMarkets',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '市场与本地化语种', permission: 'settings:view' }
      },
      {
        path: 'settings/api',
        name: 'SettingsApi',
        component: () => import('@/views/Settings.vue'),
        meta: { title: 'API 管理', permission: 'settings:view' }
      },
      {
        path: 'settings/commercial-crawler',
        name: 'SettingsCommercialCrawler',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '商业爬虫防护', permission: 'settings:view' }
      },
      {
        path: 'audit-logs',
        name: 'AuditLogs',
        component: () => import('@/views/AuditLogs.vue'),
        meta: { title: '审计日志', permission: 'logs:view' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth !== false
  const permission = to.meta.permission

  if (requiresAuth && !authStore.initialized) {
    await authStore.initAuth()
  }

  if (requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
  } else if (permission && !authStore.hasPermission(permission)) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
