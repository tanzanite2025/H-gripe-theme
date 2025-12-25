<template>
  <div>
    <h3 class="sizecharts-section__subheading">
      Why we recommend laser-engraved logos for most builds
    </h3>
    <p class="sizecharts-section__intro">
      For standard wheelset orders we usually recommend laser-engraved logos because they are fast to produce, eco-friendly and very reliable for shipping schedules. Compared with custom paint or complex decal work, a laser logo can typically shorten the overall lead time by around 5&ndash;7 days.
    </p>
    <p class="guide-section__cta-wrapper">
      <button
        type="button"
        class="wheelset-inline-button inline-flex items-center justify-center px-3 py-1.5 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
        @click="goToAboutAppearance"
      >
        Explore all weave &amp; finish options
      </button>
    </p>
    <ul class="sizecharts-section__list">
      <li>
        <strong style="color: #38bdf8;">Customized decals</strong>
        <p>
          Our skilled graphics department uses the latest technology to design and execute any
          custom decal or rim personalization you can imagine! Our most popular options are
          laser engraving, waterslide decals, or removable laser-cut vinyl stickers. Our team
          will work closely with you to design the perfect size, location, font, and color of
          your custom graphics. Whether you're looking for a bold statement or subtle
          customization, we'll help you achieve the look you want.
        </p>
      </li>
      <li>
        <strong style="color: #38bdf8;">Laser engraving: Eco-friendly personalization</strong>
        <p>
          Staying true to our sustainability mission by substantially reducing our plastic
          waste, we’re ecstatic about our new laser engraved graphic option. Our engraving
          machine delivers precision designs in a sleek light gray color, all while preserving
          the structural integrity of your carbon rim and reducing the need for plastic/vinyl
          graphics.
        </p>
      </li>
    </ul>
    <div class="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      <GuideImage
        src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO.webp"
        alt="Carbon rim with customized laser engraved logo detail"
        :zoomOnClick="true"
        caption="Carbon fiber rim laser engraving process 1"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO1.webp"
        alt="Close-up of carbon rim laser engraving process"
        :zoomOnClick="true"
        caption="Carbon fiber rim laser engraving process 2"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO2.webp"
        alt="Different examples of carbon rim laser engraved graphics"
        :zoomOnClick="true"
        caption="Carbon fiber rim laser engraving effect"
      />
    </div>
    <p class="sizecharts-section__intro mt-3">
      If you are planning a long-term build, OEM project or team order, you can also go beyond the standard logo and choose specific carbon weaves and surface finishes for your rims.
    </p>
    <div class="mt-3 grid gap-3 sm:grid-cols-2">
      <GuideImage
        src="/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp"
        alt="Tanzanite carbon rim surface finish detail 1"
        :zoomOnClick="true"
        caption="Matte carbon rim finish"
      />
      <GuideImage
        src="/company/aboutus/appearance/tanzanite-carbonrims-finish2.webp"
        alt="Tanzanite carbon rim surface finish detail 2"
        :zoomOnClick="true"
        caption="Glossy carbon rim finish"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

const { goToAboutAppearance } = defineProps<{
  goToAboutAppearance: () => void | Promise<void>
}>()
</script>
