<template>
  <div>
    <h1 class="products-page__title products-page__title--sr-only">Tire Guides</h1>
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

      <!-- Tire size (new top-level tab) -->
      <section
        v-show="activeTab === 'size'"
        id="size"
        class="sizecharts-section text-slate-100"
      >
        <h2 class="sizecharts-section__title">Tire size</h2>
        <div class="tire-size-card mt-1 rounded-2xl bg-slate-900/70 p-4 md:p-5 shadow-[4px_4px_18px_rgba(0,0,0,1)]">
          <TireSizeSection @openTireProducts="openTireProductsDrawer" />
        </div>
      </section>

      <!-- Match (tire & rim matching helpers) -->
      <section
        v-show="activeTab === 'match'"
        id="match"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">DOES THE TIRE FIT MY FRAME?</h2>
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
      </section>

      <!-- Tubeless tires -->
      <section
        v-show="activeTab === 'tubeless'"
        id="tubeless"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Tubeless tires</h2>

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
            @click="setActiveTab('choose')"
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
            @click="setActiveTab('installation')"
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
            @click="setActiveTab('choose')"
          >
            Check the tire pressure limits of different models of tubeless tires
          </button>
        </p>
      </section>

      <!-- Installation -->
      <section
        v-show="activeTab === 'installation'"
        id="installation"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Installation</h2>
        <h3 class="sizecharts-section__subheading">Before installation:</h3>
        <p class="sizecharts-section__intro">
          Before starting the first step, one thing to know is, what is the difference between using tubeless tires and clincher tires? The biggest difference is of course the lack of inner tubes.
          <br />
          Not having a tube means the entire tubeless system is airtight and inflated differently.
        </p>
        <h3 class="sizecharts-section__subheading">So focus on this first</h3>
        <p class="sizecharts-section__intro">
          The air tightness of clincher tires relies on the inner tube, while the air tightness of tubeless tires relies on the tubeless tire pad.
          <br />
          The overall process is divided into three steps. The first step is to install tubeless tires. The second step is to fill the tire with tire sealant. The third step is to cheer up.
        </p>
        <p class="sizecharts-section__intro">
          Compared with ordinary clincher tires, there is one less step to insert the inner tube. But there are some differences before installation.
        </p>
        <p class="guide-section__cta-wrapper">
          <button
            type="button"
            class="sizecharts-brand-button"
            @click="setActiveTab('tubeless')"
          >
            See details of the vacuum accessories you need
          </button>
        </p>

        <div class="sizecharts-installation-images sizecharts-installation-images--installation">
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/Bicycle tubeless tire pad.webp"
            alt="Bicycle tubeless tire pad"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/Bicycle vacuum self-replenishing fluid.webp"
            alt="Bicycle vacuum self-replenishing fluid"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/Bicycle vacuum valve.webp"
            alt="Bicycle vacuum valve"
            :zoomOnClick="true"
          />
        </div>

        <ul class="sizecharts-section__list">
          <li>So you need to carefully attach the tubeless tire pad to the rim first.</li>
          <li>Carefully attach the tubeless tire pad to the tire bed of the wheel set to seal the spoke holes.</li>
          <li>Overlap the interface by about 10CM, and then carefully press the vacuum tire pad to ensure a tight fit between the tire pad and the rim.</li>
          <li>
            Then use an awl or utility knife to make a hole at the air nozzle position and install the vacuum air nozzle. If it is a holeless vacuum ring design similar to CP wheels, you can skip the step of installing tire pads and install the valve directly.
          </li>
        </ul>

        <div class="sizecharts-installation-images sizecharts-installation-images--installation">
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/tanzanite-Bicycle-tubeless-tire-pads.webp"
            alt="Tanzanite bicycle tubeless tire pads"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/tanzanite-howtoopenthe vacuumnozzle.webp"
            alt="Tanzanite how to open the vacuum nozzle"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/tanzanite-tubeless-tirepad-interface-overlaps.webp"
            alt="Tanzanite tubeless tire pad interface overlaps"
            :zoomOnClick="true"
          />
        </div>

        <ul class="sizecharts-section__list">
          <li>
            This is the second difference between vacuum and opening, the way of inflating. The valve of clincher tires is installed on the inner tube, while the valve of tubeless tires is installed on the rim.
          </li>
          <li>
            When installing the vacuum valve, make sure that the root of the vacuum valve is tightly connected to the rim. The fixing screw of the valve only needs to be tightened by hand. Do not use a wrench or other tools to tighten it hard enough to cause miracles!
          </li>
        </ul>

        <h3 class="sizecharts-section__subheading">1. Install tubeless tires</h3>
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

        <div class="sizecharts-installation-images sizecharts-installation-images--installation">
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/3/first-step-to-install-tubelesstires.webp"
            alt="First step to install tubeless tires"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp"
            alt="Tubeless tire installation diagram 2"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/3/Picture-of-whether-the-vacuum-nozzle-is-installed-properly.webp"
            alt="Picture of whether the vacuum nozzle is installed properly"
            :zoomOnClick="true"
          />
        </div>

        <h3 class="sizecharts-section__subheading">2. Fill the tire with tire repair fluid.</h3>
        <ul class="sizecharts-section__list">
          <li>Use special tools to remove the core of the valve.</li>
          <li>
            Then pour the self-hydrating fluid into the wheel through the valve. For the amount of self-replenishing solution, refer to the instruction manual of the self-replenishing solution you purchased. The consumption of a road car is about 30-60ML per wheel.
          </li>
        </ul>

        <div class="sizecharts-installation-images sizecharts-installation-images--installation">
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant1.webp"
            alt="Pour vacuum tire sealant 1"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant2.webp"
            alt="Pour vacuum tire sealant 2"
            :zoomOnClick="true"
          />
          <GuideImage
            class="sizecharts-installation-images__item"
            src="/public/tiresizecharts/installation/4/Separate-vacuum-nozzle-tool.webp"
            alt="Separate vacuum nozzle tool"
            :zoomOnClick="true"
          />
        </div>

        <h3 class="sizecharts-section__subheading">3. Cheer up</h3>
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
        <h2 class="sizecharts-section__title">How to choose</h2>
        <!-- Tire width -> rim internal width helper placed directly under title -->
        <TireRimHelper />

        <div class="mt-4">
          <h3 class="sizecharts-section__subheading">
            Reference charts from DT Swiss
          </h3>
          <p class="sizecharts-section__intro">
            The following hookless (TSS) and hooked (TC) rim inner width charts from DT Swiss
            show recommended (dark) and possible (light) combinations. Always stay within the
            limits specified by your tire and rim manufacturers.
          </p>
          <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div class="rounded-xl bg-slate-900/70 p-2 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
              <img
                src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
                alt="DT Swiss hookless TSS rim inner width recommendation chart"
                class="block h-auto w-full"
                loading="lazy"
              />
            </div>
            <div class="rounded-xl bg-slate-900/70 p-2 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
              <img
                src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
                alt="DT Swiss hooked TC rim inner width recommendation chart"
                class="block h-auto w-full"
                loading="lazy"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- Tire pressure -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Tire pressure</h2>
        <div class="tire-pressure-card">
          <TirePressureSection @openTireProducts="openTireProductsDrawer" />
        </div>
      </section>

      <!-- Inner tube (placeholder) -->
      <section
        v-show="activeTab === 'tube'"
        id="tube"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Inner tube</h2>
        <div class="mt-2 flex justify-center">
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
            @click="openInnerTubeSearch"
          >
            Search inner tubes with advanced filters
          </button>
        </div>
        <p class="sizecharts-section__intro">
          Overview and best practices for selecting and installing inner tubes. Detailed content
          will be added here in a future update.
        </p>
        <ul class="sizecharts-section__list">
          <li>Tube sizing basics: how to match tube size to tire and rim.</li>
          <li>Valve types and rim compatibility (Presta vs Schrader).</li>
          <li>Common installation mistakes and how to avoid pinch flats.</li>
        </ul>
      </section>

      <div class="sizecharts-feedback">
        <UserFeedbackThread
          threadKey="guides-tireguides"
          title="Share your feedback about this Tire Guides guide"
        />
      </div>
    </div>
  </div>

  <WhatsAppProductSearchResultDrawer
    v-model="tireProductsDrawerVisible"
    :loading="tireProductsLoading"
    :results="tireProductsResults"
    :error="tireProductsError"
    :query="tireProductsQuery"
    @close="handleTireProductsDrawerClose"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from '#imports'
