<template>
  <div class="schwalbe-tire-circumference-guide">
    <section class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-4 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
        {{ t('guidesSchwalbeTireCircumference.circumference.title') }}
      </h2>
      <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        {{ t('guidesSchwalbeTireCircumference.circumference.intro') }}
      </p>
      <p class="mx-auto mb-8 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        {{ t('guidesSchwalbeTireCircumference.circumference.test') }}
      </p>

      <div class="schwalbe-tire-circumference-guide__table-scroll text-left">
        <table class="schwalbe-tire-circumference-guide__table">
          <caption>
            {{ t('guidesSchwalbeTireCircumference.circumference.caption') }}
          </caption>
          <thead>
            <tr>
              <th scope="col">{{ t('guidesSchwalbeTireCircumference.circumference.headers.wheelSize') }}</th>
              <th scope="col">{{ t('guidesSchwalbeTireCircumference.circumference.headers.etrto') }}</th>
              <th scope="col">{{ t('guidesSchwalbeTireCircumference.circumference.headers.circumference') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in schwalbeTireCircumferenceRows" :key="row.etrto">
              <td class="schwalbe-tire-circumference-guide__dimension-cell">{{ row.inch }}</td>
              <th scope="row">{{ row.etrto }}</th>
              <td class="schwalbe-tire-circumference-guide__dimension-cell">{{ row.circumferenceMm }} mm</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mx-auto mt-8 w-full max-w-4xl">
        <h3 class="schwalbe-tire-circumference-guide__visual-title">
          {{ t('guidesSchwalbeTireCircumference.circumference.visualTitle') }}
        </h3>
        <GuideImage
          class="mt-4 overflow-hidden rounded-xl shadow-md"
          src="/public/tiresizecharts/schwalbe-tire-circumference/exact-circumference-of-tire.webp"
          :alt="t('guidesSchwalbeTireCircumference.images.circumferenceAlt')"
          :caption="t('guidesSchwalbeTireCircumference.images.circumferenceCaption')"
          :zoomOnClick="true"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { schwalbeTireCircumferenceRows } from '~/data/tireguides/schwalbeTireCircumference'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesSchwalbeTireCircumference')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>

<style scoped>
.schwalbe-tire-circumference-guide__table-scroll {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.schwalbe-tire-circumference-guide__table {
  min-width: 38rem;
  width: 100%;
  border-collapse: collapse;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.4;
  table-layout: fixed;
}

.schwalbe-tire-circumference-guide__table caption {
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
  text-align: left;
}

.schwalbe-tire-circumference-guide__table th,
.schwalbe-tire-circumference-guide__table td {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  padding: 0.55rem 0.65rem;
  text-align: left;
  vertical-align: top;
}

.schwalbe-tire-circumference-guide__table thead th {
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1.35;
}

.schwalbe-tire-circumference-guide__table tbody tr:hover {
  background: var(--tz-surface-subtle);
}

.schwalbe-tire-circumference-guide__table tbody th {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.schwalbe-tire-circumference-guide__table th:nth-child(1),
.schwalbe-tire-circumference-guide__table td:nth-child(1) {
  width: 25%;
}

.schwalbe-tire-circumference-guide__table th:nth-child(2),
.schwalbe-tire-circumference-guide__table td:nth-child(2) {
  width: 30%;
}

.schwalbe-tire-circumference-guide__table th:nth-child(3),
.schwalbe-tire-circumference-guide__table td:nth-child(3) {
  width: 45%;
}

.schwalbe-tire-circumference-guide__dimension-cell {
  color: var(--tz-text-secondary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.schwalbe-tire-circumference-guide__visual-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  text-transform: uppercase;
}

@media (min-width: 768px) {
  .schwalbe-tire-circumference-guide__table {
    min-width: 0;
  }
}
</style>