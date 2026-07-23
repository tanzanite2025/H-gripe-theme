<template>
  <section id="protection" class="support-section">
    <h2 class="support-section__title text-center mb-8 text-white">Protection</h2>

    <div class="w-full max-w-none text-left space-y-10">
      <!-- Intro Header -->
      <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
        <h3 class="text-indigo-200 font-bold text-lg mb-1">Buyer Protection (PayPal Policy)</h3>
        <p class="tz-text-secondary text-sm">
          The following is excerpted from the <strong class="tz-text-primary">PayPal Legal Agreement</strong>, outlining the basic protections available to buyers:
        </p>
      </div>

      <div class="space-y-8">
         <!-- 1. Reasons for Dispute -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
           <h3 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Reasons for Dispute</h3>
           <p class="tz-text-secondary text-sm mb-3">Buyers may open a dispute on PayPal for the following reasons:</p>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
             <li><strong class="tz-text-primary">Item Not Received</strong></li>
             <li><strong class="tz-text-primary">Item Significantly Not as Described</strong></li>
           </ul>
         </div>

         <!-- 2. Dispute Timeframe -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Dispute Timeframe</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             A dispute must be filed within <strong class="text-white">45 days of the initial payment</strong>.
           </p>
         </div>

         <!-- 3. Claim Process -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Claim Process</h3>
           <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
             <li>If a dispute is escalated to a claim, PayPal will investigate and make a final decision.</li>
             <li>
               PayPal generally aims to resolve claims within <strong class="text-white">30 days of submission</strong>, though this period may be extended if additional investigation is required.
             </li>
           </ul>
         </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
// No emits or props needed for this tab currently
</script>
