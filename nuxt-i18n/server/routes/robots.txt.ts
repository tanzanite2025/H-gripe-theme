import { defineEventHandler, setHeader } from 'h3'
import { useRuntimeConfig } from '#imports'

const blockedCrawlers = ['AhrefsBot', 'SemrushBot', 'SimilarwebBot', 'BuiltWith']

export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event)
  const configuredSiteUrl = String(config.public?.siteUrl || '').trim().replace(/\/+$/, '')
  const siteUrl = configuredSiteUrl || 'https://learn.gripe'
  const lines = [
    ...blockedCrawlers.flatMap((crawler) => [`User-agent: ${crawler}`, 'Disallow: /', '']),
    'User-agent: *',
    'Disallow:',
    '',
    `Sitemap: ${siteUrl}/sitemap.xml`,
    '',
  ]

  setHeader(event, 'Content-Type', 'text/plain; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=3600, s-maxage=86400')
  return lines.join('\n')
})
