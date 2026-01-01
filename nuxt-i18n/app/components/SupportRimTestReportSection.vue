<template>
  <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-3 md:p-6">
    <h3 class="support-section__title text-center">Rim Test Report</h3>

    <p class="support-section__body mt-4 text-center">
      This section summarises how TANZANITE carbon rims are validated in our in-house lab,
      including strength, impact, radial load, spoke-hole tension, cutting-piece inspection,
      and X-ray flatness checks.
    </p>

    <!-- Rim test photo grid: 4 images per row, second row with 3 images -->
    <!-- Rim Test Carousel -->
    <div class="mt-8 mb-12">
      <StackedImageCarousel :items="cards" />
    </div>

    <!-- Test descriptions with global blue subtitles -->
    <div class="rim-test-card mt-6">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Strength Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Verify the rim's ability to withstand static and dynamic loads, ensuring no structural failure occurs during
        riding.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Fix the rim or complete wheelset in the testing fixture.</li>
        <li>
          Apply a vertical static load to the rim, typically in the range of 600-1000 N (approximately 60-100 kgf),
          and maintain for a specified duration (e.g., 1 minute).
        </li>
        <li>
          Conduct dynamic cyclic loading, for example applying alternating forces of 300-400 N over 100,000 cycles to
          simulate long-term riding conditions.
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Criteria for Evaluation:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The rim must not exhibit cracks, fractures, or permanent structural failure.</li>
        <li>
          Permissible deformation must remain within the specified limits (e.g., radial deviation ≤ 1.0 mm, lateral
          deviation ≤ 1.5 mm).
        </li>
        <li>After testing, the rim must retain full functional integrity.</li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Impact Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Simulates external impacts (e.g., rock strikes or landings) to assess rim resistance and structural integrity.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Secure the rim or complete wheelset in the testing fixture.</li>
        <li>Drop a standardized impact body (e.g., a 22.5 kg steel mass) from a height of 300 mm onto the rim.</li>
        <li>
          Position the impact point between spoke holes or at a representative weak section of the rim to replicate
          real-world riding impacts.
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The rim must not exhibit cracks, fractures, or severe deformation.</li>
        <li>Local dents or scratches must remain within specified limits (e.g., dent depth ≤ 2.0 mm).</li>
        <li>
          After impact, the rim must retain full functionality, ensuring safe tire mounting and riding.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Radial Load Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Tests rim deformation and load capacity under vertical forces to ensure safety and durability.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Secure the rim or complete wheelset in the testing fixture.</li>
        <li>
          Apply a vertical static load, typically 1000-1200 N (approximately 100-120 kgf), and maintain for a specified
          duration (e.g., 1 minute).
        </li>
        <li>
          Conduct dynamic cyclic loading, for example applying alternating forces of 600-800 N over 100,000 cycles to
          simulate long-term vertical stresses during riding.
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The rim must not exhibit cracks, fractures, or permanent structural failure.</li>
        <li>Radial deformation must remain within specified limits (e.g., radial deviation ≤ 1.0 mm).</li>
        <li>
          After testing, the rim must retain full functionality, ensuring safe tire mounting and riding performance.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Hole Tension Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Evaluates rim ability to withstand spoke-hole tension under static and dynamic loads, ensuring durability and
        safety of spoke interfaces.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>
          Secure the rim in the testing fixture and apply standardized bolts or tensile devices at spoke-hole
          positions.
        </li>
        <li>
          Apply a static tensile load to each spoke hole, typically 1000-1200 N (approximately 100-120 kgf), and
          maintain for a specified duration (e.g., 1 minute).
        </li>
        <li>
          Conduct dynamic cyclic tensile loading, for example applying alternating forces of 600-800 N over 100,000
          cycles to simulate long-term variations in spoke tension during riding.
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No cracks, fractures, or permanent structural failures should occur around spoke holes.</li>
        <li>Local deformation must remain within specified limits (e.g., hole diameter change ≤ 0.2 mm).</li>
        <li>
          After testing, the rim must retain full functionality, ensuring safe spoke installation and use.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Rim Cutting Pieces Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Analyze rim cross-sections by cutting a portion to evaluate both internal and external features, ensuring
        compliance with dimensional and quality standards.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>
          Cut representative sections of the rim at critical points (e.g., spoke hole area, joint section, braking
          surface).
        </li>
        <li>Examine the cross-sections under magnification or with precision measuring tools.</li>
        <li>
          Identify any burrs, sharp edges, irregularities, or deviations in wall thickness, material distribution, or
          dimensional specifications.
        </li>
        <li>Compare measurements against design tolerances (e.g., wall thickness ±0.1 mm, hole diameter ±0.05 mm).</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No burrs or sharp edges that could compromise spoke or tire safety.</li>
        <li>Dimensional deviations must remain within specified tolerances.</li>
        <li>Internal and external surfaces must be smooth and consistent, with no voids, cracks, or inclusions.</li>
        <li>
          Rim sections must demonstrate uniformity and compliance with quality standards before approval for mass
          production.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        X-ray Flatness Inspection
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Assess the internal flatness and structural uniformity of the rim using high-resolution X-ray imaging. This
        non-destructive test ensures that hidden defects do not compromise performance or safety.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Place the rim in a high-resolution X-ray inspection system.</li>
        <li>Capture cross-sectional and full-profile images of the rim to evaluate internal geometry.</li>
        <li>
          Analyze images for inconsistencies such as warping, voids, inclusions, or uneven material distribution.
        </li>
        <li>
          Compare results against design specifications (e.g., flatness tolerance ≤ 0.2 mm, uniform wall thickness
          within ± 0.1 mm).
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No internal voids, cracks, or inclusions that could weaken the rim structure.</li>
        <li>Flatness deviations must remain within specified tolerances.</li>
        <li>Material distribution must be uniform, ensuring consistent strength and durability.</li>
        <li>
          Rim must pass inspection without requiring destructive sampling, confirming production quality.
        </li>
      </ul>
    </div>

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
import { ref } from 'vue'

