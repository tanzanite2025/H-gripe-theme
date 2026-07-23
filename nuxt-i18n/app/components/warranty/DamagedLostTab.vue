<template>
  <section id="damaged-lost" class="support-section">
    <h2 class="support-section__title text-center mb-8 text-white">Damaged or Lost Goods</h2>

    <div class="w-full max-w-none text-left space-y-8">
      <!-- Intro Alert -->
      <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
        <p class="text-indigo-200 font-medium">Please carefully inspect the package before signing for it.</p>
      </div>

      <div class="space-y-10">
        <!-- 1. Normal Acceptance -->
        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
          <h3 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Normal Acceptance</h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
            <li>If the package is intact, please sign for it.</li>
            <li>If you notice damage, note it on the delivery slip or refuse the package. The goods will be returned to Tanzanite.</li>
            <li>
              If you need to file a claim with the carrier, they will issue a claim number. Please provide this number to Tanzanite so we can track the shipment. Once the returned goods are received, we will arrange re-shipment within <strong class="text-white">10 business days</strong>.
            </li>
          </ul>
        </div>

        <!-- 2. Damaged but Accepted -->
        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
          <h3 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damaged but Accepted</h3>
          <!-- Warning Box for this critical section -->
          <div class="bg-rose-500/5 rounded-lg p-4 border border-rose-500/10">
              <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-rose-500">
                <li>If the goods are damaged upon delivery but you still accept them without reporting, the carrier cannot be held responsible.</li>
                <li>In this case, you will be responsible for purchasing any replacement products.</li>
              </ul>
          </div>
        </div>

        <!-- 3. Damage During Transit -->
        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
          <h3 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damage During Transit</h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
            <li>If the package appears intact but the product is seriously damaged during transit, please keep the waybill, packaging, and related materials.</li>
            <li>Contact us within <strong class="text-white">7 days of receipt</strong> at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
          </ul>
        </div>

        <!-- 4. Lost Shipment -->
        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
          <h3 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Lost Shipment</h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
            <li>If the goods are lost during transit, please contact us immediately at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
            <li>All Tanzanite shipments are insured, and we will assist you with the claim process.</li>
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
