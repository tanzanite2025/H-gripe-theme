import type { RouteLocationNormalizedLoaded } from 'vue-router'

export type AppNavContext = 'company' | 'support' | 'guides' | 'products' | 'blog'

const isForcedProductsContext = (route: RouteLocationNormalizedLoaded): boolean => {
  const rawNav = route.query?.nav
  const value = Array.isArray(rawNav) ? rawNav[0] : rawNav

  if (typeof value !== 'string') return false

  return value.toLowerCase() === 'products'
}

export const getNavContextFromRoute = (
  route: RouteLocationNormalizedLoaded,
): AppNavContext => {
  const path = route.path || ''

  const forcedProducts = isForcedProductsContext(route)

  // 只有在限定的路由下才接受 nav=products 的覆盖
  const allowProductsOverride =
    path.startsWith('/guides/') || path === '/support/test-report'

  if (forcedProducts && allowProductsOverride) {
    return 'products'
  }

  if (path === '/company' || path.startsWith('/company/')) {
    return 'company'
  }

  if (path === '/support' || path.startsWith('/support/')) {
    return 'support'
  }

  if (path === '/guides' || path.startsWith('/guides/')) {
    return 'guides'
  }

  if (path === '/blog' || path.startsWith('/blog/')) {
    return 'blog'
  }

  if (
    path === '/products' ||
    path === '/shop' ||
    path === '/spoke-calculator' ||
    path === '/membershipandpoints' ||
    path === '/picture-warehouse'
  ) {
    return 'products'
  }

  // 兜底：其余未明确归类的路由，都视为 products 上下文
  return 'products'
}
