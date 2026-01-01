<template>
  <section id="hole-patterns" class="company-section">
      <div class="company-section__header flex flex-col items-center justify-center text-center mb-10 w-full px-4">
        <h2 class="sr-only">Hole Patterns</h2>
        <p class="text-[0.95rem] text-slate-400 max-w-4xl w-full text-center leading-relaxed">
          TANZANITE offers a variety of rim drilling patterns to meet different hub geometries and lacing requirements.
        </p>
      </div>

      <!-- Video Section -->
      <div class="flex justify-center mb-10">
         <div class="relative group cursor-pointer rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-700/50 max-w-3xl w-full" @click="showHolePatternVideo = true">
            <video 
              class="w-full h-auto object-cover opacity-90 group-hover:opacity-100 transition-opacity duration-300"
              src="/company/ourstory/holepatterns/TANZANITE-automated-wheel-rimhole-drilling–customizable-Patterns_medium.webm?v=2" 
              muted 
              loop 
              playsinline
              autoplay
            ></video>
            <!-- Play Overlay -->
            <div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/10 transition-colors duration-300">
               <div class="w-16 h-16 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center border border-white/30 group-hover:scale-110 transition-transform duration-300">
                  <svg class="w-8 h-8 text-white ml-1" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
         </div>
      </div>

      <!-- Info Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
         <!-- Card 1: Our Support -->
         <div class="bg-[#11151e] p-6 rounded-2xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50">
            <h3 class="text-lg font-bold text-slate-100 mb-4 border-b border-slate-700/50 pb-2">Our Support</h3>
            <ul class="space-y-3 text-sm text-slate-300">
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-teal-400 shrink-0"></span>
                  <span><strong class="text-white">Symmetrical Drilling:</strong> Suitable for standard hubs, providing balanced spoke tension.</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-teal-400 shrink-0"></span>
                  <span><strong class="text-white">Asymmetrical Drilling:</strong> Optimizes tension between the drive and non-drive sides, enhancing stability.</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-teal-400 shrink-0"></span>
                  <span><strong class="text-white">Custom Hole Counts:</strong> Available in 16, 20, 24, 28, 32, and 36-hole configurations.</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-teal-400 shrink-0"></span>
                  <span><strong class="text-white">Angled Drilling:</strong> Hole angles optimized for specific lacing patterns to reduce spoke stress.</span>
               </li>
            </ul>
         </div>

         <!-- Card 2: Customization Services -->
         <div class="bg-[#11151e] p-6 rounded-2xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50">
            <h3 class="text-lg font-bold text-slate-100 mb-4 border-b border-slate-700/50 pb-2">Customization Services</h3>
            <p class="text-sm text-slate-400 mb-3">We support customers with tailored drilling solutions based on specific needs:</p>
            <ul class="space-y-3 text-sm text-slate-300">
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0"></span>
                  <span>Special hole counts</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0"></span>
                  <span>Custom lacing compatibility</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0"></span>
                  <span>OEM/ODM exclusive designs</span>
               </li>
            </ul>
         </div>

         <!-- Card 3: Advantages -->
         <div class="bg-[#11151e] p-6 rounded-2xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50">
            <h3 class="text-lg font-bold text-slate-100 mb-4 border-b border-slate-700/50 pb-2">Advantages of Automated Machinery</h3>
            <p class="text-sm text-slate-400 mb-3">All drilling processes are carried out using our <strong class="text-white">self-developed automated machinery</strong>, ensuring:</p>
            <ul class="space-y-3 text-sm text-slate-300">
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></span>
                  <span>High-precision hole placement</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></span>
                  <span>Consistency and repeatability</span>
               </li>
               <li class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></span>
                  <span>Fast delivery and scalable production capacity</span>
               </li>
            </ul>
         </div>

         <!-- Card 4: Our Commitment -->
         <div class="bg-[#11151e] p-6 rounded-2xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50">
            <h3 class="text-lg font-bold text-slate-100 mb-4 border-b border-slate-700/50 pb-2">Our Commitment</h3>
             <p class="text-sm text-slate-300 leading-relaxed">
               By combining automation with customization, TANZANITE provides the most suitable rim drilling solutions for different riders and brands, guaranteeing long-term stability and reliability.
             </p>
         </div>
      </div>

    <!-- Video Modal -->
    <div v-if="showHolePatternVideo" class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/90 backdrop-blur-sm p-4" @click="showHolePatternVideo = false">
       <button class="absolute top-4 right-4 text-white hover:text-rose-500 transition-colors" @click="showHolePatternVideo = false">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
       </button>
       <div class="relative w-full max-w-5xl aspect-video bg-black rounded-lg overflow-hidden shadow-2xl" @click.stop>
          <video 
            class="w-full h-full object-contain"
            src="/company/ourstory/holepatterns/TANZANITE-automated-wheel-rimhole-drilling–customizable-Patterns_medium.webm?v=2"
            controls
            autoplay
          ></video>
       </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const showHolePatternVideo = ref(false)
</script>
