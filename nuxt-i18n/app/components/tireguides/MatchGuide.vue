<template>
  <div class="match-guide">
    <h2 class="sizecharts-section__title">Tire and Frame Clearance Guide</h2>
    <p class="sizecharts-section__intro">
      With our particularly wide tires, the question often arises as to whether the tires will still fit in the frame. Please
      understand that due to the large number of bicycle models, we cannot check all frames for compatibility with the various
      tires. Below we provide the exact diameters and widths of our extra-wide tires. You can use this information to check
      whether the installation dimensions of your frame offer enough space for the desired tire.
    </p>

    <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/match/schwalbe-tire-fit-frame1.webp"
        alt="Frame clearance overview 1 for checking whether wide tires fit in the frame"
        :zoomOnClick="true"
      />
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/match/schwalbe-tire-fit-frame2.webp"
        alt="Frame clearance overview 2 showing dimensions for wide tires"
        :zoomOnClick="true"
      />
      <GuideImage
        class="sizecharts-installation-images__item md:col-span-2"
        src="/public/tiresizecharts/match/schwalbe-tire-fit-frame3.webp"
        alt="Frame clearance overview 3 with detailed extra-wide tire measurements"
        :zoomOnClick="true"
      />
    </div>

    <h3 class="sizecharts-section__subheading mt-4">
      WHAT IS THE EXACT CIRCUMFERENCE OF MY TIRE【Schwalbe Tire】?
    </h3>
    <p class="sizecharts-section__intro">
      Exact tire circumferences are often required for precise programming of the bike computer. The wheel circumference varies
      depending on the inner rim width, puncture protection in the tire, air pressure and weight load. For this reason, we cannot
      specify the exact wheel circumferences. For precise programming of a wheel computer, we recommend a simple rolling test with
      the rider in the saddle: Align the valve from the front wheel at the bottom 6 o’clock position, make a mark on the floor or
      lay the end of a tape measure at this point, roll the bike forward in as straight a line as possible until the front has done
      one full rotation and the valve is once again the 6 o’clock position. Then take the distance measurement directly from the
      tape measure or mark the floor at this point and measure between the two marks. This distance will give you a fairly precise
      rolling circumference for the front wheel which can be entered into your computer.
    </p>

    <div class="mt-3">
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/match/exact-circumference-of-tire.webp"
        alt="Example chart for determining the exact circumference of a Schwalbe tire for bike computer setup"
        :zoomOnClick="true"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
</script>
