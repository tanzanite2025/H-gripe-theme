<template>
  <section id="change-cancel" class="support-section">
    <h2 class="support-section__title text-center mb-8 text-white">Change / Cancel</h2>
    
    <div class="w-full max-w-none text-left space-y-10">
      
      <div class="space-y-8">
         <!-- 1. Scope of Application -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Scope of Application</h3>
           <p class="text-slate-400 text-sm leading-relaxed">
             Customers may request to cancel all or part of an order before the item has shipped, regardless of whether payment has been partially or fully completed.
           </p>
         </div>

         <!-- 2. Cancellation After Shipment -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Cancellation After Shipment</h3>
           <p class="text-slate-400 text-sm leading-relaxed">
             If the product has already shipped, any cancellation request will be treated as a "return" and subject to the return process and applicable fees. Please refer to our 
             <a href="#" @click.prevent="$emit('change-tab', 'returns')" class="text-teal-400 hover:text-teal-300 transition-colors border-b border-teal-400/30 hover:border-teal-300 pb-0.5">Product Return Policy</a> 
             for details.
           </p>
         </div>

         <!-- 3. Authorization Requirement -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Authorization Requirement</h3>
           <p class="text-slate-400 text-sm leading-relaxed">
             Any order amendment or cancellation must be confirmed and authorized by TANZANITE.
           </p>
         </div>

         <!-- 4. Fee Policy -->
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
           <h3 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Fee Policy</h3>
           <p class="text-slate-400 text-sm leading-relaxed">
             Once authorized, a cancellation may incur a charge of <strong class="text-white bg-white/5 px-1.5 py-0.5 rounded text-xs font-bold border border-white/10">10% of the order value (minimum USD $50)</strong> to cover re-stocking and operational costs.
           </p>
         </div>
      </div>

      <!-- Note Box -->
      <div class="relative pl-5 py-3 border-l-2 border-slate-600 bg-slate-800/20 rounded-r-lg">
          <p class="text-slate-400 text-xs italic leading-relaxed">
            <strong class="text-slate-300 not-italic uppercase tracking-wider text-[10px] mr-1">Note:</strong> 
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
