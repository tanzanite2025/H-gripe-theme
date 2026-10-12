import { defineEventHandler, getRequestURL, setHeader } from 'h3'
import { useRuntimeConfig } from '#imports'
import commercialCrawlerRobotsExclusionPolicy from '../../config/storefront/commercialCrawlerRobotsExclusionPolicy.json'

export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event)
  const configuredSiteUrl = String(config.public?.siteUrl || '').trim().replace(/\/+$/, '')
  const siteUrl = configuredSiteUrl || getRequestURL(event).origin
  const lines = [
    ...commercialCrawlerRobotsExclusionPolicy.blocked_user_agents.flatMap((crawler) => [`User-agent: ${crawler}`, 'Disallow: /', '']),
    `User-agent: ${commercialCrawlerRobotsExclusionPolicy.wildcard_user_agent}`,
    `Disallow: ${commercialCrawlerRobotsExclusionPolicy.wildcard_disallow}`.trimEnd(),
    '',
    `Sitemap: ${siteUrl}/sitemap.xml`,
    '',
  ]

  setHeader(event, 'Content-Type', 'text/plain; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=3600, s-maxage=86400')
  return lines.join('\n')
})
