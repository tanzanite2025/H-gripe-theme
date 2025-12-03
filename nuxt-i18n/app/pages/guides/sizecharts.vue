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
        <h3 class="sizecharts-section__title">Tubeless tires</h3>
        <p class="sizecharts-section__intro">
          Notes about tubeless tire sizing, compatibility and typical use cases.
        </p>
        <ul class="sizecharts-section__list">
          <li>Basic explanation of tubeless-ready rims and tires.</li>
          <li>Common ETRTO sizes that are typically available as tubeless.</li>
          <li>Recommended pressure ranges and safety notes.</li>
        </ul>

        <TubelessProducts />
      </section>

      <!-- Installation -->
      <section
        v-show="activeTab === 'installation'"
        id="installation"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title">Installation</h3>
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

      <!-- Commonly used -->
      <section
        v-show="activeTab === 'common'"
        id="common"
        class="sizecharts-section"
      >
        <h3 class="sizecharts-section__title">Commonly used</h3>
        <p class="sizecharts-section__intro">
          Placeholder lists of the most commonly used tire sizes for different wheel diameters.
        </p>
        <ul class="sizecharts-section__list">
          <li>Typical 700c / 29&quot; combinations (e.g. 700×25C, 700×28C, 29×2.2&quot;).</li>
          <li>Typical 27.5&quot; / 650B options for gravel and MTB.</li>
          <li>Popular city / utility tire sizes and their use cases.</li>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import TubelessProducts from '~/components/TubelessProducts.vue'

type SizeChartsTabId = 'tubeless' | 'installation' | 'choose' | 'common' | 'rims'

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
  { id: 'common', label: 'Commonly used' },
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

@media (max-width: 768px) {
  .sizecharts-tabs {
    justify-content: center;
  }
}
</style>
