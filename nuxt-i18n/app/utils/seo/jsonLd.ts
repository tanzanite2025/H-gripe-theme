import type { ResolvableScript } from '@unhead/vue'

export type SeoJsonLdScript = Exclude<ResolvableScript, string>

export interface SeoFaqQuestionAnswer {
  question: string
  answer: string
}

interface CreateFaqPageJsonLdOptions {
  canonicalUrl: string
  pageId: string
  inLanguage: string
  visibleQuestionsAndAnswers: readonly SeoFaqQuestionAnswer[]
}

const escapeJsonLd = (value: string): string => value
  .replace(/</g, '\\u003c')
  .replace(/>/g, '\\u003e')
  .replace(/&/g, '\\u0026')

export const createSeoJsonLdScript = (value: unknown): SeoJsonLdScript => ({
  type: 'application/ld+json',
  textContent: escapeJsonLd(JSON.stringify(value)),
})

const stripHtmlTagsAndNormalizeFaqAnswerText = (richAnswer: string): string => richAnswer
  .replace(/<(script|style)\b[^>]*>[\s\S]*?<\/\1>/gi, ' ')
  .replace(/<br\s*\/?>|<\/(?:p|li|div|h[1-6]|tr|blockquote)>/gi, ' ')
  .replace(/<[^>]+>/g, ' ')
  .replace(/&nbsp;|&#160;|&#x0*a0;/gi, ' ')
  .replace(/&quot;/gi, '"')
  .replace(/&apos;|&#39;/gi, "'")
  .replace(/&lt;/gi, '<')
  .replace(/&gt;/gi, '>')
  .replace(/&#(\d+);/g, (_, decimalValue: string) => {
    const codePoint = Number(decimalValue)
    return Number.isInteger(codePoint) && codePoint >= 0 && codePoint <= 0x10ffff
      ? String.fromCodePoint(codePoint)
      : ' '
  })
  .replace(/&#x([\da-f]+);/gi, (_, hexadecimalValue: string) => {
    const codePoint = Number.parseInt(hexadecimalValue, 16)
    return Number.isInteger(codePoint) && codePoint >= 0 && codePoint <= 0x10ffff
      ? String.fromCodePoint(codePoint)
      : ' '
  })
  .replace(/&amp;/gi, '&')
  .replace(/\s+/g, ' ')
  .trim()

export const createFaqPageJsonLd = ({
  canonicalUrl,
  pageId,
  inLanguage,
  visibleQuestionsAndAnswers,
}: CreateFaqPageJsonLdOptions) => {
  const mainEntity = visibleQuestionsAndAnswers
    .map(item => ({
      question: item.question.trim(),
      answer: stripHtmlTagsAndNormalizeFaqAnswerText(item.answer),
    }))
    .filter(item => item.question && item.answer)
    .map(item => ({
      '@type': 'Question',
      name: item.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: item.answer,
      },
    }))

  if (!mainEntity.length) return null

  return {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    '@id': `${canonicalUrl}#faq-${encodeURIComponent(pageId)}`,
    url: canonicalUrl,
    inLanguage,
    mainEntity,
  }
}
