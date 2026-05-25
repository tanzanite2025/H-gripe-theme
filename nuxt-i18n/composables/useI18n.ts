/**
 * Nuxt 3 i18n 组合式函数
 * 
 * 使用方法:
 * 1. 将此文件放到 Nuxt 项目的 composables/ 目录
 * 2. 在组件中使用: const { locale, getLanguages, setLanguage } = useI18n()
 */

export interface Language {
  code: string
  name: string
  native_name: string
  enabled: boolean
}

export interface PostTranslation {
  id: number
  title: string
  slug: string
  locale: string
  published_at: string
  url: string
}

export interface TranslationsResponse {
  post_id: number
  translations: Record<string, PostTranslation>
  count: number
}

export const useI18n = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase || 'http://localhost:9000'
  
  // 从 Cookie 获取当前语言
  const locale = useCookie('locale', { 
    default: () => 'en',
    maxAge: 365 * 24 * 60 * 60 // 1 年
  })

  /**
   * 获取支持的语言列表
   */
  const getLanguages = async (): Promise<Language[]> => {
    try {
      const { data } = await useFetch<{ languages: Language[]; total: number }>(
        `${apiBase}/api/v1/i18n/languages`
      )
      return data.value?.languages || []
    } catch (error) {
      console.error('Failed to fetch languages:', error)
      return []
    }
  }

  /**
   * 获取文章的所有翻译版本
   */
  const getPostTranslations = async (postId: number): Promise<Record<string, PostTranslation>> => {
    try {
      const { data } = await useFetch<TranslationsResponse>(
        `${apiBase}/api/v1/i18n/translations/${postId}`
      )
      return data.value?.translations || {}
    } catch (error) {
      console.error('Failed to fetch post translations:', error)
      return {}
    }
  }

  /**
   * 设置用户语言偏好
   */
  const setLanguage = async (newLocale: string): Promise<boolean> => {
    try {
      await $fetch(`${apiBase}/api/v1/i18n/set-language`, {
        method: 'POST',
        body: { locale: newLocale }
      })
      
      // 更新 Cookie
      locale.value = newLocale
      
      return true
    } catch (error) {
      console.error('Failed to set language:', error)
      return false
    }
  }

  /**
   * 检测用户语言偏好
   */
  const detectLanguage = async (): Promise<string> => {
    try {
      const { data } = await useFetch<{ detected_locale: string; source: string }>(
        `${apiBase}/api/v1/i18n/detect`
      )
      return data.value?.detected_locale || 'en'
    } catch (error) {
      console.error('Failed to detect language:', error)
      return 'en'
    }
  }

  /**
   * 获取语言的本地化名称
   */
  const getLanguageName = async (code: string): Promise<string> => {
    const languages = await getLanguages()
    const lang = languages.find(l => l.code === code)
    return lang?.native_name || code
  }

  /**
   * 构建本地化 URL
   */
  const localizeUrl = (path: string, targetLocale?: string): string => {
    const currentLocale = targetLocale || locale.value
    
    // 英文不需要前缀
    if (currentLocale === 'en') {
      return path
    }
    
    // 其他语言添加前缀
    return `/${currentLocale}${path}`
  }

  /**
   * 切换语言并刷新页面
   */
  const switchLanguage = async (newLocale: string) => {
    const success = await setLanguage(newLocale)
    if (success) {
      // 获取当前路径
      const route = useRoute()
      const currentPath = route.path
      
      // 移除当前语言前缀
      let cleanPath = currentPath
      const languages = await getLanguages()
      for (const lang of languages) {
        if (currentPath.startsWith(`/${lang.code}/`)) {
          cleanPath = currentPath.substring(lang.code.length + 1)
          break
        }
      }
      
      // 构建新的 URL
      const newPath = localizeUrl(cleanPath, newLocale)
      
      // 导航到新 URL
      await navigateTo(newPath)
      
      // 刷新页面以重新加载数据
      window.location.reload()
    }
  }

  return {
    locale,
    getLanguages,
    getPostTranslations,
    setLanguage,
    detectLanguage,
    getLanguageName,
    localizeUrl,
    switchLanguage
  }
}
