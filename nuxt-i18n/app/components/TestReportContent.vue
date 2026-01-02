<template>
  <div class="support-test-report-content">
    <!-- Tabs header -->
    <div class="nav-pill-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === tab.id }"
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
      <SupportRimTestReportSection :open-spoke-hole-video="openSpokeHoleVideo" />
    </section>

    <!-- Wheelset Test Report -->
    <section
      v-show="activeTab === 'wheelset-test-report'"
      id="wheelset-test-report"
      class="support-section"
    >
      <SupportWheelsetTestReportSection
        :open-wheelset-video="openWheelsetVideo"
        :go-to-wheelset-assembly="() => setActiveTab('wheelset-assembly')"
      />
    </section>

    <!-- Tension -->
    <section
      v-show="activeTab === 'tension'"
      id="tension"
      class="support-section"
    >
      <TechnicalTensionSection />
    </section>

    <!-- Wheelset Assembly -->
    <section
      v-show="activeTab === 'wheelset-assembly'"
      id="wheelset-assembly"
      class="support-section"
    >
      <SupportWheelsetAssemblySection />
    </section>

    <!-- Spoke-hole strength test video modal -->
    <div
      v-if="showSpokeHoleVideo"
      class="support-video-modal"
      role="dialog"
      aria-modal="true"
    >
      <div class="support-video-modal__backdrop" @click="showSpokeHoleVideo = false" />
      <div class="support-video-modal__content">
        <button
          type="button"
          class="support-video-modal__close"
          @click="showSpokeHoleVideo = false"
        >
          ×
        </button>
        <video
          class="support-video-modal__video"
          controls
          preload="metadata"
        >
          <source src="/testreport/rimtestreport/Spoke-hole-strength-test.webm" type="video/webm" />
        </video>
      </div>
    </div>

    <div
      v-if="showWheelsetVideo"
      class="support-video-modal"
      role="dialog"
      aria-modal="true"
    >
      <div class="support-video-modal__backdrop" @click="showWheelsetVideo = false" />
      <div class="support-video-modal__content">
        <button
          type="button"
          class="support-video-modal__close"
          @click="showWheelsetVideo = false"
        >
          ×
        </button>
        <video
          class="support-video-modal__video"
          controls
          preload="metadata"
        >
          <source
            src="/testreport/wheelsettestreport/tanzanite-wheelsettestroport.webm"
            type="video/webm"
          />
        </video>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import SupportRimTestReportSection from '~/components/SupportRimTestReportSection.vue'
import SupportWheelsetTestReportSection from '~/components/SupportWheelsetTestReportSection.vue'
import SupportWheelsetAssemblySection from '~/components/SupportWheelsetAssemblySection.vue'
import TechnicalTensionSection from '~/components/TechnicalTensionSection.vue'

const props = defineProps<{
  syncWithUrl?: boolean
}>()

type TestReportTabId = 'rim-test-report' | 'wheelset-test-report' | 'tension' | 'wheelset-assembly'

const tabs: { id: TestReportTabId; label: string }[] = [
  { id: 'rim-test-report', label: 'Rim Test Report' },
  { id: 'wheelset-test-report', label: 'Wheelset Test Report' },
  { id: 'tension', label: 'Tension' },
  { id: 'wheelset-assembly', label: 'Wheelset Assembly' },
]

const activeTab = ref<TestReportTabId>('rim-test-report')
const showSpokeHoleVideo = ref(false)
const showWheelsetVideo = ref(false)

const openSpokeHoleVideo = () => {
  showSpokeHoleVideo.value = true
}

const openWheelsetVideo = () => {
  showWheelsetVideo.value = true
}

const setActiveTab = (id: TestReportTabId) => {
  activeTab.value = id
  if (props.syncWithUrl && typeof window !== 'undefined') {
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
  if (props.syncWithUrl) {
    syncTabWithHash(route.hash)
  }
})

watch(
  () => route.hash,
  (newHash) => {
    if (props.syncWithUrl) {
      syncTabWithHash(newHash)
    }
  }
)
</script>

<style scoped>
.support-section {
  margin-top: 1rem;
  text-align: center;
}

/* Titles are now handled within components mostly, but keeping global styles just in case */
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

.support-video-modal {
  position: fixed;
  inset: 0;
  z-index: 10002; /* Higher than drawer z-index 10001 */
  display: flex;
  align-items: center;
  justify-content: center;
}

.support-video-modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.85);
}

.support-video-modal__content {
  position: relative;
  z-index: 41;
  width: 100%;
  max-width: 960px;
  margin: 0 1rem;
  background: #020617;
  border-radius: 0.75rem;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.85);
  overflow: hidden;
}

.support-video-modal__close {
  position: absolute;
  top: 0.35rem;
  right: 0.6rem;
  background: transparent;
  border: none;
  color: #e5e7eb;
  font-size: 1.4rem;
  cursor: pointer;
}

.support-video-modal__video {
  display: block;
  width: 100%;
  height: auto;
}
</style>
