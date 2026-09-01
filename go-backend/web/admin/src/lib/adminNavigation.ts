import type { Component } from 'vue'
import {
  BookOpen,
  Calculator,
  ClipboardList,
  CreditCard,
  Cloud,
  Coins,
  FileSearch,
  FileText,
  FileVideo,
  Fingerprint,
  GitBranch,
  Headset,
  Images,
  Link2,
  Globe2,
  Image as ImageIcon,
  ImageDown,
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
  Search,
  ShieldAlert,
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
    id: 'resource-conversion',
    code: 'RESOURCE_CONVERSION',
    label: '资源转换',
    icon: ImageDown,
    children: [
      { id: 'image-to-webp', path: '/tools/image-to-webp', routeName: 'ImageToWebp', label: '图片转 WebP' },
      { id: 'image-vectorizer', path: '/tools/image-vectorizer', routeName: 'ImageVectorizer', label: '图片转 SVG', icon: ImageIcon },
      { id: 'mp4-to-webm', path: '/tools/mp4-to-webm', routeName: 'Mp4ToWebm', label: 'MP4 转 WebM', icon: FileVideo },
    ],
  },
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
      { id: 'catalog-home-main-products', path: '/catalog/home-main-products', routeName: 'CatalogHomeMainProducts', label: '首页主力产品', icon: Package, permission: 'content:view' },
    ],
  },
  {
    id: 'procurement',
    code: 'PROCUREMENT',
    label: '商品成本',
    icon: ClipboardList,
    permission: 'procurement:view',
    children: [
      { id: 'procurement-records', path: '/procurement/records', routeName: 'ProcurementRecords', label: '商品成本' },
    ],
  },
  {
    id: 'fitment-catalog',
    code: 'FITMENT_CATALOG',
    label: '车型适配资料库',
    icon: Waypoints,
    permission: 'fitment_catalog:view',
    children: [
      { id: 'fitment-frame-entries', path: '/fitment-catalog/frame-entries', routeName: 'FitmentFrameEntries', label: '车架 / 车型' },
      { id: 'fitment-fork-entries', path: '/fitment-catalog/fork-entries', routeName: 'FitmentForkEntries', label: '前叉' },
      { id: 'fitment-hub-specifications', path: '/fitment-catalog/hub-specifications', routeName: 'FitmentHubSpecifications', label: '花鼓规格' },
    ],
  },
  {
    id: 'spoke-calculator',
    code: 'SPOKE_CALCULATOR',
    label: '辐条计算器',
    icon: Calculator,
    permission: 'product:view',
    children: [
      { id: 'spoke-calculator-rims', path: '/spoke-calculator/rims', routeName: 'SpokeCalculatorRims', label: '轮圈' },
      { id: 'spoke-calculator-hubs', path: '/spoke-calculator/hubs', routeName: 'SpokeCalculatorHubs', label: '花鼓' },
      { id: 'spoke-calculator-builds', path: '/spoke-calculator/builds', routeName: 'SpokeCalculatorBuilds', label: '装配' },
      { id: 'spoke-calculator-import', path: '/spoke-calculator/import', routeName: 'SpokeCalculatorImport', label: '导入' },
    ],
  },
  {
    id: 'selection',
    code: 'SELECTION',
    label: '选型配置',
    icon: Zap,
    permission: 'product:view',
    children: [
      { id: 'selection-quick-buy', path: '/selection/quick-buy', routeName: 'SelectionQuickBuy', label: 'QUICK 选配流程' },
      { id: 'selection-configuration-keys', path: '/selection/configuration-keys', routeName: 'SelectionConfigurationKeys', label: 'Key 管理' },
      { id: 'selection-wheelset-fit-questionnaire', path: '/selection/wheelset-fit-questionnaire', routeName: 'SelectionWheelsetFitQuestionnaire', label: '轮组选型问卷' },
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
      { id: 'seo-categories', path: '/seo/categories', routeName: 'SEOCategories', label: '分类' },
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
      { id: 'urls-search', path: '/urls/search', routeName: 'URLSearch', label: '搜索管理', icon: Search },
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
      { id: 'orders-after-sales', path: '/orders/after-sales', routeName: 'AfterSalesCases', label: '退换货管理' },
    ],
  },
  {
    id: 'payment-methods',
    code: 'PAYMENT_METHODS',
    label: '收款方式',
    path: '/payment/methods',
    routeName: 'PaymentCollectionMethods',
    icon: CreditCard,
    permission: 'settings:view',
  },
  {
    id: 'payment-stripe',
    code: 'PAYMENT_STRIPE',
    label: 'Stripe',
    icon: Link2,
    children: [
      { id: 'payment-stripe-integration', path: '/payment/stripe/integration', routeName: 'PaymentStripeIntegration', label: '接入配置', permission: 'settings:view' },
      { id: 'payment-stripe-installments', path: '/payment/stripe/installments', routeName: 'PaymentStripeInstallments', label: '分期配置', permission: 'settings:view' },
      { id: 'payment-stripe-risk-overview', path: '/payment/stripe/risk-overview', routeName: 'StripeRiskStrategyOverview', label: 'Stripe 风控总览', permission: 'order:view' },
      { id: 'payment-stripe-3ds', path: '/payment/stripe/3ds', routeName: 'PaymentStripeThreeDS', label: '3DS / 风控策略', permission: 'order:view' },
      { id: 'payment-stripe-disputes', path: '/payment/stripe/disputes', routeName: 'PaymentStripeDisputes', label: '拒付处理', permission: 'order:view' },
    ],
  },
  {
    id: 'payment-paypal',
    code: 'PAYMENT_PAYPAL',
    label: 'PayPal',
    icon: Globe2,
    children: [
      { id: 'payment-paypal-integration', path: '/payment/paypal/integration', routeName: 'PaymentPayPalIntegration', label: '接入配置', permission: 'settings:view' },
      { id: 'payment-paypal-installments', path: '/payment/paypal/installments', routeName: 'PaymentPayPalInstallments', label: '分期配置', permission: 'settings:view' },
      { id: 'payment-paypal-risk-overview', path: '/payment/paypal/risk-overview', routeName: 'PayPalRiskStrategyOverview', label: 'PayPal 风控总览', permission: 'order:view' },
      { id: 'payment-paypal-disputes', path: '/payment/paypal/disputes', routeName: 'PaymentPayPalDisputes', label: '拒付处理', permission: 'order:view' },
      { id: 'payment-paypal-invoice', path: '/payment/paypal/invoice', routeName: 'PaymentPayPalInvoice', label: '发票方资料', permission: 'settings:view' },
    ],
  },
  {
    id: 'payment-wechat',
    code: 'PAYMENT_WECHAT',
    label: '微信支付',
    icon: MessageSquareText,
    children: [
      { id: 'payment-wechat-integration', path: '/payment/wechat/integration', routeName: 'PaymentWeChatIntegration', label: '接入配置', permission: 'settings:view' },
      { id: 'payment-wechat-installments', path: '/payment/wechat/installments', routeName: 'PaymentWeChatInstallments', label: '分期配置', permission: 'settings:view' },
    ],
  },
  {
    id: 'payment-alipay',
    code: 'PAYMENT_ALIPAY',
    label: '支付宝',
    icon: CreditCard,
    children: [
      { id: 'payment-alipay-integration', path: '/payment/alipay/integration', routeName: 'PaymentAlipayIntegration', label: '接入配置', permission: 'settings:view' },
      { id: 'payment-alipay-installments', path: '/payment/alipay/installments', routeName: 'PaymentAlipayInstallments', label: '分期配置', permission: 'settings:view' },
    ],
  },
  {
    id: 'currency-exchange',
    code: 'CURRENCY_EXCHANGE',
    label: '币种与汇率',
    icon: Coins,
    permission: 'settings:view',
    children: [
      { id: 'currency-exchange-overview', path: '/currency-exchange/overview', routeName: 'CurrencyExchangeOverview', label: '币种总览' },
      { id: 'currency-exchange-api', path: '/currency-exchange/api', routeName: 'CurrencyExchangeApi', label: '汇率 API' },
    ],
  },
  {
    id: 'risk-strategy',
    code: 'RISK_STRATEGY',
    label: '风控策略',
    icon: ShieldAlert,
    permission: 'order:view',
    children: [
      { id: 'risk-strategy-overview', path: '/payment-risk/overview', routeName: 'RiskStrategyOverview', label: '风控策略总览', permission: 'order:view' },
      { id: 'risk-strategy-3ds', path: '/payment-risk/3ds', routeName: 'RiskStrategyThreeDS', label: '3DS 策略', permission: 'order:view' },
      { id: 'risk-strategy-manual-reviews', path: '/payment-risk/reviews', routeName: 'RiskStrategyReviews', label: '人工复核', permission: 'order:view' },
      { id: 'risk-strategy-refund-recommendations', path: '/payment-risk/refund-recommendations', routeName: 'RiskStrategyRefundRecommendations', label: '退款建议', permission: 'order:view' },
      { id: 'risk-strategy-manual-protection', path: '/payment-risk/controls', routeName: 'RiskStrategyControls', label: '人工保护', permission: 'order:view' },
    ],
  },
  {
    id: 'warranty',
    code: 'WARRANTY',
    label: '保修管理',
    icon: ShieldCheck,
    permission: 'product:view',
      children: [
        { id: 'warranty-shipments', path: '/warranty/shipments', routeName: 'WarrantyShipments', label: '已发货' },
        { id: 'warranty-claims', path: '/warranty/claims', routeName: 'WarrantyClaims', label: '保修申请' },
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
    id: 'media-center',
    code: 'MEDIA_CENTER',
    label: '媒体中心',
    icon: Images,
    children: [
      { id: 'media-library', path: '/media/library', routeName: 'MediaLibrary', label: '媒体库', permission: 'media:view' },
      { id: 'media-derivatives', path: '/media/derivatives', routeName: 'MediaDerivatives', label: '图片尺寸转换', permission: 'media:configure' },
      { id: 'media-copyright', path: '/media/copyright', routeName: 'MediaCopyright', label: '图片版权', permission: 'settings:view' },
    ],
  },
  {
    id: 'site-introduction',
    code: 'SITE_INTRODUCTION',
    label: '站点介绍',
    icon: BookOpen,
    children: [
      { id: 'settings-site', path: '/site-introduction/site', routeName: 'SettingsSite', label: '站点', permission: 'settings:view' },
      { id: 'site-introduction-profile', path: '/site-introduction/website-profile', routeName: 'SiteIntroductionWebsiteProfile', label: '我与这个网站', permission: 'settings:view' },
      { id: 'site-introduction-name', path: '/site-introduction/why-this-name', routeName: 'SiteIntroductionWhyThisName', label: '为什么叫这个名字', permission: 'settings:view' },
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
      { id: 'content-visual-showcase', path: '/content/visual-showcase', routeName: 'ContentVisualShowcase', label: '首页视觉目录', icon: Images, permission: 'content:view' },
      { id: 'content-refund-return', path: '/content/refund-return', routeName: 'ContentRefundReturn', label: '退货退款', permission: 'content:view' },
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
      { id: 'social-x', path: '/social/x', routeName: 'SocialX', label: 'X' },
      { id: 'social-reddit', path: '/social/reddit', routeName: 'SocialReddit', label: 'Reddit' },
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
      { id: 'services-github', path: '/services/github', routeName: 'ServicesGitHub', label: 'GitHub / GHCR', icon: GitBranch },
      { id: 'services-connectors', path: '/services/connectors', routeName: 'ServicesConnectors', label: '连接器配置' },
    ],
  },
  {
    id: 'ops',
    code: 'OPS',
    label: '运维中心',
    icon: Server,
    permission: 'ops:view',
    children: [
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
      { id: 'settings-email', path: '/settings/email', routeName: 'SettingsEmail', label: '邮件' },
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
