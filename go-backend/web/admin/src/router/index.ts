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
        path: 'catalog/brands',
        name: 'CatalogProductBrands',
        component: () => import('@/views/ProductBrands.vue'),
        meta: { title: '商品品牌', permission: 'product:view' }
      },
      {
        path: 'catalog/categories',
        name: 'CatalogProductCategories',
        component: () => import('@/views/ProductCategories.vue'),
        meta: { title: '商品分类', permission: 'product:view' }
      },
      {
        path: 'catalog/templates',
        name: 'CatalogProductTemplates',
        component: () => import('@/views/ProductSpecificationTemplates.vue'),
        meta: { title: '商品规格模板', permission: 'product:view' }
      },
      {
        path: 'catalog/information-templates',
        name: 'CatalogProductInformationTemplates',
        component: () => import('@/views/ProductInformationTemplates.vue'),
        meta: { title: '产品信息模板', permission: 'product:view' }
      },
      {
        path: 'catalog/customs-classifications',
        name: 'CatalogCustomsClassifications',
        component: () => import('@/views/CustomsClassifications.vue'),
        meta: { title: '清关资料中心', permission: 'product:view' }
      },
      {
        path: 'content/spoke-calculator',
        name: 'ContentSpokeCalculator',
        component: () => import('@/views/SpokeCatalog.vue'),
        meta: { title: '辐条计算器数据', permission: 'product:view' }
      },
      {
        path: 'selection',
        redirect: { name: 'SelectionQuickBuy' },
        meta: { permission: 'product:view' }
      },
      {
        path: 'selection/quick-buy',
        name: 'SelectionQuickBuy',
        component: () => import('@/views/QuickBuyFlows.vue'),
        meta: { title: 'QUICK 选配流程', permission: 'product:view' }
      },
      {
        path: 'selection/configuration-keys',
        name: 'SelectionConfigurationKeys',
        component: () => import('@/views/SelectionConfigurationKeys.vue'),
        meta: { title: '选型配置 Key', permission: 'product:view' }
      },
      {
        path: 'selection/wheelset-fit-questionnaire',
        name: 'SelectionWheelsetFitQuestionnaire',
        component: () => import('@/views/WheelsetFitQuestionnaire.vue'),
        meta: { title: '轮组选型问卷', permission: 'product:view' }
      },
      {
        path: 'catalog/quick-buy',
        redirect: { name: 'SelectionQuickBuy' },
        meta: { permission: 'product:view' }
      },
      {
        path: 'catalog/wheelset-fit-questionnaire',
        redirect: { name: 'SelectionWheelsetFitQuestionnaire' },
        meta: { permission: 'product:view' }
      },
      {
        path: 'catalog/selection-assistants',
        redirect: { name: 'SelectionWheelsetFitQuestionnaire' }
      },
      {
        path: 'google-merchant',
        name: 'GoogleMerchant',
        component: () => import('@/views/GoogleMerchant.vue'),
        meta: { title: 'Google Merchant', permission: 'merchant:view' }
      },
      {
        path: 'orders',
        redirect: { name: 'OrdersList' },
        meta: { title: '订单管理', permission: 'order:view' }
      },
      {
        path: 'orders/list',
        name: 'OrdersList',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单列表', permission: 'order:view' }
      },
      {
        path: 'orders/disputes',
        name: 'OrdersDisputes',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '拒付订单', permission: 'order:view' }
      },
      {
        path: 'orders/after-sales',
        name: 'AfterSalesCases',
        component: () => import('@/views/AfterSales.vue'),
        meta: { title: '退换货管理', permission: 'order:view' }
      },
      {
        path: 'payment/methods',
        name: 'PaymentCollectionMethods',
        component: () => import('@/views/PaymentCollectionMethods.vue'),
        meta: { title: '收款方式', permission: 'settings:view' }
      },
      {
        path: 'payment/stripe/integration',
        name: 'PaymentStripeIntegration',
        component: () => import('@/views/PaymentIntegrations.vue'),
        props: { provider: 'stripe' },
        meta: { title: 'Stripe 接入', permission: 'settings:view' }
      },
      {
        path: 'payment/stripe/installments',
        name: 'PaymentStripeInstallments',
        component: () => import('@/views/PaymentProviderInstallments.vue'),
        props: { provider: 'stripe' },
        meta: { title: 'Stripe 分期', permission: 'settings:view' }
      },
      {
        path: 'payment/stripe/risk-overview',
        name: 'StripeRiskStrategyOverview',
        component: () => import('@/views/risk-strategy/StripeRiskStrategyOverviewTab.vue'),
        meta: { title: 'Stripe 风控总览', permission: 'order:view' }
      },
      {
        path: 'payment/stripe/3ds',
        name: 'PaymentStripeThreeDS',
        component: () => import('@/views/RiskStrategy.vue'),
        props: { defaultDisputeProvider: 'stripe' },
        meta: { title: 'Stripe 3DS / 风控策略', permission: 'order:view' }
      },
      {
        path: 'payment/stripe/disputes',
        name: 'PaymentStripeDisputes',
        component: () => import('@/views/RiskStrategy.vue'),
        props: { defaultDisputeProvider: 'stripe' },
        meta: { title: 'Stripe 拒付处理', permission: 'order:view' }
      },
      {
        path: 'payment/wechat/integration',
        name: 'PaymentWeChatIntegration',
        component: () => import('@/views/PaymentIntegrations.vue'),
        props: { provider: 'wechat' },
        meta: { title: '微信支付接入', permission: 'settings:view' }
      },
      {
        path: 'payment/wechat/installments',
        name: 'PaymentWeChatInstallments',
        component: () => import('@/views/PaymentProviderInstallments.vue'),
        props: { provider: 'wechat' },
        meta: { title: '微信支付分期', permission: 'settings:view' }
      },
      {
        path: 'payment/alipay/integration',
        name: 'PaymentAlipayIntegration',
        component: () => import('@/views/PaymentIntegrations.vue'),
        props: { provider: 'alipay' },
        meta: { title: '支付宝接入', permission: 'settings:view' }
      },
      {
        path: 'payment/alipay/installments',
        name: 'PaymentAlipayInstallments',
        component: () => import('@/views/PaymentProviderInstallments.vue'),
        props: { provider: 'alipay' },
        meta: { title: '支付宝分期', permission: 'settings:view' }
      },
      {
        path: 'payment/paypal/integration',
        name: 'PaymentPayPalIntegration',
        component: () => import('@/views/PaymentIntegrations.vue'),
        props: { provider: 'paypal' },
        meta: { title: 'PayPal 接入', permission: 'settings:view' }
      },
      {
        path: 'payment/paypal/installments',
        name: 'PaymentPayPalInstallments',
        component: () => import('@/views/PaymentProviderInstallments.vue'),
        props: { provider: 'paypal' },
        meta: { title: 'PayPal 分期', permission: 'settings:view' }
      },
      {
        path: 'payment/paypal/risk-overview',
        name: 'PayPalRiskStrategyOverview',
        component: () => import('@/views/risk-strategy/PayPalRiskStrategyOverviewTab.vue'),
        meta: { title: 'PayPal 风控总览', permission: 'order:view' }
      },
      {
        path: 'payment/paypal/disputes',
        name: 'PaymentPayPalDisputes',
        component: () => import('@/views/RiskStrategy.vue'),
        props: { defaultDisputeProvider: 'paypal' },
        meta: { title: 'PayPal 拒付处理', permission: 'order:view' }
      },
      {
        path: 'payment/paypal/invoice',
        name: 'PaymentPayPalInvoice',
        component: () => import('@/views/PaymentPayPalInvoice.vue'),
        meta: { title: 'PayPal 发票卖方资料', permission: 'settings:view' }
      },
      {
        path: 'payment-risk/overview',
        name: 'RiskStrategyOverview',
        component: () => import('@/views/RiskStrategy.vue'),
        meta: { title: '风控策略总览', permission: 'order:view' }
      },
      {
        path: 'payment-risk/3ds',
        name: 'RiskStrategyThreeDS',
        component: () => import('@/views/RiskStrategy.vue'),
        meta: { title: '3DS 策略', permission: 'order:view' }
      },
      {
        path: 'payment-risk/reviews',
        name: 'RiskStrategyReviews',
        component: () => import('@/views/RiskStrategy.vue'),
        meta: { title: '人工复核', permission: 'order:view' }
      },
      {
        path: 'payment-risk/refund-recommendations',
        name: 'RiskStrategyRefundRecommendations',
        component: () => import('@/views/RiskStrategy.vue'),
        meta: { title: '退款建议', permission: 'order:view' }
      },
      {
        path: 'payment-risk/controls',
        name: 'RiskStrategyControls',
        component: () => import('@/views/RiskStrategy.vue'),
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
        path: 'content',
        redirect: domainRedirect('ContentBlog', {
          blog: 'ContentBlog',
          feedback: 'ContentPageFeedback',
        }),
        meta: { permission: 'content:view' }
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
        path: 'content/showcase',
        name: 'ContentShowcase',
        component: () => import('@/views/ShowcaseModeration.vue'),
        meta: { title: '买家秀审批', permission: 'gallery:view' }
      },
      {
        path: 'content/reviews',
        name: 'ContentReviews',
        component: () => import('@/views/ReviewModeration.vue'),
        meta: { title: '评价审核', permission: 'review:view' }
      },
      {
        path: 'content/feedback',
        name: 'ContentPageFeedback',
        component: () => import('@/views/PageFeedback.vue'),
        alias: ['/content/page-feedback'],
        meta: { title: '页面留言', permission: 'content:view' }
      },
      {
        path: 'content/media-library',
        name: 'ContentMediaLibrary',
        component: () => import('@/views/MediaLibrary.vue'),
        meta: { title: '媒体库', permission: 'media:view' }
      },
      {
        path: 'content/visual-showcase',
        name: 'ContentVisualShowcase',
        component: () => import('@/views/VisualShowcase.vue'),
        meta: { title: '首页视觉目录', permission: 'content:view' }
      },
      {
        path: 'content/home-main-products',
        name: 'ContentHomeMainProducts',
        component: () => import('@/views/HomeMainProductCategories.vue'),
        meta: { title: '首页主力产品', permission: 'content:view' }
      },
      {
        path: 'content/media-derivatives',
        name: 'ContentMediaDerivatives',
        component: () => import('@/views/MediaDerivativePresets.vue'),
        meta: { title: '图片尺寸转换', permission: 'media:configure' }
      },
      {
        path: 'content/website-profile',
        name: 'ContentWebsiteProfile',
        component: () => import('@/views/SettingsWebsiteProfile.vue'),
        meta: { title: '我与这个网站', permission: 'settings:view' }
      },
      {
        path: 'support/conversations',
        name: 'SupportConversations',
        redirect: (to) => ({
          name: 'Dashboard',
          query: { ...to.query, inbox: 'customer-service' },
        }),
        meta: { title: '客服对话', permission: 'ticket:view' }
      },
      {
        path: 'support/analytics',
        name: 'SupportAnalytics',
        component: () => import('@/views/CustomerServiceAnalytics.vue'),
        meta: { title: '客服分析', permission: 'ticket:view' }
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
        path: 'urls',
        redirect: domainRedirect('URLOverview', {
          overview: 'URLOverview',
          catalog: 'URLCatalog',
          issues: 'URLIssues',
          redirects: 'URLRedirects',
          canonical: 'URLCanonical',
          operations: 'URLOperations',
        }),
        meta: { permission: 'url:view' }
      },
      {
        path: 'urls/overview',
        name: 'URLOverview',
        component: () => import('@/views/url-management/Overview.vue'),
        meta: { title: 'URL 管理 / 概览', permission: 'url:view' }
      },
      {
        path: 'urls/catalog',
        name: 'URLCatalog',
        component: () => import('@/views/url-management/RouteCatalog.vue'),
        props: { mode: 'catalog' },
        meta: { title: 'URL 管理 / 路由台账', permission: 'url:view' }
      },
      {
        path: 'urls/issues',
        name: 'URLIssues',
        component: () => import('@/views/url-management/Issues.vue'),
        meta: { title: 'URL 管理 / 问题队列', permission: 'url:view' }
      },
      {
        path: 'urls/redirects',
        name: 'URLRedirects',
        component: () => import('@/views/url-management/RedirectRules.vue'),
        meta: { title: 'URL 管理 / 重定向', permission: 'url:view' }
      },
      {
        path: 'urls/canonical',
        name: 'URLCanonical',
        component: () => import('@/views/url-management/RouteCatalog.vue'),
        props: { mode: 'canonical' },
        meta: { title: 'URL 管理 / Canonical 与冲突', permission: 'url:view' }
      },
      {
        path: 'urls/operations',
        name: 'URLOperations',
        component: () => import('@/views/url-management/Operations.vue'),
        meta: { title: 'URL 管理 / 同步与检查', permission: 'url:view' }
      },
      {
        path: 'preflight',
        redirect: { name: 'PreflightImageDimensions' }
      },
      {
        path: 'preflight/fonts',
        name: 'PreflightFonts',
        component: () => import('@/views/preflight/Fonts.vue'),
        meta: { title: '上线前检查 / 字体', permission: 'services:view' }
      },
      {
        path: 'preflight/site-quality',
        name: 'PreflightSiteQuality',
        component: () => import('@/views/preflight/SiteQuality.vue'),
        meta: { title: '上线前检查 / 页面质量', permission: 'services:view' }
      },
      {
        path: 'preflight/image-dimensions',
        name: 'PreflightImageDimensions',
        component: () => import('@/views/preflight/ImageDimensions.vue'),
        meta: { title: '上线前检查 / 图片尺寸', permission: 'media:view' }
      },
      {
        path: 'preflight/content-links',
        name: 'PreflightContentLinks',
        component: () => import('@/views/preflight/ContentLinks.vue'),
        meta: { title: '上线前检查 / 内容链接', permission: 'services:view' }
      },
      {
        path: 'social',
        redirect: { name: 'SocialOverview' },
        meta: { permission: 'settings:view' }
      },
      {
        path: 'social/overview',
        name: 'SocialOverview',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / 账号总览', permission: 'settings:view' }
      },
      {
        path: 'social/profiles',
        name: 'SocialProfiles',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / 前台展示', permission: 'settings:view' }
      },
      {
        path: 'social/youtube',
        name: 'SocialYouTube',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / YouTube', permission: 'settings:view' }
      },
      {
        path: 'social/meta',
        name: 'SocialMeta',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / Facebook / Instagram', permission: 'settings:view' }
      },
      {
        path: 'social/tiktok',
        name: 'SocialTikTok',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / TikTok', permission: 'settings:view' }
      },
      {
        path: 'social/linkedin',
        name: 'SocialLinkedIn',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / LinkedIn', permission: 'settings:view' }
      },
      {
        path: 'social/x',
        name: 'SocialX',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / X', permission: 'settings:view' }
      },
      {
        path: 'social/wechat',
        name: 'SocialWeChat',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / 微信', permission: 'settings:view' }
      },
      {
        path: 'social/publications',
        name: 'SocialPublications',
        component: () => import('@/views/SocialMedia.vue'),
        meta: { title: '社交媒体 / 发布记录', permission: 'settings:view' }
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: () => import('@/views/Analytics.vue'),
        meta: { title: 'Analytics', permission: 'analytics:view' }
      },
      {
        path: 'services',
        redirect: { name: 'ServicesOverview' },
        meta: { permission: 'services:view' }
      },
      {
        path: 'services/overview',
        name: 'ServicesOverview',
        component: () => import('@/views/services/ServiceCenter.vue'),
        meta: { title: '服务中心', permission: 'services:view' }
      },
      {
        path: 'services/cloudflare',
        name: 'ServicesCloudflare',
        component: () => import('@/views/services/CloudflareService.vue'),
        meta: { title: '服务中心 / Cloudflare', permission: 'services:view' }
      },
      {
        path: 'ops',
        redirect: { name: 'OpsOverview' },
        meta: { permission: 'ops:view' }
      },
      {
        path: 'ops/overview',
        name: 'OpsOverview',
        component: () => import('@/views/OpsOverview.vue'),
        meta: { title: '运维总览', permission: 'ops:view' }
      },
      {
        path: 'ops/admin-accounts',
        name: 'OpsAdminAccounts',
        component: () => import('@/views/OpsAdminAccounts.vue'),
        meta: { title: '后台账号管理', permission: 'system:manage' }
      },
      {
        path: 'ops/domains',
        name: 'OpsDomains',
        component: () => import('@/views/OpsDomains.vue'),
        meta: { title: '域名中心', permission: 'ops:domain:view' }
      },
      {
        path: 'ops/connectors',
        redirect: { name: 'ServicesOverview' },
        meta: { permission: 'services:view' }
      },
      {
        path: 'ops/vps',
        name: 'OpsVPS',
        component: () => import('@/views/OpsVPS.vue'),
        meta: { title: 'VPS 中心', permission: 'ops:vps:view' }
      },
      {
        path: 'ops/projects',
        name: 'OpsProjects',
        component: () => import('@/views/OpsProjects.vue'),
        meta: { title: '项目中心', permission: 'ops:project:view' }
      },
      {
        path: 'ops/deployments',
        name: 'OpsDeployments',
        component: () => import('@/views/OpsDeployments.vue'),
        meta: { title: '部署中心', permission: 'ops:deploy:view' }
      },
      {
        path: 'settings',
        redirect: domainRedirect('SettingsSite', {
          site: 'SettingsSite',
          email: 'SettingsEmail',
          social: 'SocialProfiles',
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
        redirect: { name: 'SocialProfiles' },
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
        path: 'settings/refund-return',
        name: 'SettingsRefundReturn',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '退货退款', permission: 'settings:view' }
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
