<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Tire Size Charts</h2>
    <p class="products-page__intro products-page__intro--sr-only">
      Reference charts for common tire and rim sizes. Detailed data will be added here later.
    </p>

    <div class="sizecharts-page">
      <div class="sizecharts-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="sizecharts-tabs__item"
          :class="{ 'sizecharts-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Tubeless tires -->
      <section
        v-show="activeTab === 'tubeless'"
        id="tubeless"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title sizecharts-section__title--sr-only">Tubeless tires</h3>

        <TubelessProducts />

        <div class="sizecharts-installation-images sizecharts-installation-images--tubeless">
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires1.webp"
              alt="Comparison of tubeless, inner tube and tubular tires 1"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires2.webp"
              alt="Comparison of tubeless, inner tube and tubular tires 2"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
        </div>

        <p class="sizecharts-section__subheading">What is an inner tube tire?</p>
        <p class="sizecharts-section__intro">
          Tires with inner tubes are like the bicycle tires we rode when we were children. The tires are composed of
          &quot;inner tube + outer tire&quot;. There is a frame on the outside and a pneumatic tire in the inner
          cavity. It can be applied to any type of bicycle.
        </p>

        <p class="sizecharts-section__subheading">What is tubeless?</p>
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

        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('choose')"
        >
          How to choose and specific specifications
        </button>

        <p class="sizecharts-section__subheading">
          Am I suitable for tubeless or tubeless use?
        </p>
        <p class="sizecharts-section__intro">
          It depends on your usage needs and habits. If you are doing daily commuting or relaxing holiday riding,
          tubeless tires are enough; but if you often ride on soft and slippery roads, or require more advanced,
          high-speed riding enjoyment, then tubeless tires are a good choice.
        </p>

        <p class="sizecharts-section__subheading">How to install tubeless wheels?</p>
        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('installation')"
        >
          View graphic installation details
        </button>

        <p class="sizecharts-section__subheading">
          The difference between tubeless and tubeless systems
        </p>
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

        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('choose')"
        >
          Check the tire pressure limits of different models of tubeless tires
        </button>
      </section>

      <!-- Installation -->
      <section
        v-show="activeTab === 'installation'"
        id="installation"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title sizecharts-section__title--sr-only">Installation</h3>
        <p class="sizecharts-section__intro">
          <strong>Before installation:</strong>
        </p>
        <ul class="sizecharts-section__list">
          <li>
            Before starting the first step, one thing to know is, what is the difference between using tubeless tires and clincher tires? The biggest difference is of course the lack of inner tubes.
          </li>
        </ul>
        <p class="sizecharts-section__intro">
          Not having a tube means the entire tubeless system is airtight and inflated differently.
        </p>
        <p class="sizecharts-section__intro">
          <strong>So focus on this first</strong>
        </p>
        <ul class="sizecharts-section__list">
          <li>
            The air tightness of clincher tires relies on the inner tube, while the air tightness of tubeless tires relies on the tubeless tire pad.
          </li>
        </ul>
        <p class="sizecharts-section__intro">
          The overall process is divided into three steps. The first step is to install tubeless tires. The second step is to fill the tire with tire sealant. The third step is to cheer up.
        </p>
        <p class="sizecharts-section__intro">
          Compared with ordinary clincher tires, there is one less step to insert the inner tube. But there are some differences before installation.
        </p>
        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('tubeless')"
        >
          See details of the vacuum accessories you need
        </button>

        <div class="sizecharts-installation-images">
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/Bicycle tubeless tire pad.webp"
              alt="Bicycle tubeless tire pad"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/Bicycle vacuum self-replenishing fluid.webp"
              alt="Bicycle vacuum self-replenishing fluid"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/Bicycle vacuum valve.webp"
              alt="Bicycle vacuum valve"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
        </div>

        <ul class="sizecharts-section__list">
          <li>So you need to carefully attach the tubeless tire pad to the rim first.</li>
          <li>Carefully attach the tubeless tire pad to the tire bed of the wheel set to seal the spoke holes.</li>
          <li>Overlap the interface by about 10CM, and then carefully press the vacuum tire pad to ensure a tight fit between the tire pad and the rim.</li>
          <li>
            Then use an awl or utility knife to make a hole at the air nozzle position and install the vacuum air nozzle. If it is a holeless vacuum ring design similar to CP wheels, you can skip the step of installing tire pads and install the valve directly.
          </li>
        </ul>

        <div class="sizecharts-installation-images">
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/tanzanite-Bicycle-tubeless-tire-pads.webp"
              alt="Tanzanite bicycle tubeless tire pads"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/tanzanite-howtoopenthe vacuumnozzle.webp"
              alt="Tanzanite how to open the vacuum nozzle"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/tanzanite-tubeless-tirepad-interface-overlaps.webp"
              alt="Tanzanite tubeless tire pad interface overlaps"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
        </div>

        <ul class="sizecharts-section__list">
          <li>
            This is the second difference between vacuum and opening, the way of inflating. The valve of clincher tires is installed on the inner tube, while the valve of tubeless tires is installed on the rim.
          </li>
          <li>
            When installing the vacuum valve, make sure that the root of the vacuum valve is tightly connected to the rim. The fixing screw of the valve only needs to be tightened by hand. Do not use a wrench or other tools to tighten it hard enough to cause miracles!
          </li>
        </ul>

        <p class="sizecharts-section__subheading">1. Install tubeless tires</p>
        <ul class="sizecharts-section__list">
          <li>
            Then install the tire on the rim just like a clincher tire. Generally speaking, tubeless tires are tighter and more difficult to install than clincher tires of the same type, requiring a certain amount of strength and a lot of patience. Here's a trick: Push all the bead into the concave center of the rim to make the tire easier to install.
          </li>
          <li>
            After all the tire beads on both sides are installed, use your hands to "massage" the tire from the opposite side of the valve to ensure that every position of the tire is well embedded in the rim. At the same time, push all the tire beads on both sides into the grooves of the rim, which will make it easier to inflate. Finally, carefully check the position of the valve to ensure that the bead completely covers the valve.
          </li>
          <li>
            Then it’s time to inflate. It’s also difficult to inflate a tubeless tire for the first time because the tire is not stuck in place at this time and is leaking air. You need to use an air pump with a large air intake to inflate quickly or an air pump with an air storage function. Or a CO2 cylinder can also be used. The principle is to allow a large amount of air to quickly enter the tire and push the tire into place. When I was decorating my home, I used an ordinary vertical air pump that worked wonders.
          </li>
          <li>
            After inflating, check whether the bead line is completely exposed on the rim. If the tire is not easy to place, sometimes it is necessary to lubricate the bead line.
          </li>
          <li>
            Use soapy water, dishwashing liquid, and special lubricant to apply to the bead. Or SCHWABLE has special tire fluid.
          </li>
        </ul>

        <div class="sizecharts-installation-images">
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/3/first-step-to-install-tubelesstires.webp"
              alt="First step to install tubeless tires"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp"
              alt="Tubeless tire installation diagram 2"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/3/Picture-of-whether-the-vacuum-nozzle-is-installed-properly.webp"
              alt="Picture of whether the vacuum nozzle is installed properly"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
        </div>

        <p class="sizecharts-section__subheading">2. Fill the tire with tire repair fluid.</p>
        <ul class="sizecharts-section__list">
          <li>Use special tools to remove the core of the valve.</li>
          <li>
            Then pour the self-hydrating fluid into the wheel through the valve. For the amount of self-replenishing solution, refer to the instruction manual of the self-replenishing solution you purchased. The consumption of a road car is about 30-60ML per wheel.
          </li>
        </ul>

        <div class="sizecharts-installation-images">
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant1.webp"
              alt="Pour vacuum tire sealant 1"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant2.webp"
              alt="Pour vacuum tire sealant 2"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
          <figure class="sizecharts-installation-images__item">
            <img
              src="/public/tiresizecharts/installation/4/Separate-vacuum-nozzle-tool.webp"
              alt="Separate vacuum nozzle tool"
              class="sizecharts-installation-images__img"
              loading="lazy"
            />
          </figure>
        </div>

        <p class="sizecharts-section__subheading">3. Cheer up</p>
        <ul class="sizecharts-section__list">
          <li>
            Inflate the same as ordinary clincher tires. Because the tires have been installed in place in the previous step. Pump up the air, hold your wheel up, down, left, and right, and shake the tire sealant evenly, and you're done.
          </li>
        </ul>
      </section>

      <!-- How to choose -->
      <section
        v-show="activeTab === 'choose'"
        id="choose"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title">How to choose</h3>
        <p class="sizecharts-section__intro">
          Simple guidelines for selecting tire size based on usage and frame clearance.
        </p>
        <ul class="sizecharts-section__list">
          <li>Match tire width to riding style: road, gravel, XC, trail, city.</li>
          <li>Consider rider weight, terrain and comfort preferences.</li>
          <li>Use the size charts to see compatible width ranges per rim.</li>
        </ul>
      </section>

      <!-- Suitable for rims -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title">Suitable for rims</h3>
        <p class="sizecharts-section__intro">
          How to read which tire widths are suitable for a given rim internal width.
        </p>
        <ul class="sizecharts-section__list">
          <li>Explain the relationship between rim internal width and tire width.</li>
          <li>Show recommended minimum and maximum tire widths per rim width range.</li>
          <li>Link back to the detailed size charts for exact numeric values.</li>
        </ul>
      </section>

      <div class="sizecharts-feedback">
        <UserFeedbackThread
          threadKey="guides-sizecharts"
          title="Share your feedback about this Tire Size Charts guide"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import TubelessProducts from '~/components/TubelessProducts.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

