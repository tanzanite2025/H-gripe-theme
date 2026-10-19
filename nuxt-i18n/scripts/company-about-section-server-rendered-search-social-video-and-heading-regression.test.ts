import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { parse, parseFragment, type DefaultTreeAdapterMap } from 'parse5'
import { companyAboutTabs } from '../app/utils/pageSubNavigationData.ts'
import { holePatternAutomatedDrillingVideoMetadata } from '../app/data/company/holePatternAutomatedDrillingVideoMetadata.ts'
import { carbonRimLogoLaserEngravingVideoMetadata } from '../app/data/company/carbonRimLogoLaserEngravingVideoMetadata.ts'
import { carbonRimDrillingAndLacingPhotoWallImages } from '../app/data/company/carbonRimDrillingAndLacingPhotoWallImages.ts'
import localeManifest from '../app/i18n/locales.manifest.ts'

// Run against a running storefront; the default is the repository's development port.
const companyAboutStorefrontOrigin = process.env.COMPANY_ABOUT_SSR_ORIGIN || 'http://127.0.0.1:10240'
const companyAboutBackendOrigin = process.env.COMPANY_ABOUT_FAQ_BACKEND_ORIGIN || 'http://127.0.0.1:10241'
type CompanyAboutRenderedNode = DefaultTreeAdapterMap['node']
type CompanyAboutRenderedElement = DefaultTreeAdapterMap['element']

function collectCompanyAboutServerRenderedElements(node: CompanyAboutRenderedNode): CompanyAboutRenderedElement[] {
  return [
    ...('tagName' in node ? [node] : []),
    ...('childNodes' in node ? node.childNodes.flatMap(collectCompanyAboutServerRenderedElements) : []),
  ]
}

function readCompanyAboutServerRenderedAttribute(element: CompanyAboutRenderedElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readCompanyAboutServerRenderedText(node: CompanyAboutRenderedNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readCompanyAboutServerRenderedText).join('') : ''
}

interface CompanyAboutSectionRegressionMessages {
  title: string
  seo: { title: string; description: string; imageAlt: string; videoName?: string }
  lacingPhotoWall?: { title: string }
}

const companyAboutSectionMessageNamespaces = {
  factory: 'companyAboutFactory', appearance: 'companyAboutAppearance', 'hole-patterns': 'companyAboutHolePatterns',
} as const

assert.equal(new Set(companyAboutTabs.map(tab => tab.feedbackThreadKey)).size, companyAboutTabs.length)
for (const tab of companyAboutTabs) assert.notEqual(tab.feedbackThreadKey, 'company-ourstory')

