<template>
  <div class="wheelset-assembly-section space-y-12">
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold tz-text-primary mb-4 flex items-center justify-center gap-2">
        {{ t('testReportAssembly.title') }}
      </h2>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
        {{ t('testReportAssembly.intro') }}
      </p>
    </div>

    <div class="relative space-y-8 before:absolute before:inset-0 before:ml-5 before:h-full before:w-0.5 before:-translate-x-px before:bg-slate-300 md:before:mx-auto md:before:translate-x-0">
      <div
        v-for="step in assemblySteps"
        :key="step.id"
        class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active"
      >
        <div
          class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2"
          :class="step.number === 0
            ? 'tz-surface-panel tz-text-secondary'
            : 'bg-emerald-500 tz-text-primary'"
        >
          <span class="font-bold text-sm">{{ step.number }}</span>
        </div>

        <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
          <h3 class="text-lg font-bold tz-text-primary mb-3">
            {{ t(`testReportAssembly.steps.${step.id}.title`) }}
          </h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
            <ul
              class="space-y-3 pl-4 marker:text-emerald-600"
              :class="step.number === 0 ? 'list-none pl-0 space-y-4' : 'list-disc'"
            >
              <li
                v-for="index in step.itemCount"
                :key="`${step.id}-item-${index}`"
                :class="step.number === 0 ? 'flex gap-2' : undefined"
              >
                <span
                  v-if="step.number === 0"
                  class="text-emerald-600 shrink-0"
                >
                  {{ index }}.
                </span>
                <span>{{ t(`testReportAssembly.steps.${step.id}.items.${index - 1}`) }}</span>
                <div
                  v-if="step.id === 'nipples' && index === 2"
                  class="mt-2 text-xs bg-emerald-50 p-2 rounded border border-emerald-200"
                >
                  <strong class="text-emerald-700">
                    {{ t('testReportAssembly.steps.nipples.note.jBendLabel') }}
                  </strong>
                  {{ t('testReportAssembly.steps.nipples.note.jBendBody') }}<br>
                  <strong class="text-emerald-700">
                    {{ t('testReportAssembly.steps.nipples.note.straightPullLabel') }}
                  </strong>
                  {{ t('testReportAssembly.steps.nipples.note.straightPullBody') }}
                </div>
              </li>
            </ul>

            <div
              v-if="step.imageSources.length"
              class="grid grid-cols-2 gap-3 mt-4"
            >
              <GuideImage
                v-for="(source, index) in step.imageSources"
                :key="source"
                class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                :src="source"
                :alt="t(`testReportAssembly.steps.${step.id}.images.${index}.alt`)"
                :caption="step.id === 'preparation'
                  ? t(`testReportAssembly.steps.${step.id}.images.${index}.caption`)
                  : undefined"
                :zoomOnClick="true"
              />
            </div>

            <div
              v-if="step.id === 'stressRelief'"
              class="mt-4 rounded-lg overflow-hidden shadow-md bg-black/40 cursor-pointer group/video relative aspect-video"
              @click="showStressReliefVideo = true"
            >
              <video
                class="w-full h-full object-cover opacity-80 group-hover/video:opacity-100 transition-opacity"
                muted
                preload="metadata"
              >
                <source src="/testreport/wheelsetassembly/5/Industrial-stress-relieving_high.webm" type="video/webm" />
              </video>

              <div class="absolute inset-0 flex items-center justify-center">
                <div class="w-16 h-16 rounded-full bg-emerald-600/90 tz-text-primary flex items-center justify-center shadow-lg group-hover/video:scale-110 transition-transform">
                  <svg class="w-8 h-8 ml-1" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                </div>
              </div>
              <div class="absolute bottom-3 right-3 px-2 py-1 bg-black/70 rounded text-xs tz-text-primary font-medium backdrop-blur-sm">
                {{ t('testReportAssembly.video.clickToPlay') }}
              </div>
            </div>

            <div
              v-if="step.id === 'finalInspection'"
              class="flex items-center justify-center gap-2 text-emerald-600 font-bold mt-2"
            >
              <span>✓</span>
              <span>{{ t('testReportAssembly.steps.finalInspection.complete') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="rounded-2xl bg-[var(--tz-card-surface)] border tz-border-subtle p-6 md:p-8 mt-12">
      <h3 class="text-lg font-bold tz-text-primary mb-4 flex items-center gap-2">
        <svg class="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        {{ t('testReportAssembly.notes.title') }}
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm tz-text-secondary">
        <div class="tz-surface-panel p-4 rounded-lg">
          <h4 class="tz-text-primary font-bold mb-2">
            {{ t('testReportAssembly.notes.spokeTension.title') }}
          </h4>
          <p class="leading-relaxed">
            {{ t('testReportAssembly.notes.spokeTension.bodyBefore') }}
            <strong class="tz-text-primary">{{ t('testReportAssembly.notes.spokeTension.range') }}</strong>
            {{ t('testReportAssembly.notes.spokeTension.bodyMiddle') }}
            <strong class="tz-text-primary">{{ t('testReportAssembly.notes.spokeTension.difference') }}</strong>
            {{ t('testReportAssembly.notes.spokeTension.bodyAfter') }}
          </p>
          <div class="mt-4">
            <NuxtLink
              :to="localePath('/guides/wheelset-buyers/wheel-components')"
              class="premium-button"
            >
              {{ t('testReportAssembly.notes.tensionLink') }}
            </NuxtLink>
          </div>
        </div>

        <div class="tz-surface-panel p-4 rounded-lg">
          <h4 class="tz-text-primary font-bold mb-2">
            {{ t('testReportAssembly.notes.tolerances.title') }}
          </h4>
          <ul class="list-none space-y-1">
            <li>
              {{ t('testReportAssembly.notes.tolerances.lateralRadial') }}
              <strong class="tz-text-primary">
                {{ t('testReportAssembly.notes.tolerances.maximum02') }}
              </strong>
            </li>
            <li>
              {{ t('testReportAssembly.notes.tolerances.dishOffset') }}
              <strong class="tz-text-primary">
                {{ t('testReportAssembly.notes.tolerances.maximum05') }}
              </strong>
            </li>
          </ul>
        </div>

        <div class="md:col-span-2 tz-surface-panel p-4 rounded-lg">
          <h4 class="tz-text-primary font-bold mb-2">
            {{ t('testReportAssembly.notes.measurementStandards.title') }}
          </h4>
          <p class="leading-relaxed">
            {{ t('testReportAssembly.notes.measurementStandards.body') }}
          </p>
        </div>
      </div>
    </div>

    <div
      v-if="showStressReliefVideo"
      class="support-video-modal"
      role="dialog"
      aria-modal="true"
    >
      <div class="support-video-modal__backdrop" @click="showStressReliefVideo = false" />
      <div class="support-video-modal__content">
        <button
          type="button"
          class="tz-global-close-btn support-video-modal__close"
          @click="showStressReliefVideo = false"
        >
          ×
        </button>
        <video
          class="support-video-modal__video"
          controls
          autoplay
        >
          <source src="/testreport/wheelsetassembly/5/Industrial-stress-relieving_high.webm" type="video/webm" />
        </video>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const localePath = useLocalePath()
const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('testReportAssembly')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const showStressReliefVideo = ref(false)

const assemblySteps = [
  {
    id: 'preparation',
    number: 0,
    itemCount: 2,
    imageSources: [
      '/testreport/wheelsetassembly/0/wheelsbuild-wheel-truing-stand-and-wheel-building-jig.webp',
      '/testreport/wheelsetassembly/0/wheel-building-tools.webp',
    ],
  },
  {
    id: 'lacing',
    number: 1,
    itemCount: 2,
    imageSources: [
      '/testreport/wheelsetassembly/1/carbonrim- spoke-lacing.webp',
      '/testreport/wheelsetassembly/1/carbonrim- spoke-lacing1.webp',
    ],
  },
  {
    id: 'nipples',
    number: 2,
    itemCount: 2,
    imageSources: [
      '/testreport/wheelsetassembly/2/wheel-lacing.webp',
      '/testreport/wheelsetassembly/2/wheel-lacing-end.webp',
    ],
  },
  {
    id: 'tensioning',
    number: 3,
    itemCount: 2,
    imageSources: [
      '/testreport/wheelsetassembly/3/tighten-the-spoke-nipples.webp',
      '/testreport/wheelsetassembly/3/measure-spoke-tension.webp',
    ],
  },
  {
    id: 'truing',
    number: 4,
    itemCount: 3,
    imageSources: [
      '/testreport/wheelsetassembly/4/self‑developed-wheel-building-equipment.webp',
      '/testreport/wheelsetassembly/4/wheelsbuilding-and-check-spoke-tension.webp',
    ],
  },
  {
    id: 'stressRelief',
    number: 5,
    itemCount: 2,
    imageSources: [],
  },
  {
    id: 'finalInspection',
    number: 6,
    itemCount: 3,
    imageSources: [
      '/testreport/wheelsetassembly/6/tubeless-rim-tape.webp',
      '/testreport/wheelsetassembly/6/wheelset-packaging.webp',
    ],
  },
] as const
</script>

<style scoped>
.support-video-modal {
  position: fixed;
  inset: 0;
  z-index: 10002;
  display: flex;
  align-items: center;
  justify-content: center;
  padding:
    var(--tz-safe-area-top, 0px)
    var(--tz-safe-area-right, 0px)
    var(--tz-safe-area-bottom, 0px)
    var(--tz-safe-area-left, 0px);
}

.support-video-modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(4px);
}

.support-video-modal__content {
  position: relative;
  z-index: 41;
  width: 100%;
  max-width: 960px;
  margin: 0 1rem;
  background: var(--tz-card-surface);
  border-radius: 0.75rem;
  box-shadow: 0 20px 40px rgba(20, 32, 43, 0.16);
  overflow: hidden;
}

.support-video-modal__close {
  position: absolute;
  top: max(0.5rem, calc(0.5rem + var(--tz-safe-area-top, 0px)));
  right: max(1rem, calc(1rem + var(--tz-safe-area-right, 0px)));
  line-height: 1;
  z-index: 50;
}

.support-video-modal__video {
  display: block;
  width: 100%;
  height: auto;
  max-height: min(80vh, var(--tz-mobile-safe-viewport-height, 80vh));
}
</style>