type SizeChartsTabId = 'tubeless' | 'installation' | 'choose' | 'rims'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Tire Size Charts',
})

const tabs: { id: SizeChartsTabId; label: string }[] = [
  { id: 'tubeless', label: 'Tubeless tires' },
  { id: 'installation', label: 'Installation' },
  { id: 'choose', label: 'How to choose' },
  { id: 'rims', label: 'Suitable for rims' },
]

const activeTab = ref<SizeChartsTabId>('tubeless')

const setActiveTab = (id: SizeChartsTabId) => {
  activeTab.value = id
}
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.products-page__title--sr-only {
	position: absolute;
	width: 1px;
	height: 1px;
	padding: 0;
	margin: -1px;
	overflow: hidden;
	clip: rect(0, 0, 0, 0);
	white-space: nowrap;
	border: 0;
}

.products-page__intro--sr-only {
	position: absolute;
	width: 1px;
	height: 1px;
	padding: 0;
	margin: -1px;
	overflow: hidden;
	clip: rect(0, 0, 0, 0);
	white-space: nowrap;
	border: 0;
}

.sizecharts-section__title--sr-only {
	position: absolute;
	width: 1px;
	height: 1px;
	padding: 0;
	margin: -1px;
	overflow: hidden;
	clip: rect(0, 0, 0, 0);
	white-space: nowrap;
	border: 0;
}

