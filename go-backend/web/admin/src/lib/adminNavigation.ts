import type { Component } from 'vue'
import {
  CreditCard,
  FileText,
  Fingerprint,
  Headset,
  Globe2,
  LayoutDashboard,
  Megaphone,
  MessagesSquare,
  Package,
  ScrollText,
  Settings,
  ShieldCheck,
  ShoppingCart,
  Truck,
  Users,
  Zap,
} from '@lucide/vue'

export interface AdminNavigationItem {
  id: string
  code?: string
  label: string
  path?: string
  routeName?: string
  icon?: Component
  permission?: string
  children?: AdminNavigationItem[]
}

export interface ActiveNavigationEntry {
  item: AdminNavigationItem
  parent: AdminNavigationItem | null
  score: number
}

type PermissionChecker = (permission: string) => boolean

export const adminNavigationItems: AdminNavigationItem[] = [
  { id: 'dashboard', path: '/', routeName: 'Dashboard', code: 'DASHBOARD', label: '仪表板', icon: LayoutDashboard },
  {
    id: 'catalog',
    code: 'CATALOG',
    label: '商品中心',
    icon: Package,
    permission: 'product:view',
    children: [
      { id: 'catalog-products', path: '/catalog/products', routeName: 'CatalogProducts', label: '商品管理' },
      { id: 'catalog-templates', path: '/catalog/templates', routeName: 'CatalogProductTemplates', label: '产品模板' },
      { id: 'catalog-information-templates', path: '/catalog/information-templates', routeName: 'CatalogProductInformationTemplates', label: '产品信息模板' },
      { id: 'catalog-spoke-calculator', path: '/catalog/spoke-calculator', routeName: 'CatalogSpokeCalculator', label: '辐条计算器数据' },
      { id: 'catalog-quick-buy', path: '/catalog/quick-buy', routeName: 'CatalogQuickBuy', label: 'QUICK 选配流程', icon: Zap },
    ],
  },
  {
    id: 'google-merchant',
    path: '/google-merchant',
    routeName: 'GoogleMerchant',
    code: 'GOOGLE_MERCHANT',
    label: 'Google Merchant',
    icon: Globe2,
    permission: 'merchant:view',
  },
  {
    id: 'seo',
    code: 'SEO',
    label: 'SEO',
    icon: FileText,
    permission: 'seo:view',
    children: [
      { id: 'seo-home', path: '/seo/home', routeName: 'SEOHome', label: '首页' },
      { id: 'seo-articles', path: '/seo/articles', routeName: 'SEOArticles', label: '文章' },
      { id: 'seo-products', path: '/seo/products', routeName: 'SEOProducts', label: '产品' },
    ],
  },
  { id: 'analytics', path: '/analytics', routeName: 'Analytics', code: 'ANALYTICS', label: 'Analytics', icon: Globe2, permission: 'analytics:view' },
  { id: 'orders', path: '/orders', routeName: 'Orders', code: 'ORDERS', label: '订单管理', icon: ShoppingCart, permission: 'order:view' },
  {
    id: 'payment',
    code: 'PAYMENT',
    label: '支付',
    icon: CreditCard,
    children: [
      { id: 'payment-settings', path: '/payment/settings', routeName: 'PaymentSettings', label: '支付设置', permission: 'settings:view' },
      { id: 'payment-risk-reviews', path: '/payment/risk/reviews', routeName: 'PaymentRiskReviews', label: '人工复核', permission: 'order:view' },
      { id: 'payment-risk-refund-recommendations', path: '/payment/risk/refund-recommendations', routeName: 'PaymentRiskRefundRecommendations', label: '退款建议', permission: 'order:view' },
      { id: 'payment-risk-disputes', path: '/payment/risk/disputes', routeName: 'PaymentRiskDisputes', label: 'Stripe 拒付', permission: 'order:view' },
      { id: 'payment-risk-controls', path: '/payment/risk/controls', routeName: 'PaymentRiskControls', label: '人工保护', permission: 'order:view' },
    ],
  },
  {
    id: 'warranty',
    code: 'WARRANTY',
    label: '保修管理',
    icon: ShieldCheck,
    permission: 'product:view',
    children: [
      { id: 'warranty-registrations', path: '/warranty/registrations', routeName: 'WarrantyRegistrations', label: '注册记录' },
      { id: 'warranty-claims', path: '/warranty/claims', routeName: 'WarrantyClaims', label: '保修申请' },
      { id: 'warranty-expiring', path: '/warranty/expiring', routeName: 'WarrantyExpiring', label: '即将到期' },
      { id: 'warranty-boundary', path: '/warranty/boundary', routeName: 'WarrantyBoundary', label: '数据边界' },
    ],
  },
  {
    id: 'shipping',
    code: 'SHIPPING',
    label: '物流管理',
    icon: Truck,
    permission: 'shipping:view',
    children: [
      { id: 'shipping-templates', path: '/shipping/templates', routeName: 'ShippingTemplates', label: '运费模板' },
      { id: 'shipping-zones', path: '/shipping/zones', routeName: 'ShippingZones', label: '配送区域' },
      { id: 'shipping-carriers', path: '/shipping/carriers', routeName: 'ShippingCarriers', label: '承运商' },
      { id: 'shipping-services', path: '/shipping/services', routeName: 'ShippingServices', label: '线路服务' },
      { id: 'shipping-quote', path: '/shipping/quote', routeName: 'ShippingQuote', label: '试算器' },
      { id: 'shipping-packaging', path: '/shipping/packaging', routeName: 'ShippingPackaging', label: '包装规则' },
      { id: 'shipping-tracking', path: '/shipping/tracking', routeName: 'ShippingTracking', label: '追踪配置' },
      { id: 'shipping-tracking-shipments', path: '/shipping/trackingshipments', routeName: 'ShippingTrackingShipments', label: '追踪任务' },
    ],
  },
  {
    id: 'access',
    code: 'ACCESS',
    label: '账号权限',
    icon: Users,
    permission: 'user:view',
    children: [
      { id: 'access-admin-users', path: '/access/admin-users', routeName: 'AccessAdminUsers', label: '后台账号' },
      { id: 'access-customers', path: '/access/customers', routeName: 'AccessCustomers', label: '客户账户' },
    ],
  },
  {
    id: 'content',
    code: 'CONTENT',
    label: '网站内容',
    icon: FileText,
    children: [
      { id: 'content-blog', path: '/content/blog', routeName: 'ContentBlog', label: '博客内容', permission: 'content:view' },
      { id: 'content-brand-gallery', path: '/content/brand-gallery', routeName: 'ContentBrandGallery', label: '品牌图库', permission: 'gallery:view' },
      { id: 'content-faqs', path: '/content/faqs', routeName: 'ContentFAQs', label: 'FAQ 内容', permission: 'faq:view' },
      { id: 'content-media-library', path: '/content/media-library', routeName: 'ContentMediaLibrary', label: '媒体库', permission: 'media:view' },
    ],
  },
  {
    id: 'support',
    code: 'SUPPORT',
    label: '客服中心',
    icon: Headset,
    children: [
      { id: 'support-conversations', path: '/support/conversations', routeName: 'SupportConversations', label: '客服对话', permission: 'ticket:view' },
      { id: 'support-auto-replies', path: '/support/auto-replies', routeName: 'SupportAutoReplies', label: '自动回复', permission: 'ticket:view' },
      { id: 'support-public-chat', path: '/support/public-chat', routeName: 'SupportPublicChat', label: 'Public Chat', permission: 'settings:view' },
    ],
  },
  {
    id: 'visitor-profiles',
    code: 'VISITORS',
    label: '访客画像',
    icon: Fingerprint,
    permission: 'ticket:view',
    children: [
      { id: 'visitor-profiles-profiles', path: '/visitor-profiles/profiles', routeName: 'VisitorProfilesProfiles', label: '访客画像' },
      { id: 'visitor-profiles-risk', path: '/visitor-profiles/risk', routeName: 'VisitorProfilesRisk', label: '风险事实' },
    ],
  },
  { id: 'tickets', path: '/tickets', routeName: 'Tickets', code: 'TICKETS', label: '工单管理', icon: MessagesSquare, permission: 'ticket:view' },
  {
    id: 'marketing',
    code: 'MARKETING',
    label: '营销管理',
    icon: Megaphone,
    children: [
      { id: 'marketing-coupons', path: '/marketing/coupons', routeName: 'MarketingCoupons', label: '优惠券', permission: 'marketing:view' },
      { id: 'marketing-giftcards', path: '/marketing/giftcards', routeName: 'MarketingGiftCards', label: '礼品卡', permission: 'marketing:view' },
      { id: 'marketing-loyalty-transactions', path: '/marketing/loyalty/transactions', routeName: 'MarketingLoyaltyTransactions', label: '积分流水', permission: 'marketing:view' },
      { id: 'marketing-loyalty-rules', path: '/marketing/loyalty/rules', routeName: 'MarketingLoyaltyRules', label: '积分规则', permission: 'marketing:view' },
      { id: 'marketing-levels', path: '/marketing/levels', routeName: 'MarketingLevels', label: '会员等级', permission: 'marketing:view' },
      { id: 'marketing-subscriptions', path: '/marketing/subscriptions', routeName: 'MarketingSubscriptions', label: '邮件订阅', permission: 'subscription:view' },
    ],
  },
  {
    id: 'settings',
    code: 'SETTINGS',
    label: '系统设置',
    icon: Settings,
    permission: 'settings:view',
    children: [
      { id: 'settings-site', path: '/settings/site', routeName: 'SettingsSite', label: '站点' },
      { id: 'settings-email', path: '/settings/email', routeName: 'SettingsEmail', label: '邮件' },
      { id: 'settings-social', path: '/settings/social', routeName: 'SettingsSocial', label: '社交媒体' },
      { id: 'settings-currency', path: '/settings/currency', routeName: 'SettingsCurrency', label: '价格币种' },
      { id: 'settings-markets', path: '/settings/markets', routeName: 'SettingsMarkets', label: '市场与本地化语种' },
      { id: 'settings-api', path: '/settings/api', routeName: 'SettingsApi', label: 'API 管理' },
      { id: 'settings-commercial-crawler', path: '/settings/commercial-crawler', routeName: 'SettingsCommercialCrawler', label: '商业爬虫防护' },
    ],
  },
  { id: 'audit-logs', path: '/audit-logs', routeName: 'AuditLogs', code: 'AUDIT', label: '审计日志', icon: ScrollText, permission: 'logs:view' },
]

