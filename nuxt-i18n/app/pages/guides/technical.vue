<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Technical</h2>

    <div class="technical-page">
      <div class="technical-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="technical-tabs__item"
          :class="{ 'technical-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Rims -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="technical-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Rims</h2>
        <p class="sizecharts-section__intro">
          Overview of Tanzanite rims and how rim dimensions relate to tire
          compatibility and spoke length.
        </p>
        <ul class="sizecharts-section__list">
          <li>Rim diameter standards (ISO / ETRTO such as 622, 584, etc.).</li>
          <li>Internal / external width and their impact on tire choice.</li>
          <li>Spoke hole count (28H / 32H / 36H …) and recommended use cases.</li>
          <li>ERD (Effective Rim Diameter) definition and basic measurement notes.</li>
          <li>How ERD is used when entering data into the Spoke Calculator.</li>
        </ul>
      </section>

      <!-- Spokes -->
      <section
        v-show="activeTab === 'spokes'"
        id="spokes"
        class="technical-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spokes</h2>
        <p class="sizecharts-section__intro">
          Basic spoke concepts: dimensions, materials and typical use cases for
          different riding styles.
        </p>
        <ul class="sizecharts-section__list">
          <li class="technical-list-item--no-bullet">
            Common spoke diameters and shapes (straight / J-bend, round / bladed).
            <div class="technical-inline-image technical-inline-image--spoke-shapes">
              <GuideImage
                src="/public/technical/j-bend-spokeandstraightpull-spoke.webp"
                alt="Comparison of J-bend and straight-pull bicycle spokes"
                :zoomOnClick="true"
                caption="This illustration compares J-bend and straight-pull spoke heads and how they seat in the hub flange."
              />
              <GuideImage
                src="/public/technical/Sapim-black-stainless-steel-straight-gauge-2.0mm-spoke-double-butted-2.0-1.8-2.0mm-spoke-aero-flat-bladed-spoke.webp"
                alt="Examples of Sapim stainless steel spokes in different shapes and profiles"
                :zoomOnClick="true"
                caption="Examples of stainless steel spokes in straight gauge, double-butted and aero bladed profiles used on performance wheelsets."
              />
            </div>
          </li>
        </ul>
        <div class="technical-materials">
          <p class="sizecharts-section__intro">
            The following is a detailed introduction to different spoke materials:
          </p>
          <ul class="sizecharts-section__list">
            <li>
              <strong class="technical-materials__heading">1. Stainless Steel</strong>
              <p>
                Stainless steel is currently the most mainstream and widely used spoke
                material on the market.
              </p>
              <p>The main brands currently on the market include:</p>
              <ul>
                <li>
                  <strong>Sapim</strong> &mdash; regarded as No. 1 in spokes and focused solely on
                  spokes. Several of its models are high quality, and CX-RAY is used by many
                  wheel brands.
                  <button type="button" class="technical-brand-button">
                    Sapim
                  </button>
                </li>
                <li>
                  <strong>DT Swiss</strong> &mdash; globally recognised quality, producing spokes,
                  wheels and hubs.
                  <button type="button" class="technical-brand-button">
                    DT Swiss
                  </button>
                </li>
                <li>
                  <strong>Pillar</strong> &mdash; currently the third-ranked brand and very
                  cost-effective compared with the first two.
                  <button type="button" class="technical-brand-button">
                    Pillar
                  </button>
                </li>
              </ul>
              <p>We also offer another cost-effective brand:</p>
              <ul>
                <li>
                  <strong>YT (Yongteng)</strong> &mdash; an emerging Chinese brand producing both
                  bicycle and motorcycle spokes.
                  <button type="button" class="technical-brand-button">
                    YT
                  </button>
                </li>
              </ul>
              <p><strong>Advantages:</strong></p>
              <ul>
                <li><strong>High strength and durability</strong> &mdash; able to withstand high
                tension and long-term use.</li>
                <li><strong>Strong corrosion resistance</strong> &mdash; not easy to rust and
                suitable for various weather conditions.</li>
                <li><strong>Cost-effective</strong> &mdash; relatively low cost while balancing
                weight and performance.</li>
                <li><strong>Good toughness</strong> &mdash; steel spokes have ductility and
                elasticity, more likely to deform gradually than suddenly break under impact.</li>
              </ul>
              <p><strong>Disadvantages:</strong></p>
              <ul>
                <li><strong>Relatively heavy weight</strong> compared with carbon fiber or titanium
                alloys.</li>
              </ul>
              <p>
                <strong>Applicable scenarios:</strong> suitable for almost all types of bicycles,
                from daily commuter bikes to professional mountain and road bikes.
              </p>
            </li>
            <li>
              <strong class="technical-materials__heading">2. Carbon Fiber</strong>
              <p>
                Carbon fiber spokes are typically found on the most high-end,
                performance-focused wheelsets. We also offer factory-developed carbon fiber
                spokes.
              </p>
              <p><strong>Advantages:</strong></p>
              <ul>
                <li><strong>Ultra-lightweight</strong> &mdash; can significantly reduce wheelset
                weight compared with steel spokes (often saving more than 100 grams) and improve
                acceleration.</li>
                <li><strong>Excellent rigidity</strong> &mdash; provides very good lateral rigidity
                and tensile strength, giving more direct power transfer and sharper handling.</li>
                <li><strong>Good shock absorption</strong> &mdash; can absorb part of riding
                vibration and improve comfort.</li>
              </ul>
              <p><strong>Disadvantages:</strong></p>
              <ul>
                <li><strong>Expensive</strong> &mdash; much higher price than stainless steel
                spokes.</li>
                <li><strong>Poor side impact resistance</strong> &mdash; very strong in tension but
                vulnerable to side impacts or punctures; once damaged it usually cannot be
                repaired and must be replaced.</li>
                <li><strong>Higher maintenance requirements</strong> &mdash; installation and
                service require professional skills and specific tools.</li>
              </ul>
              <p>
                <strong>Applicable scenarios:</strong> high-end road racing, time trials,
                triathlons and other situations that demand extreme light weight and stiffness.
              </p>
            </li>
            <li>
              <strong class="technical-materials__heading">3. Titanium Alloy</strong>
              <p>
                Titanium spokes sit between steel and carbon fiber in character, but are
                relatively rare on the market. We plan to launch titanium options later.
              </p>
              <p><strong>Advantages:</strong></p>
              <ul>
                <li><strong>Lightweight and high strength</strong> &mdash; very high
                strength-to-weight ratio, lighter than stainless steel while remaining strong.</li>
                <li><strong>Excellent corrosion resistance</strong> &mdash; almost no rust.</li>
                <li><strong>Smoother ride feel</strong> &mdash; good vibration absorption for a
                smoother riding experience.</li>
              </ul>
              <p><strong>Disadvantages:</strong></p>
              <ul>
                <li><strong>Very expensive</strong> compared with stainless steel.</li>
                <li><strong>Limited availability</strong> &mdash; relatively few options on the
                market.</li>
              </ul>
              <p>
                <strong>Applicable scenarios:</strong> high-end custom wheelsets and riders who
                pursue light weight, comfort and uniqueness.
              </p>
            </li>
            <li>
              <strong class="technical-materials__heading">4. Aluminum Alloy</strong>
              <p>
                Aluminum spokes are relatively rare and mainly used by a few manufacturers (such
                as Campagnolo) in certain high-end wheelsets.
              </p>
              <p><strong>Advantages:</strong></p>
              <ul>
                <li><strong>Very light</strong> &mdash; one of the lightest metal spoke
                materials.</li>
              </ul>
              <p><strong>Disadvantages:</strong></p>
              <ul>
                <li><strong>Lower strength and durability</strong> than stainless steel or titanium
                alloy, more prone to fatigue fracture.</li>
                <li><strong>More prone to corrosion</strong> and requires more careful
                maintenance.</li>
              </ul>
              <p>
                <strong>Applicable scenarios:</strong> specially designed ultra-lightweight racing
                wheelsets.
              </p>
            </li>
          </ul>
          <p class="sizecharts-section__intro">
            Which material you choose ultimately depends on your budget, riding style and
            specific performance needs such as weight, stiffness and comfort.
          </p>
        </div>
      </section>

      <!-- Spoke pattern -->
      <section
        v-show="activeTab === 'spoke-pattern'"
        id="spoke-pattern"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spoke pattern</h2>
        <p class="sizecharts-section__intro">
          Content coming soon...
        </p>
      </section>

      <!-- Spoke length -->
      <section
        v-show="activeTab === 'spoke-length'"
        id="spoke-length"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spoke length</h2>
      </section>

      <!-- Hubs -->
      <section
        v-show="activeTab === 'hubs'"
        id="hubs"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Hubs</h2>
        <p class="sizecharts-section__intro">
          Hub geometry concepts that matter for spoke length and wheel
          behaviour.
        </p>
        <ul class="sizecharts-section__list">
          <li>Left / right flange distance from the hub centerline.</li>
          <li>Flange diameter and how it interacts with rim ERD.</li>
          <li>Matching hub hole count with rim hole count.</li>
          <li>
            Basic hub standards (axle type, freehub, brake interface) as context
            for choosing the right hub.
          </li>
        </ul>
      </section>

      <!-- Tension -->
      <section
        v-show="activeTab === 'tension'"
        id="tension"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Tension &amp; formulas</h2>
        <h3 class="sizecharts-section__subheading text-sky-300 font-semibold">
          Wheel Spoke Tension and Rider Weight/Usage (Workshop Standard)
        </h3>

        <h3 class="sizecharts-section__subheading">
          <span class="text-amber-300 font-semibold">1. Effect of Rider Weight</span>
        </h3>
        <ul class="sizecharts-section__list">
          <li>Static load: A 75 kg rider increases upper spoke tension by ~15%, while lower spokes lose tension.</li>
          <li>Dynamic riding: Pedaling raises drive-side tension and lowers braking-side tension; distribution keeps changing.</li>
          <li>Heavier riders: Need higher overall tension and thicker spokes to avoid deformation or stiffness loss.</li>
        </ul>

        <h3 class="sizecharts-section__subheading">
          <span class="text-amber-300 font-semibold">2. Differences by Usage</span>
        </h3>
        <ul class="sizecharts-section__list">
          <li>
            <div>{Road racing}</div>
            <div>Focus on light weight and stiffness.</div>
            <div>Drive-side tension typically 120–130 kgf, non-drive side lower, with a ratio around 10:6–10:7.</div>
          </li>
          <li>
            <div>{Mountain biking}</div>
            <div>Emphasis on durability and shock absorption.</div>
            <div>Tension is set slightly lower to reduce risk of breakage.</div>
            <div>Rim offset (about 2.6–3 mm) is often used to balance spoke tension between sides.</div>
          </li>
          <li>
            <div>{Touring/commuting}</div>
            <div>Prioritize stability and long-term durability.</div>
            <div>Tension is kept moderate, with thicker spokes and a more balanced tension ratio.</div>
          </li>
        </ul>

        <h3 class="sizecharts-section__subheading">
          <span class="text-amber-300 font-semibold">3. Design and Building Considerations</span>
        </h3>
        <ul class="sizecharts-section__list">
          <li>
            <div>{Rim geometry}</div>
            <div>Offset designs improve tension balance between drive and non-drive sides.</div>
          </li>
          <li>
            <div>{Spoke gauge}</div>
            <div>14G (2.0 mm) is stronger than 15G (1.8 mm), recommended for heavier riders or off-road use.</div>
          </li>
          <li>
            <div>{Tension vs. stiffness}</div>
            <div>Too low → wheel feels loose, lacks stiffness.</div>
            <div>Too high → risk of metal fatigue or reduced rim stiffness, especially in lightweight carbon rims.</div>
          </li>
        </ul>

        <div class="mt-4 overflow-x-auto">
          <table class="min-w-full border-separate border-spacing-0 rounded-xl overflow-hidden bg-slate-900/70 text-slate-100">
            <thead>
              <tr class="bg-slate-800/90 text-sm font-semibold text-slate-50">
                <th class="px-4 py-3 text-left border-b border-slate-700">Rider weight range</th>
                <th class="px-4 py-3 text-left border-b border-slate-700">Road racing recommended tension (kgf)</th>
                <th class="px-4 py-3 text-left border-b border-slate-700">MTB recommended tension (kgf)</th>
                <th class="px-4 py-3 text-left border-b border-slate-700">Touring/commute recommended tension (kgf)</th>
              </tr>
            </thead>
            <tbody class="text-sm">
              <tr class="border-b border-slate-800/70">
                <td class="px-4 py-3">≤ 60 kg</td>
                <td class="px-4 py-3">115–120</td>
                <td class="px-4 py-3">95–105</td>
                <td class="px-4 py-3">90–100</td>
              </tr>
              <tr class="border-b border-slate-800/70">
                <td class="px-4 py-3">60–80 kg</td>
                <td class="px-4 py-3">120–125</td>
                <td class="px-4 py-3">100–115</td>
                <td class="px-4 py-3">95–105</td>
              </tr>
              <tr class="border-b border-slate-800/70">
                <td class="px-4 py-3">80–100 kg</td>
                <td class="px-4 py-3">125–130</td>
                <td class="px-4 py-3">110–120</td>
                <td class="px-4 py-3">100–110</td>
              </tr>
              <tr>
                <td class="px-4 py-3">≥ 100 kg</td>
                <td class="px-4 py-3">130–135</td>
                <td class="px-4 py-3">115–125</td>
                <td class="px-4 py-3">105–115</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          class="mt-4 rounded-lg border border-amber-300/80 bg-slate-900/80 px-4 py-3 text-sm leading-relaxed text-amber-100"
        >
          The above is a reference value, which is based on the geometrically
          excellent tension ratio of the hub at 10:7 or better and the value
          after the stress is completely released. The actual situation may
          depend on the number of spokes and spoke type of each hub (for
          example, flat spokes will be relatively weak laterally and need to be
          moderately increased).
        </div>

        <h3 class="sizecharts-section__subheading mt-4">
          Risks and Precautions
        </h3>
        <ul class="sizecharts-section__list">
          <li>
            <div>{Excessive tension}</div>
            <div>Can cause rim deformation or fatigue cracks.</div>
          </li>
          <li>
            <div>{Insufficient tension}</div>
            <div>Leads to loose spokes, wheel wobble, and safety issues.</div>
          </li>
          <li>
            <div>{Weight differences}</div>
            <div>Heavy riders need higher tension and stronger spokes; lighter riders can use lighter setups.</div>
          </li>
        </ul>

        <h3 class="sizecharts-section__subheading">Key Notes</h3>
        <ul class="sizecharts-section__list">
          <li>
            <div>{Non-drive side tension}</div>
            <div>Usually 20–40% lower than drive side, depending on hub flange spacing and rim geometry.</div>
          </li>
          <li>
            <div>{Carbon vs. alloy rims}</div>
            <div>Carbon rims can handle higher tension (up to 130–135 kgf). Alloy rims should be limited to 120–125 kgf.</div>
          </li>
          <li>
            <div>{Spoke choice}</div>
            <div>Light riders can use 15G spokes; heavy riders or off-road use should stick to 14G.</div>
          </li>
          <li>
            <div>{Usage summary}</div>
            <div>Road racing = stiffness and responsiveness; mountain biking = durability and cushioning; touring = stability and balance.</div>
          </li>
        </ul>

        <h3 class="sizecharts-section__subheading">Our Workshop Process (Practical Standard)</h3>
        <ul class="sizecharts-section__list">
          <li>Weight classification → determine baseline tension range.</li>
          <li>
            <div>{Usage adjustment}</div>
            <div>Road racing → higher tension for stiffness.</div>
            <div>Mountain biking → slightly lower tension for durability.</div>
            <div>Touring → balanced tension for stability.</div>
          </li>
          <li>
            <div>{Material limits}</div>
            <div>Carbon rims → allow higher tension (up to 135 kgf).</div>
            <div>Alloy rims → capped at 120–125 kgf.</div>
          </li>
          <li>Spoke selection → match spoke gauge to rider weight and usage.</li>
          <li>Final build → ensure proper drive/non-drive tension ratio (about 10:6–10:7; some hub geometries may only be 10:4) and use rim offset to balance.</li>
        </ul>

        <h3 class="sizecharts-section__subheading">Conclusion</h3>
        <ul class="sizecharts-section__list">
          <li>Rider weight sets the baseline tension.</li>
          <li>Usage defines the adjustment.</li>
          <li>Rim material sets the upper limit.</li>
          <li>Spoke gauge ensures durability and safety.</li>
          <li>Following this flow balances performance, reliability, and safety for each rider.</li>
        </ul>
      </section>
      <div class="technical-feedback">
        <UserFeedbackThread
          threadKey="guides-technical"
          title="Share your feedback about this Technical guide"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useLocalePath, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import GuideImage from '~/components/GuideImage.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Technical',
})

