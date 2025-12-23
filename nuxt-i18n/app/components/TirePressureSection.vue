<template>
  <div class="tire-pressure-section">
    <div class="mt-1 space-y-4 text-sm text-slate-100">
      <p class="sizecharts-section__intro">
        For safety and performance, the tire and rim on the bike must be perfectly matched. This applies all the more to tubeless
        tires. In addition to the classic hooked rims (crotched type), so-called "hookless" rims (straight side) are becoming
        increasingly popular on current model bikes. Hookless rims in particular place high demands on tubeless tires for safe
        operation. The Schwalbe TLR/TLE models (from model year 2020) are designed to meet these special requirements, especially in
        the bead core area, and are therefore generally compatible.
      </p>
      <p class="sizecharts-section__intro">
        Rim manufacturers often use wider inner bead widths for their hookless rims. The air pressure should therefore be reduced
        accordingly, as the real volume of the tire increases with wider inner rim widths.
      </p>
      <p class="sizecharts-section__intro">
        According to the ETRTO (European Tyre and Rim Technical Organisation), hookless rims have a maximum pressure of 5 bar for
        road bikes. This maximum tire pressure must be adhered to, or less if the maximum pressure stated by the rim/wheel
        manufacturer is lower, even if the maximum pressure printed or embossed on the tire is higher. Therefore, if the specified
        maximum pressure of the rim is lower than that of the tire, the maximum pressure of the rim must be observed. There are also
        restrictions imposed by the rim manufacturer with regard to the choice of tire, which must also be observed. Many rim and
        wheel manufacturers recommend a minimum tire width of 28 mm in order to be able to ride the low air pressure without any
        problems.
      </p>

      <div class="mt-1">
        <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
          @click="emit('openTireProducts')"
        >
          Check out tires
        </button>
      </div>

      <div>
        <h3 class="text-sm font-semibold text-slate-100">
          1. Why tire pressure matters
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>
            <span class="font-semibold">Performance:</span>
            Too high → less grip; too low → more rolling resistance.
          </li>
          <li>
            <span class="font-semibold">Safety:</span>
            Overinflation can damage rims; underinflation increases risk of pinch flats.
          </li>
          <li>
            <span class="font-semibold">Comfort:</span>
            Proper pressure reduces vibrations and improves ride quality.
          </li>
        </ul>
      </div>

      <div>
        <h3 class="text-sm font-semibold text-slate-100">
          2. Key factors affecting pressure
        </h3>
        <div class="mt-2 overflow-x-auto rounded-xl border border-slate-800 bg-slate-950/40">
          <table class="min-w-full text-left text-xs sm:text-sm text-slate-100">
            <thead class="bg-slate-900/80">
              <tr>
                <th class="px-3 py-2 font-semibold">Factor</th>
                <th class="px-3 py-2 font-semibold">Explanation</th>
                <th class="px-3 py-2 font-semibold">Adjustment direction</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr>
                <td class="px-3 py-2 align-top">Rider weight</td>
                <td class="px-3 py-2 align-top">Heavier riders need higher pressure.</td>
                <td class="px-3 py-2 align-top">Add 5–10 psi for heavier riders.</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">Tire width</td>
                <td class="px-3 py-2 align-top">Wider tires can run lower pressure.</td>
                <td class="px-3 py-2 align-top">26×2.0 MTB: 30–50 psi; 700×25C road: 90–110 psi.</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">Terrain</td>
                <td class="px-3 py-2 align-top">Soft/muddy → lower pressure for grip; rocky → higher pressure for rim protection.</td>
                <td class="px-3 py-2 align-top">Off-road lower, on-road higher.</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">Riding style</td>
                <td class="px-3 py-2 align-top">Aggressive racing → higher pressure; casual riding → lower.</td>
                <td class="px-3 py-2 align-top">Racing: 110–130 psi; leisure: 80–100 psi.</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">Rim design</td>
                <td class="px-3 py-2 align-top">Wider rims support lower pressure; narrow rims require higher.</td>
                <td class="px-3 py-2 align-top">Follow manufacturer’s range.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div>
        <h3 class="text-sm font-semibold text-slate-100">
          3. Typical pressure ranges (approximate)
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>Road bikes: 80–120 psi (up to 130 psi for racing; as low as 70 psi for cobbled classics).</li>
          <li>Mountain bikes: 30–60 psi (XC often 25–40 psi).</li>
          <li>City / commuter bikes: 40–70 psi.</li>
        </ul>
      </div>

      <div>
        <h3 class="text-sm font-semibold text-slate-100">
          4. Practical tips
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>Check the sidewall: always follow the manufacturer’s recommended range.</li>
          <li>Use a reliable gauge: don’t rely on "feel".</li>
          <li>Inspect regularly: tires lose air naturally; check weekly or before rides.</li>
          <li>Avoid extremes: never exceed max or drop below min pressure.</li>
          <li>Seasonal adjustment: cold weather can lower pressure by around 10%.</li>
        </ul>
      </div>

      <div>
        <h3 class="text-sm font-semibold text-slate-100">
          5. Risks &amp; warnings
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>Too high: harsh ride, potential rim damage.</li>
          <li>Too low: pinch flats and poor handling.</li>
          <li>Neglecting checks: gradual air loss reduces performance and safety.</li>
        </ul>
      </div>
    </div>

    <div
      class="mt-3 rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
    >
      <h3 class="mb-3 text-sm font-semibold text-slate-100">
        Perfect pressure ensures optimum grip and stability
      </h3>

      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <div
            class="mb-2 rounded bg-black px-3 py-1.5 text-xs font-semibold uppercase tracking-wide text-slate-100"
          >
            High tire pressure
          </div>
          <ul class="sizecharts-section__list text-slate-100">
            <li>+ Stability</li>
            <li>+ Puncture protection</li>
            <li>- Less grip</li>
            <li>- Uncomfortable</li>
            <li>- More rolling resistance on rough terrain</li>
          </ul>
        </div>

        <div>
          <div
            class="mb-2 rounded bg-black px-3 py-1.5 text-xs font-semibold uppercase tracking-wide text-slate-100"
          >
            Low tire pressure
          </div>
          <ul class="sizecharts-section__list text-slate-100">
            <li>+ More grip</li>
            <li>+ Better off-road rolling performance</li>
            <li>+ Comfortable</li>
            <li>- Less stability</li>
            <li>- Less puncture protection</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const emit = defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>
