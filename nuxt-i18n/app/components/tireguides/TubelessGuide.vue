<template>
  <div class="tubeless-guide">
    <h2 class="sizecharts-section__title">Tubeless Tires</h2>

    <TubelessProducts />

    <div class="sizecharts-installation-images sizecharts-installation-images--tubeless">
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires1.webp"
        alt="Comparison of tubeless, inner tube and tubular road bike tires with their construction and basic features"
        :zoomOnClick="true"
        caption="The overview above shows the basic construction differences between tubeless, inner tube and tubular tires. Understanding how each system seals air and interfaces with the rim helps you choose the right setup for your riding style and maintenance preferences."
      />
      <GuideImage
        class="sizecharts-installation-images__item"
        src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires2.webp"
        alt="Pros and cons chart comparing tubeless, inner tube and tubular tires for puncture protection, rolling resistance and comfort"
        :zoomOnClick="true"
        caption="This comparison chart summarizes the main pros and cons of tubeless vs inner tube vs tubular tires, focusing on puncture protection, rolling resistance, comfort and ease of installation. Use it as a quick reference when deciding which tire system best matches your daily rides and performance goals."
      />
    </div>

    <h3 class="sizecharts-section__subheading">What is an inner tube tire?</h3>
    <p class="sizecharts-section__intro">
      Tires with inner tubes are like the bicycle tires we rode when we were children. The tires are composed of
      &quot;inner tube + outer tire&quot;. There is a frame on the outside and a pneumatic tire in the inner
      cavity. It can be applied to any type of bicycle.
    </p>

    <h3 class="sizecharts-section__subheading">What is tubeless?</h3>
    <p class="sizecharts-section__intro">
      As the name suggests, tubeless tubes are clincher tires that are inflated on a rim without a tube. Unlike
      tubeless tubes, which need to keep air inside the tire, the tubeless tire, rim, and valve create a closed
      air-tight space, just like a car or motorcycle tire.
    </p>
    <p class="sizecharts-section__intro">
      The airtight chamber of a tubeless tire is made from a specialized tire using a specially made (usually
      carbon fiber) bead and compatible rim, which locks into the rim and creates an airtight seal that maintains
      pressure.
    </p>
    <p class="guide-section__cta-wrapper">
      <button
        type="button"
        class="sizecharts-brand-button"
        @click="$emit('changeTab', 'choose')"
      >
        How to choose and specific specifications
      </button>
    </p>

    <h3 class="sizecharts-section__subheading">
      Am I suitable for tubeless or tubeless use?
    </h3>
    <p class="sizecharts-section__intro">
      It depends on your usage needs and habits. If you are doing daily commuting or relaxing holiday riding,
      tubeless tires are enough; but if you often ride on soft and slippery roads, or require more advanced,
      high-speed riding enjoyment, then tubeless tires are a good choice.
    </p>

    <h3 class="sizecharts-section__subheading">How to install tubeless wheels?</h3>
    <p class="guide-section__cta-wrapper">
      <button
        type="button"
        class="sizecharts-brand-button"
        @click="$emit('changeTab', 'installation')"
      >
        View graphic installation details
      </button>
    </p>

    <h3 class="sizecharts-section__subheading">
      The difference between tubeless and tubeless systems
    </h3>
    <p class="sizecharts-section__intro">
      The biggest difference between the two is that tubeless tires do not require inner tubes, while both clincher
      and tubular tires require the support of an inner tube structure. When some small punctures occur, the tire
      sealing fluid in the tubeless tire will fill it up on its own. However, once the puncture is larger, it
      cannot be repaired. At this time, there is no need to worry. You can plug an inner tube in to solve the
      problem. At the same time, since there is no inner tube, there is no risk of snake bite and puncture in
      tubeless tires, so the chance of puncture will be greatly reduced.
    </p>
    <p class="sizecharts-section__intro">
      Compared with ordinary tires, tubeless tires are wider and can hold more air inside. The safe tire pressure
      required to maintain the shape is relatively small, so the tire pressure of tubeless tires does not need to
      be above 100psi. The best rolling resistance and riding comfort can be obtained at around 80psi.
    </p>

    <p class="guide-section__cta-wrapper">
      <button
        type="button"
        class="sizecharts-brand-button"
        @click="$emit('changeTab', 'choose')"
      >
        Check the tire pressure limits of different models of tubeless tires
      </button>
    </p>
  </div>
</template>

<script setup lang="ts">
import TubelessProducts from '~/components/TubelessProducts.vue'
import GuideImage from '~/components/GuideImage.vue'

// Match SizeChartsTabId from parent
type GuideTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

defineEmits<{
  (e: 'changeTab', tab: GuideTabId): void
}>()
</script>
