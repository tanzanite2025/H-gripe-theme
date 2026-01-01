<template>
  <div class="support-warranty">
    <!-- SEO-friendly hidden H1 -->
    <h1 class="support-page__title support-page__title--sr-only">Warranty</h1>

    <!-- Tabs header -->
    <div class="nav-pill-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 1. Change / Cancel -->
    <section
      v-show="activeTab === 'change-cancel'"
      id="change-cancel"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Change / Cancel</h3>
      
      <div class="max-w-3xl mx-auto text-left space-y-10">
        
        <div class="space-y-8">
           <!-- 1. Scope of Application -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Scope of Application</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               Customers may request to cancel all or part of an order before the item has shipped, regardless of whether payment has been partially or fully completed.
             </p>
           </div>

           <!-- 2. Cancellation After Shipment -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Cancellation After Shipment</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               If the product has already shipped, any cancellation request will be treated as a "return" and subject to the return process and applicable fees. Please refer to our 
               <a href="#" @click.prevent="setActiveTab('returns')" class="text-teal-400 hover:text-teal-300 transition-colors border-b border-teal-400/30 hover:border-teal-300 pb-0.5">Product Return Policy</a> 
               for details.
             </p>
           </div>

           <!-- 3. Authorization Requirement -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Authorization Requirement</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               Any order amendment or cancellation must be confirmed and authorized by TANZANITE.
             </p>
           </div>

           <!-- 4. Fee Policy -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Fee Policy</h4>
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

    <!-- 2. Damaged or Lost Goods -->
    <section
      v-show="activeTab === 'damaged-lost'"
      id="damaged-lost"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Damaged or Lost Goods</h3>

      <div class="max-w-3xl mx-auto text-left space-y-8">
        <!-- Intro Alert -->
        <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
          <p class="text-indigo-200 font-medium">Please carefully inspect the package before signing for it.</p>
        </div>

        <div class="space-y-10">
          <!-- 1. Normal Acceptance -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Normal Acceptance</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
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
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damaged but Accepted</h4>
            <!-- Warning Box for this critical section -->
            <div class="bg-rose-500/5 rounded-lg p-4 border border-rose-500/10">
                <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-rose-500">
                  <li>If the goods are damaged upon delivery but you still accept them without reporting, the carrier cannot be held responsible.</li>
                  <li>In this case, you will be responsible for purchasing any replacement products.</li>
                </ul>
            </div>
          </div>

          <!-- 3. Damage During Transit -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damage During Transit</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
              <li>If the package appears intact but the product is seriously damaged during transit, please keep the waybill, packaging, and related materials.</li>
              <li>Contact us within <strong class="text-white">7 days of receipt</strong> at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
            </ul>
          </div>

          <!-- 4. Lost Shipment -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Lost Shipment</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
              <li>If the goods are lost during transit, please contact us immediately at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
              <li>All Tanzanite shipments are insured, and we will assist you with the claim process.</li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- 3. Returns -->
    <section
      v-show="activeTab === 'returns'"
      id="returns"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Returns</h3>

      <div class="max-w-3xl mx-auto text-left space-y-10">
        <!-- Intro -->
        <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
          <p class="text-slate-300 leading-relaxed">
            If the product you receive does not match your order, please do not use it and contact our customer service team (<a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 font-medium transition-colors">support@tanzanite.site</a>).
          </p>
          <p class="text-slate-400 text-sm mt-2">
            If a return and refund request is approved, unless otherwise agreed by Tanzanite, the following terms apply:
          </p>
        </div>

        <!-- Return & Refund Conditions -->
        <div>
           <h4 class="text-xl font-bold text-white mb-6 border-b border-white/10 pb-3">Return & Refund Conditions</h4>
           <div class="space-y-8">
              <!-- 1. Eligibility -->
              <div class="relative pl-12">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
                <h5 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Eligibility</h5>
                <p class="text-slate-400 text-sm leading-relaxed">
                  Only products shipped from Tanzanite, delivered within the past <strong class="text-white">30 days</strong>, and in resalable condition will be considered for return and refund.
                </p>
              </div>

              <!-- 2. Special Orders -->
              <div class="relative pl-12">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
                <h5 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Special Orders</h5>
                <p class="text-slate-400 text-sm leading-relaxed">
                  Non-stock or custom-configured products (special orders) are not eligible for return or refund, unless the issue is caused by <strong class="text-white">Tanzanite's error</strong>.
                </p>
              </div>

              <!-- 3. Refund Timeline -->
              <div class="relative pl-12">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
                <h5 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Refund Timeline</h5>
                <p class="text-slate-400 text-sm leading-relaxed">
                  If Tanzanite approves the return and refund, the refund will be issued within <strong class="text-white">3 business days</strong> after the goods are received and inspected by Tanzanite.
                </p>
              </div>
           </div>
        </div>

        <!-- Fees & Charges -->
        <div>
           <h4 class="text-xl font-bold text-white mb-6 border-b border-white/10 pb-3">Fees & Charges</h4>
           <p class="text-slate-400 text-sm mb-6">All costs related to returns are the responsibility of the customer, including:</p>
           <div class="space-y-8">
              <!-- 1. Restocking Fee -->
              <div class="relative pl-12">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
                <h5 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Restocking & Refurbishment Fee</h5>
                <p class="text-slate-400 text-sm leading-relaxed">
                  Charged at <strong class="text-white bg-white/5 px-1.5 py-0.5 rounded text-xs font-bold border border-white/10">20% of the original purchase value (minimum USD $100)</strong>.
                </p>
              </div>

              <!-- 2. Other Costs -->
              <div class="relative pl-12">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
                <h5 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Other Costs</h5>
                <p class="text-slate-400 text-sm leading-relaxed">
                  Unless otherwise specified under the warranty policy, shipping fees, duties, taxes, and any additional charges are borne by the customer.
                </p>
              </div>
           </div>
        </div>
      </div>
    </section>

    <!-- 4. Warranty -->
    <section
      v-show="activeTab === 'warranty'"
      id="warranty"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Advanced Warranty Policy</h3>

      <div class="max-w-3xl mx-auto text-left space-y-10">
        <div class="space-y-8">
           <!-- 1. Applicable Products -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Applicable Products</h4>
             <ul class="space-y-2 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
                <li>This policy applies to all <strong class="text-white">Tanzanite series Wheels/Rims</strong>.</li>
                <li>You can use the product filter on our website to identify which series your wheelset belongs to.</li>
             </ul>
           </div>

           <!-- 2. Coverage Options -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Coverage Options</h4>
             <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-decimal pl-4 marker:text-slate-500">
                <li>
                    <strong class="text-slate-200">5-Year Warranty:</strong> Valid for five years from the date of purchase, at no additional cost.
                </li>
                <li>
                    <strong class="text-slate-200">Lifetime Warranty:</strong> Upgrade to lifetime coverage for <strong class="text-white">USD $100 per rim</strong>.
                </li>
             </ul>
           </div>

           <!-- 3. Policy Conditions -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
             <h4 class="text-slate-100 font-bold text-lg mb-4 pt-0.5">Policy Conditions</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- 3.1 Manufacturing Defects -->
                <div class="bg-slate-800/20 rounded-lg p-5 border border-slate-700/50 md:col-span-2">
                    <h5 class="text-teal-400 font-bold text-sm mb-3">1. Manufacturing Defects</h5>
                    <ul class="space-y-2 text-slate-400 text-xs leading-relaxed list-disc pl-4 marker:text-teal-500/40">
                        <li>Within <strong class="text-white">30 days of receipt</strong>: Tanzanite will replace the rim and cover shipping costs.</li>
                        <li>After <strong class="text-white">31 days</strong> (within warranty period): Tanzanite will replace the rim, but shipping costs are the customer's responsibility.</li>
                    </ul>
                </div>
                <!-- 3.2 Product Upgrade -->
                 <div class="bg-slate-800/20 rounded-lg p-5 border border-slate-700/50">
                    <h5 class="text-teal-400 font-bold text-sm mb-3">2. Product Upgrade</h5>
                    <ul class="space-y-2 text-slate-400 text-xs leading-relaxed list-disc pl-4 marker:text-teal-500/40">
                        <li>Tanzanite reserves the right to replace your rim with a stronger, more durable model.</li>
                        <li>If discontinued, we will upgrade you to the latest equivalent model at no cost.</li>
                    </ul>
                </div>
                <!-- 3.3 Non-Transferable -->
                <div class="bg-slate-800/20 rounded-lg p-5 border border-slate-700/50">
                    <h5 class="text-teal-400 font-bold text-sm mb-3">3. Non-Transferable</h5>
                    <p class="text-slate-400 text-xs leading-relaxed">
                        This warranty plan applies only to the original purchaser and cannot be transferred.
                    </p>
                </div>
                 <!-- 3.4 Claim Timeframe -->
                <div class="bg-slate-800/20 rounded-lg p-5 border border-slate-700/50 md:col-span-2">
                    <h5 class="text-teal-400 font-bold text-sm mb-3">4. Claim Timeframe</h5>
                    <ul class="space-y-2 text-slate-400 text-xs leading-relaxed list-disc pl-4 marker:text-teal-500/40">
                         <li>Any issues with replacement products must be reported within <strong class="text-white">7 days of receipt</strong>.</li>
                         <li>Warranty coverage is calculated from the original purchase date (does not reset after replacement).</li>
                    </ul>
                </div>
             </div>
           </div>

           <!-- 4. Exclusions -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Exclusions</h4>
             <p class="text-slate-400 text-sm mb-3">This warranty does not cover:</p>
             <ul class="space-y-2 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-red-400/50">
                <li>Improper assembly or maintenance by unauthorized technicians.</li>
                <li>Installation of incompatible parts, accessories, or tires.</li>
                <li>Damage caused by crashes, accidental impact, misuse, heat damage, or non-compliance with usage guidelines.</li>
                <li>Issues related to spokes/nipples/assembly if not built by Tanzanite.</li>
                <li>Normal wear and tear (bearings, decals, etc.).</li>
             </ul>
           </div>

           <!-- 5. Contact -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">5</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Contact</h4>
             <div class="space-y-3 text-slate-400 text-sm">
                <p>Email: <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a></p>
                <p>
                    <a href="#" @click.prevent="setActiveTab('submit-warranty')" class="inline-flex items-center gap-1 text-teal-400 hover:text-teal-300 font-bold transition-colors">
                   Submit a Claim 
                   <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
                 </a>
                </p>
             </div>
           </div>
        </div>
      </div>
    </section>

    <!-- 5. Accidental Damage -->
    <section
      v-show="activeTab === 'accidental-damage'"
      id="accidental-damage"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Accidental Damage</h3>

      <div class="max-w-3xl mx-auto text-left space-y-10">
        <div class="space-y-8">
           <!-- 1. Scope -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Scope</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               This policy covers accidental crashes or misuse, such as jumps or falls, collisions, or impacts with rocks.
             </p>
           </div>

           <!-- 2. Product Group -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Product Group</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               Applies to all <strong class="text-white">Tanzanite series Wheels/Rims</strong>. You can use the product filter on our website to identify which series your wheelset belongs to.
             </p>
           </div>

           <!-- 3. Coverage -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Coverage</h4>
             <ul class="space-y-2 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
               <li>Crash replacement is available within <strong class="text-white">three years from the date of purchase</strong>.</li>
               <li>Replacement products are offered at a <strong class="text-white">10% discount</strong>.</li>
             </ul>
           </div>

           <!-- 4. Conditions -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Conditions</h4>
             <ul class="space-y-4 text-slate-400 text-sm leading-relaxed list-decimal pl-4 marker:text-slate-500">
               <li>
                 The Standard Crash Replacement Policy is non-transferable and applies only to the original purchaser.
               </li>
               <li>
                 If there are any issues with the replacement product, claims must be submitted within <strong class="text-white">7 days of receipt</strong>; otherwise, the case will be considered resolved.
                 <p class="mt-2 text-xs bg-white/5 p-2 rounded border border-white/5">
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
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">5</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Claim Process</h4>
             <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
               <li>Once a claim is submitted, our after-sales team will contact you within <strong class="text-white">24 hours</strong> to verify details.</li>
               <li>In most cases, you will not be required to return the damaged equipment to receive a replacement.</li>
               <li class="list-none -ml-4 mt-2">
                 <a href="#" @click.prevent="setActiveTab('submit-warranty')" class="inline-flex items-center gap-1 text-teal-400 hover:text-teal-300 font-bold transition-colors">
                   Submit a Claim 
                   <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
                 </a>
               </li>
             </ul>
           </div>
        </div>
      </div>
    </section>

    <!-- 6. Protection -->
    <section
      v-show="activeTab === 'protection'"
      id="protection"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Protection</h3>

      <div class="max-w-3xl mx-auto text-left space-y-10">
        <!-- Intro Header -->
        <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
          <h4 class="text-indigo-200 font-bold text-lg mb-1">Buyer Protection (PayPal Policy)</h4>
          <p class="text-slate-400 text-sm">
            The following is excerpted from the <strong class="text-slate-200">PayPal Legal Agreement</strong>, outlining the basic protections available to buyers:
          </p>
        </div>

        <div class="space-y-8">
           <!-- 1. Reasons for Dispute -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
             <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Reasons for Dispute</h4>
             <p class="text-slate-400 text-sm mb-3">Buyers may open a dispute on PayPal for the following reasons:</p>
             <ul class="space-y-2 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
               <li><strong class="text-slate-200">Item Not Received</strong></li>
               <li><strong class="text-slate-200">Item Significantly Not as Described</strong></li>
             </ul>
           </div>

           <!-- 2. Dispute Timeframe -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Dispute Timeframe</h4>
             <p class="text-slate-400 text-sm leading-relaxed">
               A dispute must be filed within <strong class="text-white">45 days of the initial payment</strong>.
             </p>
           </div>

           <!-- 3. Claim Process -->
           <div class="relative pl-12">
             <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
             <h4 class="text-slate-100 font-bold text-lg mb-2 pt-0.5">Claim Process</h4>
             <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
               <li>If a dispute is escalated to a claim, PayPal will investigate and make a final decision.</li>
               <li>
                 PayPal generally aims to resolve claims within <strong class="text-white">30 days of submission</strong>, though this period may be extended if additional investigation is required.
               </li>
             </ul>
           </div>
        </div>
      </div>
    </section>

    <!-- 7. Submit Warranty -->
    <section
      v-show="activeTab === 'submit-warranty'"
      id="submit-warranty"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-6">Submit Warranty Claim</h3>
      
      <div class="max-w-xl mx-auto">
        <div v-if="submitMessage" :class="['p-4 rounded mb-6 text-center', submitStatus === 'success' ? 'bg-green-500/20 text-green-200 border border-green-500/30' : 'bg-red-500/20 text-red-200 border border-red-500/30']">
          {{ submitMessage }}
        </div>

        <form @submit.prevent="submitClaim" class="space-y-4">
          <!-- Order Number -->
          <div>
            <label class="block text-slate-300 text-sm font-bold mb-1">Order Number</label>
            <input 
              v-model="form.order_number" 
              type="text" 
              required
              class="w-full bg-slate-800/50 border border-slate-600 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-teal-500 transition-colors"
              placeholder="e.g. TANZ-12345"
            />
          </div>

          <!-- Email -->
          <div>
            <label class="block text-slate-300 text-sm font-bold mb-1">Email Address</label>
            <input 
              v-model="form.email" 
              type="email" 
              required
              class="w-full bg-slate-800/50 border border-slate-600 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-teal-500 transition-colors"
              placeholder="Your email address"
            />
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Tire Pressure -->
            <div>
              <label class="block text-slate-300 text-sm font-bold mb-1">Tire Pressure (PSI)</label>
              <input 
                v-model="form.tire_pressure" 
                type="text" 
                class="w-full bg-slate-800/50 border border-slate-600 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-teal-500 transition-colors"
                placeholder="e.g. 80"
              />
            </div>

            <!-- Tubeless -->
            <div>
              <label class="block text-slate-300 text-sm font-bold mb-1">Tubeless Setup?</label>
              <div class="flex items-center space-x-6 mt-2">
                 <label class="flex items-center cursor-pointer">
                   <input type="radio" v-model="form.is_tubeless" value="yes" class="form-radio text-teal-500 bg-slate-800 border-slate-600 focus:ring-teal-500/50">
                   <span class="ml-2 text-slate-300">Yes</span>
                 </label>
                 <label class="flex items-center cursor-pointer">
                   <input type="radio" v-model="form.is_tubeless" value="no" class="form-radio text-teal-500 bg-slate-800 border-slate-600 focus:ring-teal-500/50">
                   <span class="ml-2 text-slate-300">No</span>
                 </label>
              </div>
            </div>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-slate-300 text-sm font-bold mb-1">Issue Description</label>
            <textarea 
              v-model="form.issue_description" 
              rows="4"
              required
              class="w-full bg-slate-800/50 border border-slate-600 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-teal-500 transition-colors"
              placeholder="Please describe the issue in detail..."
            ></textarea>
          </div>

          <!-- Images -->
          <div>
            <label class="block text-slate-300 text-sm font-bold mb-1">Upload Images</label>
            <div class="flex items-center gap-3">
              <button 
                type="button" 
                @click="triggerImageUpload"
                class="bg-slate-700 hover:bg-slate-600 text-teal-400 font-semibold py-1.5 px-4 rounded text-sm transition-colors"
              >
                Choose Files
              </button>
              <span class="text-slate-400 text-sm">
                {{ imageFiles.length > 0 ? `${imageFiles.length} file(s) selected` : 'No file chosen' }}
              </span>
            </div>
            <input 
              ref="imageInput"
              type="file" 
              multiple 
              accept="image/*" 
              @change="handleImages"
              class="hidden"
            />
            <p class="mt-1 text-xs text-slate-500">Max 5MB per image.</p>
          </div>

          <!-- Video -->
          <div>
            <label class="block text-slate-300 text-sm font-bold mb-1">Upload Video (Optional)</label>
            <div class="flex items-center gap-3">
              <button 
                type="button" 
                @click="triggerVideoUpload"
                class="bg-slate-700 hover:bg-slate-600 text-teal-400 font-semibold py-1.5 px-4 rounded text-sm transition-colors"
              >
                Choose File
              </button>
              <span class="text-slate-400 text-sm">
                {{ videoFile ? videoFile.name : 'No file chosen' }}
              </span>
            </div>
            <input 
              ref="videoInput"
              type="file" 
              accept="video/*" 
              @change="handleVideo"
              class="hidden"
            />
            <p class="mt-1 text-xs text-slate-500">Max 20MB. For larger videos, please provide a link in the description.</p>
          </div>

          <!-- Submit Button -->
          <div class="pt-2">
            <button 
              type="submit" 
              :disabled="isSubmitting"
              class="w-full bg-teal-600 hover:bg-teal-500 text-white font-bold py-2.5 px-4 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-teal-900/20"
            >
              <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                <svg class="animate-spin h-5 w-5 text-white" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                Processing...
              </span>
              <span v-else>Submit Claim</span>
            </button>
          </div>
        </form>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="support-section mt-8">
      <PageFaq 
        page-id="support-warranty"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Warranty',
})

