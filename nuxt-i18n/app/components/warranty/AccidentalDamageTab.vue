<template>
  <section id="accidental-damage" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">Accidental Damage</h2>

    <div class="w-full max-w-none text-left space-y-10">
      <div class="space-y-8">
         <!-- 1. Scope -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Scope</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             This policy covers accidental crashes or misuse, such as jumps or falls, collisions, or impacts with rocks.
           </p>
         </div>

         <!-- 2. Product Group -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Product Group</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             Applies to all <strong class="tz-text-primary">eligible Wheels/Rims</strong>. You can use the product filter on our website to identify which series your wheelset belongs to.
           </p>
         </div>

         <!-- 3. Coverage -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Coverage</h3>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li>Crash replacement is available within <strong class="tz-text-primary">three years from the date of purchase</strong>.</li>
             <li>Replacement products are offered at a <strong class="tz-text-primary">10% discount</strong>.</li>
           </ul>
         </div>

         <!-- 4. Conditions -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">4</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Conditions</h3>
           <ul class="space-y-4 tz-text-secondary text-sm leading-relaxed list-decimal pl-4 marker:tz-text-muted">
             <li>
               The Standard Crash Replacement Policy is non-transferable and applies only to the original purchaser.
             </li>
             <li>
               If there are any issues with the replacement product, claims must be submitted within <strong class="tz-text-primary">7 days of receipt</strong>; otherwise, the case will be considered resolved.
               <p class="mt-2 text-xs tz-surface-subtle p-2 rounded border tz-border-strong/5">
                 Warranty coverage is calculated from the original purchase date. The warranty period does not reset after a rim replacement.
               </p>
             </li>
             <li>
               Please note that all shipping fees, duties, taxes, and related costs are the responsibility of the customer.
             </li>
           </ul>
         </div>

         <!-- 5. Claim Process -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">5</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Claim Process</h3>
           <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li>Once a claim is submitted, our after-sales team will contact you within <strong class="tz-text-primary">24 hours</strong> to verify details.</li>
             <li>In most cases, you will not be required to return the damaged equipment to receive a replacement.</li>
             <li class="list-none -ml-4 mt-2">
               <a href="#" @click.prevent="$emit('change-tab', 'submit-warranty')" class="inline-flex items-center gap-1 text-emerald-600 hover:text-emerald-700 font-bold transition-colors">
                 Submit a Claim 
                 <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
               </a>
             </li>
           </ul>
         </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineEmits<{
  (e: 'change-tab', id: string): void
}>()
</script>
