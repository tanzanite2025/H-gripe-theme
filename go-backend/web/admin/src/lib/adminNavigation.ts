import type { Component } from 'vue'
import {
  CreditCard,
  Cloud,
  FileSearch,
  FileText,
  Fingerprint,
  Headset,
  Link2,
  Globe2,
  LayoutDashboard,
  ListChecks,
  MessageSquareText,
  Megaphone,
  Package,
  Rocket,
  Server,
  Share2,
  ScrollText,
  Settings,
  ShieldCheck,
  ShoppingCart,
  Star,
  Truck,
  Type,
  Users,
  Waypoints,
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
      { id: 'catalog-brands', path: '/catalog/brands', routeName: 'CatalogProductBrands', label: '商品品牌' },
      { id: 'catalog-categories', path: '/catalog/categories', routeName: 'CatalogProductCategories', label: '商品分类' },
      { id: 'catalog-templates', path: '/catalog/templates', routeName: 'CatalogProductTemplates', label: '商品规格模板' },
      { id: 'catalog-information-templates', path: '/catalog/information-templates', routeName: 'CatalogProductInformationTemplates', label: '产品信息模板' },
      { id: 'catalog-customs-classifications', path: '/catalog/customs-classifications', routeName: 'CatalogCustomsClassifications', label: '清关资料中心', icon: FileSearch },
      { id: 'catalog-quick-buy', path: '/catalog/quick-buy', routeName: 'CatalogQuickBuy', label: 'QUICK 选配流程', icon: Zap },
      { id: 'catalog-wheelset-fit-questionnaire', path: '/catalog/wheelset-fit-questionnaire', routeName: 'CatalogWheelsetFitQuestionnaire', label: '轮组选型问卷', icon: ListChecks },
    ],
  },
  {
    id: 'google',
    code: 'GOOGLE',
    label: 'GOOGLE',
    icon: Globe2,
    children: [
      { id: 'google-merchant', path: '/google-merchant', routeName: 'GoogleMerchant', label: 'Google Merchant', permission: 'merchant:view' },
      { id: 'analytics', path: '/analytics', routeName: 'Analytics', label: 'Analytics', permission: 'analytics:view' },
    ],
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
  {
    id: 'urls',
    code: 'URLS',
    label: 'URL 管理',
    icon: Waypoints,
    permission: 'url:view',
    children: [
      { id: 'urls-overview', path: '/urls/overview', routeName: 'URLOverview', label: '概览' },
      { id: 'urls-catalog', path: '/urls/catalog', routeName: 'URLCatalog', label: '路由台账' },
      { id: 'urls-issues', path: '/urls/issues', routeName: 'URLIssues', label: '问题队列' },
      { id: 'urls-redirects', path: '/urls/redirects', routeName: 'URLRedirects', label: '重定向' },
      { id: 'urls-canonical', path: '/urls/canonical', routeName: 'URLCanonical', label: 'Canonical 与冲突' },
      { id: 'urls-operations', path: '/urls/operations', routeName: 'URLOperations', label: '同步与检查' },
    ],
  },
  {
    id: 'preflight',
    code: 'PREFLIGHT',
    label: '上线前检查',
    icon: ListChecks,
    children: [
      { id: 'preflight-fonts', path: '/preflight/fonts', routeName: 'PreflightFonts', label: '字体', icon: Type, permission: 'services:view' },
      { id: 'preflight-site-quality', path: '/preflight/site-quality', routeName: 'PreflightSiteQuality', label: '页面质量', permission: 'services:view' },
      { id: 'preflight-image-dimensions', path: '/preflight/image-dimensions', routeName: 'PreflightImageDimensions', label: '图片尺寸', permission: 'media:view' },
      { id: 'preflight-content-links', path: '/preflight/content-links', routeName: 'PreflightContentLinks', label: '内容链接', icon: Link2, permission: 'services:view' },
    ],
  },
  {
    id: 'orders',
    code: 'ORDERS',
    label: '订单管理',
    icon: ShoppingCart,
    permission: 'order:view',
    children: [
      { id: 'orders-list', path: '/orders/list', routeName: 'OrdersList', label: '订单列表' },
      { id: 'orders-disputes', path: '/orders/disputes', routeName: 'OrdersDisputes', label: '拒付订单' },
    ],
  },
  {
    id: 'payment',
    code: 'PAYMENT',
    label: '支付',
    icon: CreditCard,
    children: [
      { id: 'payment-settings', path: '/payment/settings', routeName: 'PaymentSettings', label: '支付设置', permission: 'settings:view' },
      { id: 'payment-paypal-invoice', path: '/payment/paypal-invoice', routeName: 'PaymentPayPalInvoice', label: 'PayPal 发票卖方资料', permission: 'settings:view' },
      { id: 'payment-risk-overview', path: '/payment/risk/overview', routeName: 'PaymentRiskOverview', label: '风控总览', permission: 'order:view' },
      { id: 'payment-risk-3ds', path: '/payment/risk/3ds', routeName: 'PaymentRiskThreeDS', label: '3DS 策略', permission: 'order:view' },
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
      { id: 'content-showcase', path: '/content/showcase', routeName: 'ContentShowcase', label: '买家秀审批', permission: 'gallery:view' },
      { id: 'content-reviews', path: '/content/reviews', routeName: 'ContentReviews', label: '评价审核', icon: Star, permission: 'review:view' },
      { id: 'content-page-feedback', path: '/content/feedback', routeName: 'ContentPageFeedback', label: '页面留言', icon: MessageSquareText, permission: 'content:view' },
      { id: 'content-faqs', path: '/content/faqs', routeName: 'ContentFAQs', label: 'FAQ 内容', permission: 'faq:view' },
      { id: 'content-media-library', path: '/content/media-library', routeName: 'ContentMediaLibrary', label: '媒体库', permission: 'media:view' },
      { id: 'content-media-derivatives', path: '/content/media-derivatives', routeName: 'ContentMediaDerivatives', label: '图片尺寸转换', permission: 'media:configure' },
      { id: 'content-spoke-calculator', path: '/content/spoke-calculator', routeName: 'ContentSpokeCalculator', label: '辐条计算器数据', permission: 'product:view' },
      { id: 'content-website-profile', path: '/content/website-profile', routeName: 'ContentWebsiteProfile', label: '我与这个网站', permission: 'settings:view' },
    ],
  },
  {
    id: 'social-media',
    code: 'SOCIAL_MEDIA',
    label: '社交媒体',
    icon: Share2,
    permission: 'settings:view',
    children: [
      { id: 'social-overview', path: '/social/overview', routeName: 'SocialOverview', label: '账号总览' },
      { id: 'social-profiles', path: '/social/profiles', routeName: 'SocialProfiles', label: '前台展示' },
      { id: 'social-youtube', path: '/social/youtube', routeName: 'SocialYouTube', label: 'YouTube' },
      { id: 'social-meta', path: '/social/meta', routeName: 'SocialMeta', label: 'Facebook / Instagram' },
      { id: 'social-tiktok', path: '/social/tiktok', routeName: 'SocialTikTok', label: 'TikTok' },
      { id: 'social-linkedin', path: '/social/linkedin', routeName: 'SocialLinkedIn', label: 'LinkedIn' },
      { id: 'social-x', path: '/social/x', routeName: 'SocialX', label: 'X' },
      { id: 'social-wechat', path: '/social/wechat', routeName: 'SocialWeChat', label: '微信' },
      { id: 'social-publications', path: '/social/publications', routeName: 'SocialPublications', label: '发布记录' },
    ],
  },
  {
    id: 'support',
    code: 'SUPPORT',
    label: '客服中心',
    icon: Headset,
    children: [
      { id: 'support-analytics', path: '/support/analytics', routeName: 'SupportAnalytics', label: '客服分析', permission: 'ticket:view' },
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
    id: 'services',
    code: 'SERVICES',
    label: '服务中心',
    icon: Cloud,
    permission: 'services:view',
    children: [
      { id: 'services-overview', path: '/services/overview', routeName: 'ServicesOverview', label: '服务总览' },
      { id: 'services-cloudflare', path: '/services/cloudflare', routeName: 'ServicesCloudflare', label: 'Cloudflare' },
    ],
  },
  {
    id: 'ops',
    code: 'OPS',
    label: '运维中心',
    icon: Server,
    permission: 'ops:view',
    children: [
      { id: 'ops-overview', path: '/ops/overview', routeName: 'OpsOverview', label: '运维总览', permission: 'ops:view' },
      { id: 'ops-admin-accounts', path: '/ops/admin-accounts', routeName: 'OpsAdminAccounts', label: '后台账号', permission: 'system:manage' },
      { id: 'ops-domains', path: '/ops/domains', routeName: 'OpsDomains', label: '域名中心', permission: 'ops:domain:view' },
      { id: 'ops-vps', path: '/ops/vps', routeName: 'OpsVPS', label: 'VPS 中心', permission: 'ops:vps:view' },
      { id: 'ops-projects', path: '/ops/projects', routeName: 'OpsProjects', label: '项目中心', permission: 'ops:project:view' },
      { id: 'ops-deployments', path: '/ops/deployments', routeName: 'OpsDeployments', label: '部署中心', icon: Rocket, permission: 'ops:deploy:view' },
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
