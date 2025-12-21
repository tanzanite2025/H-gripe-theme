<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">About us</h1>

    <div class="company-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="company-tabs__item"
        :class="{ 'company-tabs__item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <section
      v-show="activeTab === 'factory'"
      id="factory"
      class="company-section"
    >
      <h2 class="company-section__title">Factory</h2>
      <ul class="company-section__list">
        <li>
          began in 2013 in Xiamen,China.We specialize in manufacturing carbon fiber bicycle rims and
          wheels since then,which has successfully innovated and industrialised the supply of
          lightweight carbon fibre rims and wheels to the bicycle industry.
        </li>
        <li>
          More than 20+experienced engineers in our R&D Department and 250+ effective workers,with the
          sufficient manufacturing facilities,makes monthly capacity of 25,000 pcs rims possible.With
          strict high-quality standards on the manufacturing processes and 100% quality test&check to
          guarantee all the rims and wheels greater than UCI/ISO Test standard.
        </li>
        <li>
          We have a number of unique patented technology in production,making the carbon rims lighter and
          more strength far above the industry standard.Besides we put a strong emphasis on research and
          development,Amoy Wheels keeps innovating. We mostly use Toray carbon fiber,especially T700 and
          rest of T800,T1000 mixed.Both of T800 and T1000 are too rigid to be used alone.
        </li>
        <li>
          Xiamen city is leader of carbon fiber rims in China,We are the leader of carbon fiber rims in
          Xiamen.We have built a strong relationships with some famous bicycle brands over the world as
          the carbon rims or wheels supplier.Good product can talk,choosing us,choosing the better carbon
          rims &wheels.
        </li>
      </ul>

      <div class="factory-flow-grid">
        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-fanctory-EpoxyResinWorkshop1.webp"
            alt="Epoxy resin workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 01 - Epoxy Resin
            </p>
            <h3 class="factory-flow-card__title">Prepare and control epoxy resin</h3>
            <p class="factory-flow-card__text">
              We purchase raw materials from reliable suppliers and produce the optimal epoxy resin for carbon rims.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-carbonprepregsworkshop2.webp"
            alt="Carbon prepregs workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 02 - Carbon Prepregs
            </p>
            <h3 class="factory-flow-card__title">Produce carbon prepregs</h3>
            <p class="factory-flow-card__text">
              We purchase different kinds of carbon fibers from Toray Japan. The carbon prepregs are carbon fibers saturated with epoxy resin through the weaving machines.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-carbonprepregsstoreroom3.webp"
            alt="Carbon prepregs storeroom"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 03 - Carbon prepregs Storeroom
            </p>
            <h3 class="factory-flow-card__title">Store prepregs at stable temperature</h3>
            <p class="factory-flow-card__text">
              Carbon prepregs are carbon fibers saturated with epoxy resin and kept in optimal condition at -5 degrees Celsius, a stable temperature for the resin.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-cuttingworkshop4.webp"
            alt="Cutting workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 04 - Cutting Workshop
            </p>
            <h3 class="factory-flow-card__title">Cut carbon sheets to precise angles</h3>
            <p class="factory-flow-card__text">
              According to the production schedule we select the necessary carbon sheets. Sheets are cut into desired angles of 0 degrees, plus/minus 45 degrees, plus/minus 30 degrees and plus/minus 90 degrees, then precisely machine-cut into pieces.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-laminationpreparationworkshop5.webp"
            alt="Lamination preparation workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 05 - Lamination Preparation
            </p>
            <h3 class="factory-flow-card__title">Prepare laminations for rim sections</h3>
            <p class="factory-flow-card__text">
              This step speeds up carbon rim production. Lamination prepares different parts of the rim with carbon sheets so we can arrange the product more easily and balance rim weight, durability and rigidity.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-premoldlayupworkshop6.webp"
            alt="Premold and layup workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 06 - Premold & Layup
            </p>
            <h3 class="factory-flow-card__title">Premold and hand layup</h3>
            <p class="factory-flow-card__text">
              Premolds eliminate the risk of fiber movement during layup transfer. Technicians hand lay fabric layers from outer to inner, smoothing each overlap to avoid wrinkles, fractures and air bubbles before closing the premold.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-moldingworkshop7.webp"
            alt="Molding workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 07 - Molding
            </p>
            <h3 class="factory-flow-card__title">Cure rims under heat and pressure</h3>
            <p class="factory-flow-card__text">
              After checks, complete molds are moved to molding machines where heat and pressure are accurately controlled and monitored in curing ovens.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-Appearance%20inspection8.webp"
            alt="Appearance inspection"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 08 - Appearance inspection
            </p>
            <h3 class="factory-flow-card__title">Inspect every rim visually</h3>
            <p class="factory-flow-card__text">
              All rims are 100% inspected for appearance and obvious defects before moving on.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-cncmachiningworkshop9.webp"
            alt="CNC machining workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 09 - CNC Machining
            </p>
            <h3 class="factory-flow-card__title">Drill spoke and valve holes</h3>
            <p class="factory-flow-card__text">
              After inspection, rims are drilled on aviation-grade CNC machines using bits designed for composites to ensure correct angles and clean holes.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-Weightdetection10.webp"
            alt="Weight detection station"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 10 - Weight detection
            </p>
            <h3 class="factory-flow-card__title">Control rim weight precisely</h3>
            <p class="factory-flow-card__text">
              Weight standards are strictly controlled, with measurements accurate to 1 gram.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-sandingworshop11.webp"
            alt="Sanding workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 11 - Sanding
            </p>
            <h3 class="factory-flow-card__title">Prepare surfaces by sanding</h3>
            <p class="factory-flow-card__text">
              After inspection, rims are transferred to the sanding workshop to prepare surfaces for the next processes.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-grinding12.webp"
            alt="Grinding workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 12 - Grinding
            </p>
            <h3 class="factory-flow-card__title">Refine critical surfaces</h3>
            <p class="factory-flow-card__text">
              After sanding, rims move to grinding to refine key surfaces and ensure consistency.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-detailstreatment13.webp"
            alt="Details treatment"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 13 - Details Treatment
            </p>
            <h3 class="factory-flow-card__title">Handle fine details</h3>
            <p class="factory-flow-card__text">
              Details determine success or failure. A dedicated workshop focuses on rim details to keep quality high.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-otherdetection14.webp"
            alt="Other rim tests"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 14 - Other tests
            </p>
            <h3 class="factory-flow-card__title">Perform additional tests</h3>
            <p class="factory-flow-card__text">
              Drainage holes, spoke holes and appearance are checked again before shipment.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzantie-factory-painting15.webp"
            alt="Painting workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 15 - Painting
            </p>
            <h3 class="factory-flow-card__title">Apply finishes where required</h3>
            <p class="factory-flow-card__text">
              Ultralight rims are paintless, while traditional rims receive glossy, matte or custom finishes in the painting workshop.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tananite-factory-logolidedecals16.webp"
            alt="Waterslide decals and logos"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 16 - Waterslide Decals
            </p>
            <h3 class="factory-flow-card__title">Apply waterslide decals</h3>
            <p class="factory-flow-card__text">
              After painting, rims that require waterslide decals are processed here for branding and graphics.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-lasercarving17.webp"
            alt="Laser carving logo workshop"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 17 - Laser Carving
            </p>
            <h3 class="factory-flow-card__title">Laser carve logos and marks</h3>
            <p class="factory-flow-card__text">
              Rims that require laser-carved logos are processed here for a durable, precise finish.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-inspectionpacking18.webp"
            alt="Inspection and packing"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 18 - Inspection & Packing
            </p>
            <h3 class="factory-flow-card__title">Final inspection and packing</h3>
            <p class="factory-flow-card__text">
              All rims are inspected again, protected with foam bags and carefully packed into cartons.
            </p>
          </div>
        </article>

        <article class="factory-flow-card">
          <img
            class="factory-flow-card__image"
            src="/company/ourstory/factory/tanzanite-factory-ourwarehouse19.webp"
            alt="Rim warehouse"
          />
          <div class="factory-flow-card__overlay">
            <p class="factory-flow-card__step">
              <span class="factory-flow-card__step-dot"></span>
              Step 19 - Our Warehouse
            </p>
            <h3 class="factory-flow-card__title">Store rims in organized warehouse</h3>
            <p class="factory-flow-card__text">
              Compliant operation and clear classification in the warehouse ensure model and rule accuracy.
            </p>
          </div>
        </article>
      </div>
    </section>

    <section
      v-show="activeTab === 'appearance'"
      id="appearance"
      class="company-section"
    >
      <h2 class="company-section__title">Appearance</h2>
      <p class="company-section__body">
        Carbon Fiber Weave &amp; Surface Finish Options
      </p>
      <p class="company-section__body">
        We offer a wide range of carbon fiber weave patterns and surface finishes to meet the aesthetic and performance preferences of every rider.
      </p>
      <div class="company-appearance-images">
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

      <h3 class="company-section__title">Weave patterns</h3>
      <ul class="company-section__list">
        <li><strong>UD (Unidirectional)</strong> &mdash; Matte &mdash; Minimal straight grain, ultra-light.</li>
        <li><strong>Butterfly</strong> &mdash; Matte &mdash; Unique organic pattern, high visual impact.</li>
        <li><strong>Marble</strong> &mdash; Matte &mdash; Stone-like texture, bold and distinctive.</li>
        <li><strong>3K</strong> &mdash; Satin &mdash; Classic checkerboard, balanced look.</li>
        <li><strong>3K Twill (Vertical)</strong> &mdash; Satin &mdash; Diagonal weave, dynamic appearance.</li>
        <li><strong>3K Twill (Horizontal)</strong> &mdash; Satin &mdash; Horizontal twill, modern aesthetic.</li>
        <li><strong>12K</strong> &mdash; Satin &mdash; Coarse weave, powerful visual texture.</li>
        <li><strong>12K Twill (Vertical)</strong> &mdash; Satin &mdash; Bold diagonal weave, vertical layout.</li>
        <li><strong>12K Twill (Horizontal)</strong> &mdash; Satin &mdash; Horizontal twill, aggressive styling.</li>
      </ul>
      <p class="company-section__body">
        All weaves are structural carbon fiber, not printed or laminated, ensuring both strength and visual integrity.
      </p>

      <h3 class="company-section__title">Brake track surface options (rim brake models only)</h3>
      <p class="company-section__body">
        We provide multiple brake track surface treatments to enhance braking performance and heat resistance:
      </p>
      <ul class="company-section__list">
        <li>High TG Resin UD Brake Track (UD weave only).</li>
        <li>High TG Resin Graphene Brake Track.</li>
        <li>AC3 Grooved Brake Track (two groove styles available).</li>
        <li>UD C-Grooved Brake Track (UD weave only).</li>
        <li>C-Grooved Graphene Brake Track.</li>
      </ul>
      <p class="company-section__body">
        All brake tracks are heat-treated and abrasion-tested for compatibility with mainstream brake pads.
      </p>

      <h3 class="company-section__title">Finish &amp; painting effects</h3>
      <p class="company-section__body">
        Choose from a variety of surface gloss levels and special paint effects to personalize your rims:
      </p>
      <p class="company-section__body">
        <strong>Gloss levels</strong>
      </p>
      <ul class="company-section__list">
        <li><strong>Matte:</strong> Subtle, scratch-resistant.</li>
        <li><strong>Satin:</strong> Soft sheen, balanced texture.</li>
        <li><strong>Glossy:</strong> High shine, premium look.</li>
        <li><strong>Paintless:</strong> Raw carbon fiber, ultra-lightweight.</li>
      </ul>
      <p class="company-section__body">
        <strong>Crystal effect finishes</strong>
      </p>
      <ul class="company-section__list">
        <li>Crystal Effect Black.</li>
        <li>Crystal Effect Silver.</li>
        <li>Crystal Effect Orange.</li>
        <li>Crystal Effect Purple.</li>
        <li>Crystal Effect Blue.</li>
      </ul>
      <p class="company-section__body">
        Crystal finishes use micro-layer coating technology for unique light reflection and color depth.
      </p>

      <h3 class="company-section__title">Laser-engraved logos: faster and more sustainable</h3>
      <p class="company-section__body">
        In addition to traditional decals and painted graphics, we also support laser-engraved logos as a more sustainable appearance option.
      </p>
      <p class="company-section__body">
        Laser logos remove the need for large vinyl decals and extra paint masks, which helps reduce plastic waste and speeds up production. For many standard builds, choosing a laser-engraved logo can typically shorten the overall lead time by about 5&ndash;7 days compared to complex custom paint or multi-layer decals.
      </p>
      <p class="company-section__body">
        For practical guidance on when to choose laser engraving versus decals for a specific wheelset, please see the Appearance Logo section in our Wheelset Buyers Guide.
      </p>
      <div
        class="company-video-thumbnail"
        @click="showLaserVideo = true"
      >
        <img
          class="company-video-thumbnail__image"
          src="/company/aboutus/appearance/tanzanite-carbonrims-finish2.webp"
          alt="Play laser engraving video showing Tanzanite carbon rim logo finish"
          loading="lazy"
        />
        <div class="company-video-thumbnail__overlay">
          <span class="company-video-thumbnail__icon">▶</span>
          <span class="company-video-thumbnail__label">Watch laser engraving video</span>
        </div>
      </div>

      <h3 class="company-section__title">Presentation suggestions</h3>
      <ul class="company-section__list">
        <li>Organize options into three clear sections: Weave, Brake Track, Finish.</li>
        <li>Pair each option with a high-resolution image and concise label.</li>
        <li>Include &ldquo;Recommended combinations&rdquo; or &ldquo;Customer favorites&rdquo; to guide decision-making.</li>
      </ul>
      <p class="company-section__body">
        For OEM clients, team orders, or custom branding requests, we offer logo integration, laser etching, and color matching services. Contact us for tailored solutions.
      </p>
    </section>

    <section
      v-show="activeTab === 'hole-patterns'"
      id="hole-patterns"
      class="company-section"
    >
      <h2 class="company-section__title">Hole Patterns</h2>
      <p class="company-section__body">
        Placeholder content for Hole Patterns.
      </p>
      <p class="company-section__body">
        Here you can describe the various drilling patterns available for the rims.
      </p>
    </section>

    <section
      v-show="activeTab === 'facility'"
      id="facility"
      class="company-section"
    >
      <h2 class="company-section__title">Facility</h2>
      <p class="company-section__body">
        Describe the key facilities that support the factory: warehousing, assembly
        areas, test labs, packing zones, and any dedicated spaces for training or
        demonstrations.
      </p>
      <p class="company-section__body">
        You can later replace this placeholder text with concrete details about
        climate control, storage systems, and how the layout helps keep builds
        organized and predictable.
      </p>
    </section>

    <section
      v-show="activeTab === 'manufacture'"
      id="manufacture"
      class="company-section"
    >
      <h2 class="company-section__title">Manufacture</h2>
      <p class="company-section__body">
        Use this tab to walk visitors through how a Tanzanite wheel is manufactured,
        from selecting rims and hubs to lacing, tensioning, truing, and final
        inspection.
      </p>
      <p class="company-section__body">
        Later you can add step-by-step photos, short videos, or diagrams so
        builders understand what level of craftsmanship and tooling goes into
        every wheel.
      </p>
    </section>

    <section
      v-show="activeTab === 'qualitycontrol'"
      id="qualitycontrol"
      class="company-section"
    >
      <h2 class="company-section__title">Quality control</h2>
      <p class="company-section__body">
        Summarize how quality control works at Tanzanite: incoming material checks,
        in-process measurements, and final wheel verification before shipping.
      </p>
      <p class="company-section__body">
        This placeholder can later be replaced with your actual test procedures,
        measurement tolerances, and any certifications or standards that the
        factory follows.
      </p>
    </section>

    <div class="company-feedback">
      <UserFeedbackThread
        threadKey="company-ourstory"
        title="Share your feedback about Our Story and the factory"
      />
    </div>

    <div
      v-if="showLaserVideo"
      class="company-video-modal"
      role="dialog"
      aria-modal="true"
    >
      <div class="company-video-modal__backdrop" @click="showLaserVideo = false" />
      <div class="company-video-modal__content">
        <button
          type="button"
          class="company-video-modal__close"
          @click="showLaserVideo = false"
        >
          ×
        </button>
        <video
          class="company-video-modal__video"
          controls
          preload="metadata"
        >
          <source src="/company/aboutus/appearance/tanzanite-carbonrim-appearancelogo.mp4" type="video/mp4" />
        </video>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useHead, definePageMeta, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import GuideImage from '~/components/GuideImage.vue'

