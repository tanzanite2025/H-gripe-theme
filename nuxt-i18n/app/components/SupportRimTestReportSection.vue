<template>
  <div>
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
        @click="openSpokeHoleVideo"
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
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

const { openSpokeHoleVideo } = defineProps<{
  openSpokeHoleVideo: () => void | Promise<void>
}>()
</script>

<style scoped>
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
</style>