import TubelessProducts from '~/components/TubelessProducts.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import GuideImage from '~/components/GuideImage.vue'
import TireRimHelper from '~/components/TireRimHelper.vue'
import TirePressureSection from '~/components/TirePressureSection.vue'
import TireSizeSection from '~/components/TireSizeSection.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'

type SizeChartsTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

definePageMeta({
  layout: 'products',
  path: '/guides/tireguides',
})

useHead({
  title: 'Tire Guides',
})

const tabs: { id: SizeChartsTabId; label: string }[] = [
  { id: 'size', label: 'Tire size' },
  { id: 'match', label: 'Match' },
  { id: 'tubeless', label: 'Tubeless tires' },
  { id: 'installation', label: 'Installation' },
  { id: 'choose', label: 'How to choose' },
  { id: 'rims', label: 'Tire pressure' },
  { id: 'tube', label: 'Inner tube' },
]

const activeTab = ref<SizeChartsTabId>('tubeless')

const route = useRoute()

const { open: openShopSearchSheet } = useShopSearchSheet()

const openInnerTubeSearch = () => {
  openShopSearchSheet({
    presetCategorySlug: 'inner-tube',
    presetKeywords: ['Inner tube'],
  })
}

const getTabFromHash = (hash: string): SizeChartsTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: SizeChartsTabId[] = ['size', 'match', 'tubeless', 'installation', 'choose', 'rims', 'tube']
  return (allowed as string[]).includes(raw) ? (raw as SizeChartsTabId) : null
}

