import type { ResolvableScript } from '@unhead/vue'

export type SeoJsonLdScript = Exclude<ResolvableScript, string>

const escapeJsonLd = (value: string): string => value
  .replace(/</g, '\\u003c')
  .replace(/>/g, '\\u003e')
  .replace(/&/g, '\\u0026')

export const createSeoJsonLdScript = (value: unknown): SeoJsonLdScript => ({
  type: 'application/ld+json',
  textContent: escapeJsonLd(JSON.stringify(value)),
})
