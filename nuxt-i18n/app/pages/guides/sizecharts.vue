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
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Tire size</h2>

        <div class="mt-1">
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
            @click="openTireProductsDrawer"
          >
            Check out tires
          </button>
        </div>

        <div class="mt-4">
          <GuideImage
            src="/public/tiresizecharts/tiresize/tanzanite-schwalbe-tiresize.webp"
            alt="Tanzanite and Schwalbe tire size chart overview"
            :zoomOnClick="true"
          />
        </div>
        <ul class="sizecharts-section__list space-y-2">
          <li>
            The sizes of bicycle tires are now designated according to the European tire and rim standard ETRTO
            (European Tire and Rim Technical Organization). In practice, however, the older English and French names are
            also used.
          </li>
          <li>
            The ETRTO size designation (e.g. 37-622) indicates the width (37 mm) and the inner diameter of the tire
            (622 mm). This designation is unambiguous and allows a clear assignment to the rim size.
          </li>
          <li>
            The inch designation (e.g. 28 x 1.40) indicates the approximate outside diameter (28 inches) and the tire
            width (1.40 inches). There is also the inch designation in the form 28 x 1 5⁄8 x 1 3⁄8 (approximate outer
            diameter x tire height x tire width). The inch information is not precise and not clear. For example,
            diameters of 559 mm (MTB), 571 mm (triathlon) and 590 mm (Dutch touring bikes) are all labeled 26-inch.
            Tires with a diameter of 622 and 635 mm are both referred to as 28 inches. Curiously, tires with an inner
            diameter of 630 mm are called 27 inches. These designations have their origin – and their meaning – from the
            times of the tire brake that acted directly onto the tire. At that time, the exact outer diameter of the
            wheel and tire was dictated by the brake. Depending on the tire width, there were different standards for
            the inner diameter. In English-language regions and in MTB sports, the inch designations are still
            widespread.
          </li>
          <li>
            The French sizing (e.g. 700 x 35C) indicates the approximate outside diameter (700 mm) and tire width
            (35 mm). The letter at the end gives an indication of the inside diameter of the tire. In this case, the C
            stands for 622 mm. The French designation does not exist for all tire sizes. For example, B is not used for
            MTB sizes in the 700 series.
          </li>
        </ul>

        <div class="mt-4 overflow-x-auto rounded-2xl border border-slate-800 bg-slate-950/40">
          <table class="min-w-full text-left text-xs sm:text-sm text-slate-100">
            <thead class="bg-slate-900/80">
              <tr>
                <th class="px-3 py-2 font-semibold"></th>
                <th class="px-3 py-2 font-semibold">ETRTO</th>
                <th class="px-3 py-2 font-semibold">Inch</th>
                <th class="px-3 py-2 font-semibold">French</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
              <tr>
                <td class="px-3 py-2 font-semibold align-top">Sizes</td>
                <td class="px-3 py-2 align-top">37-622</td>
                <td class="px-3 py-2 align-top">
                  28 x 1,40<br />
                  28 x 1 5⁄8 x 1 3⁄8
                </td>
                <td class="px-3 py-2 align-top">700 x 35C</td>
              </tr>
              <tr>
                <td class="px-3 py-2 font-semibold align-top">Outer diameter</td>
                <td class="px-3 py-2 align-top">-</td>
                <td class="px-3 py-2 align-top">approx. 28 Inch</td>
                <td class="px-3 py-2 align-top">approx. 700 mm</td>
              </tr>
              <tr>
                <td class="px-3 py-2 font-semibold align-top">Inner diameter</td>
                <td class="px-3 py-2 align-top">622 mm</td>
                <td class="px-3 py-2 align-top">-</td>
                <td class="px-3 py-2 align-top">-</td>
              </tr>
              <tr>
                <td class="px-3 py-2 font-semibold align-top">Tire width</td>
                <td class="px-3 py-2 align-top">approx. 37 mm</td>
                <td class="px-3 py-2 align-top">
                  ca. 1 3⁄8 Inch<br />
                  bzw. 1,40 Inch
                </td>
                <td class="px-3 py-2 align-top">approx. 35 mm</td>
              </tr>
              <tr>
                <td class="px-3 py-2 font-semibold align-top">Tire height</td>
                <td class="px-3 py-2 align-top">-</td>
                <td class="px-3 py-2 align-top">approx. 1 5⁄8 Inch</td>
                <td class="px-3 py-2 align-top">-</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 class="sizecharts-section__subheading mt-4">WHAT TIRE SIZES ARE THERE?</h3>
        <p class="sizecharts-section__intro">
          You should find almost every known size designation in this list. We have tried to include all size
          designations that are used in the market or have been used in the past. Tire sizes with a classic fractional
          size designation are often found on old Schwalbe tires, and sometimes even on current models if that size
          designation is still in common usage.
        </p>

        <div
          class="mt-2 rounded-2xl bg-slate-900/80 shadow-[4px_4px_18px_rgba(0,0,0,1)] overflow-hidden"
        >
          <img
            src="/public/tiresizecharts/tiresize/what-tire-sizes-are-available-schwalbe.svg"
            alt="Overview table comparing ETRTO, inch and French tire size designations"
            class="block h-auto w-[110%] max-w-none -translate-x-[4.5%]"
            loading="lazy"
          />
        </div>
      </section>

      <!-- Tubeless tires -->
      <section
        v-show="activeTab === 'tubeless'"
        id="tubeless"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title sizecharts-section__title--sr-only">Tubeless tires</h2>

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

        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('choose')"
        >
          How to choose and specific specifications
        </button>

        <h3 class="sizecharts-section__subheading">
          Am I suitable for tubeless or tubeless use?
        </h3>
        <p class="sizecharts-section__intro">
          It depends on your usage needs and habits. If you are doing daily commuting or relaxing holiday riding,
          tubeless tires are enough; but if you often ride on soft and slippery roads, or require more advanced,
          high-speed riding enjoyment, then tubeless tires are a good choice.
        </p>

        <h3 class="sizecharts-section__subheading">How to install tubeless wheels?</h3>
        <button
          type="button"
          class="sizecharts-brand-button"
          @click="setActiveTab('installation')"
        >
          View graphic installation details
        </button>

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
        <h2 class="sizecharts-section__title sizecharts-section__title--sr-only">Installation</h2>
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
        <p class="sizecharts-section__intro">
          Simple guidelines for selecting tire size based on usage and frame clearance.
        </p>
        <ul class="sizecharts-section__list">
          <li>Match tire width to riding style: road, gravel, XC, trail, city.</li>
          <li>Consider rider weight, terrain and comfort preferences.</li>
          <li>Use the size charts to see compatible width ranges per rim.</li>
        </ul>

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

        <!-- Tire width -> rim internal width helper -->
        <TireRimHelper />
      </section>

      <!-- Tire pressure -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="sizecharts-section"
      >
        <TirePressureSection @openTireProducts="openTireProductsDrawer" />
      </section>

      <!-- Inner tube (placeholder) -->
      <section
        v-show="activeTab === 'tube'"
        id="tube"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Inner tube</h2>
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
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'

type SizeChartsTabId = 'size' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

definePageMeta({
  layout: 'products',
  path: '/guides/tireguides',
})

useHead({
  title: 'Tire Guides',
})

const tabs: { id: SizeChartsTabId; label: string }[] = [
  { id: 'size', label: 'Tire size' },
  { id: 'tubeless', label: 'Tubeless tires' },
  { id: 'installation', label: 'Installation' },
  { id: 'choose', label: 'How to choose' },
  { id: 'rims', label: 'Tire pressure' },
  { id: 'tube', label: 'Inner tube' },
]

const activeTab = ref<SizeChartsTabId>('tubeless')

const route = useRoute()

const getTabFromHash = (hash: string): SizeChartsTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: SizeChartsTabId[] = ['size', 'tubeless', 'installation', 'choose', 'rims', 'tube']
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
