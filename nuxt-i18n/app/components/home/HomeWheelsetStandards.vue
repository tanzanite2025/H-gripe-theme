<template>
  <section
    class="wheelset-standards"
    aria-labelledby="wheelset-standards-title"
  >
    <div class="wheelset-standards__head">
      <div>
        <p class="wheelset-standards__eyebrow">{{ t('home.wheelsetStandards.eyebrow') }}</p>
        <h3 id="wheelset-standards-title" class="wheelset-standards__title">
          {{ t('home.wheelsetStandards.title') }}
        </h3>
      </div>

      <div class="wheelset-standards__controls">
        <button
          type="button"
          class="wheelset-standards__arrow"
          :aria-label="t('home.wheelsetStandards.previous')"
          @click="scrollSteps(-1)"
        >
          <Icon name="lucide:arrow-left" class="h-4 w-4" aria-hidden="true" />
        </button>
        <button
          type="button"
          class="wheelset-standards__arrow"
          :aria-label="t('home.wheelsetStandards.next')"
          @click="scrollSteps(1)"
        >
          <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
        </button>
      </div>
    </div>

    <div
      ref="rail"
      class="wheelset-standards__rail"
      role="region"
      :aria-label="t('home.wheelsetStandards.ariaLabel')"
      @scroll.passive="updateActiveStep"
    >
      <article
        v-for="(step, index) in steps"
        :key="step.title"
        :data-wheelset-standard-step="index"
        class="wheelset-standards__card"
      >
        <div class="wheelset-standards__visual" aria-hidden="true">
          <Icon :name="step.icon" class="wheelset-standards__visual-icon" />
          <span class="wheelset-standards__visual-label">{{ t(step.visualLabel) }}</span>
          <strong>{{ t(step.visualValue) }}</strong>
          <span class="wheelset-standards__visual-line" />
        </div>

        <div class="wheelset-standards__copy">
          <span class="wheelset-standards__step">{{ t(step.step) }}</span>
          <h4>{{ t(step.title) }}</h4>
          <p>{{ t(step.description) }}</p>
        </div>
      </article>
    </div>

    <div
      class="tz-carousel-pagination wheelset-standards__pagination"
      role="tablist"
      :aria-label="t('home.wheelsetStandards.progress')"
    >
      <button
        v-for="(step, index) in steps"
        :key="step.step"
        type="button"
        class="tz-carousel-pagination__dot"
        :class="{ 'is-active': activeStep === index }"
        :aria-label="t('home.wheelsetStandards.showStep', { step: index + 1 })"
        :aria-selected="activeStep === index"
        role="tab"
        @click="scrollToStep(index)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()
const rail = ref<HTMLElement | null>(null)
const activeStep = ref(0)

const steps = [
  {
    step: 'home.wheelsetStandards.steps.0.step',
    title: 'home.wheelsetStandards.steps.0.title',
    description: 'home.wheelsetStandards.steps.0.description',
    icon: 'lucide:crosshair',
    visualLabel: 'home.wheelsetStandards.steps.0.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.0.visualValue',
  },
  {
    step: 'home.wheelsetStandards.steps.1.step',
    title: 'home.wheelsetStandards.steps.1.title',
    description: 'home.wheelsetStandards.steps.1.description',
    icon: 'lucide:scale',
    visualLabel: 'home.wheelsetStandards.steps.1.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.1.visualValue',
  },
  {
    step: 'home.wheelsetStandards.steps.2.step',
    title: 'home.wheelsetStandards.steps.2.title',
    description: 'home.wheelsetStandards.steps.2.description',
    icon: 'lucide:refresh-cw',
    visualLabel: 'home.wheelsetStandards.steps.2.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.2.visualValue',
  },
  {
    step: 'home.wheelsetStandards.steps.3.step',
    title: 'home.wheelsetStandards.steps.3.title',
    description: 'home.wheelsetStandards.steps.3.description',
    icon: 'lucide:move-horizontal',
    visualLabel: 'home.wheelsetStandards.steps.3.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.3.visualValue',
  },
  {
    step: 'home.wheelsetStandards.steps.4.step',
    title: 'home.wheelsetStandards.steps.4.title',
    description: 'home.wheelsetStandards.steps.4.description',
    icon: 'lucide:sliders-horizontal',
    visualLabel: 'home.wheelsetStandards.steps.4.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.4.visualValue',
  },
  {
    step: 'home.wheelsetStandards.steps.5.step',
    title: 'home.wheelsetStandards.steps.5.title',
    description: 'home.wheelsetStandards.steps.5.description',
    icon: 'lucide:badge-check',
    visualLabel: 'home.wheelsetStandards.steps.5.visualLabel',
    visualValue: 'home.wheelsetStandards.steps.5.visualValue',
  },
] as const

const scrollSteps = (direction: -1 | 1) => {
  const currentRail = rail.value
  if (!currentRail) return

  const firstCard = currentRail.querySelector<HTMLElement>('[data-wheelset-standard-step]')
  const cardWidth = firstCard?.getBoundingClientRect().width ?? currentRail.clientWidth * 0.8
  const gap = Number.parseFloat(
    getComputedStyle(currentRail).columnGap || getComputedStyle(currentRail).gap,
  ) || 0

  currentRail.scrollBy({
    left: direction * (cardWidth + gap),
    behavior: 'smooth',
  })
}