type OurStoryTabId = 'factory' | 'appearance' | 'hole-patterns' | 'facility' | 'manufacture' | 'qualitycontrol'

const tabs: { id: OurStoryTabId; label: string }[] = [
  { id: 'factory', label: 'Factory' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'hole-patterns', label: 'Hole Patterns' },
  { id: 'facility', label: 'Facility' },
  { id: 'manufacture', label: 'Manufacture' },
  { id: 'qualitycontrol', label: 'Quality control' },
]

const activeTab = ref<OurStoryTabId>('factory')
const route = useRoute()
const showLaserVideo = ref(false)

const getTabFromHash = (hash: string): OurStoryTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: OurStoryTabId[] = ['factory', 'appearance', 'hole-patterns', 'facility', 'manufacture', 'qualitycontrol']
  return (allowed as string[]).includes(raw) ? (raw as OurStoryTabId) : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: OurStoryTabId) => {
  activeTab.value = id
}

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'About us',
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__title--sr-only {
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

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.company-tabs {
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

.company-tabs::-webkit-scrollbar {
  display: none;
}

.company-tabs__item {
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
  box-shadow: 6px 8px 18px -12px rgba(0, 0, 0, 0.85);
}

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: #ffffff;
  color: #0f172a;
  border: none;
  font-weight: 600;
  box-shadow: 8px 10px 22px -10px rgba(0, 0, 0, 0.9);
}

.company-section {
  margin-top: 0;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section__list li + li {
  margin-top: 0.25rem;
}

.company-video-button {
  margin-top: 0.5rem;
  padding: 0.35rem 0.85rem;
  border-radius: 9999px;
  border: 1px solid rgba(148, 163, 184, 0.6);
  background: rgba(15, 23, 42, 0.9);
  color: #e5e7eb;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}

.company-video-thumbnail {
  margin-top: 0.75rem;
  position: relative;
  border-radius: 0.75rem;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.85);
}

.company-video-thumbnail__image {
  display: block;
  width: 100%;
  height: auto;
}

.company-video-thumbnail__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.85),
    rgba(15, 23, 42, 0.45)
  );
  color: #e5e7eb;
}

