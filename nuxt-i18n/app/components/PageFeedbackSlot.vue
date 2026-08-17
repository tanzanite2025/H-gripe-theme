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

const activeFeedbackTab = computed<PageSubNavigationTab | null>(() => {
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
