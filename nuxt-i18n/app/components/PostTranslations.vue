<template>
  <div v-if="hasTranslations" class="post-translations">
    <h3 class="translations-title">
      {{ title || 'Available in other languages' }}
    </h3>
    
    <ul class="translations-list">
      <li 
        v-for="trans in visibleTranslations"
        :key="trans.locale"
        :class="{ 'current-locale': trans.locale === currentLocaleCode }"
      >
        <NuxtLink 
          :to="trans.url" 
          class="translation-link"
          :aria-current="trans.locale === currentLocaleCode ? 'page' : undefined"
        >
          <span class="locale-flag">{{ getFlagEmoji(trans.locale) }}</span>
          <span class="locale-name" :lang="getLocaleLanguageTag(trans.locale)">{{ getLocaleName(trans.locale) }}</span>
          <span v-if="trans.locale === currentLocaleCode" class="current-badge">
            (Current)
          </span>
        </NuxtLink>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useBlogApi } from '~/composables/useBlogApi'
import { useI18n } from '#imports'
import {
  getStorefrontLocaleFlag,
  getStorefrontLocaleLanguageTag,
  getStorefrontLocaleName,
  normalizeStorefrontLocaleCode,
} from '~/utils/storefrontLocales'
import type { BlogLocalizedRoute } from '~/utils/blog/types'

interface PostTranslation {
  id?: number
  title?: string
  slug: string
  locale: string
  published_at?: string
  url: string
}

interface Props {
  postId: number
  title?: string
  showCurrentLocale?: boolean
  localizedRoutes?: BlogLocalizedRoute[]
}

const props = withDefaults(defineProps<Props>(), {
  showCurrentLocale: true,
  localizedRoutes: () => [],
})

const { locale: currentLocale } = useI18n()
const { getPostTranslations } = useBlogApi()

const translations = ref<Record<string, PostTranslation>>({})
const currentLocaleCode = computed(() => normalizeStorefrontLocaleCode(currentLocale.value) || 'en')
const serverTranslations = computed<Record<string, PostTranslation>>(() => {
  return props.localizedRoutes.reduce<Record<string, PostTranslation>>((result, route) => {
    const locale = normalizeStorefrontLocaleCode(route.locale)
    if (!locale || !route.slug || !route.path) return result

    result[locale] = {
      id: route.id,
      slug: route.slug,
      locale,
      url: route.path,
    }
    return result
  }, {})
})

const visibleTranslations = computed<PostTranslation[]>(() => {
  const byLocale = new Map<string, PostTranslation>()
  const source = Object.keys(serverTranslations.value).length
    ? serverTranslations.value
    : translations.value

  for (const [rawLocale, translation] of Object.entries(source)) {
    const canonicalLocale = normalizeStorefrontLocaleCode(translation.locale || rawLocale)
    if (!canonicalLocale) continue
    if (!props.showCurrentLocale && canonicalLocale === currentLocaleCode.value) continue

    byLocale.set(canonicalLocale, {
      ...translation,
      locale: canonicalLocale,
    })
  }

  return Array.from(byLocale.values())
})

// 计算是否有翻译
const hasTranslations = computed(() => {
  return props.showCurrentLocale ? visibleTranslations.value.length > 1 : visibleTranslations.value.length > 0
})

onMounted(async () => {
  if (Object.keys(serverTranslations.value).length) return

  try {
    translations.value = await getPostTranslations(props.postId)
  } catch {
    translations.value = {}
  }
})

// 获取语言名称
const getLocaleName = (localeCode: string): string => {
  return getStorefrontLocaleName(localeCode, localeCode)
}

const getLocaleLanguageTag = (localeCode: string): string => {
  return getStorefrontLocaleLanguageTag(localeCode, localeCode.replace('_', '-'))
}

// 获取国旗 Emoji
const getFlagEmoji = (locale: string): string => {
  return getStorefrontLocaleFlag(locale)
}
</script>

<style scoped>
.post-translations {
  margin: 2rem 0;
  padding: 1.5rem;
  background-color: #f9fafb;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}

.translations-title {
  margin: 0 0 1rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
}

.translations-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.translations-list li {
  margin: 0;
}

.translation-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  text-decoration: none;
  color: #374151;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.translation-link:hover {
  background-color: #eff6ff;
  border-color: #3b82f6;
  color: #3b82f6;
}

.current-locale .translation-link {
  background-color: #3b82f6;
  border-color: #3b82f6;
  color: white;
  cursor: default;
}

.locale-flag {
  font-size: 1.25rem;
  line-height: 1;
}

.locale-name {
  font-weight: 500;
}

.current-badge {
  font-size: 0.75rem;
  opacity: 0.8;
}

/* 响应式 */
@media (max-width: 640px) {
  .translations-list {
    flex-direction: column;
  }
  
  .translation-link {
    width: 100%;
  }
}
</style>