type WarrantyTabId = 
  | 'change-cancel' 
  | 'damaged-lost' 
  | 'returns' 
  | 'warranty' 
  | 'accidental-damage' 
  | 'protection' 
  | 'submit-warranty'

const tabs: { id: WarrantyTabId; label: string }[] = [
  { id: 'change-cancel', label: 'Change / Cancel' },
  { id: 'damaged-lost', label: 'Damaged or Lost Goods' },
  { id: 'returns', label: 'Returns' },
  { id: 'warranty', label: 'Warranty' },
  { id: 'accidental-damage', label: 'Accidental Damage' },
  { id: 'protection', label: 'Protection' },
  { id: 'submit-warranty', label: 'Submit Warranty' },
]

const activeTab = ref<WarrantyTabId>('warranty') // Default to Warranty tab as it's the main topic? Or 'change-cancel' as first? 
// User listed tabs in order: Change/Cancel... but page title is Warranty.
// Usually default is first tab unless deep linked. I'll set default to 'change-cancel' (first tab) or 'warranty' if contextually appropriate.
// User listed "Change / Cancel" first. I will default to 'change-cancel' to match list order.
// But wait, the previous page was "Warranty". It might be confusing if "Warranty" tab is hidden.
// I'll stick to the first tab in the list: 'change-cancel'.

