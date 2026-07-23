<template>
  <div class="space-y-8">
    <!-- Intro Banner -->
    <div class="relative overflow-hidden rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-6 text-center">
       <div class="flex flex-col items-center">

         <h3 class="text-xl font-bold text-slate-100 mb-2">Professional Wheel Building Service</h3>
         <p class="tz-text-secondary text-sm max-w-2xl mx-auto leading-relaxed">
           We provide professional assembly services, supporting customers who wish to send in hubs or rims.
           Our skilled technicians will complete the assembly to the highest standards.
         </p>
         
         <div class="mt-6 flex flex-wrap justify-center gap-3">
            <div class="inline-flex items-center px-4 py-2 rounded-full border-none bg-sky-500/5 text-sky-400 text-xs font-bold uppercase tracking-wider shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               Assembly Fee: USD 30 / Set
            </div>
            <button
               type="button"
               class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-slate-700 hover:bg-slate-600 text-white text-xs font-bold uppercase tracking-wider transition-colors shadow-lg"
               @click="openWhatsAppChat"
             >
               <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
               Get Quote via Chat
             </button>
         </div>
       </div>
    </div>

    <!-- Process & Logistics Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

       <!-- 1. The Process (Emerald) -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
          <div class="flex items-center gap-3 mb-5 pb-3 border-b border-emerald-500/10">
            <h3 class="text-lg font-bold tz-text-primary">Service Process</h3>
         </div>

         <div class="space-y-6 relative pl-2">
            <!-- Timeline Line -->
            <div class="absolute left-[19px] top-2 bottom-4 w-px bg-slate-700/50"></div>

            <!-- Steps -->
            <div class="relative pl-8">
               <span class="absolute left-0 top-1 w-2.5 h-2.5 rounded-full bg-emerald-500 ring-4 ring-[#11151e]"></span>
               <strong class="block text-emerald-400 text-xs uppercase tracking-wider mb-1">1. Shipping</strong>
               <p class="text-xs tz-text-secondary">Send components to our designated address.</p>
            </div>
            <div class="relative pl-8">
               <span class="absolute left-0 top-1 w-2.5 h-2.5 rounded-full bg-slate-600 ring-4 ring-[#11151e]"></span>
               <strong class="block tz-text-primary text-xs uppercase tracking-wider mb-1">2. Confirmation</strong>
               <p class="text-xs tz-text-secondary">We inspect and confirm receipt of all items.</p>
            </div>
             <div class="relative pl-8">
               <span class="absolute left-0 top-1 w-2.5 h-2.5 rounded-full bg-slate-600 ring-4 ring-[#11151e]"></span>
               <strong class="block tz-text-primary text-xs uppercase tracking-wider mb-1">3. Assembly & Testing</strong>
               <p class="text-xs tz-text-secondary">Professional build + tension/safety analysis.</p>
            </div>
             <div class="relative pl-8">
               <span class="absolute left-0 top-1 w-2.5 h-2.5 rounded-full bg-emerald-500 ring-4 ring-[#11151e]"></span>
               <strong class="block text-emerald-400 text-xs uppercase tracking-wider mb-1">4. Delivery</strong>
               <p class="text-xs tz-text-secondary">Finished wheelset shipped back or ready for pickup.</p>
            </div>
         </div>
       </div>

       <!-- 2. Scope & Notes (Amber) -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
          <div class="flex items-center gap-3 mb-5 pb-3 border-b border-amber-500/10">
            <h3 class="text-lg font-bold tz-text-primary">Scope & Notes</h3>
         </div>
         
         <div class="space-y-4">
             <div class="bg-amber-500/5 rounded-lg p-3 text-sm tz-text-secondary shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-amber-400 text-xs uppercase tracking-wider mb-1">Accepted Items</strong>
               <p class="text-xs">Wheelsets (rims, hubs, spokes), or related components. Please confirm compatibility with us first.</p>
            </div>
             <div class="bg-amber-500/5 rounded-lg p-3 text-sm tz-text-secondary shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-amber-400 text-xs uppercase tracking-wider mb-1">Important</strong>
               <ul class="space-y-1.5 mt-2 list-none pl-0 text-xs">
                  <li class="flex gap-2"><span class="text-amber-500">-</span> Components must be new or in good condition.</li>
                  <li class="flex gap-2"><span class="text-amber-500">-</span> Include detailed packing list.</li>
                  <li class="flex gap-2"><span class="text-amber-500">-</span> Turnaround: 3-7 business days.</li>
                  <li class="flex gap-2"><span class="text-amber-500">-</span> Shipping costs are customer responsibility.</li>
               </ul>
            </div>
         </div>
       </div>

    </div>

    <!-- DIY Support Section -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold tz-text-primary mb-3">Building Your Own Wheels?</h3>
       <p class="text-sm tz-text-secondary mb-6 max-w-2xl mx-auto">
          If you prefer to assemble yourself using TANZANITE parts, we support you! access our technical guides and customization options below.
       </p>
       <div class="flex flex-wrap justify-center gap-3">
          <button
            type="button"
            class="px-4 py-2 rounded-lg border border-slate-600 hover:bg-slate-700 tz-text-secondary text-xs font-bold uppercase tracking-wider transition-colors"
            @click="goToTechnicalSpokePattern"
          >
            Viwe Spoke Patterns
          </button>
           <button
            type="button"
            class="px-4 py-2 rounded-lg border border-slate-600 hover:bg-slate-700 tz-text-secondary text-xs font-bold uppercase tracking-wider transition-colors"
            @click="goToHolePatterns"
          >
            View Hole Options
          </button>
       </div>
    </div>

  </div>
</template>

<script setup lang="ts">
const { openWhatsAppChat, goToTechnicalSpokePattern, goToHolePatterns } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  goToTechnicalSpokePattern: () => void | Promise<void>
  goToHolePatterns: () => void | Promise<void>
}>()
</script>
