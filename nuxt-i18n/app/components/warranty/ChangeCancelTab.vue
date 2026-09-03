<template>
  <section id="change-cancel" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">Change / Cancel</h2>

    <div class="w-full max-w-none text-left space-y-10">

      <div class="space-y-8">
         <!-- 1. Scope of Application -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Scope of Application</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             Customers may request to cancel all or part of an order before the item has shipped, regardless of whether payment has been partially or fully completed.
           </p>
         </div>

         <!-- 2. Cancellation After Shipment -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Cancellation After Shipment</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             If the product has already shipped, any cancellation request will be treated as a "return" and subject to the return process and applicable fees. Please refer to our
             <a href="#" @click.prevent="$emit('change-tab', 'returns')" class="text-emerald-600 hover:text-emerald-700 transition-colors border-b border-emerald-400/30 hover:border-emerald-300 pb-0.5">Refund & Cancellation Policy</a>
             for details.
           </p>
         </div>

         <!-- 3. Authorization Requirement -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Authorization Requirement</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             Any order amendment or cancellation must be confirmed and authorized by our team.
           </p>
         </div>

         <!-- 4. Fee Policy -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">4</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">Fee Policy</h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             Once authorized, a cancellation may incur a charge of <strong class="tz-text-primary tz-surface-subtle px-1.5 py-0.5 rounded text-xs font-bold border tz-border-subtle">10% of the order value (minimum USD $50)</strong> to cover re-stocking and operational costs.
           </p>
         </div>
      </div>

      <!-- Note Box -->
      <div class="relative pl-5 py-3 border-l-2 tz-border-subtle tz-surface-panel rounded-r-lg">
          <p class="tz-text-muted text-xs italic leading-relaxed">
              <strong class="tz-text-secondary not-italic uppercase tracking-wider tz-micro-label mr-1">Note:</strong>
            Please carefully review product specifications and quantities before placing an order to avoid unnecessary fees and procedures.
          </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineEmits<{
  (e: 'change-tab', id: string): void
}>()
</script>