.sizecharts-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.sizecharts-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  padding: 0.3rem;
  border-radius: 9999px;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid rgba(148, 163, 184, 0.4);
  margin-bottom: 1rem;
  max-width: 100%;
  box-sizing: border-box;
  overflow-x: hidden;
}

.sizecharts-tabs__item {
  border: none;
  border-radius: 9999px;
  padding: 0.25rem 0.75rem;
  font-size: 0.8rem;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.9);
  background: transparent;
  cursor: pointer;
}

.sizecharts-tabs__item--active {
  background: rgba(56, 189, 248, 0.15);
  color: #e5e7eb;
  border: 1px solid rgba(56, 189, 248, 0.8);
}

.sizecharts-section {
  margin-top: 0.75rem;
}

.sizecharts-section__title {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.sizecharts-section__intro {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.9);
}

.sizecharts-section__subheading {
  margin: 0.85rem 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #38bdf8;
}

.sizecharts-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.88rem;
  color: rgba(148, 163, 184, 0.9);
}

.sizecharts-section__list li + li {
  margin-top: 0.25rem;
}

.sizecharts-section__list--numbered {
  list-style-type: decimal;
}

.sizecharts-brand-button {
  margin-left: 0.5rem;
  margin-top: 0.5rem;
  padding: 0.25rem 0.8rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  background-image: linear-gradient(
    135deg,
    rgba(56, 189, 248, 0.9),
    rgba(59, 130, 246, 0.95)
  );
  color: #0b1020;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.sizecharts-feedback {
  margin-top: 2.5rem;
}

.sizecharts-installation-images {
  margin-top: 0.75rem;
  display: flex;
  gap: 0.5rem;
}

.sizecharts-installation-images__item {
  flex: 1 1 0;
}

.sizecharts-installation-images__img {
  width: 100%;
  height: 140px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid rgba(30, 64, 175, 0.7);
}

.sizecharts-installation-images--tubeless .sizecharts-installation-images__img {
  object-fit: contain;
}

@media (max-width: 768px) {
  .sizecharts-tabs {
    justify-content: center;
  }
}
</style>
