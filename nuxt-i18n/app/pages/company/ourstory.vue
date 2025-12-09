<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">Our Story</h1>

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
          More than 20+experienced engineers in our R&amp;D Department and 250+ effective workers,with the
          sufficient manufacturing facilities,makes monthly capacity of 25,000 pcs rims possible.With
          strict high-quality standards on the manufacturing processes and 100% quality test&amp;check to
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
          rims &amp;wheels.
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
              Step 06 - Premold &amp; Layup
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
              Step 18 - Inspection &amp; Packing
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
      <h3 class="company-section__title">Appearance</h3>
      <p class="company-section__body">
        Use this tab to highlight the visual details of Tanzanite rims and wheelsets: finishes, decals, logo styles, and paint or paintless options.
      </p>
      <p class="company-section__body">
        Later you can replace this placeholder with your own photos and descriptions of different appearance packages so riders can quickly understand what their wheels will look like on the bike.
      </p>
    </section>

    <section
      v-show="activeTab === 'facility'"
      id="facility"
      class="company-section"
    >
      <h3 class="company-section__title">Facility</h3>
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
      <h3 class="company-section__title">Manufacture</h3>
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
      <h3 class="company-section__title">Quality control</h3>
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
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useHead, definePageMeta } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

type OurStoryTabId = 'factory' | 'appearance' | 'facility' | 'manufacture' | 'qualitycontrol'

const tabs: { id: OurStoryTabId; label: string }[] = [
  { id: 'factory', label: 'Factory' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'facility', label: 'Facility' },
  { id: 'manufacture', label: 'Manufacture' },
  { id: 'qualitycontrol', label: 'Quality control' },
]

const activeTab = ref<OurStoryTabId>('factory')

const setActiveTab = (id: OurStoryTabId) => {
  activeTab.value = id
}

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
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
  box-shadow:
    0 3px 9px -6px rgba(0, 0, 0, 0.9),
    0 0 9px rgba(0, 0, 0, 0.85);
}

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
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
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: #020617;
  box-shadow: 0 0 25px rgba(15, 23, 42, 0.9);
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
