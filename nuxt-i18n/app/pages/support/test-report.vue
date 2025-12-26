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

      <p class="support-section__body mt-4 text-center">
        This section summarises how TANZANITE carbon rims are validated in our in-house lab,
        including strength, impact, radial load, spoke-hole tension, cutting-piece inspection,
        and X-ray flatness checks.
      </p>

      <!-- Rim test photo grid: 4 images per row, second row with 3 images -->
      <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-rim-testreport.webp"
          alt="Overview of Tanzanite rim laboratory test setup"
          :zoomOnClick="true"
          caption="Overall view of the rim test setup used in TANZANITE's lab."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-carbonrim-strength-test1.webp"
          alt="Tanzanite carbon rim strength test under static load"
          :zoomOnClick="true"
          caption="Strength Test – evaluating rim performance under static and dynamic loads."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-carbonrim-strength-test1 (2).webp"
          alt="Additional view of Tanzanite carbon rim strength testing"
          :zoomOnClick="true"
          caption="Strength Test – additional view of clamping and load application."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-carbonrim-radialload-test.webp"
          alt="Tanzanite carbon rim radial load test showing vertical loading"
          :zoomOnClick="true"
          caption="Radial Load Test – checking deformation and load capacity under vertical forces."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-hole-tension-test.webp"
          alt="Spoke hole tension test setup for a Tanzanite carbon rim"
          :zoomOnClick="true"
          caption="Hole Tension Test – measuring how spoke holes withstand spoke tension and repeated loading."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanznaite-rim-cutting-pieces-Test.webp"
          alt="Cross-section cutting piece of a Tanzanite carbon rim"
          :zoomOnClick="true"
          caption="Rim Cutting Pieces Test – inspecting internal layup, wall thickness, and burrs."
        />
        <GuideImage
          src="/testreport/rimtestreport/tanzanite-xray-flatness-Inspection-test.webp"
          alt="X-ray flatness inspection of a Tanzanite carbon rim"
          :zoomOnClick="true"
          caption="X-ray Flatness Inspection – non-destructive check for internal uniformity and flatness."
        />
      </div>

      <!-- Test descriptions with global blue subtitles -->
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-6">
        Strength Test
      </h4>
      <p class="support-section__body">
        Evaluates rim ability to withstand static or dynamic loads.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Impact Test
      </h4>
      <p class="support-section__body">
        Simulates external impacts (for example rock strikes or landings) to test rim resistance.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Radial Load Test
      </h4>
      <p class="support-section__body">
        Tests rim deformation and load capacity under vertical forces.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Hole Tension Test
      </h4>
      <p class="support-section__body">
        Evaluates rim ability to withstand spoke-hole tension under static and dynamic loads.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        Rim Cutting Pieces Test
      </h4>
      <p class="support-section__body">
        The rim cutting pieces test involves analyzing cross-sections of the rim by cutting a portion to evaluate
        internal and external features. This inspection focuses on identifying any burrs, irregularities, or deviations
        in dimensional specifications to ensure the rim meets our stringent quality standards.
      </p>

      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
        X-ray Flatness Inspection
      </h4>
      <p class="support-section__body">
        Uses high-resolution X-ray imaging to assess the internal flatness and structural uniformity of the rim. This
        non-destructive test detects internal inconsistencies, warping, or voids that may compromise performance or
        safety.
      </p>

      <!-- Spoke-hole strength test video thumbnail (positioned above disclaimer) -->
      <div class="mt-4 flex justify-center">
        <div
          class="support-video-thumbnail"
          @click="showSpokeHoleVideo = true"
        >
          <img
            class="support-video-thumbnail__image"
            src="/testreport/rimtestreport/tanzanite-rim-testreport.webp"
            alt="Play spoke-hole strength test video for Tanzanite carbon rims"
            loading="lazy"
          />
          <div class="support-video-thumbnail__overlay">
            <span class="support-video-thumbnail__icon">▶</span>
            <span class="support-video-thumbnail__label">Watch spoke-hole strength test video</span>
          </div>
        </div>
      </div>

      <!-- Disclaimer block styled like Technical Tension amber card -->
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'
import GuideImage from '~/components/GuideImage.vue'

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
const showSpokeHoleVideo = ref(false)

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
