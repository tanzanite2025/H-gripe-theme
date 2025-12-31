<template>
  <div class="space-y-8">
    <!-- Intro Warning Card -->
    <div class="relative overflow-hidden rounded-2xl bg-rose-500/10 border-l-4 border-rose-500 p-5 md:p-6 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
       <div class="flex gap-4">
         <div class="shrink-0 text-rose-500 mt-1">
           <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
         </div>
         <div>
           <h3 class="text-lg font-bold text-rose-500 mb-2">Important Safety Notice</h3>
           <p class="text-sm text-rose-200/80 leading-relaxed mb-3">
            Improper installation, use, or maintenance may cause serious damage to the wheels and pose a danger to the rider.
            To ensure a reliable riding experience with TANZANITE products, please read and follow the instructions below.
          </p>
          <p class="text-xs text-rose-200/60 leading-relaxed italic">
             Even if you have years of cycling or wheelset assembly experience, these instructions apply specifically to TANZANITE carbon fiber products.
          </p>
         </div>
       </div>
    </div>

    <!-- Contact Support Strip -->
    <div class="flex flex-col md:flex-row items-center justify-center gap-2 text-sm text-slate-400 bg-slate-800/40 p-3 rounded-lg border border-slate-700/30">
      <span>Have questions? Contact us at</span>
      <a class="text-teal-400 font-medium hover:text-teal-300 transition-colors" href="mailto:support@tanzanite.site">support@tanzanite.site</a>
      <span class="hidden md:inline text-slate-600">|</span>
      <button
        type="button"
        class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-slate-700/50 hover:bg-slate-700 text-slate-200 text-xs font-semibold transition-all"
        @click="openWhatsAppChat"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        Instant Chat
      </button>
    </div>

    <!-- Grid Layout -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

      <!-- 1. Installation & Inspection (Sky) -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
         <div class="flex items-center gap-3 mb-5 pb-3 border-b border-sky-500/10">
            <h3 class="text-lg font-bold text-slate-200">Installation & Inspection</h3>
         </div>

         <div class="space-y-4">
            <div class="bg-sky-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-sky-400 text-xs uppercase tracking-wider mb-1">Professional Installation</strong>
               <p>Have the wheelset installed by a qualified mechanic. Incompatibility with brakes or drivetrain can damage wheels or frame.</p>
            </div>

            <div class="bg-sky-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-sky-400 text-xs uppercase tracking-wider mb-1">Tire Installation</strong>
               <p class="mb-2">Use provided plastic tire levers only. Metal levers will permanently damage the rim.</p>
               <button type="button" class="text-[10px] font-bold uppercase tracking-wider text-sky-400 border border-sky-500/30 px-2 py-1 rounded hover:bg-sky-500/10 transition-colors" @click="goToTubelessInstallation">
                  View Tubeless Guide
               </button>
            </div>

            <div class="grid grid-cols-2 gap-3">
               <div class="bg-sky-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                  <strong class="block text-sky-400 text-xs uppercase tracking-wider mb-1">Torque Specs</strong>
                  <p class="text-xs">Follow manufacturer recommended torque for all bolts and axles.</p>
               </div>
                <div class="bg-sky-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                  <strong class="block text-sky-400 text-xs uppercase tracking-wider mb-1">Brake Pads</strong>
                  <p class="text-xs">Use carbon-specific pads ONLY to prevent rim damage.</p>
               </div>
            </div>
         </div>
      </div>

      <!-- 2. Usage & Riding (Emerald) -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
         <div class="flex items-center gap-3 mb-5 pb-3 border-b border-emerald-500/10">
            <h3 class="text-lg font-bold text-slate-200">Usage & Riding</h3>
         </div>

         <div class="space-y-4">
             <div class="bg-emerald-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-emerald-400 text-xs uppercase tracking-wider mb-1">Weight Limits</strong>
               <p>Strictly follow max rider weight/load limits. Overloading causes rim cracking and injury risk.</p>
            </div>

            <div class="bg-emerald-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-emerald-400 text-xs uppercase tracking-wider mb-1">Road Conditions</strong>
               <p class="mb-2">Avoid severe potholes and gravel. Impact damage may not be immediately visible.</p>
               <div class="flex items-start gap-2 text-xs text-emerald-500/80 mt-2 bg-emerald-500/10 p-2 rounded">
                  <svg class="shrink-0 mt-0.5" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
                  <span>After impact: Clean rim, remove tire, inspect carefully.</span>
               </div>
            </div>

            <div class="bg-emerald-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-emerald-400 text-xs uppercase tracking-wider mb-1">Braking Technique</strong>
               <p>Avoid continuous drag braking on long descents. Use intermittent braking to prevent carbon overheating.</p>
            </div>
         </div>
      </div>

       <!-- 3. Regular Maintenance (Amber) -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
         <div class="flex items-center gap-3 mb-5 pb-3 border-b border-amber-500/10">
            <h3 class="text-lg font-bold text-slate-200">Regular Maintenance</h3>
         </div>

         <div class="space-y-4">
            <div class="bg-amber-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-amber-400 text-xs uppercase tracking-wider mb-1">Pre-ride Inspection</strong>
               <ul class="space-y-2 mt-2 list-none pl-0">
                  <li class="flex gap-2"><span class="text-amber-500">●</span> Check rims for cracks, bulges, or delamination.</li>
                  <li class="flex gap-2"><span class="text-amber-500">●</span> Inspect brake pads; remove metal shavings/stones.</li>
               </ul>
            </div>

            <div class="bg-amber-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <div class="flex justify-between items-start">
                   <div>
                       <strong class="block text-amber-400 text-xs uppercase tracking-wider mb-1">Spoke Tension</strong>
                       <p class="mb-2">Regularly check for even tension distribution.</p>
                   </div>
                   <button type="button" class="text-[10px] font-bold uppercase tracking-wider text-amber-400 border border-amber-500/30 px-2 py-1 rounded hover:bg-amber-500/10 transition-colors whitespace-nowrap" @click="goToTechnicalTension">
                      View Settings
                   </button>
               </div>
            </div>

            <div class="bg-amber-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-amber-400 text-xs uppercase tracking-wider mb-1">Cleaning</strong>
               <p>Use neutral detergent and soft cloth only. NO strong acids, alkalis, or high-pressure water.</p>
            </div>
         </div>
      </div>

       <!-- 4. Critical Warnings (Rose/Indigo) -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
         <div class="flex items-center gap-3 mb-5 pb-3 border-b border-indigo-500/10">
            <span class="text-indigo-500 bg-indigo-500/10 p-2 rounded-lg">
               <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
            </span>
            <h3 class="text-lg font-bold text-slate-200">Prohibited & Emergency</h3>
         </div>

         <div class="space-y-4">
             <!-- Prohibited -->
            <div class="bg-indigo-500/5 rounded-lg p-3 text-sm text-slate-400 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-indigo-400 text-xs uppercase tracking-wider mb-2">Strictly Prohibited</strong>
               <ul class="space-y-2 list-none pl-0">
                  <li class="flex gap-2 items-center"><span class="text-rose-500 font-bold text-lg leading-none">×</span> <span>No non-original parts modification.</span></li>
                  <li class="flex gap-2 items-center"><span class="text-rose-500 font-bold text-lg leading-none">×</span> <span>No sanding/grinding with metal tools.</span></li>
                  <li class="flex gap-2 items-center"><span class="text-rose-500 font-bold text-lg leading-none">×</span> <span>No extreme temp storage (ovens/freezing).</span></li>
               </ul>
            </div>

             <!-- Emergency -->
             <div class="bg-indigo-500/5 rounded-lg p-3 text-sm text-slate-400 border-none shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
               <strong class="block text-indigo-400 text-xs uppercase tracking-wider mb-1">Emergency Protocol</strong>
               <p class="mb-2">If cracks/abnormal noises occur: <strong>STOP RIDING IMMEDIATELY</strong>.</p>
               <p class="text-xs text-slate-500">After any crash, perform thorough inspection even if no visible damage.</p>
            </div>

             <div class="pt-2 text-center">
                 <button
                    type="button"
                    class="w-full py-2 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 text-xs font-bold uppercase tracking-wider transition-colors"
                    @click="goToAfterSales"
                  >
                    View After-sales Instructions
                  </button>
             </div>
         </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
const { openWhatsAppChat, goToTubelessInstallation, goToTechnicalTension, goToAfterSales } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  goToTubelessInstallation: () => void | Promise<void>
  goToTechnicalTension: () => void | Promise<void>
  goToAfterSales: () => void | Promise<void>
  }>()
</script>
