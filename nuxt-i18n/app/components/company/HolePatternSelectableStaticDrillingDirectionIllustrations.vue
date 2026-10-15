<template>
  <section class="hole-pattern-static-illustrations" aria-labelledby="hole-pattern-static-illustrations-title">
    <div class="hole-pattern-illustrations-heading">
      <div>
        <h2 id="hole-pattern-static-illustrations-title">{{ t('companyAboutHolePatterns.staticDrillingIllustrations.title') }}</h2>
        <p>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.introduction') }}</p>
      </div>
      <NuxtLink class="hole-pattern-topology-simulator-link" :to="localePath('/tools/spoke-lacing-topology-simulator')">
        <span>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.openSimulator') }}</span>
        <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
          <path d="M5 12h14m-6-6 6 6-6 6" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </NuxtLink>
    </div>

    <div class="hole-pattern-illustration-selection" role="group" :aria-label="t('companyAboutHolePatterns.staticDrillingIllustrations.selectionLabel')">
      <button
        v-for="drillingIllustrationExample in holePatternStaticDrillingIllustrationExamples"
        :key="drillingIllustrationExample.key"
        type="button"
        :aria-pressed="selectedHolePatternDrillingIllustrationKey === drillingIllustrationExample.key"
        aria-controls="hole-pattern-selected-static-illustration"
        @click="selectedHolePatternDrillingIllustrationKey = drillingIllustrationExample.key"
      >
        {{ t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${drillingIllustrationExample.key}.label`) }}
      </button>
    </div>

    <p class="hole-pattern-additional-supported-lacing-patterns">
      {{ t('companyAboutHolePatterns.staticDrillingIllustrations.additionalSupportedPatterns', { simulatorButtonLabel: t('companyAboutHolePatterns.staticDrillingIllustrations.openSimulator') }) }}
    </p>

    <div id="hole-pattern-selected-static-illustration" class="hole-pattern-illustration-grid">
      <figure class="hole-pattern-illustration-card hole-pattern-wheel-lacing-figure">
        <figcaption>
          <h3>{{ t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${selectedHolePatternDrillingIllustration.key}.label`) }}</h3>
          <p>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.wheelViewLabel') }}</p>
          <p>{{ t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${selectedHolePatternDrillingIllustration.key}.crossingGuide`) }}</p>
        </figcaption>
        <div class="hole-pattern-illustration-image-frame">
          <img
            :src="selectedHolePatternDrillingIllustration.wheelImagePath"
            :alt="t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${selectedHolePatternDrillingIllustration.key}.wheelImageAlt`)"
            width="560"
            height="560"
            loading="lazy"
            decoding="async"
          >
        </div>
      </figure>

      <figure class="hole-pattern-illustration-card">
        <figcaption>
          <h3>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.longitudinalViewTitle') }}</h3>
          <p>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.longitudinalViewDescription') }}</p>
        </figcaption>
        <div class="hole-pattern-illustration-image-frame">
          <img
            :src="selectedHolePatternDrillingIllustration.longitudinalImagePath"
            :alt="t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${selectedHolePatternDrillingIllustration.key}.longitudinalImageAlt`)"
            width="420"
            height="200"
            loading="lazy"
            decoding="async"
          >
        </div>
      </figure>

      <figure class="hole-pattern-illustration-card">
        <figcaption>
          <h3>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.axialViewTitle') }}</h3>
          <p>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.axialViewDescription') }}</p>
        </figcaption>
        <div class="hole-pattern-illustration-image-frame">
          <img
            :src="holePatternAxialDrillingIllustrationImagePath"
            :alt="t('companyAboutHolePatterns.staticDrillingIllustrations.axialImageAlt')"
            width="200"
            height="200"
            loading="lazy"
            decoding="async"
          >
        </div>
      </figure>
    </div>

    <ul class="hole-pattern-illustration-legend">
      <li><span class="hole-pattern-spoke-direction-legend-line" aria-hidden="true"></span>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.spokeDirectionLegend') }}</li>
      <li><span class="hole-pattern-drilling-axis-legend-line" aria-hidden="true"></span>{{ t('companyAboutHolePatterns.staticDrillingIllustrations.drillingAxisLegend') }}</li>
    </ul>
    <p class="hole-pattern-selected-lacing-description" aria-live="polite" aria-atomic="true">{{ t(`companyAboutHolePatterns.staticDrillingIllustrations.examples.${selectedHolePatternDrillingIllustration.key}.description`) }}</p>
    <p class="hole-pattern-static-illustration-drawing-limit">{{ t('companyAboutHolePatterns.staticDrillingIllustrations.drawingLimit') }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n, useLocalePath } from '#imports'

const { t } = useI18n()
const localePath = useLocalePath()
const holePatternAxialDrillingIllustrationImagePath = '/company/ourstory/holepatterns/left-right-offset-rim-holes-and-hub-flange-axial-drilling-projection.svg'

const holePatternStaticDrillingIllustrationExamples = [
  {
    key: 'radial',
    wheelImagePath: '/company/ourstory/holepatterns/32-hole-radial-single-flange-lacing-with-counted-crossings.svg',
    longitudinalImagePath: '/company/ourstory/holepatterns/radial-lacing-longitudinal-rim-hole-drilling-detail.svg',
  },
  {
    key: 'twoCross',
    wheelImagePath: '/company/ourstory/holepatterns/32-hole-two-cross-single-flange-lacing-with-counted-crossings.svg',
    longitudinalImagePath: '/company/ourstory/holepatterns/two-cross-lacing-longitudinal-rim-hole-drilling-detail.svg',
  },
  {
    key: 'threeCross',
    wheelImagePath: '/company/ourstory/holepatterns/32-hole-three-cross-single-flange-lacing-with-counted-crossings.svg',
    longitudinalImagePath: '/company/ourstory/holepatterns/three-cross-lacing-longitudinal-rim-hole-drilling-detail.svg',
  },
] as const

const selectedHolePatternDrillingIllustrationKey = ref<typeof holePatternStaticDrillingIllustrationExamples[number]['key']>('twoCross')
const selectedHolePatternDrillingIllustration = computed(() => (
  holePatternStaticDrillingIllustrationExamples.find((drillingIllustrationExample) => (
    drillingIllustrationExample.key === selectedHolePatternDrillingIllustrationKey.value
  )) ?? holePatternStaticDrillingIllustrationExamples[1]
))
</script>

<style scoped>
.hole-pattern-static-illustrations {
  margin-bottom: 2.5rem;
  padding: clamp(1rem, 2.5vw, 1.75rem);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
}

.hole-pattern-illustrations-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem 1.5rem;
}

