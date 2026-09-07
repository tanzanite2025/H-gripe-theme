<template>
  <div class="space-y-8">
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-slate-500">
      <div class="flex flex-col items-center">
        <h3 class="text-lg font-bold tz-text-primary mb-2">
          {{ t('guidesWheelsetBuyersChooseFreehub.intro.title') }}
        </h3>
        <p class="text-sm tz-text-secondary leading-relaxed max-w-2xl mx-auto mb-4">
          {{ t('guidesWheelsetBuyersChooseFreehub.intro.body') }}
        </p>
        <FreehubGroupsetHelper />
      </div>
    </div>

    <article
      v-for="ecosystem in ecosystems"
      :key="ecosystem.id"
      class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300"
    >
      <div
        class="flex items-center gap-3 mb-6 pb-3 border-b"
        :class="ecosystem.borderClass"
      >
        <h3 class="text-lg font-bold tz-text-primary">{{ ecosystem.title }}</h3>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6" :class="ecosystem.gridClass">
        <div
          v-for="card in ecosystem.cards"
          :key="card.id"
          class="rounded-xl p-3 border"
          :class="card.surfaceClass"
        >
          <GuideImage
            :src="card.src"
            :alt="t(card.altKey)"
            :zoomOnClick="true"
            :caption="t(card.captionKey)"
            class="rounded-lg mb-2"
          />
          <strong class="block text-sm mb-1 text-center" :class="card.accentClass">
            {{ t(card.nameKey) }}
          </strong>
          <p class="text-xs tz-text-muted text-center">{{ t(card.bodyKey) }}</p>
        </div>
      </div>

      <div class="tz-surface-panel rounded-xl p-4">
        <h4 class="text-xs font-bold tz-text-muted uppercase tracking-widest mb-3">
          {{ t(ecosystem.cheatsheetTitleKey) }}
        </h4>
        <ul class="space-y-3">
          <li
            v-for="item in ecosystem.cheatsheet"
            :key="item.key"
            class="flex items-start gap-3 text-xs tz-text-secondary"
          >
            <span class="mt-0.5 w-1.5 h-1.5 rounded-full shrink-0" :class="item.dotClass"></span>
            <span>{{ t(item.key) }}</span>
          </li>
        </ul>
      </div>
    </article>

    <div class="pt-6 border-t tz-border-subtle">
      <h3 class="text-base font-bold tz-text-primary mb-4 px-2 border-l-4 border-slate-500">
        {{ t('guidesWheelsetBuyersChooseFreehub.reference.standardsTitle') }}
      </h3>
      <div class="overflow-x-auto rounded-xl shadow-md bg-[var(--tz-card-surface)]">
        <table class="min-w-full text-left text-xs sm:text-sm tz-text-secondary">
          <thead class="tz-surface-panel font-bold tz-text-primary uppercase tracking-wider tz-micro-label sm:text-xs">
            <tr>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.standardsHeaders.type') }}</th>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.standardsHeaders.groupsets') }}</th>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.standardsHeaders.features') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200/50">
            <tr
              v-for="row in standardsRows"
              :key="`${row.type}-${row.groupsets}`"
              class="hover:tz-surface-panel transition-colors"
            >
              <td class="px-4 py-3 font-medium tz-text-primary">{{ row.type }}</td>
              <td class="px-4 py-3">{{ row.groupsets }}</td>
              <td class="px-4 py-3">{{ row.features }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h3 class="text-base font-bold tz-text-primary mt-8 mb-4 px-2 border-l-4 border-slate-500">
        {{ t('guidesWheelsetBuyersChooseFreehub.reference.matrixTitle') }}
      </h3>
      <div class="overflow-x-auto rounded-xl shadow-md bg-[var(--tz-card-surface)]">
        <table class="min-w-full text-left text-xs sm:text-sm tz-text-secondary">
          <thead class="tz-surface-panel font-bold tz-text-primary uppercase tracking-wider tz-micro-label sm:text-xs">
            <tr>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.matrixHeaders.brand') }}</th>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.matrixHeaders.freehub') }}</th>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.matrixHeaders.cassette') }}</th>
              <th class="px-4 py-3">{{ t('guidesWheelsetBuyersChooseFreehub.reference.matrixHeaders.spacer') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200/50">
            <tr
              v-for="row in matrixRows"
              :key="`${row.brand}-${row.freehub}-${row.cassette}`"
              class="hover:tz-surface-panel transition-colors"
            >
              <td class="px-4 py-3 font-medium" :class="brandClass(row.tone)">{{ row.brand }}</td>
              <td class="px-4 py-3">{{ row.freehub }}</td>
              <td class="px-4 py-3">{{ row.cassette }}</td>
              <td class="px-4 py-3" :class="spacerClass(row.spacerTone)">{{ row.spacer }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import FreehubGroupsetHelper from '~/components/FreehubGroupsetHelper.vue'
import { usePageMessages } from '~/composables/usePageMessages'

type EcosystemId = 'shimano' | 'sram' | 'campagnolo'
type Tone = 'shimano' | 'sram' | 'campagnolo' | 'mavic'
type SpacerTone = 'ok' | 'warning'

interface EcosystemCard {
  id: string
  src: string
  altKey: string
  captionKey: string
  nameKey: string
  bodyKey: string
  accentClass: string
  surfaceClass: string
}

interface Ecosystem {
  id: EcosystemId
  title: string
  borderClass: string
  gridClass: string
  cards: EcosystemCard[]
  cheatsheetTitleKey: string
  cheatsheet: { key: string; dotClass: string }[]
}

interface MatrixRow {
  brand: string
  freehub: string
  cassette: string
  spacer: string
  tone: Tone
  spacerTone: SpacerTone
}

const { locale, t, tm, rt } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetBuyersChooseFreehub')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const ecosystems = computed<Ecosystem[]>(() => [
  {
    id: 'shimano',
    title: t('guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.title'),
    borderClass: 'border-emerald-200',
    gridClass: 'md:grid-cols-3',
    cards: [
      {
        id: 'micro-spline',
        src: '/public/wheelsetbuyersguide/choose%20freehub/shimano-micro-spline-11-12-speed-mountain-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.microSpline.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.microSpline.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.microSpline.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.microSpline.body',
        accentClass: 'text-emerald-600',
        surfaceClass: 'bg-emerald-50 border-emerald-200',
      },
      {
        id: 'hg-road',
        src: '/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-road-hyper-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgRoad.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgRoad.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgRoad.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgRoad.body',
        accentClass: 'text-emerald-600',
        surfaceClass: 'bg-emerald-50 border-emerald-200',
      },
      {
        id: 'hg-mountain',
        src: '/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-mountain-hyper-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgMountain.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgMountain.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgMountain.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cards.hgMountain.body',
        accentClass: 'text-emerald-600',
        surfaceClass: 'bg-emerald-50 border-emerald-200',
      },
    ],
    cheatsheetTitleKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cheatsheetTitle',
    cheatsheet: [
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cheatsheet.microSpline',
        dotClass: 'bg-emerald-500',
      },
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.shimano.cheatsheet.hyperglide',
        dotClass: 'bg-emerald-500',
      },
    ],
  },
  {
    id: 'sram',
    title: t('guidesWheelsetBuyersChooseFreehub.ecosystems.sram.title'),
    borderClass: 'border-red-500/10',
    gridClass: 'md:grid-cols-2',
    cards: [
      {
        id: 'xd',
        src: '/public/wheelsetbuyersguide/choose%20freehub/sram-xd-11-12-speed-mountain-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xd.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xd.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xd.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xd.body',
        accentClass: 'text-red-400',
        surfaceClass: 'bg-red-500/5 border-red-500/10',
      },
      {
        id: 'xdr',
        src: '/public/wheelsetbuyersguide/choose%20freehub/sram-xdr-road-11-12-speed-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xdr.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xdr.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xdr.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cards.xdr.body',
        accentClass: 'text-red-400',
        surfaceClass: 'bg-red-500/5 border-red-500/10',
      },
    ],
    cheatsheetTitleKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cheatsheetTitle',
    cheatsheet: [
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cheatsheet.xd',
        dotClass: 'bg-red-500',
      },
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.sram.cheatsheet.xdr',
        dotClass: 'bg-red-500',
      },
    ],
  },
  {
    id: 'campagnolo',
    title: t('guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.title'),
    borderClass: 'border-emerald-200',
    gridClass: 'md:grid-cols-2',
    cards: [
      {
        id: 'n3w',
        src: '/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-N3W-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.n3w.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.n3w.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.n3w.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.n3w.body',
        accentClass: 'text-emerald-600',
        surfaceClass: 'bg-emerald-50 border-emerald-200',
      },
      {
        id: 'classic',
        src: '/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-spd-freehub.webp',
        altKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.classic.name',
        captionKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.classic.caption',
        nameKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.classic.name',
        bodyKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cards.classic.body',
        accentClass: 'text-emerald-600',
        surfaceClass: 'bg-emerald-50 border-emerald-200',
      },
    ],
    cheatsheetTitleKey: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cheatsheetTitle',
    cheatsheet: [
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cheatsheet.n3w',
        dotClass: 'bg-emerald-500',
      },
      {
        key: 'guidesWheelsetBuyersChooseFreehub.ecosystems.campagnolo.cheatsheet.classic',
        dotClass: 'bg-emerald-500',
      },
    ],
  },
])

const standardsRows = computed(() => {
  const rows = tm('guidesWheelsetBuyersChooseFreehub.reference.standardsRows') as {
    type: string
    groupsets: string
    features: string
  }[]
  return rows.map(row => ({
    type: rt(row.type),
    groupsets: rt(row.groupsets),
    features: rt(row.features),
  }))
})

const matrixRows = computed(() => {
  const rows = tm('guidesWheelsetBuyersChooseFreehub.reference.matrixRows') as MatrixRow[]
  return rows.map(row => ({
    ...row,
    brand: rt(row.brand),
    freehub: rt(row.freehub),
    cassette: rt(row.cassette),
    spacer: rt(row.spacer),
  }))
})

const brandClass = (tone: Tone) => ({
  shimano: 'text-emerald-600',
  sram: 'text-red-400',
  campagnolo: 'text-emerald-600',
  mavic: 'text-yellow-400',
}[tone])

const spacerClass = (tone: SpacerTone) => (
  tone === 'warning' ? 'text-amber-400' : 'text-emerald-600'
)
</script>
