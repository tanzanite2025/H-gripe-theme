<template>
  <div class="installation-guide space-y-12">
    <!-- Intro / Overview Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold tz-text-secondary mb-4 flex items-center justify-center gap-2">
        {{ t('guidesTireInstallation.title') }}
      </h2>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
         {{ t('guidesTireInstallation.intro') }}
      </p>
      
      <div class="flex justify-center mb-4">
         <button
            type="button"
            class="inline-flex items-center justify-center rounded-full tz-surface-panel border tz-border-subtle px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-secondary hover:tz-surface-subtle transition-colors"
            @click="$emit('changeTab', 'tubeless')"
          >
            {{ t('guidesTireInstallation.actions.viewAccessories') }}
          </button>
      </div>

       <div class="mt-8">
        <h4 class="text-xs font-bold uppercase tracking-wider tz-text-muted mb-4">{{ t('guidesTireInstallation.essential.title') }}</h4>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle tubeless tire pad.webp"
            :alt="t('guidesTireInstallation.essential.images.tirePadAlt')"
            :zoomOnClick="true"
            :caption="t('guidesTireInstallation.essential.images.tirePadLabel')"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle vacuum self-replenishing fluid.webp"
            :alt="t('guidesTireInstallation.essential.images.sealantAlt')"
            :zoomOnClick="true"
            :caption="t('guidesTireInstallation.essential.images.sealantLabel')"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle vacuum valve.webp"
            :alt="t('guidesTireInstallation.essential.images.valveAlt')"
            :zoomOnClick="true"
             :caption="t('guidesTireInstallation.essential.images.valveLabel')"
          />
        </div>
      </div>
    </div>
    
    <!-- Timeline Steps -->
    <div class="relative space-y-8 before:absolute before:inset-0 before:ml-5 before:h-full before:w-0.5 before:-translate-x-px before:bg-slate-300 md:before:mx-auto md:before:translate-x-0">
      
      <!-- Step 0: Preparation -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] tz-surface-panel tz-text-secondary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">0</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">{{ t('guidesTireInstallation.steps.preparation.title') }}</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <ul class="space-y-2 list-none pl-0">
                 <li class="flex gap-2"><span class="text-emerald-600">1.</span> {{ t('guidesTireInstallation.steps.preparation.items.0') }}</li>
                 <li class="flex gap-2"><span class="text-emerald-600">2.</span> {{ t('guidesTireInstallation.steps.preparation.items.1') }}</li>
                 <li class="flex gap-2"><span class="text-emerald-600">3.</span> {{ t('guidesTireInstallation.steps.preparation.items.2') }}</li>
                 <li class="flex gap-2"><span class="text-emerald-600">4.</span> {{ t('guidesTireInstallation.steps.preparation.items.3') }}</li>
               </ul>

               <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/Bicycle-tubeless-tire-pads.webp"
                    :alt="t('guidesTireInstallation.steps.preparation.images.tapingAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/howtoopenthe vacuumnozzle.webp"
                    :alt="t('guidesTireInstallation.steps.preparation.images.valveHoleAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/tubeless-tirepad-interface-overlaps.webp"
                    :alt="t('guidesTireInstallation.steps.preparation.images.overlapAlt')"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 1: Install Tire -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">1</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">{{ t('guidesTireInstallation.steps.install.title') }}</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <p>{{ t('guidesTireInstallation.steps.install.body') }}</p>
               <div class="bg-emerald-50 p-3 rounded-lg border border-emerald-200 text-xs">
                  <strong>{{ t('guidesTireInstallation.steps.install.proTipTitle') }}:</strong> {{ t('guidesTireInstallation.steps.install.proTip') }}
               </div>

                <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/first-step-to-install-tubelesstires.webp"
                    :alt="t('guidesTireInstallation.steps.install.images.beadAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp"
                    :alt="t('guidesTireInstallation.steps.install.images.seatingAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/Picture-of-whether-the-vacuum-nozzle-is-installed-properly.webp"
                    :alt="t('guidesTireInstallation.steps.install.images.valveCheckAlt')"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 2: Add Sealant -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">2</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">{{ t('guidesTireInstallation.steps.sealant.title') }}</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <ul class="space-y-2 list-none pl-0">
                 <li class="flex gap-2"><span class="text-emerald-600">1.</span> {{ t('guidesTireInstallation.steps.sealant.items.0') }}</li>
                 <li class="flex gap-2"><span class="text-emerald-600">2.</span> {{ t('guidesTireInstallation.steps.sealant.items.1') }}</li>
                 <li class="flex gap-2"><span class="text-emerald-600">3.</span> {{ t('guidesTireInstallation.steps.sealant.items.2') }}</li>
               </ul>

                <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant1.webp"
                    :alt="t('guidesTireInstallation.steps.sealant.images.removeCoreAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant2.webp"
                    :alt="t('guidesTireInstallation.steps.sealant.images.injectAlt')"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Separate-vacuum-nozzle-tool.webp"
                    :alt="t('guidesTireInstallation.steps.sealant.images.toolAlt')"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 3: Inflate -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">3</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">{{ t('guidesTireInstallation.steps.inflate.title') }}</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <p>{{ t('guidesTireInstallation.steps.inflate.inflate') }}</p>
               <p>{{ t('guidesTireInstallation.steps.inflate.shake') }}</p>
               <div class="flex items-center justify-center gap-2 text-emerald-600 font-bold mt-2">
                  <span>✓</span> <span>{{ t('guidesTireInstallation.steps.inflate.ready') }}</span>
               </div>
            </div>
         </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireInstallation')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

// Match SizeChartsTabId from parent
type GuideTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'tire-pressure' | 'tube'

defineEmits<{
  (e: 'changeTab', tab: GuideTabId): void
}>()
</script>
