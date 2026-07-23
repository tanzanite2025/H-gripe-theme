<template>
  <div class="space-y-6">
    <!-- Intro Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center">
       <div class="flex flex-col items-center">
         <p class="tz-text-secondary text-sm leading-relaxed max-w-2xl mx-auto">
            Example riding scenarios where "Mixed Rim Profiles" (different front/rear specs) provide a tactical advantage.
            This approach balances aerodynamics, handling, and durability.
         </p>
         
         <div class="mt-4 flex flex-wrap justify-center gap-3">
             <button
               type="button"
               class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-slate-600 hover:bg-slate-700 tz-text-secondary text-xs font-bold uppercase tracking-wider transition-colors"
               @click="openQuickBuy"
             >
               Quick Custom
             </button>
             <button
               type="button"
               class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-slate-700 hover:bg-slate-600 text-white text-xs font-bold uppercase tracking-wider transition-colors shadow-lg"
               @click="openWhatsAppChat"
             >
               <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
               Consult Expert
             </button>
         </div>
       </div>
    </div>

    <!-- Scenarios Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- 1. Diff Models -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-fuchsia-500/10">
            <h3 class="text-sm font-bold tz-text-primary uppercase tracking-wide">Different Models</h3>
          </div>
          <p class="text-xs tz-text-secondary leading-relaxed">
             Using completely different rim models (e.g., Lightweight Front + Heavy Duty Rear). Common in custom builds to balance specific performance needs.
          </p>
       </div>

       <!-- 2. Diff Heights -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-fuchsia-500/10">
            <h3 class="text-sm font-bold tz-text-primary uppercase tracking-wide">Different Heights</h3>
          </div>
           <p class="text-xs tz-text-secondary leading-relaxed">
             E.g., 50mm Front (handling) + 60mm Rear (aero/stiffness). The most common "Mixed" setup to optimize crosswind stability and speed.
          </p>
       </div>
    </div>

    <!-- Mobile-First Comparison Cards (Replacing Table) -->
    <div class="space-y-4">
       <div class="flex items-center gap-2 mb-2 px-2">
          <div class="h-4 w-1 bg-fuchsia-500 rounded-full"></div>
          <h3 class="text-sm font-bold tz-text-primary uppercase tracking-wider">Configuration Cheatsheet</h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Card 1 -->
          <div class="bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] rounded-xl p-4 border-l-4 border-fuchsia-500/50">
             <div class="flex justify-between items-start mb-2">
                <strong class="text-fuchsia-400 text-sm">Climbing</strong>
                <span class="text-[10px] bg-slate-800 tz-text-muted px-2 py-0.5 rounded">Lightweight</span>
             </div>
             <div class="grid grid-cols-2 gap-4 text-xs tz-text-secondary mb-2">
                <div><span class="block tz-text-muted text-[10px] uppercase">Front</span> 30-40mm</div>
                <div><span class="block tz-text-muted text-[10px] uppercase">Rear</span> 40-50mm</div>
             </div>
             <p class="text-xs tz-text-muted italic">Front for handling/weight, Rear for stiffness.</p>
          </div>

          <!-- Card 2 -->
          <div class="bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] rounded-xl p-4 border-l-4 border-fuchsia-500/50">
             <div class="flex justify-between items-start mb-2">
                <strong class="text-fuchsia-400 text-sm">Aerodynamics</strong>
                <span class="text-[10px] bg-slate-800 tz-text-muted px-2 py-0.5 rounded">Speed</span>
             </div>
             <div class="grid grid-cols-2 gap-4 text-xs tz-text-secondary mb-2">
                <div><span class="block tz-text-muted text-[10px] uppercase">Front</span> 50-60mm</div>
                <div><span class="block tz-text-muted text-[10px] uppercase">Rear</span> 60-80mm</div>
             </div>
             <p class="text-xs tz-text-muted italic">Front balances crosswinds, Rear maxes speed.</p>
          </div>

           <!-- Card 3 -->
           <div class="bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] rounded-xl p-4 border-l-4 border-fuchsia-500/50">
             <div class="flex justify-between items-start mb-2">
                <strong class="text-fuchsia-400 text-sm">Training / Durability</strong>
                <span class="text-[10px] bg-slate-800 tz-text-muted px-2 py-0.5 rounded">Stable</span>
             </div>
             <div class="grid grid-cols-2 gap-4 text-xs tz-text-secondary mb-2">
                <div><span class="block tz-text-muted text-[10px] uppercase">Front</span> Mid Alloy/Carbon</div>
                <div><span class="block tz-text-muted text-[10px] uppercase">Rear</span> High/Wide Profile</div>
             </div>
             <p class="text-xs tz-text-muted italic">Rear bears load, Front ensures comfort.</p>
          </div>

           <!-- Card 4 -->
           <div class="bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] rounded-xl p-4 border-l-4 border-fuchsia-500/50">
             <div class="flex justify-between items-start mb-2">
                <strong class="text-fuchsia-400 text-sm">Mixed / Versatile</strong>
                <span class="text-[10px] bg-slate-800 tz-text-muted px-2 py-0.5 rounded">All-around</span>
             </div>
             <div class="grid grid-cols-2 gap-4 text-xs tz-text-secondary mb-2">
                <div><span class="block tz-text-muted text-[10px] uppercase">Front</span> 40-50mm</div>
                <div><span class="block tz-text-muted text-[10px] uppercase">Rear</span> 55-65mm</div>
             </div>
             <p class="text-xs tz-text-muted italic">A great balance for racing and training.</p>
          </div>
       </div>
    </div>

  </div>
</template>

<script setup lang="ts">
const { openQuickBuy, openWhatsAppChat } = defineProps<{
  openQuickBuy: () => void | Promise<void>
  openWhatsAppChat: () => void | Promise<void>
}>()
</script>

<style scoped>
.guide-section__cta-wrapper {
  color: var(--tz-text-primary);
}

.wheelset-link {
  color: #2dd4bf;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: #60a5fa;
}

.wheelset-inline-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.65rem;
  margin: 0 0.25rem;
  border-radius: 9999px;
  border: none;
  background: rgba(31, 41, 55, 0.65);
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.85);
  transition: background-color 0.15s ease, transform 0.08s ease;
}

.wheelset-inline-button:hover {
  background: rgba(51, 65, 85, 0.75);
}

.wheelset-inline-button:active {
  transform: scale(0.98);
}
</style>
