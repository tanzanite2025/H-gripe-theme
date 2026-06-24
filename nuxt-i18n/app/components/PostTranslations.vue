<template>
  <div v-if="hasTranslations" class="post-translations">
    <h3 class="translations-title">
      {{ title || 'Available in other languages' }}
    </h3>
    
    <ul class="translations-list">
      <li 
        v-for="(trans, locale) in translations" 
        :key="locale"
        :class="{ 'current-locale': locale === currentLocale }"
      >
        <NuxtLink 
          :to="trans.url" 
          class="translation-link"
          :aria-current="locale === currentLocale ? 'page' : undefined"
        >
          <span class="locale-flag">{{ getFlagEmoji(locale) }}</span>
          <span class="locale-name">{{ getLocaleName(locale) }}</span>
          <span v-if="locale === currentLocale" class="current-badge">
            (Current)
          </span>
        </NuxtLink>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useBlogApi } from '~/composables/useBlogApi'
import { useI18n } from '#imports'

interface PostTranslation {
  id: number
  title: string
  slug: string
  locale: string
  published_at: string
  url: string
}

interface Props {
  postId: number
  title?: string
  showCurrentLocale?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showCurrentLocale: true
})

const { locale: currentLocale, locales } = useI18n()
const { getPostTranslations } = useBlogApi()

const translations = ref<Record<string, PostTranslation>>({})

// 计算是否有翻译
const hasTranslations = computed(() => {
  const count = Object.keys(translations.value).length
  return props.showCurrentLocale ? count > 1 : count > 0
})

// 加载翻译数据
onMounted(async () => {
  translations.value = await getPostTranslations(props.postId)
})

// 获取语言名称
const getLocaleName = (localeCode: string): string => {
  const allLocales = locales.value as Array<{ code: string; name?: string; nativeName?: string }> | undefined
  const found = allLocales?.find(l => l.code === localeCode)
  return found?.name || found?.nativeName || localeCode
}

// 获取国旗 Emoji
const getFlagEmoji = (locale: string): string => {
  const flagMap: Record<string, string> = {
    'en': '🇬🇧',
    'zh': '🇨🇳',
    'zh-TW': '🇹🇼',
    'ja': '🇯🇵',
    'ko': '🇰🇷',
    'fr': '🇫🇷',
    'de': '🇩🇪',
    'es': '🇪🇸',
    'it': '🇮🇹',
    'pt': '🇵🇹',
    'ru': '🇷🇺',
    'ar': '🇸🇦',
    'nl': '🇳🇱',
    'pl': '🇵🇱',
    'tr': '🇹🇷',
    'vi': '🇻🇳',
    'th': '🇹🇭',
    'id': '🇮🇩',
    'ms': '🇲🇾',
    'hi': '🇮🇳',
    'bn': '🇧🇩',
    'ta': '🇮🇳',
    'te': '🇮🇳',
    'mr': '🇮🇳',
    'ur': '🇵🇰',
    'fa': '🇮🇷',
    'he': '🇮🇱',
    'sv': '🇸🇪',
    'no': '🇳🇴',
    'da': '🇩🇰',
    'fi': '🇫🇮',
    'cs': '🇨🇿',
    'hu': '🇭🇺',
    'ro': '🇷🇴',
  }
  return flagMap[locale] || '🌐'
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