.hole-pattern-illustrations-heading h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.25rem, 1.8vw, 1.65rem);
  font-weight: 600;
  line-height: 1.45;
}

.hole-pattern-illustrations-heading p {
  max-width: 38rem;
  margin: 0.5rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.7;
}

.hole-pattern-topology-simulator-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  min-height: var(--tz-pill-button-min-height);
  padding: var(--tz-pill-button-padding-block) 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 600;
  line-height: var(--tz-pill-button-line-height);
  text-decoration: none;
}

.hole-pattern-topology-simulator-link svg {
  flex-shrink: 0;
  width: 1.1rem;
  height: 1.1rem;
}

.hole-pattern-topology-simulator-link:hover {
  border-color: var(--tz-site-accent);
  color: var(--tz-text-accent);
}

.hole-pattern-illustration-selection {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 1.5rem 0;
}

.hole-pattern-illustration-selection button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: var(--tz-pill-button-min-height);
  padding: var(--tz-pill-button-padding-block) 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  color: var(--tz-text-secondary);
  background: transparent;
  font-size: 0.875rem;
  line-height: var(--tz-pill-button-line-height);
  font-weight: 600;
  cursor: pointer;
}

.hole-pattern-illustration-selection button[aria-pressed='true'] {
  border-color: var(--tz-site-accent-hover);
  color: var(--tz-action-primary-foreground);
  background: var(--tz-site-accent-hover);
}

.hole-pattern-illustration-selection button:focus-visible,
.hole-pattern-topology-simulator-link:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 3px;
}

.hole-pattern-additional-supported-lacing-patterns {
  margin: -0.5rem 0 1.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.75;
}

.hole-pattern-illustration-grid {
  display: grid;
  gap: 1rem;
}

.hole-pattern-illustration-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
}

.hole-pattern-illustration-card figcaption {
  flex: 1;
  padding: 1rem;
}

.hole-pattern-illustration-grid h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 600;
  line-height: 1.5;
}

.hole-pattern-illustration-grid figcaption p {
  margin: 0.4rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.7;
}

.hole-pattern-illustration-image-frame {
  position: relative;
  flex-shrink: 0;
  aspect-ratio: 4 / 3;
  border-top: 1px solid var(--tz-border-subtle);
  background: #f8fafc;
}

.hole-pattern-illustration-image-frame img {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  padding: 0.75rem;
  object-fit: contain;
}

.hole-pattern-illustration-legend {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 0.5rem 1.25rem;
  padding: 0;
  margin: 1rem 0 0;
  list-style: none;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.hole-pattern-illustration-legend li {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.hole-pattern-illustration-legend span {
  width: 1.3rem;
  border-top: 2px solid var(--tz-site-accent);
}

.hole-pattern-illustration-legend .hole-pattern-drilling-axis-legend-line {
  border-top: 2px dashed var(--tz-site-accent);
}

.hole-pattern-selected-lacing-description {
  max-width: 70rem;
  margin: 1rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.7;
}

.hole-pattern-static-illustration-drawing-limit {
  margin: 1.5rem 0 0;
  padding-top: 1rem;
  border-top: 1px solid var(--tz-border-subtle);
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.7;
}

@media (min-width: 640px) and (max-width: 959px) {
  .hole-pattern-illustration-card {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-items: center;
  }

  .hole-pattern-illustration-image-frame {
    border-top: 0;
    border-inline-start: 1px solid var(--tz-border-subtle);
  }
}

@media (min-width: 960px) {
  .hole-pattern-illustration-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
