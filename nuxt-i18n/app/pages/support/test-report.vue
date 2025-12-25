<template>
  <div class="support-test-report">
    <h1 class="support-page__title support-page__title--sr-only">Test report</h1>

    <!-- Tabs header -->
    <div class="support-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="support-tabs__item"
        :class="{ 'support-tabs__item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Rim Test Report -->
    <section
      v-show="activeTab === 'rim-test-report'"
      id="rim-test-report"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Rim Test Report</h3>
      <p class="support-section__body text-center">
        Rim Test Report content placeholder. Detailed rim test data and downloadable reports
        will be added here.
      </p>
    </section>

    <!-- Wheelset Test Report -->
    <section
      v-show="activeTab === 'wheelset-test-report'"
      id="wheelset-test-report"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Wheelset Test Report</h3>
      <p class="support-section__body text-center">
        Wheelset Test Report content placeholder. Detailed wheelset test data and
        downloadable reports will be added here.
      </p>
    </section>

    <!-- FAQ Section - 放在所有 tab 内容之后 -->
    <section class="support-section">
      <PageFaq
        page-id="support-test-report"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Test report',
})

type TestReportTabId = 'rim-test-report' | 'wheelset-test-report'

const tabs: { id: TestReportTabId; label: string }[] = [
  { id: 'rim-test-report', label: 'Rim Test Report' },
  { id: 'wheelset-test-report', label: 'Wheelset Test Report' },
]

const activeTab = ref<TestReportTabId>('rim-test-report')

const setActiveTab = (id: TestReportTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
}

const route = useRoute()

const syncTabWithHash = (hash: string | null | undefined) => {
  if (!hash) return
  const clean = hash.startsWith('#') ? hash.slice(1) : hash
  if (tabs.some((tab) => tab.id === clean)) {
    activeTab.value = clean as TestReportTabId
  }
}

onMounted(() => {
  syncTabWithHash(route.hash)
})

watch(
  () => route.hash,
  (newHash) => {
    syncTabWithHash(newHash)
  }
)
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.support-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-test-report {
  margin-top: -1rem;
}

.support-section {
  margin-top: 1rem;
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}

/* Tabs – 结构参考 /guides/wheelset-buyers */
.support-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 0;
  margin: 0 0 1rem;
  justify-content: center;
}

.support-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
}

.support-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}
</style>