const scrollToStep = (index: number) => {
  const card = rail.value?.querySelector<HTMLElement>(
    `[data-wheelset-standard-step="${index}"]`,
  )
  const currentRail = rail.value
  if (!card || !currentRail) return

  const railRect = currentRail.getBoundingClientRect()
  const cardRect = card.getBoundingClientRect()
  const maxScrollLeft = Math.max(0, currentRail.scrollWidth - currentRail.clientWidth)
  const targetScrollLeft =
    currentRail.scrollLeft +
    (cardRect.left - railRect.left) -
    (currentRail.clientWidth - cardRect.width) / 2

  currentRail.scrollTo({
    left: Math.min(maxScrollLeft, Math.max(0, targetScrollLeft)),
    behavior: 'smooth',
  })
}

const updateActiveStep = () => {
  const currentRail = rail.value
  if (!currentRail) return

  const railCenter = currentRail.getBoundingClientRect().left + currentRail.clientWidth / 2
  const cards = [...currentRail.querySelectorAll<HTMLElement>('[data-wheelset-standard-step]')]
  if (!cards.length) return

  activeStep.value = cards.reduce((closestIndex, card, index) => {
    const current = cards[closestIndex]
    if (!current) return index

    const cardCenter = card.getBoundingClientRect().left + card.clientWidth / 2
    const currentCenter = current.getBoundingClientRect().left + current.clientWidth / 2
    return Math.abs(cardCenter - railCenter) < Math.abs(currentCenter - railCenter)
      ? index
      : closestIndex
  }, 0)
}
</script>

<style scoped>
.wheelset-standards {
  padding-top: 0.25rem;
  overflow-x: hidden;
}

.wheelset-standards__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 0.875rem;
  border-bottom: 1px solid rgba(5, 150, 105, 0.22);
}

.wheelset-standards__eyebrow {
  margin: 0;
  color: #059669;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1.2;
  text-transform: uppercase;
}

.wheelset-standards__title {
  margin: 0.375rem 0 0;
  color: var(--tz-text-primary);
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.2;
}

.wheelset-standards__controls {
  display: flex;
  flex: none;
  gap: 0.5rem;
}

.wheelset-standards__arrow {
  display: inline-grid;
  width: 2rem;
  height: 2rem;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.wheelset-standards__arrow:hover,
.wheelset-standards__arrow:focus-visible {
  border-color: rgba(5, 150, 105, 0.75);
  background: rgba(5, 150, 105, 0.08);
  color: #059669;
}

.wheelset-standards__rail {
  display: flex;
  gap: 0.875rem;
  margin: 0 -0.5rem;
  padding: 1rem 0.5rem 0.75rem;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-padding-inline: 0.5rem;
  scroll-behavior: smooth;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-x;
}

.wheelset-standards__rail::-webkit-scrollbar {
  display: none;
}

.wheelset-standards__card {
  display: grid;
  flex: 0 0 clamp(16rem, 29vw, 19.5rem);
  grid-template-rows: minmax(0, 0.56fr) minmax(0, 0.44fr);
  aspect-ratio: 1 / 1;
  overflow: hidden;
  scroll-snap-align: start;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
}

.wheelset-standards__visual {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  min-height: 0;
  overflow: hidden;
  padding: 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  background:
    linear-gradient(135deg, rgba(5, 150, 105, 0.12), transparent 42%),
    repeating-linear-gradient(0deg, transparent 0, transparent 1.65rem, rgba(20, 32, 43, 0.05) 1.7rem),
    var(--tz-surface-muted);
}

.wheelset-standards__visual::after {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 2.25rem;
  height: 2.25rem;
  border-top: 1px solid rgba(5, 150, 105, 0.8);
  border-right: 1px solid rgba(5, 150, 105, 0.8);
  content: '';
}

.wheelset-standards__visual-icon {
  position: absolute;
  top: 1rem;
  left: 1rem;
  width: 1.25rem;
  height: 1.25rem;
  color: #059669;
}

.wheelset-standards__visual-label {
  position: relative;
  color: var(--tz-text-secondary);
  font-size: 0.625rem;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1.2;
  text-transform: uppercase;
}

.wheelset-standards__visual strong {
  position: relative;
  display: block;
  max-width: 85%;
  margin-top: 0.35rem;
  color: var(--tz-text-primary);
  font-size: clamp(1.25rem, 2.4vw, 1.75rem);
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1;
}

.wheelset-standards__visual-line {
  position: relative;
  display: block;
  width: 42%;
  height: 0.2rem;
  margin-top: 0.8rem;
  border-radius: 999px;
  background: #059669;
}

.wheelset-standards__copy {
  min-height: 0;
  padding: 0.875rem;
}

.wheelset-standards__step {
  display: inline-flex;
  color: #059669;
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.2;
  text-transform: uppercase;
}

.wheelset-standards__copy h4 {
  margin: 0.4rem 0 0;
  color: var(--tz-text-primary);
  font-size: 0.9375rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.25;
}

.wheelset-standards__copy p {
  display: -webkit-box;
  margin: 0.4rem 0 0;
  overflow: hidden;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  letter-spacing: 0;
  line-height: 1.4;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.wheelset-standards__pagination {
  margin-top: 0.125rem;
}

@media (max-width: 767px) {
  .wheelset-standards__card {
    flex-basis: min(82vw, 20rem);
  }

  .wheelset-standards__rail {
    padding-inline-end: 1rem;
    scroll-padding-inline-end: 1rem;
  }
}
</style>
