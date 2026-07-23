<template>
  <section id="factory" class="company-section">
    <!-- Intro Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-12 lg:mb-16">
      <div v-for="item in factoryIntro" :key="item.title" class="premium-card p-6 flex flex-col gap-3">

        <div>
          <h3 class="text-lg font-bold text-white mb-2">{{ item.title }}</h3>
          <p class="text-sm tz-text-secondary leading-relaxed">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Process Section Header -->
    <div class="text-center mb-8 lg:mb-12 relative">
        <div class="inline-flex items-center justify-center w-12 h-1 mb-4 rounded-full bg-gradient-to-r from-teal-500 to-blue-500"></div>
        <h2 class="text-2xl sm:text-3xl font-bold text-white mb-2">Manufacturing Process</h2>
        <p class="tz-text-secondary max-w-2xl mx-auto">From raw material to final packing—precision at every step.</p>
    </div>

    <!-- Process Steps Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <article 
        v-for="step in factorySteps" 
        :key="step.step" 
        class="premium-card overflow-hidden group flex flex-col h-full"
      >
          <!-- Image -->
          <div class="relative aspect-[4/3] overflow-hidden bg-slate-800 shrink-0">
            <img 
              :src="step.img" 
              :alt="step.fullTitle" 
              class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105" 
              loading="lazy"
              decoding="async"
            >
            <div class="absolute inset-0 bg-gradient-to-t from-slate-900/40 to-transparent"></div>
            <!-- Step Badge -->
            <div class="absolute top-3 left-3 bg-slate-900/80 backdrop-blur border border-white/10 text-white text-xs font-mono font-bold px-2 py-1 rounded">
              Step {{ step.step }}
            </div>
          </div>
          
          <!-- Content -->
          <div class="p-5 flex flex-col flex-1">
            <h3 class="text-base font-bold text-white mb-2 group-hover:text-teal-400 transition-colors">
              {{ step.fullTitle }}
            </h3>
            <p class="text-sm tz-text-secondary leading-relaxed">
              {{ step.desc }}
            </p>
          </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
const factoryIntro = [
  {
    title: 'History & Expertise',
    desc: 'Began in 2013 in Xiamen, China. We specialize in manufacturing carbon fiber bicycle rims and wheels, successfully innovating and industrialising the supply of lightweight carbon fibre rims to the bicycle industry.'
  },
  {
    title: 'Scale & Capacity',
    desc: 'More than 20+ experienced engineers in R&D and 250+ effective workers. Monthly capacity of 25,000 pcs rims. 100% quality test & check guarantees all rims meet UCI/ISO standards.'
  },
  {
    title: 'Innovation & Tech',
    desc: 'Multiple patented technologies in production. We use mostly Toray carbon fiber (T700 mixed with T800/T1000) to balance weight, stiffnes, and durability.'
  },
  {
    title: 'Industry Leader',
    desc: 'A leader in carbon fiber rim industry. We have built strong relationships with famous bicycle brands worldwide as a trusted OEM/ODM supplier.'
  }
]

