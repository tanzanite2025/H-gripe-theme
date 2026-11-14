import { computed, onServerPrefetch, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useI18n, useNuxtApp, useState } from '#imports'
import type { NuxtApp } from '#app'
import { usePageMessages } from '~/composables/usePageMessages'
import englishTestReportDialogMessages from '~/i18n/page-messages/testReportDialog/en.json'
import simplifiedChineseTestReportDialogMessages from '~/i18n/page-messages/testReportDialog/zh_cn.json'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'

// 每个 Nuxt 应用（SSR 时每个请求）独立缓存，多个测试实例共用同一次翻译加载。
const testReportMessageLoadersByNuxtApplication = new WeakMap<NuxtApp, Map<string, ReturnType<typeof usePageMessages>>>()

export function useIndependentTestReportMessages(messageNamespace: MaybeRefOrGetter<string>) {
  const nuxtApp = useNuxtApp()
  const { locale } = useI18n()
  const hydratedTestReportMessages = useState<Record<string, Record<string, unknown>>>(
    'independent-test-report-messages', () => ({}),
  )
  let namespaceLoaders = testReportMessageLoadersByNuxtApplication.get(nuxtApp)
  if (!namespaceLoaders) {
    namespaceLoaders = new Map()
    testReportMessageLoadersByNuxtApplication.set(nuxtApp, namespaceLoaders)
  }
  const applicationNamespaceLoaders = namespaceLoaders

  function getCurrentTestReportMessageLoader() {
    const namespace = toValue(messageNamespace)
    const loader = applicationNamespaceLoaders.get(namespace)
    if (loader) return loader
    const newLoader = nuxtApp.vueApp.runWithContext(() => usePageMessages(namespace))
    applicationNamespaceLoaders.set(namespace, newLoader)
    return newLoader
  }

  const currentTestReportMessageLoader = computed(getCurrentTestReportMessageLoader)

  async function loadCurrentTestReportMessages(): Promise<void> {
    const requestedLocale = locale.value
    const namespace = toValue(messageNamespace)
    const messageCacheKey = `${namespace}:${requestedLocale}`
    const hydratedMessages = hydratedTestReportMessages.value[messageCacheKey]
    if (hydratedMessages) {
      nuxtApp.$i18n.mergeLocaleMessage(requestedLocale, { [namespace]: hydratedMessages })
      return
    }
    // 弹窗外壳的短文案随组件加载，内容或翻译请求失败时仍可关闭和重试。
    if (namespace === 'testReportDialog') {
      const interfaceMessages = isSimplifiedChineseStorefrontLocale(requestedLocale)
        ? simplifiedChineseTestReportDialogMessages
        : englishTestReportDialogMessages
      nuxtApp.$i18n.mergeLocaleMessage(requestedLocale, { [namespace]: interfaceMessages })
      hydratedTestReportMessages.value[messageCacheKey] = interfaceMessages
      return
    }
    await currentTestReportMessageLoader.value.loadPageMessages(requestedLocale)
    const localeMessages = nuxtApp.$i18n.getLocaleMessage(requestedLocale) as Record<string, unknown>
    hydratedTestReportMessages.value[messageCacheKey] = localeMessages[namespace] as Record<string, unknown>
  }

  async function retryLoadingTestReportMessages(): Promise<void> {
    try {
      await loadCurrentTestReportMessages()
    } catch {
      // 错误保留在共享加载器中，由当前组件显示重试状态。
    }
  }

  // 保持组件 setup 同步，支持在没有 Suspense 的商品弹层中动态挂载。
  // SSR 在渲染前等待翻译；客户端先显示通用加载文案。
  onServerPrefetch(loadCurrentTestReportMessages)
  if (import.meta.client) {
    watch([locale, () => toValue(messageNamespace)], () => {
      void loadCurrentTestReportMessages().catch(error => {
        console.error('[test-report] Failed to load independent component messages', error)
      })
    }, { immediate: true })
  }

  return {
    areTestReportMessagesReady: computed(() => Boolean(
      hydratedTestReportMessages.value[`${toValue(messageNamespace)}:${locale.value}`],
    )),
    testReportMessageLoadingError: computed(() => currentTestReportMessageLoader.value.error.value),
    retryLoadingTestReportMessages,
  }
}
