<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Wheelset Buyers Guide</h2>

    <div class="wheelset-page">
      <div class="wheelset-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="wheelset-tabs__item"
          :class="{ 'wheelset-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Safety instructions -->
      <section
        v-show="activeTab === 'safety-instructions'"
        id="safety-instructions"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Safety instructions</h3>
      </section>

      <!-- Mullet wheelsets -->
      <section
        v-show="activeTab === 'mullet-wheelsets'"
        id="mullet-wheelsets"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Mullet wheelsets</h3>
      </section>

      <!-- Sample assembly -->
      <section
        v-show="activeTab === 'sample-assembly'"
        id="sample-assembly"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Sample assembly</h3>
      </section>

      <!-- Mixed rim -->
      <section
        v-show="activeTab === 'mixed-rim'"
        id="mixed-rim"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Mixed rim</h3>
      </section>

      <!-- Appearance Logo -->
      <section
        v-show="activeTab === 'appearance-logo'"
        id="appearance-logo"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Appearance Logo</h3>
        <ul class="wheelset-section__list">
          <li>
            <strong>Customized decals</strong>
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
            <strong>Laser engraving: Eco-friendly personalization</strong>
            <p>
              Staying true to our sustainability mission by substantially reducing our plastic
              waste, we’re ecstatic about our new laser engraved graphic option. Our engraving
              machine delivers precision designs in a sleek light gray color, all while preserving
              the structural integrity of your carbon rim and reducing the need for plastic/vinyl
              graphics.
            </p>
          </li>
        </ul>
        <div class="wheelset-appearance-images">
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO.webp"
              alt="Carbon rim with customized laser engraved logo detail"
              class="wheelset-appearance-images__image"
            />
          </div>
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO1.webp"
              alt="Close-up of carbon rim laser engraving process"
              class="wheelset-appearance-images__image"
            />
          </div>
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO2.webp"
              alt="Different examples of carbon rim laser engraved graphics"
              class="wheelset-appearance-images__image"
            />
          </div>
        </div>
      </section>

      <!-- Optional -->
      <section
        v-show="activeTab === 'optional'"
        id="optional"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Optional</h3>
      </section>

      <!-- FAQ Section - 放在所有 tab 内容之后 -->
      <section class="wheelset-section wheelset-faq">
        <PageFaq 
          page-id="guides-wheelset-buyers"
          theme="dark"
          :show-categories="true"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Wheelset Buyers Guide',
})

type WheelsetTabId =
  | 'safety-instructions'
  | 'mullet-wheelsets'
  | 'sample-assembly'
  | 'mixed-rim'
  | 'appearance-logo'
  | 'optional'

const tabs: { id: WheelsetTabId; label: string }[] = [
  { id: 'safety-instructions', label: 'Safety instructions' },
  { id: 'mullet-wheelsets', label: 'Mullet wheelsets' },
  { id: 'sample-assembly', label: 'Sample assembly' },
  { id: 'mixed-rim', label: 'Mixed rim' },
  { id: 'appearance-logo', label: 'Appearance Logo' },
  { id: 'optional', label: 'Optional' },
]

const activeTab = ref<WheelsetTabId>('safety-instructions')

const setActiveTab = (id: WheelsetTabId) => {
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

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.wheelset-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.wheelset-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.wheelset-tabs::-webkit-scrollbar {
  display: none;
}

.wheelset-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow:
    0 3px 9px -6px rgba(0, 0, 0, 0.9),
    0 0 9px rgba(0, 0, 0, 0.85);
}

.wheelset-tabs__item:active {
  transform: scale(0.96);
}

.wheelset-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.wheelset-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .wheelset-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}


.wheelset-section {
  margin-top: 0.75rem;
}

.wheelset-section__title {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.wheelset-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.88rem;
  color: rgba(148, 163, 184, 0.9);
}

.wheelset-section__list ul {
  margin-top: 0.15rem;
  padding-left: 1.1rem;
  list-style-type: disc;
}

.wheelset-section__list li + li {
  margin-top: 0.25rem;
}

.wheelset-appearance-images {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.wheelset-appearance-images__item {
  flex: 1 1 260px;
  max-width: 100%;
  border-radius: 0.75rem;
  overflow: hidden;
  background: rgba(15, 23, 42, 0.9);
}

.wheelset-appearance-images__image {
  width: 100%;
  height: auto;
  display: block;
  aspect-ratio: 2 / 1;
  object-fit: cover;
}

@media (min-width: 768px) {
  .wheelset-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .wheelset-tabs {
    justify-content: flex-start;
  }
}
</style>
