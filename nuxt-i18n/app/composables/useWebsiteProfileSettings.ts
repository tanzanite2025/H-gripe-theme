import { computed, useAsyncData } from '#imports'
import type { Ref } from 'vue'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'
import { usePublicApiBase } from '~/composables/usePublicApiBase'

export interface WebsiteProfileSettings {
  locale: string
  eyebrow: string
  title: string
  lead: string
  scope: string
  avatarUrl: string
  avatarLabel: string
  avatarMark: string
  profileLabel: string
  profileRole: string
  profileContext: string
  statementEyebrow: string
  statementTitle: string
  statementParagraphs: string[]
  factoryImageUrl: string
  factoryImageAlt: string
  factoryImageCaption: string
  factoryEyebrow: string
  factoryTitle: string
  factoryBody: string
  factoryCta: string
  factoryLink: string
}

type RawWebsiteProfileSettings = Record<string, unknown>

const defaultFactoryImageUrl = '/company/ourstory/factory/tanzanite-factory-premoldlayupworkshop6.webp'

const asString = (value: unknown, fallback = '') => {
  const result = typeof value === 'string' ? value : value === null || value === undefined ? '' : String(value)
  return result.trim() || fallback
}

export const defaultWebsiteProfileSettings = (locale: unknown): WebsiteProfileSettings => {
  if (isSimplifiedChineseStorefrontLocale(locale)) {
    return {
      locale: String(locale || 'en'),
      eyebrow: '网站管理者 / 工厂成员',
      title: '我与这个网站',
      lead: '我负责这个网站的内容、结构和持续维护，也属于我们工厂正在做的事情。这个域名承载的是我的工作视角，而不是把我从工厂之外单独分离出来。',
      scope: '这个网站由我负责管理，也代表我们工厂的一部分工作',
      avatarUrl: '',
      avatarLabel: '网站管理者头像位置',
      avatarMark: '我',
      profileLabel: '网站管理者',
      profileRole: '网站内容与方向',
      profileContext: '我们工厂的一员',
      statementEyebrow: '为什么有这一页',
      statementTitle: '让网站背后的人被看见',
      statementParagraphs: [
        '这个域名属于我，但它表达的并不是一个脱离工厂的个人身份。相反，我希望用更接近个人的方式，说明我如何理解我们的工厂、产品和长期方向。',
        '这里会记录网站背后的判断、正在推进的事情，以及我认为应该被准确表达的内容。它不是客服窗口，也不是单独成立的另一家公司，而是我们工厂工作中的一个管理和表达入口。',
      ],
      factoryImageUrl: defaultFactoryImageUrl,
      factoryImageAlt: '我们工厂的碳纤维手工铺层工序',
      factoryImageCaption: '我负责表达的网站，来自我们真实的制造工作。',
      factoryEyebrow: '我们共同的工作',
      factoryTitle: '查看我们的工厂',
      factoryBody: '网站上的内容最终要回到真实的产品、制造、研发和质量控制。这里是我的视角，但它所指向的仍然是我们正在一起建设的工厂。',
      factoryCta: '查看工厂与制造流程',
      factoryLink: '/company/about#factory',
    }
  }

  return {
    locale: 'en',
    eyebrow: 'THE PERSON BEHIND THIS WEBSITE',
    title: 'Me & This Website',
    lead: 'I manage the content, structure, and ongoing maintenance of this website while remaining part of our factory and the work it represents. This domain carries my working perspective; it does not separate me from the factory.',
    scope: 'Managed by me, grounded in the work of our factory',
    avatarUrl: '',
    avatarLabel: 'Website manager avatar placeholder',
    avatarMark: 'ME',
    profileLabel: 'Website manager',
    profileRole: 'Content and direction',
    profileContext: 'Part of our factory',
    statementEyebrow: 'WHY THIS PAGE EXISTS',
    statementTitle: 'Let the person behind the site be visible',
    statementParagraphs: [
      'This domain belongs to me, but it does not describe a personal identity outside the factory. It gives me a more direct way to explain how I see our factory, our products, and the direction we are building toward.',
      'This is where I can record the decisions behind the website, the work in progress, and the things I believe should be represented accurately. It is not a support desk or a separate company. It is one management and expression point within our factory work.',
    ],
    factoryImageUrl: defaultFactoryImageUrl,
    factoryImageAlt: 'Carbon fiber hand layup work inside our factory',
    factoryImageCaption: 'The site I manage is grounded in our real manufacturing work.',
    factoryEyebrow: 'THE WORK WE SHARE',
    factoryTitle: 'See our factory',
    factoryBody: 'The site should always lead back to real products, manufacturing, engineering, and quality control. This is my perspective, but it points to the factory we are building together.',
    factoryCta: 'View the factory and process',
    factoryLink: '/company/about#factory',
  }
}

