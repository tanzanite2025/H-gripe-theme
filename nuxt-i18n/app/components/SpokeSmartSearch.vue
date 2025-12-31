<template>
  <div class="spoke-smart-search w-full max-w-2xl mx-auto">
    <!-- Search Header -->
    <div class="relative mb-6">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <svg class="h-5 w-5 text-slate-400" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
        </svg>
      </div>
      <input
        v-model="query"
        type="text"
        placeholder="Type a hub model (e.g. '350', '240', 'Mavic')..."
        class="block w-full pl-10 pr-4 py-3 bg-[radial-gradient(ellipse_at_top,rgba(30,41,59,0.95),rgba(15,23,42,0.98))] border border-slate-700/50 rounded-xl text-slate-50 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-500/50 shadow-[0_4px_16px_-4px_rgba(0,0,0,0.5)] transition-all duration-300"
      />
      <!-- Build Count Badge -->
      <div v-if="query.length > 1 && matchingConfigs.length > 0" class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-sky-900/50 text-sky-400 border border-sky-800/30">
          {{ matchingConfigs.length }} builds found
        </span>
      </div>
    </div>

    <!-- Results List -->
    <TransitionGroup 
      name="list-results" 
      tag="div" 
      class="space-y-4"
    >
      <div 
        v-for="config in matchingConfigs" 
        :key="config.id"
        class="group relative overflow-hidden rounded-xl bg-slate-900/50 border border-slate-800/50 hover:bg-slate-900/80 hover:border-slate-700/80 hover:shadow-[0_8px_30px_rgb(0,0,0,0.3)] transition-all duration-300"
      >
        <!-- Background Gradient Glow -->
        <div class="absolute -inset-0.5 bg-gradient-to-r from-sky-500/10 via-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500 blur-md pointer-events-none"></div>

        <div class="relative p-5 grid gap-4 md:grid-cols-[1fr,auto]">
          <!-- Setup Info -->
          <div>
            <h3 class="text-sm font-semibold text-slate-200 mb-1 group-hover:text-sky-300 transition-colors">
              {{ config.name }}
            </h3>
            <p v-if="config.description" class="text-xs text-slate-400 mb-3 line-clamp-1">
              {{ config.description }}
            </p>
            
            <!-- Specs Tag Cloud -->
            <div class="flex flex-wrap gap-2">
              <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-400 border border-slate-700">
                {{ config.spokeCount }}H
              </span>
              <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-400 border border-slate-700">
                {{ config.crossing }}X
              </span>
               <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-400 border border-slate-700">
                {{ config.nippleType === 'hidden' ? 'Hidden Nipple' : 'Standard Nipple' }}
              </span>
            </div>
          </div>

          <!-- Calculated Results -->
          <div class="flex items-center gap-6 border-t border-slate-800 pt-4 md:border-t-0 md:border-l md:pl-6 md:pt-0">
             <!-- Left -->
             <div class="text-center">
                <div class="text-[10px] uppercase tracking-wider text-slate-500 mb-0.5">Left</div>
                <div class="text-lg font-mono font-bold text-sky-400">
                   {{ calculateResult(config).left }}<span class="text-xs text-sky-600/70 ml-0.5">mm</span>
                </div>
             </div>
             <!-- Right -->
             <div class="text-center">
                <div class="text-[10px] uppercase tracking-wider text-slate-500 mb-0.5">Right</div>
                 <div class="text-lg font-mono font-bold text-emerald-400">
                   {{ calculateResult(config).right }}<span class="text-xs text-emerald-600/70 ml-0.5">mm</span>
                </div>
             </div>
          </div>
        </div>
      </div>
      
      <!-- Empty State -->
      <div 
        v-if="query.length > 1 && matchingConfigs.length === 0"
        class="text-center py-12 rounded-xl border border-dashed border-slate-800 text-slate-500"
      >
        <p class="text-sm">No verified builds found for "{{ query }}".</p>
        <p class="text-xs mt-1 text-slate-600">Try searching for generic terms like "350" or "Mavic".</p>
      </div>

    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { PRESET_BUILDS, RIM_DATABASE, HUB_DATABASE, type WheelBuildPreset } from '~/data/spoke-calculator/database'
import { computeSpokeLength } from '~/utils/spokeMath'

const query = ref('')

const matchingConfigs = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (q.length < 2) return [] // Minimum 2 chars to search

  return PRESET_BUILDS.filter(preset => {
    // Match name or keywords
    const matchName = preset.name.toLowerCase().includes(q)
    const matchKeywords = preset.keywords.some(k => k.toLowerCase().includes(q))
    return matchName || matchKeywords
  })
})

function calculateResult(preset: WheelBuildPreset) {
  // 1. Find Rim Data
  const rimBrand = RIM_DATABASE.find(b => b.id === preset.rimBrandId)
  const rimModel = rimBrand?.items.find(r => r.id === preset.rimModelId)

  // 2. Find Hub Data
  const hubBrand = HUB_DATABASE.find(b => b.id === preset.hubBrandId)
  const hubModel = hubBrand?.items.find(h => h.id === preset.hubModelId)

  if (!rimModel || !hubModel) {
    return { left: 'N/A', right: 'N/A' }
  }

  // Determine Hub Geometry (Front vs Rear)
  // Usually the Preset ID/Name implies Front/Rear, but let's try to auto-detect
  // For this Demo, we assume the hub model has 'front' or 'rear' properties
  // And we prioritise Front geometry if the preset implies Front, or Rear otherwise.
  // Actually, let's simplify: check if it has Front or Rear geometry available.
  
  // Logic: In a real app, the Preset should specify if it's a Front or Rear build explicitly.
  // We didn't add that field to WheelBuildPreset in step 1, but we can infer or default.
  // Let's assume Front for now if not specified, OR try to match based on hub capabilities.
  
  // Improved Logic: Check what geometry the hub HAS.
  let geo = hubModel.front
  if (!geo && hubModel.rear) geo = hubModel.rear
  // If both exist (e.g. disk hubs often come in pair sets, but usually separate IDs),
  // ideally the preset ID tells us. e.g. 'tz_ar45_dt350_fr' -> Front?
  // Let's fallback to Front for this demo if ambiguous.
  
  if (!geo) return { left: '?', right: '?' }

  const leftLen = computeSpokeLength(
    rimModel.erd,
    geo.leftFlangePcd,
    geo.leftFlange,
    preset.spokeCount,
    preset.crossing,
    preset.nippleType,
    preset.nippleLength
  )

  const rightLen = computeSpokeLength(
    rimModel.erd,
    geo.rightFlangePcd,
    geo.rightFlange,
    preset.spokeCount,
    preset.crossing,
    preset.nippleType,
    preset.nippleLength
  )

  return { left: leftLen.toFixed(1), right: rightLen.toFixed(1) }
}
</script>

<style scoped>
.list-results-enter-active,
.list-results-leave-active {
  transition: all 0.3s ease;
}
.list-results-enter-from,
.list-results-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
