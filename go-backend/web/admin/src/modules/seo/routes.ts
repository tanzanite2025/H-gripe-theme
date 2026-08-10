const storefrontBaseUrl = String(import.meta.env.VITE_STOREFRONT_URL || '').replace(/\/+$/, '')

export const storefrontHref = (path: string): string => `${storefrontBaseUrl}${path.startsWith('/') ? path : `/${path}`}`

export const buildBlogHubPath = (): string => '/blog'

export const buildArticlePath = (resource: { route_path?: string | null }): string => {
  return String(resource.route_path || '').trim()
}

export const buildProductPath = (resource: { route_path?: string | null }): string => {
  return String(resource.route_path || '').trim()
}