// Tire products drawer (filtered by tire category)
const tireProductsDrawerVisible = ref(false)
const tireProductsLoading = ref(false)
const tireProductsResults = ref<any[]>([])
const tireProductsError = ref<string | null>(null)
const tireProductsQuery = ref('')

const openTireProductsDrawer = async () => {
  const categorySlug = 'tire'

  tireProductsQuery.value = 'Tire category'
  tireProductsError.value = null
  tireProductsDrawerVisible.value = true
  tireProductsLoading.value = true

  try {
    const response = await $fetch<any>('/wp-json/tanzanite/v1/products', {
      params: {
        category: categorySlug,
        per_page: 20,
        status: 'publish',
      },
      credentials: 'include',
    })

    if (response && Array.isArray(response.items)) {
      tireProductsResults.value = response.items.map((item: any) => ({
        id: item.id,
        title: item.title,
        url: item.preview_url || `/shop/${item.slug || item.id}`,
        thumbnail: item.thumbnail,
        price:
          item.prices?.sale > 0
            ? `$${item.prices.sale}`
            : item.prices?.regular > 0
              ? `$${item.prices.regular}`
              : '',
      }))
    } else {
      tireProductsResults.value = []
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load tire products', error)
    tireProductsError.value = 'Failed to load tire products. Please try again.'
    tireProductsResults.value = []
  } finally {
    tireProductsLoading.value = false
  }
}

const handleTireProductsDrawerClose = () => {
  tireProductsDrawerVisible.value = false
  tireProductsError.value = null
  tireProductsQuery.value = ''
  tireProductsResults.value = []
  tireProductsLoading.value = false
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

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
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.sizecharts-tabs::-webkit-scrollbar {
  display: none;
}

.sizecharts-tabs__item {
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

.sizecharts-tabs__item:active {
  transform: scale(0.96);
}

.sizecharts-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.sizecharts-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .sizecharts-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
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
  border: none;
  box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
}

.sizecharts-installation-images--tubeless {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.5rem;
}

.sizecharts-installation-images--installation {
  /* AUTO-FIT GRID for GuideImage rows: 1–3 images share the row evenly */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.5rem;
}

.sizecharts-installation-images--tubeless .sizecharts-installation-images__item {
  flex: 1 1 0;
}

.sizecharts-installation-images--tubeless .sizecharts-installation-images__img {
  height: auto;
  aspect-ratio: 16 / 9;
  object-fit: contain;
}

@media (max-width: 768px) {
  .sizecharts-tabs {
    justify-content: flex-start;
  }
}
</style>