const { openSpokeHoleVideo } = defineProps<{
  openSpokeHoleVideo: () => void | Promise<void>
}>()

// --- Stacked Carousel Logic ---

const cards = [
  {
    src: '/testreport/rimtestreport/tanzanite-rim-testreport.webp',
    alt: 'Overview of Tanzanite rim laboratory test setup',
    caption: "Overall view of the rim test setup used in TANZANITE's lab."
  },
  {
    src: '/testreport/rimtestreport/tanzanite-carbonrim-strength-test1.webp',
    alt: 'Tanzanite carbon rim strength test under static load',
    caption: "Strength Test – evaluating rim performance under static and dynamic loads."
  },
  {
    src: '/testreport/rimtestreport/tanzanite-carbonrim-strength-test1 (2).webp',
    alt: 'Additional view of Tanzanite carbon rim strength testing',
    caption: "Strength Test – additional view of clamping and load application."
  },
  {
    src: '/testreport/rimtestreport/tanzanite-carbonrim-radialload-test.webp',
    alt: 'Tanzanite carbon rim radial load test showing vertical loading',
    caption: "Radial Load Test – checking deformation and load capacity under vertical forces."
  },
  {
    src: '/testreport/rimtestreport/tanzanite-hole-tension-test.webp',
    alt: 'Spoke hole tension test setup for a Tanzanite carbon rim',
    caption: "Hole Tension Test – measuring how spoke holes withstand spoke tension and repeated loading."
  },
  {
    src: '/testreport/rimtestreport/tanznaite-rim-cutting-pieces-Test.webp',
    alt: 'Cross-section cutting piece of a Tanzanite carbon rim',
    caption: "Rim Cutting Pieces Test – inspecting internal layup, wall thickness, and burrs."
  },
  {
    src: '/testreport/rimtestreport/tanzanite-xray-flatness-Inspection-test.webp',
    alt: 'X-ray flatness inspection of a Tanzanite carbon rim',
    caption: "X-ray Flatness Inspection – non-destructive check for internal uniformity and flatness."
  }
]


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

.rim-test-card {
  padding: 1.25rem;
  border-radius: 0.75rem;
  background: rgba(2, 6, 23, 0.4); /* Darker, more subtle for inner nested card */
  border: 1px solid rgba(255, 255, 255, 0.03); /* Very subtle border for definition */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
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