const route = useRoute()

// --- Form Logic ---
const form = ref({
  order_number: '',
  email: '',
  tire_pressure: '',
  is_tubeless: 'no',
  issue_description: '',
})

const imageFiles = ref<File[]>([])
const videoFile = ref<File | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const isSubmitting = ref(false)
const submitMessage = ref('')
const submitStatus = ref<'success' | 'error' | ''>('')

const triggerImageUpload = () => {
  imageInput.value?.click()
}

const triggerVideoUpload = () => {
  videoInput.value?.click()
}

const handleImages = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files) {
    imageFiles.value = Array.from(target.files)
  }
}

const handleVideo = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    videoFile.value = target.files[0]
  }
}

const submitClaim = async () => {
  isSubmitting.value = true
  submitMessage.value = ''
  submitStatus.value = ''

  try {
    const formData = new FormData()
    formData.append('order_number', form.value.order_number)
    formData.append('email', form.value.email)
    formData.append('tire_pressure', form.value.tire_pressure)
    formData.append('is_tubeless', form.value.is_tubeless === 'yes' ? 'yes' : 'no')
    formData.append('issue_description', form.value.issue_description)

    imageFiles.value.forEach((file) => {
      formData.append('images[]', file)
    })

    if (videoFile.value) {
      formData.append('video', videoFile.value)
    }

    // Assuming WP REST API is at /wp-json/
    // Adjust base URL as needed for the environment
    const { data, error } = await useFetch('/wp-json/tanzanite/v1/warranty/claim', {
      method: 'POST',
      body: formData,
    })

    if (error.value) {
      throw new Error(error.value.message || 'Submission failed')
    }

    submitStatus.value = 'success'
    submitMessage.value = 'Your claim has been submitted successfully. We will contact you shortly.'
    
    // Reset form
    form.value = {
      order_number: '',
      email: '',
      tire_pressure: '',
      is_tubeless: 'no',
      issue_description: '',
    }
    imageFiles.value = []
    videoFile.value = null
    if (imageInput.value) imageInput.value.value = ''
    if (videoInput.value) videoInput.value.value = ''

  } catch (err: any) {
    console.error(err)
    submitStatus.value = 'error'
    submitMessage.value = err.message || 'An error occurred. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}

const setActiveTab = (id: WarrantyTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
}

const syncTabWithHash = (hash: string | null | undefined) => {
  if (!hash) return
  const clean = hash.startsWith('#') ? hash.slice(1) : hash
  if (tabs.some((tab) => tab.id === clean)) {
    activeTab.value = clean as WarrantyTabId
  }
}

onMounted(() => {
  syncTabWithHash(route.hash)
})

watch(
  () => route.hash,
  (newHash) => {
    syncTabWithHash(newHash)
  }
)
</script>

<style scoped>
.support-warranty {
  margin-top: -1rem;
}

.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.support-section {
  margin-top: 2rem;
  /* text-align: center; Optional, keeping typical flow */
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0 auto;
  max-width: 65ch; /* Readable line length */
  font-size: 0.95rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}
</style>
