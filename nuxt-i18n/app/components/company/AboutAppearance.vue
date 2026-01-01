<template>
  <section id="appearance" class="company-section">
    <!-- Intro Header -->
    <div class="text-center mb-10 max-w-3xl mx-auto">
      <h2 class="text-xl sm:text-2xl font-bold text-white mb-4">Appearance & Customization</h2>
      <p class="text-slate-400">
        We offer a wide range of carbon fiber weave patterns and surface finishes to meet the aesthetic and performance preferences of every rider.
        All weaves are structural carbon fiber, not printed or laminated.
      </p>
    </div>

    <!-- Surface Visuals -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-16">
       <GuideImage
          src="/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp"
          alt="Tanzanite carbon rim surface finish detail 1"
          :zoomOnClick="true"
          caption="Matte Finish Detail"
          class="rounded-xl overflow-hidden shadow-2xl border border-slate-800"
        />
        <GuideImage
          src="/company/aboutus/appearance/tanzanite-carbonrims-finish2.webp"
          alt="Tanzanite carbon rim surface finish detail 2"
          :zoomOnClick="true"
          caption="Glossy Finish Detail"
          class="rounded-xl overflow-hidden shadow-2xl border border-slate-800"
        />
    </div>

    <!-- Weave Patterns -->
    <div class="mb-16">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-1 h-6 bg-teal-500 rounded-full"></div>
        <h3 class="text-xl font-bold text-white">Weave Patterns</h3>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="item in weavePatterns" :key="item.name" class="premium-card p-5 group hover:bg-slate-800/80 transition-colors">
          <div class="flex justify-between items-start mb-2">
            <h4 class="font-bold text-white group-hover:text-teal-400 transition-colors">{{ item.name }}</h4>
            <span class="text-xs font-mono uppercase tracking-wider text-slate-500 bg-slate-900 px-2 py-1 rounded border border-slate-800">{{ item.finish }}</span>
          </div>
          <p class="text-sm text-slate-400">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Finish Options: Gloss & Crystal -->
    <div class="mb-16 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12">
      <!-- Gloss Levels -->
      <div>
        <div class="flex items-center gap-3 mb-6">
          <div class="w-1 h-6 bg-blue-500 rounded-full"></div>
          <h3 class="text-xl font-bold text-white">Gloss Levels</h3>
        </div>
        <div class="grid grid-cols-1 gap-3">
          <div v-for="item in glossLevels" :key="item.name" class="premium-card p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
             <div class="font-bold text-white">{{ item.name }}</div>
             <div class="text-sm text-slate-400">{{ item.desc }}</div>
          </div>
        </div>
      </div>

      <!-- Crystal Effects -->
      <div>
        <div class="flex items-center gap-3 mb-6">
          <div class="w-1 h-6 bg-purple-500 rounded-full"></div>
          <h3 class="text-xl font-bold text-white">Crystal Effects</h3>
        </div>
        <div class="premium-card p-6">
          <p class="text-sm text-slate-400 mb-4">Crystal finishes use micro-layer coating technology for unique light reflection and color depth.</p>
          <div class="flex flex-wrap gap-2">
            <span v-for="color in crystalEffects" :key="color" class="px-3 py-1.5 rounded-full text-sm font-medium bg-slate-900 border border-slate-700 text-slate-300">
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
        <h3 class="text-xl font-bold text-white">Brake Track Options <span class="text-sm font-normal text-slate-500 ml-2">(Rim Brake Only)</span></h3>
      </div>
      <div class="premium-card p-0 overflow-hidden">
         <div class="grid divide-y divide-slate-800">
            <div v-for="track in brakeTracks" :key="track" class="p-4 hover:bg-slate-800/50 transition-colors text-slate-300 text-sm">
               {{ track }}
            </div>
         </div>
         <div class="bg-slate-900/50 p-4 text-xs text-slate-500 border-t border-slate-800">
           All brake tracks are heat-treated and abrasion-tested for compatibility with mainstream brake pads.
         </div>
      </div>
    </div>

    <!-- Laser Engraving -->
    <div class="mb-16 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12 items-center">
      <div>
         <div class="flex items-center gap-3 mb-4">
            <div class="w-1 h-6 bg-teal-400 rounded-full"></div>
            <h3 class="text-xl font-bold text-white">Laser Engraving</h3>
         </div>
         <div class="prose prose-invert prose-sm max-w-none text-slate-400">
           <p>In addition to traditional decals and painted graphics, we support laser-engraved logos as a <strong>sustainable, faster</strong> option.</p>
           <ul class="my-4 space-y-2 list-none pl-0">
             <li class="flex gap-2">
               <span class="text-teal-400">✔</span> Reduces plastic waste (no vinyl decals/masks).
             </li>
             <li class="flex gap-2">
               <span class="text-teal-400">✔</span> Shortens lead time by 5-7 days.
             </li>
             <li class="flex gap-2">
               <span class="text-teal-400">✔</span> Permanent, durable finish.
             </li>
           </ul>
           <p class="text-xs bg-slate-800/50 p-3 rounded border border-slate-700/50">
             See our <strong>Wheelset Buyers Guide</strong> for more advice on choosing between laser and decals.
           </p>
         </div>
      </div>
      
      <!-- Video Thumbnail -->
      <div class="premium-card p-2 group cursor-pointer" @click="showLaserVideo = true">
         <div class="relative rounded-lg overflow-hidden aspect-video bg-slate-900">
            <img
              src="/company/aboutus/appearance/tanzanite-carbon-rim-laser-engraving-LOGO.webp"
              alt="Laser engraving video"
              class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity duration-300"
              loading="lazy"
            />
            <div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/10 transition-colors">
               <div class="w-14 h-14 flex items-center justify-center rounded-full bg-teal-500 text-white pl-1 shadow-[0_0_30px_rgba(20,184,166,0.5)] group-hover:scale-110 transition-transform">
                 <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
            <div class="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/90 to-transparent">
              <span class="text-sm font-medium text-white">Watch Process Video</span>
            </div>
         </div>
      </div>
    </div>

    <!-- Presentation Suggestions (Tips) -->
    <div class="premium-card p-6 border-l-4 border-l-blue-500 bg-blue-500/5">
      <h3 class="text-lg font-bold text-white mb-3">Presentation Suggestions</h3>
      <ul class="space-y-2 text-sm text-slate-300">
        <li>• Organize options into three clear sections: Weave, Brake Track, Finish.</li>
        <li>• Pair each option with a high-resolution image and concise label.</li>
        <li>• Include “Recommended combinations” or “Customer favorites” to guide decision-making.</li>
      </ul>
      <p class="mt-4 text-xs text-slate-500 pt-3 border-t border-slate-700/30">
        For OEM clients, team orders, or custom branding requests, we offer logo integration, laser etching, and color matching services. Contact us for tailored solutions.
      </p>
    </div>

    <!-- Video Modal -->
    <div
      v-if="showLaserVideo"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
    >
      <div class="absolute inset-0 bg-slate-900/90 backdrop-blur-sm" @click="showLaserVideo = false" />
      <div class="relative w-full max-w-4xl bg-black rounded-xl overflow-hidden shadow-2xl ring-1 ring-white/10">
        <button
          type="button"
          class="absolute top-4 right-4 z-10 w-8 h-8 flex items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20 transition-colors"
          @click="showLaserVideo = false"
        >
          ×
        </button>
        <video
          class="w-full h-auto aspect-video"
          controls
          autoplay
        >
          <source src="/company/aboutus/appearance/tanzanite-carbonrim-appearancelogo.mp4" type="video/mp4" />
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