for (const locale of ['en', 'zh_cn', 'fr']) {
  const localePrefix = locale === 'en' ? '' : `/${locale}`
  const contentLocale = locale === 'zh_cn' ? 'zh_cn' : 'en'

  for (const section of ['factory', 'appearance', 'hole-patterns', 'base'] as const) {
    const sectionId = section === 'base' ? 'factory' : section
    const routePath = `${localePrefix}/company/about${section === 'base' ? '' : `/${section}`}`
    const namespace = companyAboutSectionMessageNamespaces[sectionId]
    const messages: CompanyAboutSectionRegressionMessages = JSON.parse(await readFile(
      new URL(`../app/i18n/page-messages/${namespace}/${contentLocale}.json`, import.meta.url), 'utf8',
    ))
    const response = await fetch(new URL(routePath, companyAboutStorefrontOrigin))
    assert.equal(response.status, 200, routePath)
    const html = await response.text()
    const elements = collectCompanyAboutServerRenderedElements(parse(html))
    assert.equal(elements.filter(element => readCompanyAboutServerRenderedAttribute(element, 'class')
      .split(/\s+/).includes('share-your-feedback-card')).length, 1, `${routePath}: one layout-owned feedback thread`)
    const headings = elements.filter(element => element.tagName === 'h1')
    assert.equal(headings.length, 1, `${routePath}: one page heading`)
    assert.equal(readCompanyAboutServerRenderedText(headings[0]!).trim(), messages.title)
    assert.doesNotMatch(readCompanyAboutServerRenderedAttribute(headings[0]!, 'class'), /sr-only/)
    const titles = elements.filter(element => element.tagName === 'title')
    assert.equal(titles.length, 1)
    const title = readCompanyAboutServerRenderedText(titles[0]!)
    assert.ok(title.startsWith(messages.seo.title))

    function readCompanyAboutUniqueServerRenderedMetadata(metadataKey: string): string {
      const matches = elements.filter(element => element.tagName === 'meta'
        && (readCompanyAboutServerRenderedAttribute(element, 'name') === metadataKey
          || readCompanyAboutServerRenderedAttribute(element, 'property') === metadataKey))
      assert.equal(matches.length, 1, `${routePath}: ${metadataKey}`)
      const content = readCompanyAboutServerRenderedAttribute(matches[0]!, 'content')
      assert.ok(content.length > 0)
      return content
    }

    assert.equal(readCompanyAboutUniqueServerRenderedMetadata('description'), messages.seo.description)
    for (const metadataKey of ['og:title', 'twitter:title']) {
      assert.equal(readCompanyAboutUniqueServerRenderedMetadata(metadataKey), title)
    }
    for (const metadataKey of ['og:description', 'twitter:description']) {
      assert.equal(readCompanyAboutUniqueServerRenderedMetadata(metadataKey), messages.seo.description)
    }
    const socialImageUrl = readCompanyAboutUniqueServerRenderedMetadata('og:image')
    assert.equal(readCompanyAboutUniqueServerRenderedMetadata('twitter:image'), socialImageUrl)
    assert.equal(readCompanyAboutUniqueServerRenderedMetadata('twitter:card'), 'summary_large_image')
    const canonicalLinks = elements.filter(element => element.tagName === 'link'
      && readCompanyAboutServerRenderedAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1)
    const canonicalUrl = readCompanyAboutServerRenderedAttribute(canonicalLinks[0]!, 'href')
    assert.equal(new URL(canonicalUrl).pathname, routePath)
    assert.equal(readCompanyAboutUniqueServerRenderedMetadata('og:url'), canonicalUrl)

    const structuredData = elements.filter(element => element.tagName === 'script'
      && readCompanyAboutServerRenderedAttribute(element, 'type') === 'application/ld+json')
      .flatMap(element => {
        const data = JSON.parse(readCompanyAboutServerRenderedText(element))
        return data['@graph'] || [data]
      })
    const pages = structuredData.filter(data => data['@type'] === 'WebPage')
    assert.equal(pages.length, 1)
    assert.equal(pages[0].name, messages.title)
    assert.equal(pages[0].description, messages.seo.description)
    assert.equal(pages[0].inLanguage, contentLocale === 'zh_cn' ? 'zh-CN' : 'en-US')
    const breadcrumbs = structuredData.filter(data => data['@type'] === 'BreadcrumbList')
    assert.equal(breadcrumbs.length, 1)
    assert.equal(breadcrumbs[0]['@id'], pages[0].breadcrumb['@id'])
    assert.equal(breadcrumbs[0].itemListElement.at(-1).item, canonicalUrl)
    assert.equal(new URL(breadcrumbs[0].itemListElement[0].item).pathname, localePrefix || '/')
    const videos = structuredData.filter(data => data['@type'] === 'VideoObject')
    assert.equal(videos.length, sectionId === 'factory' ? 0 : 1)
    const imageGalleries = structuredData.filter(data => data['@type'] === 'ImageGallery')
    const photoImageObjects = structuredData.filter(data => data['@type'] === 'ImageObject')
    assert.equal(imageGalleries.length, sectionId === 'hole-patterns' ? 1 : 0, `${routePath}: gallery belongs to its own section`)
    assert.equal(photoImageObjects.length, sectionId === 'hole-patterns' ? carbonRimDrillingAndLacingPhotoWallImages.length : 0)
    const entityIds = structuredData.map(data => data['@id']).filter(Boolean)
    assert.equal(new Set(entityIds).size, entityIds.length, `${routePath}: no duplicate schema entities`)
    const faqs = structuredData.filter(data => data['@type'] === 'FAQPage')
    assert.doesNotMatch(html, /companyAbout(?:Appearance|HolePatterns)\.(?:navigation|seo)\./)
    assert.doesNotMatch(html, /company\.aboutTabs\./)

    if (sectionId !== 'factory') {
      const videoMetadata = sectionId === 'appearance' ? carbonRimLogoLaserEngravingVideoMetadata : holePatternAutomatedDrillingVideoMetadata
      const video = videos[0]
      assert.equal(video.name, messages.seo.videoName)
      assert.equal(video.duration, videoMetadata.duration)
      assert.equal(video.uploadDate, videoMetadata.uploadDate)
      assert.equal(new URL(video.contentUrl).pathname, videoMetadata.contentPath)
      assert.deepEqual(video.thumbnailUrl, [socialImageUrl])
      assert.equal(video['@id'], pages[0].video['@id'])
    }

    if (sectionId === 'appearance') {
      const backendFaqUrl = new URL('/api/v1/content/faq-pages/by-route', companyAboutBackendOrigin)
      backendFaqUrl.searchParams.set('route_path', '/company/about/appearance')
      backendFaqUrl.searchParams.set('locale', locale)
      const backendFaqResponse = await fetch(backendFaqUrl)
      assert.ok(backendFaqResponse.status === 200 || backendFaqResponse.status === 404)
      const backendFaqPage = backendFaqResponse.ok ? (await backendFaqResponse.json()).page : null
      const backendFaqItems: { question: string; answer: string }[] = backendFaqPage?.items || []
      assert.equal(faqs.length, backendFaqItems.length ? 1 : 0, 'Only backend-published questions generate FAQ schema')
      assert.ok(!elements.some(element => readCompanyAboutServerRenderedAttribute(element, 'id') === 'appearance-frequently-asked-questions-title'))
      if (backendFaqItems.length) {
        assert.equal(faqs[0].inLanguage, localeManifest.find(entry => entry.code === locale)?.iso)
        assert.equal(faqs[0].mainEntity.length, backendFaqItems.length)
        const visibleFaqSection = elements.find(element => readCompanyAboutServerRenderedAttribute(element, 'class').split(/\s+/).includes('page-faq'))!
        assert.ok(visibleFaqSection, 'The shared backend-configured FAQ module must render')
        const visibleFaqText = readCompanyAboutServerRenderedText(visibleFaqSection).replace(/\s+/g, ' ')
        for (const item of backendFaqItems) {
          const backendAnswerText = readCompanyAboutServerRenderedText(parseFragment(item.answer)).replace(/\s+/g, ' ').trim()
          assert.ok(visibleFaqText.includes(item.question), item.question)
          assert.ok(visibleFaqText.includes(backendAnswerText), 'The complete answer must be present without JavaScript or expansion')
          assert.ok(faqs[0].mainEntity.some((question: { name: string; acceptedAnswer: { text: string } }) => (
            question.name === item.question && question.acceptedAnswer.text === backendAnswerText
          )))
        }
      }
      const customOptionsSection = elements.find(element => readCompanyAboutServerRenderedAttribute(element, 'aria-labelledby') === 'appearance-custom-options-title')!
      assert.equal(customOptionsSection?.tagName, 'section', 'Weave and finish options must be visible without opening details')
      assert.ok(elements.some(element => readCompanyAboutServerRenderedAttribute(element, 'id') === 'appearance-weave-patterns'))
      assert.ok(elements.some(element => readCompanyAboutServerRenderedAttribute(element, 'id') === 'appearance-gloss-levels'))
      const posterImages = elements.filter(element => element.tagName === 'img'
        && readCompanyAboutServerRenderedAttribute(element, 'src') === carbonRimLogoLaserEngravingVideoMetadata.poster.path)
      assert.equal(posterImages.length, 1)
    }

    if (sectionId === 'hole-patterns') {
      const galleryElementId = 'hole-pattern-drilled-rim-lacing-photo-gallery'
      const visibleGallery = elements.find(element => readCompanyAboutServerRenderedAttribute(element, 'id') === galleryElementId)
      assert.ok(visibleGallery)
      const galleryElements = collectCompanyAboutServerRenderedElements(visibleGallery)
      const galleryPhotos = galleryElements.filter(element => element.tagName === 'img')
      const galleryCaptions = galleryElements.filter(element => element.tagName === 'figcaption')
      assert.equal(galleryPhotos.length, carbonRimDrillingAndLacingPhotoWallImages.length,
        `${routePath}: every photo, including later groups, is present without JavaScript`)
      assert.equal(galleryCaptions.length, galleryPhotos.length)
      const gallery = imageGalleries[0]
      assert.equal(gallery['@id'], `${canonicalUrl}#${galleryElementId}`)
      assert.equal(gallery.name, messages.lacingPhotoWall?.title)
      assert.equal(gallery.inLanguage, pages[0].inLanguage)
      assert.equal(gallery.isPartOf['@id'], pages[0]['@id'])
      assert.equal(pages[0].hasPart['@id'], gallery['@id'])
      assert.deepEqual(pages[0].image, gallery.associatedMedia)
      assert.deepEqual(gallery.associatedMedia.map((photo: { '@id': string }) => photo['@id']),
        photoImageObjects.map(photo => photo['@id']))

      for (const [photoIndex, photo] of carbonRimDrillingAndLacingPhotoWallImages.entries()) {
        const renderedPhoto = galleryPhotos[photoIndex]!
        const renderedCaption = readCompanyAboutServerRenderedText(galleryCaptions[photoIndex]!).trim()
        assert.ok(readCompanyAboutServerRenderedAttribute(renderedPhoto, 'src').includes(photo.src.split('/').at(-1)!))
        assert.equal(readCompanyAboutServerRenderedAttribute(renderedPhoto, 'alt'), photo.alt[contentLocale])
        assert.equal(renderedCaption, photo.caption[contentLocale])
        const imageObject = photoImageObjects[photoIndex]
        assert.equal(imageObject.contentUrl, new URL(photo.src, canonicalUrl).toString())
        assert.equal(imageObject.url, imageObject.contentUrl)
        assert.equal(imageObject.name, renderedCaption)
        assert.equal(imageObject.caption, renderedCaption)
        assert.equal(imageObject.description, readCompanyAboutServerRenderedAttribute(renderedPhoto, 'alt'))
        assert.equal(imageObject.width, photo.width)
        assert.equal(imageObject.height, photo.height)
        assert.equal(imageObject.encodingFormat, photo.mimeType)
        assert.equal(imageObject.inLanguage, pages[0].inLanguage)
        assert.equal(imageObject.isPartOf['@id'], gallery['@id'])
      }
      const previews = elements.filter(element => element.tagName === 'video')
      assert.equal(previews.length, 0, 'The video player loads only after clicking its static poster')
      const posterImages = elements.filter(element => element.tagName === 'img'
        && readCompanyAboutServerRenderedAttribute(element, 'src') === new URL(socialImageUrl).pathname)
      assert.equal(posterImages.length, 1)
      assert.equal(readCompanyAboutServerRenderedAttribute(posterImages[0]!, 'alt'), messages.seo.imageAlt)
      const consultation = elements.find(element => readCompanyAboutServerRenderedAttribute(element, 'class') === 'hole-pattern-drawing-consultation')!
      assert.ok(consultation)
      const actions = collectCompanyAboutServerRenderedElements(consultation)
      assert.ok(actions.some(element => element.tagName === 'a'
        && readCompanyAboutServerRenderedAttribute(element, 'href') === `${localePrefix}/company/contact`))
      assert.ok(actions.some(element => element.tagName === 'button'))
      assert.ok(elements.some(element => element.tagName === 'a'
        && readCompanyAboutServerRenderedAttribute(element, 'href') === `${localePrefix}/tools/spoke-lacing-topology-simulator`))
    }
    console.log(`PASS ${routePath}: visible heading, localized metadata and section-specific JSON-LD`)
  }
}

for (const mediaPath of [
  holePatternAutomatedDrillingVideoMetadata.contentPath, holePatternAutomatedDrillingVideoMetadata.poster.path,
  carbonRimLogoLaserEngravingVideoMetadata.contentPath, carbonRimLogoLaserEngravingVideoMetadata.poster.path,
  ...carbonRimDrillingAndLacingPhotoWallImages.map(photo => photo.src),
]) {
  const response = await fetch(new URL(mediaPath, companyAboutStorefrontOrigin), { method: 'HEAD' })
  assert.equal(response.status, 200, mediaPath)
  assert.match(response.headers.get('content-type') || '', mediaPath.endsWith('.webm') ? /^video\/webm/ : mediaPath.endsWith('.mp4') ? /^video\/mp4/ : /^image\/webp/)
}
console.log('PASS manufacturing videos, posters and all lacing gallery photographs are directly accessible')
