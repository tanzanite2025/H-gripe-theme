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
      <SupportRimTestReportSection :open-spoke-hole-video="openSpokeHoleVideo" />
    </section>

    <!-- Wheelset Test Report -->
    <section
      v-show="activeTab === 'wheelset-test-report'"
      id="wheelset-test-report"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Wheelset Test Report</h3>
      <p class="support-section__body mt-4 text-center">
        This section summarises how TANZANITE wheelsets are validated in our in-house lab,
        covering lateral load, torsional stiffness, environmental durability, dynamic balance,
        fatigue, and braking performance tests.
      </p>
      <div class="mt-3 flex justify-center">
        <div
          class="inline-flex items-center rounded-full border border-sky-500/70 bg-sky-500/5 px-3 py-1 text-xs sm:text-sm text-sky-300"
        >
          <span>To learn more about wheelset shipping and assembly tests,</span>
          <button
            type="button"
            class="ml-1 underline decoration-sky-400/80 underline-offset-2 hover:text-sky-200 hover:decoration-sky-200 transition-colors duration-150"
            @click="setActiveTab('wheelset-assembly')"
          >
            see Wheelset Assembly.
          </button>
        </div>
      </div>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-6">
        Lateral Load Test
      </h4>
      <p class="support-section__body">
        Tests wheelset stability and stiffness under lateral forces.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Torsional Stiffness Test
      </h4>
      <p class="support-section__body">
        Evaluates stiffness and response when transmitting torque.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Environmental Durability Test
      </h4>
      <p class="support-section__body">
        Simulates humidity, temperature, salt spray, and other environmental conditions.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Dynamic Balance Test
      </h4>
      <p class="support-section__body">
        Tests wheelset balance and stability during high-speed rotation.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Fatigue Test
      </h4>
      <p class="support-section__body">
        Simulates repeated loads over long-term riding to verify rim durability.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Braking Performance Test
      </h4>
      <p class="support-section__body">
        For V-brake and disc brake wheelsets, tests stability and heat resistance during braking.
      </p>

      <div class="mt-4 flex justify-center">
        <div
          class="support-video-thumbnail"
          @click="openWheelsetVideo"
        >
          <img
            class="support-video-thumbnail__image"
            src="/testreport/wheelsettestreport/tanzanite-wheelssettestroport-video-firstpicture.webp"
            alt="Play wheelset test report video for Tanzanite wheelsets"
            loading="lazy"
          />
          <div class="support-video-thumbnail__overlay">
            <span class="support-video-thumbnail__icon">▶</span>
            <span class="support-video-thumbnail__label">Watch wheelset test report video</span>
          </div>
        </div>
      </div>

      <div
        class="mt-6 rounded-lg border border-amber-300/80 bg-slate-900/80 px-4 py-3 text-sm leading-relaxed text-amber-100"
      >
        <h4 class="mb-1 font-semibold text-amber-300">
          Disclaimer
        </h4>
        <p class="mb-2">
          We pick only one sample for every test report and the results will likely vary as the rim diameters change.
          Please note that the differences between models in test results are specially designed by our engineers for
          the intended uses.
        </p>
        <p>
          All the test results of this section are based on the lab criteria of TANZANITE and are implemented at our
          well-established testing facilities. TANZANITE is only responsible for the test results themselves which are
          not set for any comparison to other brands or such regards.
        </p>
      </div>
    </section>

    <!-- Wheelset Assembly -->
    <section
      v-show="activeTab === 'wheelset-assembly'"
      id="wheelset-assembly"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Wheelset Assembly</h3>
      <p class="support-section__body text-center">
        Wheelset Assembly content placeholder. Detailed assembly guides and checklists will be added here.
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

    <!-- Spoke-hole strength test video modal (same interaction pattern as /company/about#appearance) -->
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
import PageFaq from '~/components/PageFaq.vue'
import SupportRimTestReportSection from '~/components/SupportRimTestReportSection.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Test report',
})

type TestReportTabId = 'rim-test-report' | 'wheelset-test-report' | 'wheelset-assembly'

const tabs: { id: TestReportTabId; label: string }[] = [
  { id: 'rim-test-report', label: 'Rim Test Report' },
  { id: 'wheelset-test-report', label: 'Wheelset Test Report' },
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
  text-align: center;
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

.support-video-thumbnail {
  margin-top: 0.75rem;
  position: relative;
  border-radius: 0.75rem;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.85);
}

.support-video-thumbnail__image {
  display: block;
  width: 100%;
  height: auto;
}

.support-video-thumbnail__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.85),
    rgba(15, 23, 42, 0.45)
  );
  color: #e5e7eb;
}

.support-video-thumbnail__icon {
  font-size: 1.8rem;
}

.support-video-thumbnail__label {
  font-size: 0.9rem;
  font-weight: 500;
}

.support-video-modal {
  position: fixed;
  inset: 0;
  z-index: 40;
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
