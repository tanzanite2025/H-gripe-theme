<template>
  <div class="match-guide space-y-8">
    <!-- Card 1: Tire and Frame Clearance -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold text-slate-100 mb-4 flex items-center justify-center gap-2">
        Tire and Frame Clearance Guide
      </h2>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
        With our particularly wide tires, the question often arises as to whether the tires will still fit in the frame. Please
        understand that due to the large number of bicycle models, we cannot check all frames for compatibility with the various
        tires. Below we provide the exact diameters and widths of our extra-wide tires. You can use this information to check
        whether the installation dimensions of your frame offer enough space for the desired tire.
      </p>

      <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
        <GuideImage
          class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
          src="/public/tiresizecharts/match/schwalbe-tire-fit-frame1.webp"
          alt="Frame clearance overview 1 for checking whether wide tires fit in the frame"
          :zoomOnClick="true"
        />
        <GuideImage
          class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
          src="/public/tiresizecharts/match/schwalbe-tire-fit-frame2.webp"
          alt="Frame clearance overview 2 showing dimensions for wide tires"
          :zoomOnClick="true"
        />
        <GuideImage
          class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] md:col-span-2"
          src="/public/tiresizecharts/match/schwalbe-tire-fit-frame3.webp"
          alt="Frame clearance overview 3 with detailed extra-wide tire measurements"
          :zoomOnClick="true"
        />
      </div>
    </div>

    <!-- Card 2: Circumference -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
      <h3 class="text-lg font-bold text-slate-100 mb-4 flex items-center justify-center gap-2 uppercase tracking-wide">
        Exact Circumference Guide
      </h3>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
        Exact tire circumferences are often required for precise programming of the bike computer. The wheel circumference varies
        depending on the inner rim width, puncture protection in the tire, air pressure and weight load. For this reason, we cannot
        specify the exact wheel circumferences.
      </p> 
      <p class="tz-text-secondary text-sm leading-relaxed mb-8 max-w-3xl mx-auto">
        For precise programming of a wheel computer, we recommend a simple rolling test with
        the rider in the saddle: Align the valve from the front wheel at the bottom 6 o’clock position, make a mark on the floor or
        lay the end of a tape measure at this point, roll the bike forward in as straight a line as possible until the front has done
        one full rotation and the valve is once again the 6 o’clock position. Then take the distance measurement directly from the
        tape measure or mark the floor at this point and measure between the two marks.
      </p>

      <div class="mt-3 flex justify-center">
        <div class="w-full max-w-4xl">
           <GuideImage
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
            src="/public/tiresizecharts/match/exact-circumference-of-tire.webp"
            alt="Example chart for determining the exact circumference of a Schwalbe tire for bike computer setup"
            :zoomOnClick="true"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
</script>
