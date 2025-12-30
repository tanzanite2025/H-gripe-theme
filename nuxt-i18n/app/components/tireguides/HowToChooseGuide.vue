<template>
  <div class="how-to-choose-guide">
    <h2 class="sizecharts-section__title">How to Choose</h2>
    <!-- Tire width -> rim internal width helper placed directly under title -->
    <TireRimHelper />

    <div class="mt-4">
      <h3 class="sizecharts-section__subheading">
        Reference charts from DT Swiss
      </h3>
      <p class="sizecharts-section__intro">
        The following hookless (TSS) and hooked (TC) rim inner width charts from DT Swiss
        show recommended (dark) and possible (light) combinations. Always stay within the
        limits specified by your tire and rim manufacturers.
      </p>
      <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="rounded-xl bg-slate-900/70 p-2 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
          <img
            src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
            alt="DT Swiss hookless TSS rim inner width recommendation chart"
            class="block h-auto w-full"
            loading="lazy"
          />
        </div>
        <div class="rounded-xl bg-slate-900/70 p-2 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
          <img
            src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
            alt="DT Swiss hooked TC rim inner width recommendation chart"
            class="block h-auto w-full"
            loading="lazy"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TireRimHelper from '~/components/TireRimHelper.vue'
</script>
