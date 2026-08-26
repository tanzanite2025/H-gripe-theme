<template>
  <section id="appearance" class="company-section">
    <!-- Intro Header -->
    <div class="text-center mb-10 max-w-3xl mx-auto">
      <h2 class="text-xl sm:text-2xl font-bold tz-text-primary mb-4">Appearance & Customization</h2>
      <p class="tz-text-secondary">
        We offer a wide range of carbon fiber weave patterns and surface finishes to meet the aesthetic and performance preferences of every rider.
        All weaves are structural carbon fiber, not printed or laminated.
      </p>
    </div>

    <!-- Surface Visuals -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-16">
       <GuideImage
          src="/company/aboutus/appearance/carbon-rim-finish1.webp"
          alt="carbon rim surface finish detail 1"
          :zoomOnClick="true"
          caption="Matte Finish Detail"
          class="rounded-xl overflow-hidden shadow-lg border tz-border-subtle"
        />
        <GuideImage
          src="/company/aboutus/appearance/carbonrims-finish2.webp"
          alt="carbon rim surface finish detail 2"
          :zoomOnClick="true"
          caption="Glossy Finish Detail"
          class="rounded-xl overflow-hidden shadow-lg border tz-border-subtle"
        />
    </div>

    <!-- Weave Patterns -->
    <div class="mb-16">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-1 h-6 bg-emerald-500 rounded-full"></div>
        <h3 class="text-xl font-bold tz-text-primary">Weave Patterns</h3>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="item in weavePatterns" :key="item.name" class="premium-card p-5 group hover:tz-surface-panel transition-colors">
          <div class="flex justify-between items-start mb-2">
            <h4 class="font-bold tz-text-primary group-hover:text-emerald-700 transition-colors">{{ item.name }}</h4>
            <span class="text-xs font-mono uppercase tracking-wider tz-text-muted tz-surface-panel px-2 py-1 rounded border tz-border-subtle">{{ item.finish }}</span>
          </div>
          <p class="text-sm tz-text-secondary">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Finish Options: Gloss & Crystal -->
    <div class="mb-16 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12">
      <!-- Gloss Levels -->
      <div>
        <div class="flex items-center gap-3 mb-6">
          <div class="w-1 h-6 bg-emerald-500 rounded-full"></div>
          <h3 class="text-xl font-bold tz-text-primary">Gloss Levels</h3>
        </div>
        <div class="grid grid-cols-1 gap-3">
          <div v-for="item in glossLevels" :key="item.name" class="premium-card p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
             <div class="font-bold tz-text-primary">{{ item.name }}</div>
             <div class="text-sm tz-text-secondary">{{ item.desc }}</div>
          </div>
        </div>
      </div>

      <!-- Crystal Effects -->
      <div>
        <div class="flex items-center gap-3 mb-6">
          <div class="w-1 h-6 bg-emerald-500 rounded-full"></div>
          <h3 class="text-xl font-bold tz-text-primary">Crystal Effects</h3>
        </div>
        <div class="premium-card p-6">
          <p class="text-sm tz-text-secondary mb-4">Crystal finishes use micro-layer coating technology for unique light reflection and color depth.</p>
          <div class="flex flex-wrap gap-2">
            <span v-for="color in crystalEffects" :key="color" class="px-3 py-1.5 rounded-full text-sm font-medium tz-surface-panel border tz-border-subtle tz-text-secondary">
              {{ color }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Brake Tracks -->
    <div class="mb-16">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-1 h-6 bg-orange-500 rounded-full"></div>
        <h3 class="text-xl font-bold tz-text-primary">Brake Track Options <span class="text-sm font-normal tz-text-muted ml-2">(Rim Brake Only)</span></h3>
      </div>
      <div class="premium-card p-0 overflow-hidden">
         <div class="grid divide-y divide-slate-200">
            <div v-for="track in brakeTracks" :key="track" class="p-4 hover:tz-surface-panel transition-colors tz-text-secondary text-sm">
               {{ track }}
            </div>
         </div>
         <div class="tz-surface-panel p-4 text-xs tz-text-muted border-t tz-border-subtle">
           All brake tracks are heat-treated and abrasion-tested for compatibility with mainstream brake pads.
         </div>
      </div>
    </div>

    <!-- Laser Engraving -->
    <div class="mb-16 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12 items-center">
      <div>
         <div class="flex items-center gap-3 mb-4">
            <div class="w-1 h-6 bg-emerald-400 rounded-full"></div>
            <h3 class="text-xl font-bold tz-text-primary">Laser Engraving</h3>
         </div>
        <div class="prose prose-invert prose-sm max-w-none tz-text-secondary">
           <p>In addition to traditional decals and painted graphics, we support laser-engraved logos as a <strong>sustainable, faster</strong> option.</p>
           <ul class="my-4 space-y-2 list-none pl-0">
             <li class="flex gap-2">
               <span class="text-emerald-600">✔</span> Reduces plastic waste (no vinyl decals/masks).
             </li>
             <li class="flex gap-2">
               <span class="text-emerald-600">✔</span> Shortens lead time by 5-7 days.
             </li>
             <li class="flex gap-2">
               <span class="text-emerald-600">✔</span> Permanent, durable finish.
             </li>
           </ul>
           <p class="text-xs tz-surface-panel p-3 rounded border tz-border-subtle">
             See our <strong>Wheelset Buyers Guide</strong> for more advice on choosing between laser and decals.
           </p>
         </div>
      </div>
      
      <!-- Video Thumbnail -->
      <div class="premium-card p-2 group cursor-pointer" @click="showLaserVideo = true">
         <div class="relative rounded-lg overflow-hidden aspect-video tz-surface-panel">
            <img
              src="/company/aboutus/appearance/carbon-rim-laser-engraving-LOGO.webp"
              alt="Laser engraving video"
              class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity duration-300"
              loading="lazy"
            />
            <div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/10 transition-colors">
               <div class="w-14 h-14 flex items-center justify-center rounded-full bg-emerald-500 tz-text-primary pl-1 shadow-[0_0_30px_rgba(20,184,166,0.5)] group-hover:scale-110 transition-transform">
                 <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
            <div class="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/90 to-transparent">
              <span class="text-sm font-medium tz-text-primary">Watch Process Video</span>
            </div>
         </div>
      </div>
    </div>

    <!-- Presentation Suggestions (Tips) -->
    <div class="premium-card p-6 border-l-4 border-l-emerald-500 bg-emerald-50">
      <h3 class="text-lg font-bold tz-text-primary mb-3">Presentation Suggestions</h3>
      <ul class="space-y-2 text-sm tz-text-secondary">
        <li>• Organize options into three clear sections: Weave, Brake Track, Finish.</li>
        <li>• Pair each option with a high-resolution image and concise label.</li>
        <li>• Include “Recommended combinations” or “Customer favorites” to guide decision-making.</li>
      </ul>
      <p class="mt-4 text-xs tz-text-muted pt-3 border-t tz-border-subtle">
        For OEM clients, team orders, or custom branding requests, we offer logo integration, laser etching, and color matching services. Contact us for tailored solutions.
      </p>
    </div>

    <!-- Video Modal -->
    <div
      v-if="showLaserVideo"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 tz-mobile-safe-modal-mask"
      role="dialog"
      aria-modal="true"
    >
      <div class="tz-standard-modal-backdrop absolute inset-0" @click="showLaserVideo = false" />
      <div class="appearance-video-modal-content tz-standard-modal-surface relative w-full max-w-4xl tz-surface-card overflow-hidden ring-1 ring-black/10">
        <button
          type="button"
          class="tz-global-close-btn absolute top-4 right-4 z-10"
          @click="showLaserVideo = false"
        >
          ×
        </button>
        <video
          class="w-full h-auto aspect-video"
          controls
          autoplay
        >
          <source src="/company/aboutus/appearance/carbonrim-appearancelogo.mp4" type="video/mp4" />
        </video>
      </div>
    </div>

  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import GuideImage from '~/components/GuideImage.vue'

const showLaserVideo = ref(false)

const weavePatterns = [
  { name: 'UD (Unidirectional)', finish: 'Matte', desc: 'Minimal straight grain, ultra-light.' },
  { name: 'Butterfly', finish: 'Matte', desc: 'Unique organic pattern, high visual impact.' },
  { name: 'Marble', finish: 'Matte', desc: 'Stone-like texture, bold and distinctive.' },
  { name: '3K', finish: 'Satin', desc: 'Classic checkerboard, balanced look.' },
  { name: '3K Twill (Vertical)', finish: 'Satin', desc: 'Diagonal weave, dynamic appearance.' },
  { name: '3K Twill (Horizontal)', finish: 'Satin', desc: 'Horizontal twill, modern aesthetic.' },
  { name: '12K', finish: 'Satin', desc: 'Coarse weave, powerful visual texture.' },
  { name: '12K Twill (Vertical)', finish: 'Satin', desc: 'Bold diagonal weave, vertical layout.' },
  { name: '12K Twill (Horizontal)', finish: 'Satin', desc: 'Horizontal twill, aggressive styling.' },
]

const glossLevels = [
  { name: 'Matte', desc: 'Subtle, scratch-resistant.' },
  { name: 'Satin', desc: 'Soft sheen, balanced texture.' },
  { name: 'Glossy', desc: 'High shine, premium look.' },
  { name: 'Paintless', desc: 'Raw carbon fiber, ultra-lightweight.' },
]

const crystalEffects = [
  'Black', 'Silver', 'Orange', 'Purple', 'Blue'
]

const brakeTracks = [
  'High TG Resin UD (UD weave only)',
  'High TG Resin Graphene',
  'AC3 Grooved (two styles)',
  'UD C-Grooved (UD weave only)',
  'C-Grooved Graphene'
]
</script>

<style scoped>
.appearance-video-modal-content {
  max-height: var(--tz-mobile-safe-viewport-height, 100vh);
}
</style>
