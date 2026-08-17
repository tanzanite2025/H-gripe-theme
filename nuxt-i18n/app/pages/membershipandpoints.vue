<template>
  <div class="membership-page">
    <section class="page-content-shell px-0 pb-3 md:px-4">
      <h1 class="company-page__title">{{ pageTitle }}</h1>
      <p class="company-page__intro">{{ pageIntro }}</p>
    </section>

    <MembershipAndPointsTabs />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { definePageMeta, useHead, useI18n, useRoute } from '#imports'
import MembershipAndPointsTabs from '~/components/MembershipAndPointsTabs.vue'
import {
  getPageSubNavigationTabFromPath,
  membershipAndPointsTabs,
  type MembershipTabId,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const { t } = useI18n()

const defaultTabId: MembershipTabId = 'myinfo'

const activeTabId = computed<MembershipTabId>(() => {
  return getPageSubNavigationTabFromPath(
    membershipAndPointsTabs,
    '/membershipandpoints',
    route.path,
    { match: 'nested' }
  ) || defaultTabId
})

const activeTab = computed<PageSubNavigationTab>(() => {
  return membershipAndPointsTabs.find(tab => tab.id === activeTabId.value) || membershipAndPointsTabs[0]
})

const translateTabText = (key: string | undefined, fallback: string) => {
  return key ? t(key, fallback) : fallback
}

const pageTitle = computed(() => translateTabText(
  activeTab.value.pageTitleKey,
  activeTab.value.pageTitle || activeTab.value.fallback || 'Membership and Points'
))

const pageIntro = computed(() => translateTabText(
  activeTab.value.pageIntroKey,
  activeTab.value.pageIntro || activeTab.value.description || ''
))

const seoTitle = computed(() => translateTabText(
  activeTab.value.seoTitleKey,
  activeTab.value.seoTitle || pageTitle.value
))

const seoDescription = computed(() => translateTabText(
  activeTab.value.seoDescriptionKey,
  activeTab.value.seoDescription || pageIntro.value
))

useHead(() => ({
  title: seoTitle.value,
  meta: [
    { name: 'description', content: seoDescription.value },
    { property: 'og:title', content: seoTitle.value },
    { property: 'og:description', content: seoDescription.value },
  ],
}))
</script>

<style scoped>
.membership-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}
</style>
