<template>
  <section v-if="feedbackThreadKey" class="page-feedback-slot">
    <UserFeedbackThread
      :key="feedbackThreadKey"
      :thread-key="feedbackThreadKey"
      :title="feedbackTitle"
      :subtitle="feedbackSubtitle"
      :page-path="feedbackPagePath"
      :page-title="feedbackPageTitle"
    />
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import { normalizeFaqRoutePath } from '~/data/faq/routing'
import {
  getPageSubNavigationTabFromPath,
  pageSubNavigationEntries,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'

const route = useRoute()
const { t } = useI18n()

const routeFeedbackTab = computed<PageSubNavigationTab | null>(() => {
  const routeMeta = route.meta as Record<string, unknown>
  const threadKey = typeof routeMeta.feedbackThreadKey === 'string'
    ? routeMeta.feedbackThreadKey.trim()
    : ''

  if (!threadKey) return null

  return {
    id: 'route-feedback',
    feedbackThreadKey: threadKey,
    feedbackTitleKey: typeof routeMeta.feedbackTitleKey === 'string'
      ? routeMeta.feedbackTitleKey
      : undefined,
    feedbackTitle: typeof routeMeta.feedbackTitle === 'string'
      ? routeMeta.feedbackTitle
      : undefined,
    feedbackSubtitleKey: typeof routeMeta.feedbackSubtitleKey === 'string'
      ? routeMeta.feedbackSubtitleKey
      : undefined,
    feedbackSubtitle: typeof routeMeta.feedbackSubtitle === 'string'
      ? routeMeta.feedbackSubtitle
      : undefined,
    pageTitleKey: typeof routeMeta.pageTitleKey === 'string'
      ? routeMeta.pageTitleKey
      : undefined,
    pageTitle: typeof routeMeta.pageTitle === 'string'
      ? routeMeta.pageTitle
      : undefined,
  }
})

const activeFeedbackTab = computed<PageSubNavigationTab | null>(() => {
  if (routeFeedbackTab.value) return routeFeedbackTab.value

  const normalizedPath = normalizeFaqRoutePath(route.path)

  for (const entry of pageSubNavigationEntries) {
    const tabId = getPageSubNavigationTabFromPath(
      entry.tabs,
      entry.path,
      route.path,
      { match: 'nested' }
    )
    const activeTab: PageSubNavigationTab | undefined = tabId
      ? entry.tabs.find(tab => tab.id === tabId)
      : normalizeFaqRoutePath(entry.path) === normalizedPath
        ? entry.tabs[0]
        : undefined

    if (activeTab?.feedbackThreadKey) {
      return activeTab
    }
  }

  return null
})

const translateTabText = (key: string | undefined, fallback: string) => {
  return key ? t(key, fallback) : fallback
}

const feedbackThreadKey = computed(() => activeFeedbackTab.value?.feedbackThreadKey || '')

const feedbackTitle = computed(() => {
  const tab = activeFeedbackTab.value
  if (!tab) return ''

  return translateTabText(
    tab.feedbackTitleKey,
    tab.feedbackTitle || tab.pageTitle || tab.fallback || 'Share your feedback'
  )
})

const feedbackSubtitle = computed(() => {
  const tab = activeFeedbackTab.value
  if (!tab) return ''

  return translateTabText(
    tab.feedbackSubtitleKey,
    tab.feedbackSubtitle || tab.pageIntro || tab.description || ''
  )
})

const feedbackPagePath = computed(() => normalizeFaqRoutePath(route.path))

const feedbackPageTitle = computed(() => {
  const tab = activeFeedbackTab.value
  if (!tab) return ''

  return translateTabText(
    tab.pageTitleKey,
    tab.pageTitle || tab.fallback || tab.label || feedbackPagePath.value
  )
})
</script>

<style scoped>
.page-feedback-slot {
  width: 100%;
}
</style>