.company-video-thumbnail__icon {
  font-size: 1.8rem;
}

.company-video-thumbnail__label {
  font-size: 0.9rem;
  font-weight: 500;
}

.company-video-modal {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: center;
}

.company-video-modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.85);
}

.company-video-modal__content {
  position: relative;
  z-index: 41;
  width: 100%;
  max-width: 960px;
  margin: 0 1rem;
  background: #020617;
  border-radius: 0.75rem;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.85);
  overflow: hidden;
}

.company-video-modal__close {
  position: absolute;
  top: 0.35rem;
  right: 0.6rem;
  border: none;
  background: transparent;
  color: #e5e7eb;
  font-size: 1.25rem;
  cursor: pointer;
}

.company-video-modal__video {
  display: block;
  width: 100%;
  height: auto;
}

.company-appearance-images {
  margin-top: 0.75rem;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.75rem;
}

.factory-flow-grid {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.factory-flow-card {
  position: relative;
  aspect-ratio: 1 / 1;
  border-radius: 0.9rem;
  overflow: hidden;
  background: #020617;
  box-shadow: 4px 4px 0 rgba(0, 0, 0, 1);
}

.factory-flow-card__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform-origin: center;
  transition: transform 0.35s ease;
}

.factory-flow-card:hover .factory-flow-card__image {
  transform: scale(1.05);
}

