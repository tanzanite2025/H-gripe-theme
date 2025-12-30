<template>
  <div class="inner-tube-guide">
    <h2 class="sizecharts-section__title">Inner Tube</h2>
    <div class="mt-2 flex justify-center">
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
        @click="openInnerTubeSearch"
      >
        Search inner tubes with advanced filters
      </button>
    </div>
    <h3 class="sizecharts-section__subheading mt-3">Step 1: Choose your tire size</h3>
    <p class="sizecharts-section__intro">
      All information can be found on your tire.
    </p>

    <h3 class="sizecharts-section__subheading mt-4">Step 2: Choose your valve type</h3>
    <ul class="sizecharts-section__list">
      <li><strong>AV</strong> Schrader, Auto, Moto. Length 40 mm.</li>
      <li><strong>DV</strong> Bicycle, Woods, Hollands. Length 32 mm or 40 mm.</li>
      <li><strong>SV</strong> Presta, Sclaverand, French. Length 40 mm.</li>
      <li><strong>SV Extra long</strong> Presta, Sclaverand, French. Length 50 mm / 60 mm / 80 mm.</li>
    </ul>

    <div class="mt-3">
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/innertube/tanzanite-schwalbe-innertube-valve.webp"
        alt="Schwalbe inner tube valve types and lengths overview"
        :zoomOnClick="true"
      />
    </div>

    <div class="mt-6 space-y-6">
      <div>
        <h4 class="sizecharts-section__subheading flex justify-center items-center gap-2 text-sky-400">
          <span>🔧</span> Pump Compatibility
        </h4>
        <ul class="sizecharts-section__list">
          <li><strong>Standard pumps</strong>: Suitable for ≤45mm rims, 40–60mm valves work fine.</li>
          <li><strong>Locking-head pumps</strong>: Better for mid/high rims, securely grip shorter valves.</li>
          <li><strong>Valve extenders</strong>: Essential for ≥70mm rims, especially inner-type extenders for airtightness.</li>
        </ul>
      </div>

      <div class="overflow-x-auto rounded-lg border border-slate-700/50 shadow-lg">
        <table class="w-full text-sm text-left text-slate-300">
          <thead class="text-xs uppercase bg-slate-800/80 text-slate-200">
            <tr>
              <th scope="col" class="px-4 py-3 font-semibold whitespace-nowrap">Rim Height</th>
              <th scope="col" class="px-4 py-3 font-semibold whitespace-nowrap">Rec. Valve Length</th>
              <th scope="col" class="px-4 py-3 font-semibold">Pump Compatibility</th>
              <th scope="col" class="px-4 py-3 font-semibold min-w-[200px]">Notes</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700/50 bg-slate-900/40">
            <tr class="hover:bg-slate-800/30 transition-colors">
              <td class="px-4 py-3 font-medium text-white whitespace-nowrap">≤25mm</td>
              <td class="px-4 py-3 text-sky-300 font-mono">40mm</td>
              <td class="px-4 py-3">Compatible with most pumps</td>
              <td class="px-4 py-3 text-slate-400 text-xs">Standard low-profile rims, regular valves are sufficient</td>
            </tr>
            <tr class="hover:bg-slate-800/30 transition-colors">
              <td class="px-4 py-3 font-medium text-white whitespace-nowrap">30–45mm</td>
              <td class="px-4 py-3 text-sky-300 font-mono">60mm</td>
              <td class="px-4 py-3">Recommended to use pumps with locking heads</td>
              <td class="px-4 py-3 text-slate-400 text-xs">Mid-profile rims, standard valves may be too short</td>
            </tr>
            <tr class="hover:bg-slate-800/30 transition-colors">
              <td class="px-4 py-3 font-medium text-white whitespace-nowrap">50–65mm</td>
              <td class="px-4 py-3 text-sky-300 font-mono">80mm</td>
              <td class="px-4 py-3">Some pumps require valve extenders</td>
              <td class="px-4 py-3 text-slate-400 text-xs">High-profile rims, longer valves needed</td>
            </tr>
            <tr class="hover:bg-slate-800/30 transition-colors">
              <td class="px-4 py-3 font-medium text-white whitespace-nowrap">≥70mm</td>
              <td class="px-4 py-3 text-amber-300 font-mono">80mm + extender</td>
              <td class="px-4 py-3">Extender required, inner type recommended</td>
              <td class="px-4 py-3 text-slate-400 text-xs">Ultra-high rims, no extra-long inner tube valves available</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div>
        <h4 class="sizecharts-section__subheading flex justify-center items-center gap-2 text-amber-400">
          <span>⚠️</span> Notes
        </h4>
        <ul class="sizecharts-section__list">
          <li><strong>Rule of thumb</strong>: Valve length should exceed rim height by at least 15mm to ensure secure pump connection.</li>
          <li>
            <strong>Valve extenders</strong>:
            <span class="block mt-1 text-xs text-slate-400">
              • <strong>External type</strong>: Easy to install, but less airtight.<br>
              • <strong>Internal type</strong>: Requires removing valve core, better airtightness, recommended for tubeless setups.
            </span>
          </li>
          <li><strong>Brand differences</strong>: Continental and Schwalbe support internal extenders; cheaper inner tubes often do not.</li>
        </ul>
      </div>
    </div>

    <h3 class="sizecharts-section__subheading mt-4">Step 3: Choose your execution.</h3>
    <ul class="sizecharts-section__list">
      <li>
        <div><strong class="text-amber-300">Standard</strong></div>
        <div>A standard Schwalbe tube holds air considerably longer. Due to the purity of the rubber compound and the uniform wall thickness.</div>
      </li>
      <li>
        <div><strong class="text-amber-300">Air Plus "NEW"</strong></div>
        <div>Holds air even longer thanks to greater wall thickness. Less frequent re-inflation is necessary. Also provides above-average protection against punctures.</div>
      </li>
      <li>
        <div><strong class="text-amber-300">Extralight</strong></div>
        <div>Reduced weight. Same reliability as the standard version. Not available in all sizes.</div>
      </li>
      <li>
        <div><strong class="text-amber-300">XXLight</strong></div>
        <div>Extremely light. Only for a very few sizes.</div>
      </li>
      <li>
        <div><strong class="text-amber-300">Freeride</strong></div>
        <div>Large volume. Low Weight.</div>
      </li>
      <li>
        <div><strong class="text-amber-300">Downhill</strong></div>
        <div>Extremely robust. A reinforced rubber base protects against the dreaded valve shear off. Only in Downhill sizes. If you want large volume with low weight, choose Extralight.</div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'

const { open: openShopSearchSheet } = useShopSearchSheet()

const openInnerTubeSearch = () => {
  openShopSearchSheet({
    presetCategorySlug: 'inner-tube',
    presetKeywords: ['Inner tube'],
  })
}
</script>

<style scoped>
/* Scoped styles specific to this component can go here if needed */
/* Most styles are global via .sizecharts-section classes in guide-sections.css */
</style>