const normalizeWebsiteProfileSettings = (
  raw: RawWebsiteProfileSettings | null | undefined,
  locale: unknown,
): WebsiteProfileSettings => {
  const fallback = defaultWebsiteProfileSettings(locale)
  const statementParagraphs = [
    asString(raw?.statement_paragraph_1, fallback.statementParagraphs[0]),
    asString(raw?.statement_paragraph_2, fallback.statementParagraphs[1]),
  ]

  return {
    locale: asString(raw?.locale, fallback.locale),
    eyebrow: asString(raw?.eyebrow, fallback.eyebrow),
    title: asString(raw?.title, fallback.title),
    lead: asString(raw?.lead, fallback.lead),
    scope: asString(raw?.scope, fallback.scope),
    avatarUrl: asString(raw?.avatar_url, fallback.avatarUrl),
    avatarLabel: asString(raw?.avatar_label, fallback.avatarLabel),
    avatarMark: asString(raw?.avatar_mark, fallback.avatarMark),
    profileLabel: asString(raw?.profile_label, fallback.profileLabel),
    profileRole: asString(raw?.profile_role, fallback.profileRole),
    profileContext: asString(raw?.profile_context, fallback.profileContext),
    statementEyebrow: asString(raw?.statement_eyebrow, fallback.statementEyebrow),
    statementTitle: asString(raw?.statement_title, fallback.statementTitle),
    statementParagraphs,
    factoryImageUrl: asString(raw?.factory_image_url, fallback.factoryImageUrl),
    factoryImageAlt: asString(raw?.factory_image_alt, fallback.factoryImageAlt),
    factoryImageCaption: asString(raw?.factory_image_caption, fallback.factoryImageCaption),
    factoryEyebrow: asString(raw?.factory_eyebrow, fallback.factoryEyebrow),
    factoryTitle: asString(raw?.factory_title, fallback.factoryTitle),
    factoryBody: asString(raw?.factory_body, fallback.factoryBody),
    factoryCta: asString(raw?.factory_cta, fallback.factoryCta),
    factoryLink: asString(raw?.factory_link, fallback.factoryLink),
  }
}

export function useWebsiteProfileSettings(locale: Ref<string> | string) {
  const apiBase = usePublicApiBase()
  const localeValue = computed(() => String(typeof locale === 'string' ? locale : locale.value || 'en'))
  const key = computed(() => `mytheme-website-profile-${localeValue.value}`)

  const { data, pending, error } = useAsyncData<WebsiteProfileSettings>(
    key,
    async () => {
      const fallback = defaultWebsiteProfileSettings(localeValue.value)
      if (!apiBase.value) return fallback

      try {
        const raw = await $fetch<RawWebsiteProfileSettings>(
          `${apiBase.value}/settings/website-profile`,
          {
            query: { locale: localeValue.value },
            headers: { accept: 'application/json' },
          },
        )
        return normalizeWebsiteProfileSettings(raw, localeValue.value)
      } catch (fetchError) {
        console.warn('Failed to load website profile settings:', fetchError)
        return fallback
      }
    },
    {
      default: () => defaultWebsiteProfileSettings(localeValue.value),
      watch: [localeValue],
    },
  )

  const websiteProfileSettings = computed(() =>
    data.value || defaultWebsiteProfileSettings(localeValue.value),
  )

  return {
    websiteProfileSettings,
    pending,
    error,
  }
}
