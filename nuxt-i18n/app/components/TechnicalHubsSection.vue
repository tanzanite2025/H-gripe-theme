<template>
  <div>
    <p class="sizecharts-section__intro text-center">
      Bicycle Hub Key Specifications
    </p>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-2"
    >
      1. Spoke Type
    </h3>
    <div class="mt-2 text-sm text-slate-100">
      <div class="text-center font-semibold">J-bend (Elbowed spokes)</div>
      <ul class="sizecharts-section__list mt-1">
        <li>The spoke head is bent to fit into the hub flange holes.</li>
        <li>Advantages: Highly compatible, easy to service.</li>
        <li>Disadvantages: Slightly lower spoke tension, more traditional appearance.</li>
      </ul>

      <div class="mt-3 text-center font-semibold">Straight-pull</div>
      <ul class="sizecharts-section__list mt-1">
        <li>The spokes are inserted directly into the hub without bending.</li>
        <li>Common hole count: 24H (arranged in 1:1 or 2:1 patterns).</li>
        <li>Advantages: Higher spoke tension, cleaner aesthetics.</li>
        <li>Disadvantages: Limited compatibility when replacing spokes, less convenient for long-distance maintenance.</li>
      </ul>
    </div>

    <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
      <GuideImage
        src="/public/technical/hubs/hubs/tanzanite-dt240-6bolt-jbend-hub.webp"
        alt="Tanzanite DT240 6-bolt J-bend hub"
        :zoomOnClick="true"
        caption="tanzanite-dt240-6bolt-jbend-hub"
      />
      <GuideImage
        src="/public/technical/hubs/hubs/tanzanite-dt240-6bolt-straightpull-hub.webp"
        alt="Tanzanite DT240 6-bolt straight-pull hub"
        :zoomOnClick="true"
        caption="tanzanite-dt240-6bolt-straightpull-hub"
      />
      <GuideImage
        src="/public/technical/hubs/hubs/tanzanite-dt240-centerlock-jbend-hub.webp"
        alt="Tanzanite DT240 Center Lock J-bend hub"
        :zoomOnClick="true"
        caption="tanzanite-dt240-centerlock-jbend-hub"
      />
      <GuideImage
        src="/public/technical/hubs/hubs/tanzanite-dt240-centerlock-straightpull-hub.webp"
        alt="Tanzanite DT240 Center Lock straight-pull hub"
        :zoomOnClick="true"
        caption="tanzanite-dt240-centerlock-straightpull-hub"
      />
    </div>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-6"
    >
      2. Hole Count
    </h3>
    <div class="mt-2 text-sm text-slate-100 text-center">
      <div>
        <span class="font-semibold text-amber-300">Common specifications:</span>
      </div>
      <div class="mt-1">16H, 18H, 21H, 24H, 28H, 32H, 36H, etc.</div>
      <div class="mt-3">
        <span class="font-semibold text-amber-300">Selection logic:</span>
      </div>
    </div>
    <ul class="sizecharts-section__list mt-1 text-sm text-slate-100">
      <li>
        Fewer holes (e.g., 24H): Lightweight, suitable for racing. Similarly, 16H and 18H are generally
        used for riders with lighter body weight and lower riding intensity, typically on road bikes.
        Mountain bikes rarely use such a low spoke count.
      </li>
      <li>
        More holes (e.g., 36H): Higher strength, suitable for touring or heavy load requirements. Currently,
        MTB hubs are most commonly 28-32H. With appropriate spokes, 36H is more often applied in BMX wheels.
        Common road disc brake hubs are 24H / 28H.
      </li>
    </ul>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-6"
    >
      3. Freehub Body
    </h3>
    <ul class="sizecharts-section__list mt-2 text-sm text-slate-100">
      <li>HG (Shimano HyperGlide): Compatible with 8-11 speed cassettes.</li>
      <li>MS (Micro Spline): MTB-specific, compatible with 12-speed cassettes.</li>
      <li>XD (SRAM XD): Compatible with 11-12 speed cassettes.</li>
      <li>XDR (SRAM XDR): Road-specific, 1.85 mm longer than XD; can be converted to XD with a spacer.</li>
    </ul>

    <p class="guide-section__cta-wrapper mt-3">
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
      >
        How to choose
      </button>
    </p>
    <p class="mt-1 text-center text-xs text-sky-200">
      Click to learn how to choose the right freehub body
    </p>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-6"
    >
      4. Brake Interface
    </h3>
    <ul class="sizecharts-section__list mt-2 text-sm text-slate-100">
      <li>6-bolt: Common, secured with six bolts.</li>
      <li>Center Lock: Fixed with splines and a lockring, faster installation.</li>
    </ul>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-6"
    >
      5. Axle Specifications
    </h3>
    <ul class="sizecharts-section__list mt-2 text-sm text-slate-100">
      <li>
        <div><span class="font-semibold text-amber-300">Front hubs:</span></div>
        <div>
          Φ9×100 mm (quick release), Φ15×100 mm, Φ15×110 mm [BOOST], Φ20×110 mm, etc.
        </div>
      </li>
      <li class="mt-2">
        <div><span class="font-semibold text-amber-300">Rear hubs:</span></div>
        <div>
          Φ9×135 mm, Φ12×142 mm, Φ12×148 mm [BOOST], etc.
        </div>
      </li>
      <li class="mt-2">
        Quick release vs thru-axle: Quick release is lighter and more convenient, while thru-axle provides greater
        stiffness and stability. Currently, the most commonly used standards are MTB BOOST 15×110 mm front and
        12×148 mm rear thru-axles. For road disc brakes, 12×100 mm front and 12×142 mm rear thru-axles are typical.
      </li>
    </ul>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-6"
    >
      6. Ratchet System
    </h3>
    <ul class="sizecharts-section__list mt-2 text-sm text-slate-100">
      <li>
        At present, most ratchet systems on the market are DT-style. As shown in typical illustrations, many modern wheels
        use this structure, and our shop also sells the full range of DT hubs, with upgrade options available.
      </li>
      <li>
        Common tooth counts include 18T, 36T, 54T (and even larger options such as 60T). A higher tooth count provides a
        faster engagement angle, resulting in quicker pedaling response.
      </li>
      <li>
        From a purely physical perspective, with the same radius, more teeth mean each tooth is smaller, which leads to
        faster wear.
      </li>
      <li>
        If you are a high-power rider, it is recommended to choose 36T as a balance between durability and performance.
      </li>
      <li>
        The design intent of these products is always to ensure good stability and longevity; the discussion here is
        relative, comparing 18T, 36T, 54T, or even larger options.
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
</script>
