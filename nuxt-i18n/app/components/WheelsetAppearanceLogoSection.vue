<template>
  <div class="space-y-6">
    <!-- Intro / Recommendation Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-sky-500">
       <h3 class="text-lg font-bold text-slate-100 mb-3">
         Why we recommend laser-engraved logos
       </h3>
       <p class="text-sm text-slate-400 leading-relaxed max-w-2xl mx-auto mb-4">
          For most standard orders, <strong>Laser Engraving</strong> is our top recommendation. 
          It's eco-friendly, permanent, and significantly faster to produce—shortening lead times by <strong>5–7 days</strong> compared to custom paint or water decals.
       </p>
       
       <button
        type="button"
        class="inline-flex items-center gap-2 px-6 py-2 rounded-full bg-slate-800 hover:bg-slate-700 text-sky-400 text-xs font-bold uppercase tracking-wider transition-all shadow-lg hover:shadow-sky-500/20"
        @click="goToAboutAppearance"
      >
        <span>Explore Weave & Finish Options</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
      </button>
    </div>

    <!-- Options Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- Option 1: Custom Decals (Violet) -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300 group">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-violet-500/10">
            <h3 class="text-base font-bold text-slate-200">Customized Decals</h3>
          </div>
          
          <p class="text-sm text-slate-400 leading-relaxed mb-4">
             Our graphics department can execute any vision using:
          </p>
          <ul class="space-y-2 mb-4">
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-violet-500"></span> Waterslide Decals (complex colors)
             </li>
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-violet-500"></span> Removable Vinyl Stickers
             </li>
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-violet-500"></span> Custom Fonts & Locations
             </li>
          </ul>
          <p class="text-xs text-slate-500 italic">
             Ideal for bold statements and specific team colors.
          </p>
       </div>

       <!-- Option 2: Laser Engraving (Teal) -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300 group">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-teal-500/10">
            <h3 class="text-base font-bold text-slate-200">Laser Engraving</h3>
          </div>
          
          <div class="absolute top-5 right-5">
             <span class="inline-flex items-center gap-1 px-2 py-1 rounded text-[10px] font-bold uppercase tracking-wider bg-teal-500/10 text-teal-400 border border-teal-500/20">
                <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z"/></svg>
                Eco-Friendly
             </span>
          </div>

          <p class="text-sm text-slate-400 leading-relaxed mb-4">
             Our sustainability choice. Precision etching in a sleek <span class="text-slate-200">light gray</span> tone.
          </p>
          <ul class="space-y-2 mb-4">
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-teal-500"></span> Preserves structural integrity
             </li>
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-teal-500"></span> No plastic/vinyl waste
             </li>
             <li class="flex items-center gap-2 text-xs text-slate-300">
                <span class="w-1.5 h-1.5 rounded-full bg-teal-500"></span> Fastest production time
             </li>
          </ul>
          <p class="text-xs text-slate-500 italic">
             Subtle, premium, and permanent.
          </p>
       </div>
    </div>

    <!-- Gallery 1: Laser Examples -->
    <div>
        <h4 class="text-xs font-bold text-slate-500 uppercase tracking-widest mb-3 pl-1">Laser Engraving Gallery</h4>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO.webp"
            alt="Carbon rim with customized laser engraved logo detail"
            :zoomOnClick="true"
            caption="Detail View"
            class="rounded-xl overflow-hidden shadow-lg"
          />
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO1.webp"
            alt="Close-up of carbon rim laser engraving process"
            :zoomOnClick="true"
            caption="Process Close-up"
            class="rounded-xl overflow-hidden shadow-lg"
          />
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO2.webp"
            alt="Different examples of carbon rim laser engraved graphics"
            :zoomOnClick="true"
            caption="Finished Effect"
            class="rounded-xl overflow-hidden shadow-lg"
          />
        </div>
    </div>

    <!-- Advanced Finishes Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 mt-6">
       <div class="flex items-center justify-between mb-4 pb-3 border-b border-slate-700/50">
          <h3 class="text-base font-bold text-slate-200">Pro Options: Weave & Finish</h3>
          <span class="text-[10px] bg-slate-800 text-slate-400 px-2 py-1 rounded border border-slate-700">For OEM / Teams</span>
       </div>
       
       <p class="text-sm text-slate-400 mb-6">
          Planning a long-term build or team fleet? Go beyond standard logos with custom carbon weaves and surface treatments.
       </p>

       <div class="grid gap-4 sm:grid-cols-2">
          <GuideImage
            src="/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp"
            alt="Tanzanite carbon rim surface finish detail 1"
            :zoomOnClick="true"
            caption="Matte Finish"
            class="rounded-xl overflow-hidden border border-slate-700/50"
          />
          <GuideImage
            src="/company/aboutus/appearance/tanzanite-carbonrims-finish2.webp"
            alt="Tanzanite carbon rim surface finish detail 2"
            :zoomOnClick="true"
            caption="Glossy Finish"
            class="rounded-xl overflow-hidden border border-slate-700/50"
          />
       </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

const { goToAboutAppearance } = defineProps<{
  goToAboutAppearance: () => void | Promise<void>
}>()
</script>
