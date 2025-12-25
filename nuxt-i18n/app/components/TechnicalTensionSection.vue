<template>
  <div>
    <h3 class="sizecharts-section__subheading text-sky-300 font-semibold">
      Wheel Spoke Tension and Rider Weight/Usage (Workshop Standard)
    </h3>

    <h3 class="sizecharts-section__subheading mt-2">
      <span class="text-amber-300 font-semibold">1. Effect of Rider Weight</span>
    </h3>
    <ul class="sizecharts-section__list">
      <li>Static load: A 75 kg rider increases upper spoke tension by ~15%, while lower spokes lose tension.</li>
      <li>Dynamic riding: Pedaling raises drive-side tension and lowers braking-side tension; distribution keeps changing.</li>
      <li>Heavier riders: Need higher overall tension and thicker spokes to avoid deformation or stiffness loss.</li>
    </ul>

    <h3 class="sizecharts-section__subheading mt-4">
      <span class="text-amber-300 font-semibold">2. Differences by Usage</span>
    </h3>
    <ul class="sizecharts-section__list">
      <li>
        <div>{Road racing}</div>
        <div>Focus on light weight and stiffness.</div>
        <div>Drive-side tension typically 120–130 kgf, non-drive side lower, with a ratio around 10:6–10:7.</div>
      </li>
      <li>
        <div>{Mountain biking}</div>
        <div>Emphasis on durability and shock absorption.</div>
        <div>Tension is set slightly lower to reduce risk of breakage.</div>
        <div>Rim offset (about 2.6–3 mm) is often used to balance spoke tension between sides.</div>
      </li>
      <li>
        <div>{Touring/commuting}</div>
        <div>Prioritize stability and long-term durability.</div>
        <div>Tension is kept moderate, with thicker spokes and a more balanced tension ratio.</div>
      </li>
    </ul>

    <h3 class="sizecharts-section__subheading mt-4">
      <span class="text-amber-300 font-semibold">3. Design and Building Considerations</span>
    </h3>
    <ul class="sizecharts-section__list">
      <li>
        <div>{Rim geometry}</div>
        <div>Offset designs improve tension balance between drive and non-drive sides.</div>
      </li>
      <li>
        <div>{Spoke gauge}</div>
        <div>14G (2.0 mm) is stronger than 15G (1.8 mm), recommended for heavier riders or off-road use.</div>
      </li>
      <li>
        <div>{Tension vs. stiffness}</div>
        <div>Too low → wheel feels loose, lacks stiffness.</div>
        <div>Too high → risk of metal fatigue or reduced rim stiffness, especially in lightweight carbon rims.</div>
      </li>
    </ul>

    <div class="mt-4 overflow-x-auto">
      <table class="min-w-full border-separate border-spacing-0 rounded-xl overflow-hidden bg-slate-900/70 text-slate-100">
        <thead>
          <tr class="bg-slate-800/90 text-sm font-semibold text-slate-50">
            <th class="px-4 py-3 text-left border-b border-slate-700">Rider weight range</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">Road racing recommended tension (kgf)</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">MTB recommended tension (kgf)</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">Touring/commute recommended tension (kgf)</th>
          </tr>
        </thead>
        <tbody class="text-sm">
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3">≤ 60 kg</td>
            <td class="px-4 py-3">115–120</td>
            <td class="px-4 py-3">95–105</td>
            <td class="px-4 py-3">90–100</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3">60–80 kg</td>
            <td class="px-4 py-3">120–125</td>
            <td class="px-4 py-3">100–115</td>
            <td class="px-4 py-3">95–105</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3">80–100 kg</td>
            <td class="px-4 py-3">125–130</td>
            <td class="px-4 py-3">110–120</td>
            <td class="px-4 py-3">100–110</td>
          </tr>
          <tr>
            <td class="px-4 py-3">≥ 100 kg</td>
            <td class="px-4 py-3">130–135</td>
            <td class="px-4 py-3">115–125</td>
            <td class="px-4 py-3">105–115</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div
      class="mt-4 rounded-lg border border-amber-300/80 bg-slate-900/80 px-4 py-3 text-sm leading-relaxed text-amber-100"
    >
      The above is a reference value, which is based on the geometrically
      excellent tension ratio of the hub at 10:7 or better and the value
      after the stress is completely released. The actual situation may
      depend on the number of spokes and spoke type of each hub (for
      example, flat spokes will be relatively weak laterally and need to be
      moderately increased).
    </div>

    <h3 class="sizecharts-section__subheading mt-4">
      Risks and Precautions
    </h3>
    <ul class="sizecharts-section__list">
      <li>
        <div>{Excessive tension}</div>
        <div>Can cause rim deformation or fatigue cracks.</div>
      </li>
      <li>
        <div>{Insufficient tension}</div>
        <div>Leads to loose spokes, wheel wobble, and safety issues.</div>
      </li>
      <li>
        <div>{Weight differences}</div>
        <div>Heavy riders need higher tension and stronger spokes; lighter riders can use lighter setups.</div>
      </li>
    </ul>

    <h3 class="sizecharts-section__subheading mt-4">Key Notes</h3>
    <ul class="sizecharts-section__list">
      <li>
        <div>{Non-drive side tension}</div>
        <div>Usually 20–40% lower than drive side, depending on hub flange spacing and rim geometry.</div>
      </li>
      <li>
        <div>{Carbon vs. alloy rims}</div>
        <div>Carbon rims can handle higher tension (up to 130–135 kgf). Alloy rims should be limited to 120–125 kgf.</div>
      </li>
      <li>
        <div>{Spoke choice}</div>
        <div>Light riders can use 15G spokes; heavy riders or off-road use should stick to 14G.</div>
      </li>
      <li>
        <div>{Usage summary}</div>
        <div>Road racing = stiffness and responsiveness; mountain biking = durability and cushioning; touring = stability and balance.</div>
      </li>
    </ul>

    <h3 class="sizecharts-section__subheading mt-4">Our Workshop Process (Practical Standard)</h3>
    <ul class="sizecharts-section__list">
      <li>Weight classification → determine baseline tension range.</li>
      <li>
        <div>{Usage adjustment}</div>
        <div>Road racing → higher tension for stiffness.</div>
        <div>Mountain biking → slightly lower tension for durability.</div>
        <div>Touring → balanced tension for stability.</div>
      </li>
      <li>
        <div>{Material limits}</div>
        <div>Carbon rims → allow higher tension (up to 135 kgf).</div>
        <div>Alloy rims → capped at 120–125 kgf.</div>
      </li>
      <li>Spoke selection → match spoke gauge to rider weight and usage.</li>
      <li>Final build → ensure proper drive/non-drive tension ratio (about 10:6–10:7; some hub geometries may only be 10:4) and use rim offset to balance.</li>
    </ul>

    <h3 class="sizecharts-section__subheading mt-4">Conclusion</h3>
    <ul class="sizecharts-section__list">
      <li>Rider weight sets the baseline tension.</li>
      <li>Usage defines the adjustment.</li>
      <li>Rim material sets the upper limit.</li>
      <li>Spoke gauge ensures durability and safety.</li>
      <li>Following this flow balances performance, reliability, and safety for each rider.</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
// no extra imports needed
</script>