.factory-flow-card__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 0.75rem 0.8rem 0.65rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.96) 0%,
    rgba(15, 23, 42, 0.9) 35%,
    rgba(15, 23, 42, 0.6) 55%,
    transparent 100%
  );
  color: #f9fafb;
}

.factory-flow-card__step {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #22d3ee;
}

.factory-flow-card__step-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: radial-gradient(circle at 30% 30%, #a5f3fc, #22d3ee 55%, #0ea5e9 100%);
  box-shadow: 0 0 9px rgba(56, 189, 248, 0.9);
}

.factory-flow-card__title {
  margin: 0 0 0.1rem;
  font-size: 0.9rem;
  font-weight: 600;
}

.factory-flow-card__text {
  margin: 0;
  font-size: 0.78rem;
  color: rgba(226, 232, 240, 0.9);
}

.company-section--values {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--timeline {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--cta {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-values {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.company-values__item {
  padding: 0.75rem 0.9rem;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.24);
}

.company-values__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-timeline {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.company-timeline__item {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  align-items: flex-start;
}

.company-timeline__year {
  font-size: 0.85rem;
  font-weight: 600;
  color: #e5e7eb;
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-feedback {
  margin-top: 2.5rem;
}

@media (min-width: 768px) {
  .company-page__title {
    font-size: 1.75rem;
  }

  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .company-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .company-tabs {
    justify-content: flex-start;
  }

  .factory-flow-grid {
    grid-template-columns: 1fr;
  }
}
</style>