const factorySteps = [
  { step: '01', title: 'Epoxy Resin Prep', fullTitle: 'Prepare and control epoxy resin', desc: 'We purchase raw materials from reliable suppliers and produce the optimal epoxy resin for carbon rims.', img: '/company/ourstory/factory/tanzanite-fanctory-EpoxyResinWorkshop1.webp' },
  { step: '02', title: 'Carbon Prepregs', fullTitle: 'Produce carbon prepregs', desc: 'Different kinds of carbon fibers from Toray Japan. The carbon prepregs are carbon fibers saturated with epoxy resin through the weaving machines.', img: '/company/ourstory/factory/tanzanite-factory-carbonprepregsworkshop2.webp' },
  { step: '03', title: 'Material Storage', fullTitle: 'Store prepregs at stable temperature', desc: 'Carbon prepregs are saturated with epoxy resin and kept in optimal condition at -5°C, a stable temperature for the resin.', img: '/company/ourstory/factory/tanzanite-factory-carbonprepregsstoreroom3.webp' },
  { step: '04', title: 'Precision Cutting', fullTitle: 'Cut carbon sheets to precise angles', desc: 'Sheets are cut into desired angles of 0, ±45, ±30, and ±90 degrees, then precisely machine-cut into pieces.', img: '/company/ourstory/factory/tanzanite-factory-cuttingworkshop4.webp' },
  { step: '05', title: 'Lamination Prep', fullTitle: 'Prepare laminations for rim sections', desc: 'Lamination prepares different parts of the rim with carbon sheets so we can arrange the product more easily and balance rim weight, durability and rigidity.', img: '/company/ourstory/factory/tanzanite-factory-laminationpreparationworkshop5.webp' },
  { step: '06', title: 'Premold & Layup', fullTitle: 'Premold and hand layup', desc: 'Technicians hand lay fabric layers from outer to inner, smoothing each overlap to avoid wrinkles, fractures and air bubbles before closing the premold.', img: '/company/ourstory/factory/tanzanite-factory-premoldlayupworkshop6.webp' },
  { step: '07', title: 'Molding & Curing', fullTitle: 'Cure rims under heat and pressure', desc: 'Complete molds are moved to molding machines where heat and pressure are accurately controlled and monitored in curing ovens.', img: '/company/ourstory/factory/tanzanite-factory-moldingworkshop7.webp' },
  { step: '08', title: 'Visual Inspection', fullTitle: 'Inspect every rim visually', desc: 'All rims are 100% inspected for appearance and obvious defects before moving on.', img: '/company/ourstory/factory/tanzanite-factory-Appearance%20inspection8.webp' },
  { step: '09', title: 'CNC Machining', fullTitle: 'Drill spoke and valve holes', desc: 'Rims are drilled on aviation-grade CNC machines using bits designed for composites to ensure correct angles and clean holes.', img: '/company/ourstory/factory/tanzanite-factory-cncmachiningworkshop9.webp' },
  { step: '10', title: 'Weight Check', fullTitle: 'Control rim weight precisely', desc: 'Weight standards are strictly controlled, with measurements accurate to 1 gram.', img: '/company/ourstory/factory/tanzanite-factory-Weightdetection10.webp' },
  { step: '11', title: 'Sanding', fullTitle: 'Prepare surfaces by sanding', desc: 'After inspection, rims are transferred to the sanding workshop to prepare surfaces for the next processes.', img: '/company/ourstory/factory/tanzanite-factory-sandingworshop11.webp' },
  { step: '12', title: 'Grinding', fullTitle: 'Refine critical surfaces', desc: 'After sanding, rims move to grinding to refine key surfaces and ensure consistency.', img: '/company/ourstory/factory/tanzanite-factory-grinding12.webp' },
  { step: '13', title: 'Detailing', fullTitle: 'Handle fine details', desc: 'Details determine success or failure. A dedicated workshop focuses on rim details to keep quality high.', img: '/company/ourstory/factory/tanzanite-factory-detailstreatment13.webp' },
  { step: '14', title: 'Final Tests', fullTitle: 'Perform additional tests', desc: 'Drainage holes, spoke holes and appearance are checked again before shipment.', img: '/company/ourstory/factory/tanzanite-factory-otherdetection14.webp' },
  { step: '15', title: 'Painting', fullTitle: 'Apply finishes where required', desc: 'Ultralight rims are paintless, while traditional rims receive glossy, matte or custom finishes in the painting workshop.', img: '/company/ourstory/factory/tanzantie-factory-painting15.webp' },
  { step: '16', title: 'Decals', fullTitle: 'Apply waterslide decals', desc: 'After painting, rims that require waterslide decals are processed here for branding and graphics.', img: '/company/ourstory/factory/tananite-factory-logolidedecals16.webp' },
  { step: '17', title: 'Laser Engraving', fullTitle: 'Laser carve logos and marks', desc: 'Rims that require laser-carved logos are processed here for a durable, precise finish.', img: '/company/ourstory/factory/tanzanite-factory-lasercarving17.webp' },
  { step: '18', title: 'Packing', fullTitle: 'Final inspection and packing', desc: 'All rims are inspected again, protected with foam bags and carefully packed into cartons.', img: '/company/ourstory/factory/tanzanite-factory-inspectionpacking18.webp' },
  { step: '19', title: 'Warehouse', fullTitle: 'Store rims in organized warehouse', desc: 'Compliant operation and clear classification in the warehouse ensure model and rule accuracy.', img: '/company/ourstory/factory/tanzanite-factory-ourwarehouse19.webp' },
]
</script>