const matchesPath = (itemPath: string | undefined, activePath: string): boolean => {
  if (!itemPath) return false
  if (itemPath === '/') return activePath === '/'
  return activePath === itemPath || activePath.startsWith(`${itemPath}/`)
}

export const filterNavigationItems = (items: AdminNavigationItem[], hasPermission: PermissionChecker): AdminNavigationItem[] => {
  return items
    .map((item): AdminNavigationItem | null => {
      if (item.permission && !hasPermission(item.permission)) return null

      const children = Array.isArray(item.children)
        ? filterNavigationItems(item.children, hasPermission)
        : []

      if (item.children && children.length === 0) return null

      return children.length
        ? { ...item, children }
        : { ...item, children: undefined }
    })
    .filter((item): item is AdminNavigationItem => Boolean(item))
}

export const findActiveNavigationEntry = (items: AdminNavigationItem[], activePath: string): ActiveNavigationEntry | null => {
  let matchedEntry: ActiveNavigationEntry | null = null

  const consider = (item: AdminNavigationItem, parent: AdminNavigationItem | null = null) => {
    const itemPath = item.path || parent?.path || ''
    if (!matchesPath(itemPath, activePath)) return

    const score = itemPath.length + (parent ? 1 : 0)
    if (!matchedEntry || score > matchedEntry.score) {
      matchedEntry = { item, parent, score }
    }
  }

  items.forEach((item) => {
    consider(item)
    ;(item.children || []).forEach((child) => consider(child, item))
  })

  return matchedEntry
}