type TechnicalTabId = 'rims' | 'spokes' | 'spoke-pattern' | 'spoke-length' | 'hubs' | 'tension'

const localePath = useLocalePath()
const route = useRoute()

const tabs: { id: TechnicalTabId; label: string }[] = [
  { id: 'rims', label: 'Rims' },
  { id: 'spokes', label: 'Spokes' },
  { id: 'spoke-pattern', label: 'Spoke pattern' },
  { id: 'spoke-length', label: 'Spoke length' },
  { id: 'hubs', label: 'Hubs' },
  { id: 'tension', label: 'Tension' },
]

const activeTab = ref<TechnicalTabId>('rims')

const getTabFromHash = (hash: string): TechnicalTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: TechnicalTabId[] = ['rims', 'spokes', 'spoke-pattern', 'spoke-length', 'hubs', 'tension']
  return (allowed as string[]).includes(raw) ? (raw as TechnicalTabId) : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: TechnicalTabId) => {
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
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.technical-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.technical-tabs {
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

.technical-tabs::-webkit-scrollbar {
  display: none;
}

.technical-tabs__item {
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

.technical-tabs__item:active {
  transform: scale(0.96);
}

.technical-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.technical-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .technical-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

.technical-section {
  margin-top: 0.75rem;
}

.technical-materials {
  margin-top: 0.75rem;
}

.technical-materials > .sizecharts-section__list > li {
  padding: 0.75rem 0.85rem;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.75);
  border: 1px solid rgba(148, 163, 184, 0.35);
}

.technical-materials > .sizecharts-section__list > li + li {
  margin-top: 0.6rem;
}

.technical-materials__heading {
  display: block;
  margin-bottom: 0.25rem;
  color: #fbbf24;
}

.technical-link-button {
  margin-left: 0.5rem;
  padding: 0.15rem 0.6rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  font-size: 0.8rem;
  font-weight: 500;
  color: #e0f2fe;
  text-decoration: none;
  background: rgba(15, 23, 42, 0.9);
  transition: background-color 0.15s ease, border-color 0.15s ease,
    color 0.15s ease;
}

.technical-link-button:hover {
  background: rgba(56, 189, 248, 0.2);
  border-color: rgba(56, 189, 248, 1);
  color: #ffffff;
}

.technical-brand-button {
  margin-left: 0.5rem;
  margin-top: 0.15rem;
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
}

.technical-inline-image {
  margin-top: 0.4rem;
}

.technical-inline-image img {
  display: block;
  max-width: 260px;
  width: 100%;
  height: auto;
  border-radius: 0.5rem;
}

.technical-inline-image--spoke-length img {
  max-width: 380px;
}

.technical-inline-image--spoke-shapes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.technical-inline-image--spoke-length {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.75rem;
}

.technical-inline-image--erd {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.75rem;
}

.technical-spoke-length-heading {
  color: #38bdf8;
}

.technical-list-item--no-bullet {
  padding-left: 0;
}

.technical-list-item--no-bullet::before {
  content: none;
}

.technical-feedback {
  margin-top: 2.5rem;
}

/* Nested lists inside shared guide lists on this page */
.technical-page .sizecharts-section__list ul {
  margin-top: 0.15rem;
  padding-left: 1.1rem;
  list-style-type: none;
}

@media (min-width: 768px) {
  .technical-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .technical-tabs {
    justify-content: flex-start;
  }
}
</style>
