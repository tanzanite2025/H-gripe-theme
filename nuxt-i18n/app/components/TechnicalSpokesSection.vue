<template>
  <div>
    <!-- Tabs Header -->
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'basic-concepts' }"
        @click="activeTab = 'basic-concepts'"
      >
        Basic Concepts
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'choose-spoke' }"
        @click="activeTab = 'choose-spoke'"
      >
        Choose spoke
      </button>
    </div>

    <!-- Tab: Basic Concepts -->
    <div v-show="activeTab === 'basic-concepts'">
    <!-- Intro Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center mb-8">
      <h3 class="text-xl font-bold text-slate-100 mb-4 flex items-center justify-center gap-2">
        Basic Concepts
      </h3>
      
      <p class="text-slate-400 text-sm leading-relaxed mb-6 max-w-2xl mx-auto hidden">
        Understanding spoke dimensions, materials, and typical use cases is crucial for building a reliable wheelset that matches your riding style.
      </p>

      <div class="w-full mb-8">
        <p class="text-xs font-bold uppercase tracking-wider text-slate-500 mb-3">Common Profiles & Types</p>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <GuideImage
            src="/public/technical/j-bend-spokeandstraightpull-spoke.webp"
            alt="Comparison of J-bend and straight-pull bicycle spokes"
            :zoomOnClick="true"
            caption="J-bend (Classic) vs Straight-pull"
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
          />
          <GuideImage
            src="/public/technical/Sapim-black-stainless-steel-straight-gauge-2.0mm-spoke-double-butted-2.0-1.8-2.0mm-spoke-aero-flat-bladed-spoke.webp"
            alt="Examples of Sapim stainless steel spokes in different shapes and profiles"
            :zoomOnClick="true"
            caption="Straight Gauge vs Double-Butted vs Aero"
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
          />
        </div>
      </div>

      <!-- New Content: Major Brands -->
      <div class="w-full mb-8 text-center">
         <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 bg-slate-700"></span>
            <h4 class="text-slate-200 font-bold text-base uppercase tracking-wider">Major Brands & Positioning</h4>
            <span class="h-px w-8 bg-slate-700"></span>
         </div>
         <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
             <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                <strong class="block text-sky-400 text-sm mb-1">DT Swiss (Switzerland)</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">Renowned for stable craftsmanship and a comprehensive product matrix (straight, butted, aero), with complete engineering support.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-sky-500/30 text-[10px] font-bold uppercase tracking-wider text-sky-400 hover:bg-sky-500/10 transition-colors">
                  View DT Swiss Spoke
                </button>
             </div>
              <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                <strong class="block text-sky-400 text-sm mb-1">Sapim (Belgium)</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">Leading in high-end aero spokes (e.g. CX-Ray). Known for lightweight performance and excellent fatigue life.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-sky-500/30 text-[10px] font-bold uppercase tracking-wider text-sky-400 hover:bg-sky-500/10 transition-colors">
                  View Sapim Spoke
                </button>
             </div>
             <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                <strong class="block text-amber-400 text-sm mb-1">Pillar (Taiwan)</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">Strong cost-performance with a wide range including reinforced (PSR) designs. Suitable for custom and mass production.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-amber-500/30 text-[10px] font-bold uppercase tracking-wider text-amber-400 hover:bg-amber-500/10 transition-colors">
                  View Pillar Spoke
                </button>
             </div>
             <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                <strong class="block text-slate-300 text-sm mb-1">Mac / CNSpoke</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">Fast supply chain response. Covers all types balancing cost and performance for diverse applications.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-slate-500/30 text-[10px] font-bold uppercase tracking-wider text-slate-400 hover:bg-slate-500/10 transition-colors">
                  View Mac / CNSpoke Spoke
                </button>
             </div>
             <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                <strong class="block text-slate-300 text-sm mb-1">YT Spokes (China)</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">Fast supply chain response, suitable for cost and performance balance.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-slate-500/30 text-[10px] font-bold uppercase tracking-wider text-slate-400 hover:bg-slate-500/10 transition-colors">
                  View YT Spokes Spoke
                </button>
             </div>
             <!-- New: Tanzanite Carbon -->
             <div class="bg-slate-800/40 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center ring-1 ring-indigo-500/20">
                <strong class="block text-indigo-400 text-sm mb-1">TANZANITE CARBON SPOKE</strong>
                <p class="text-xs text-slate-400 leading-relaxed mb-3">The pinnacle of lightweight performance. Designed for ultimate responsiveness and rigidity, drastically reducing rotational weight.</p>
                <button class="mt-auto px-4 py-1.5 rounded-full border border-indigo-500/30 text-[10px] font-bold uppercase tracking-wider text-indigo-400 hover:bg-indigo-500/10 transition-colors">
                  View Tanzanite Carbon Spoke
                </button>
             </div>
         </div>
      </div>

      <!-- New Content: Spoke Types -->
      <div class="w-full mb-8 text-center">
          <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 bg-slate-700"></span>
            <h4 class="text-slate-200 font-bold text-base uppercase tracking-wider">Spoke Types</h4>
            <span class="h-px w-8 bg-slate-700"></span>
         </div>

         <div class="space-y-4">
             <!-- Straight Gauge -->
              <div class="bg-slate-800/30 rounded-xl p-4 shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                 <h5 class="text-slate-100 font-bold mb-2 flex items-center justify-center gap-2">
                    <span class="w-2 h-2 rounded-full bg-slate-500"></span> Straight Gauge Spokes
                 </h5>
                 <div class="grid md:grid-cols-2 gap-x-6 gap-y-2 text-sm text-slate-400 w-full">
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Specs:</span> Commonly 14G (2.0mm), 13G (2.3mm).</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Features:</span> High strength margin, low cost, easy maintenance.</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Examples:</span> DT Champion, Sapim Leader.</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Use:</span> Commuting, touring, cargo, high-tension drive sides.</p>
                 </div>
             </div>

             <!-- Butted -->
               <div class="bg-slate-800/30 rounded-xl p-4 shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                 <h5 class="text-slate-100 font-bold mb-2 flex items-center justify-center gap-2">
                    <span class="w-2 h-2 rounded-full bg-sky-500"></span> Double/Triple Butted
                 </h5>
                 <div class="grid md:grid-cols-2 gap-x-6 gap-y-2 text-sm text-slate-400 w-full">
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Specs:</span> e.g. 2.0-1.8-2.0mm, 2.0-1.5-2.0mm.</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Features:</span> Lighter, better fatigue resistance (flex in midsection).</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Examples:</span> DT Competition/Revolution, Sapim Race/Laser.</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Use:</span> Endurance road, XC/Gravel, versatile durable builds.</p>
                 </div>
             </div>

              <!-- Aero -->
               <div class="bg-slate-800/30 rounded-xl p-4 shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center text-center">
                 <h5 class="text-slate-100 font-bold mb-2 flex items-center justify-center gap-2">
                    <span class="w-2 h-2 rounded-full bg-indigo-500"></span> Bladed / Aero
                 </h5>
                 <div class="grid md:grid-cols-2 gap-x-6 gap-y-2 text-sm text-slate-400 w-full">
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Specs:</span> Flat ~0.9-1.0mm, width 2.0-2.3mm (or wider for TT).</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Features:</span> Aerodynamic, easier torque control (anti-twist).</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Examples:</span> Sapim CX-Ray, DT Aerolite, Pillar PSR X-TRA.</p>
                    <p><span class="text-slate-500 font-mono text-xs uppercase mr-1">Use:</span> Racing road, disc brake high-tension, aero rims.</p>
                 </div>
             </div>
         </div>
      </div>

       <!-- New Content: Head Types -->
      <div class="w-full text-center">
          <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 bg-slate-700"></span>
            <h4 class="text-slate-200 font-bold text-base uppercase tracking-wider">Interfaces</h4>
             <span class="h-px w-8 bg-slate-700"></span>
         </div>
         <div class="grid gap-3 sm:grid-cols-3">
            <div class="bg-slate-800/20 p-3 rounded-lg border border-slate-700/20 text-center">
                 <strong class="block text-slate-300 text-sm mb-1">J-bend</strong>
                 <p class="text-xs text-slate-500">Highly versatile, standard, easy sourcing.</p>
            </div>
             <div class="bg-slate-800/20 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] text-center">
                 <strong class="block text-slate-300 text-sm mb-1">Straight-pull</strong>
                 <p class="text-xs text-slate-500">No elbow stress, fast building, specific hubs.</p>
            </div>
             <div class="bg-slate-800/20 p-3 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] text-center">
                 <strong class="block text-slate-300 text-sm mb-1">Nipples</strong>
                 <p class="text-xs text-slate-500">Brass (durable) vs Aluminum (light). Washer recommended.</p>
            </div>
         </div>
      </div>
    </div>

    <div class="mt-4">
      <div class="flex items-center justify-center gap-2 mb-4">
         <div class="h-px w-12 bg-slate-800"></div>
         <p class="text-xs font-bold uppercase tracking-widest text-slate-500">Material Comparison</p>
         <div class="h-px w-12 bg-slate-800"></div>
      </div>

      <!-- Disclaimer Note -->
      <div class="bg-amber-900/10 border-l-2 border-amber-600/50 p-3 mb-6 mx-auto max-w-3xl rounded-r-lg">
         <div class="flex gap-3">
            <span class="text-amber-500 shrink-0 mt-0.5">
               <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            </span>
            <p class="text-xs text-amber-500/80 leading-relaxed text-left">
               <strong class="font-bold text-amber-500">Note:</strong> The following comparison reflects relative performance characteristics across different spoke materials. Each material has its own strengths depending on the application. The listed pros and cons are not absolute judgments or defect classifications.
            </p>
         </div>
      </div>

    <ul class="technical-spokes-list mt-8 space-y-8 w-full">
        <!-- 1. Stainless Steel -->
        <li class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 transition-transform hover:-translate-y-1 duration-300 text-center">
          <div class="flex items-center justify-center gap-3 mb-4">
            <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-500 text-lg font-bold">1</span>
            <strong class="text-xl text-slate-100">Stainless Steel</strong>
          </div>
          
          <div class="w-full max-w-none text-slate-400 text-sm leading-relaxed flex flex-col items-center">
            <p class="mb-4 text-center">
              Currently the most mainstream and widely used spoke material on the market, offering the best balance of durability and cost.
            </p>

            <!-- Market Context -->
            <div class="mb-6 w-full text-center">
              <strong class="text-xs uppercase tracking-wider text-slate-500 mb-2 block">Market Landscape</strong>
              <p class="text-sm text-slate-400 max-w-2xl mx-auto">
                As detailed in the Basic Concepts above, the stainless steel spoke market is defined by the "Big 2" (DT Swiss, Sapim) setting the quality benchmark, with Pillar and others offering competitive high-value alternatives. All these major brands primarily utilize high-grade stainless steel (e.g., 18/8 or proprietary blends) for their main product lines.
              </p>
            </div>

            <div class="grid gap-4 md:grid-cols-2 w-full">
              <div class="bg-emerald-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-emerald-400 mb-2 text-center">Pros</strong>
                <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">High strength & durability</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Excellent corrosion resistance</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Cost-effective</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Good toughness (gradual deformation)</span>
                  </li>
                </ul>
              </div>

              <div class="bg-rose-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-rose-400 mb-2 text-center">Cons</strong>
                 <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Heavier than carbon/titanium</span>
                  </li>
                </ul>
              </div>
            </div>

            <p class="mt-4 text-xs font-medium text-slate-500 uppercase tracking-widest text-center">Best For</p>
            <p class="mt-1 text-slate-300 text-center">Almost all bicycles, from commuters to pro MTB/Road bikes.</p>
          </div>
        </li>

        <!-- 2. Carbon Fiber -->
        <li class="mt-6 rounded-2xl bg-[#0b0e14] shadow-[0_8px_30px_rgba(0,0,0,0.8)] border-l-4 border-indigo-500 p-5 md:p-6 transition-transform hover:-translate-y-1 duration-300 text-center">
           <div class="flex items-center justify-center gap-3 mb-4">
            <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-400 text-lg font-bold">2</span>
            <strong class="text-xl text-slate-100">Carbon Fiber</strong>
            <span class="text-[10px] font-bold uppercase tracking-wider bg-indigo-500 text-white px-2 py-1 rounded-full">Elite</span>
          </div>

           <div class="w-full max-w-none text-slate-400 text-sm leading-relaxed flex flex-col items-center">
            <p class="mb-4 text-center">
              Found on high-end performance wheelsets. Offers extreme weight savings and rigidity.
            </p>

             <div class="grid gap-4 md:grid-cols-2 w-full">
              <div class="bg-indigo-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-indigo-400 mb-2 text-center">Pros</strong>
                <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-indigo-400 shrink-0 text-center w-4">✓</span> <span class="text-center">Ultra-lightweight (-100g+)</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-indigo-400 shrink-0 text-center w-4">✓</span>  <span class="text-center">Excellent rigidity & acceleration</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-indigo-400 shrink-0 text-center w-4">✓</span>  <span class="text-center">Vibration absorption</span>
                  </li>
                </ul>
              </div>

               <div class="bg-rose-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-rose-400 mb-2 text-center">Cons</strong>
                 <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Expensive</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Poor side impact resistance</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Harder maintenance</span>
                  </li>
                </ul>
              </div>
            </div>

             <p class="mt-4 text-xs font-medium text-slate-500 uppercase tracking-widest text-center">Best For</p>
            <p class="mt-1 text-slate-300 text-center">High-end road racing, time trials, and weight weenies.</p>
           </div>
        </li>

        <!-- 3. Titanium Alloy (Compact) -->
        <!-- 3. Titanium Alloy -->
        <li class="mt-6 rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 transition-transform hover:-translate-y-1 duration-300 opacity-90 text-center">
           <div class="flex items-center justify-center gap-3 mb-4">
             <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700/30 text-slate-400 text-lg font-bold">3</span>
            <strong class="text-xl text-slate-200">Titanium Alloy</strong>
            <span class="text-[10px] border border-slate-700 text-slate-500 px-2 py-0.5 rounded">Rare</span>
          </div>

          <div class="w-full max-w-none text-slate-400 text-sm leading-relaxed flex flex-col items-center">
             <p class="mb-4 text-center">
               Middle ground between steel and carbon. High strength-to-weight ratio and no rust.
             </p>

            <div class="grid gap-4 md:grid-cols-2 w-full">
              <!-- Pros -->
              <div class="bg-emerald-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-emerald-400 mb-2 text-center">Pros</strong>
                <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Lightweight</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Smooth Ride Quality</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">No Rust (Corrosion Free)</span>
                  </li>
                </ul>
              </div>

              <!-- Cons -->
              <div class="bg-rose-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-rose-400 mb-2 text-center">Cons</strong>
                 <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Very Expensive</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Limited Availability</span>
                  </li>
                </ul>
              </div>
            </div>

            <p class="mt-4 text-xs font-medium text-slate-500 uppercase tracking-widest text-center">Best For</p>
            <p class="mt-1 text-slate-300 text-center">Custom builds, riders valuing comfort & aesthetics.</p>
          </div>
        </li>

         <!-- 4. Aluminum Alloy -->
        <li class="mt-4 rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 transition-transform hover:-translate-y-1 duration-300 opacity-90 text-center">
           <div class="flex items-center justify-center gap-3 mb-4">
             <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-700/30 text-slate-400 text-lg font-bold">4</span>
            <strong class="text-xl text-slate-200">Aluminum Alloy</strong>
             <span class="text-[10px] border border-slate-700 text-slate-500 px-2 py-0.5 rounded">Rare</span>
          </div>

          <div class="w-full max-w-none text-slate-400 text-sm leading-relaxed flex flex-col items-center">
             <p class="mb-4 text-center">
               Used by specific manufacturers (e.g. Campagnolo). Very light but can be problematic.
             </p>

             <div class="grid gap-4 md:grid-cols-2 w-full">
              <!-- Pros -->
              <div class="bg-emerald-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-emerald-400 mb-2 text-center">Pros</strong>
                <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Very Light</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                    <span class="text-emerald-500 shrink-0 text-center w-4">✓</span> <span class="text-center">Stiff</span>
                  </li>
                </ul>
              </div>

              <!-- Cons -->
              <div class="bg-rose-500/5 rounded-lg p-3 shadow-[0_4px_16px_rgba(0,0,0,0.5)] w-full">
                <strong class="block text-xs uppercase tracking-wider text-rose-400 mb-2 text-center">Cons</strong>
                 <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Fatigue Prone</span>
                  </li>
                  <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Corrosion Issues</span>
                  </li>
                   <li class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                     <span class="text-rose-500 shrink-0 text-center w-4">×</span> <span class="text-center">Proprietary (Hard to Replace)</span>
                  </li>
                </ul>
              </div>
            </div>

             <p class="mt-4 text-xs font-medium text-slate-500 uppercase tracking-widest text-center">Best For</p>
            <p class="mt-1 text-slate-300 text-center">Specific factory racing wheelsets.</p>
          </div>
        </li>
      </ul>
      <p class="sizecharts-section__intro mt-3">
        Which material you choose ultimately depends on your budget, riding style and
        specific performance needs such as weight, stiffness and comfort.
      </p>
    </div>
    </div>

    <!-- Tab: Choose spoke -->
    <div v-show="activeTab === 'choose-spoke'">
       <div class="space-y-8">
         <div class="bg-[#11151e] rounded-2xl shadow-[0_8px_30px_rgba(0,0,0,0.6)] border-t-4 border-emerald-500 p-6 text-center">
            <h3 class="text-xl font-bold text-slate-100 mb-2">Spoke Selection Matrix</h3>
            <p class="text-slate-400 text-sm max-w-2xl mx-auto">
               Compare our supported spoke models to find the perfect match for your building goals—whether it's maximum durability, lightweight performance, or aerodynamic efficiency.
            </p>
         </div>

         <div class="overflow-x-auto rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800">
            <table class="w-full text-sm text-left text-slate-400">
               <thead class="text-xs text-slate-200 uppercase bg-slate-800/80">
                  <tr>
                     <th scope="col" class="px-6 py-4 font-bold">Brand / Model</th>
                     <th scope="col" class="px-6 py-4 font-bold text-center">Type</th>
                     <th scope="col" class="px-6 py-4 font-bold text-center">Weight</th>
                     <th scope="col" class="px-6 py-4 font-bold text-center">Section (mm)</th>
                     <th scope="col" class="px-6 py-4 font-bold">Details</th>
                  </tr>
               </thead>
               <tbody class="divide-y divide-slate-800">
                  
                  <!-- SAPIM Models -->
                   <!-- 1. Leader -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">SAPIM Leader</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-slate-800 text-slate-300 text-xs font-bold border border-slate-700">Straight</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-slate-300">6.8g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">2.0</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                        <strong>Stainless (18/8)</strong>. Strength: 1080-1180 N/mm². Durable straight gauge choice.
                     </td>
                  </tr>

                  <!-- 2. Race -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">SAPIM Race</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-sky-500/10 text-sky-400 text-xs font-bold border border-sky-500/20">Butted</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-slate-300">5.7g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">2.0 - 1.8 - 2.0</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                        <strong>Strength: 1300 N/mm²</strong>. Evolution of classic Competition. Good fatigue adaptability.
                     </td>
                  </tr>

                   <!-- 3. D-Light -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">SAPIM D-Light</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-sky-500/10 text-sky-400 text-xs font-bold border border-sky-500/20">Butted</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-slate-300">4.8g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">2.0 - 1.65 - 2.0</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                         <strong>Strength: 1370 N/mm²</strong>. Designed for top quality MTB/Disc wheels. Max strength, min weight.
                     </td>
                  </tr>

                   <!-- 4. CX-Sprint -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">SAPIM CX-Sprint</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-indigo-500/10 text-indigo-400 text-xs font-bold border border-indigo-500/20">Aero</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-slate-300">5.2g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">2.25 x 1.25</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                        <strong>Strength: 1430 N/mm²</strong>. Aero bladed. High stiffness for drive-side applications.
                     </td>
                  </tr>

                   <!-- 5. CX-Ray -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <div class="flex items-center gap-2">
                            <strong class="block text-slate-200 text-base">SAPIM CX-Ray</strong>
                            <span class="text-[10px] bg-amber-500 text-slate-900 px-1.5 py-0.5 rounded font-bold">TOP</span>
                        </div>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-indigo-500/10 text-indigo-400 text-xs font-bold border border-indigo-500/20">Aero</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-emerald-400 font-bold">4.4g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">0.9 x 2.2</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                        <strong>Strength: 1600 N/mm²</strong>. The gold standard. Light as titanium, strong as steel. Highest fatigue life.
                     </td>
                  </tr>

                   <!-- 6. Super-CX -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors opacity-75">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">SAPIM Super-CX</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-indigo-500/10 text-indigo-400 text-xs font-bold border border-indigo-500/20">Aero</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-emerald-400 font-bold">3.6g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">0.9 x 2.0</td>
                     <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px]">
                         <strong>Strength: 1430 N/mm²</strong>. Lightest stainless spoke ever. Extra light wheels.
                     </td>
                  </tr>

                  <!-- DT Swiss Comp Race -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors border-t-4 border-slate-800">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">DT Swiss Comp Race</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-sky-500/10 text-sky-400 text-xs font-bold border border-sky-500/20">Butted</span>
                     </td>
                     <td class="px-6 py-4 text-center font-mono text-slate-300">5.7g</td>
                     <td class="px-6 py-4 text-center font-mono text-xs">2.0 - 1.6 - 2.0</td>
                     <td class="px-6 py-4 text-xs italic text-slate-500">
                        Coming Soon
                     </td>
                  </tr>

                  <!-- YT Aero -->
                  <tr class="bg-[#11151e] hover:bg-slate-800/30 transition-colors">
                     <td class="px-6 py-4">
                        <strong class="block text-slate-200 text-base">YT Aero</strong>
                     </td>
                     <td class="px-6 py-4 text-center">
                        <span class="inline-flex px-2 py-1 rounded bg-amber-500/10 text-amber-400 text-xs font-bold border border-amber-500/20">Aero</span>
                     </td>
                     <td class="px-6 py-4 text-center text-slate-500 italic">Coming Soon</td>
                     <td class="px-6 py-4 text-center text-slate-500 italic">Coming Soon</td>
                     <td class="px-6 py-4 text-xs italic text-slate-500">
                        Coming Soon
                     </td>
                  </tr>

                  <!-- TANZANITE Carbon -->
                  <tr class="bg-[#0b0e14] hover:bg-slate-900 transition-colors relative overflow-hidden group border-t border-indigo-900/50">
                     <td class="px-6 py-4 relative z-10">
                        <div class="flex items-center gap-3">
                           <div class="w-1 h-8 bg-indigo-500 rounded-full shadow-[0_0_10px_rgba(99,102,241,0.5)]"></div>
                           <div>
                              <strong class="flex items-center gap-2 text-indigo-400 text-base uppercase tracking-wide">
                                 TANZANITE Carbon
                              </strong>
                           </div>
                        </div>
                     </td>
                     <td class="px-6 py-4 text-center relative z-10">
                        <span class="inline-flex px-2 py-1 rounded bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-500/30">Aero</span>
                     </td>
                     <td class="px-6 py-4 text-center text-indigo-300 font-mono font-bold relative z-10">Coming Soon</td>
                     <td class="px-6 py-4 text-center font-mono text-xs text-indigo-300 relative z-10">Coming Soon</td>
                     <td class="px-6 py-4 text-xs leading-relaxed max-w-xs text-indigo-200/80 relative z-10 italic">
                        Proprietary carbon manufacturing process. Provides 3x specificity stiffness of steel.
                     </td>
                     <!-- Shine effect -->
                     <td class="absolute inset-0 bg-indigo-500/5 group-hover:bg-indigo-500/10 transition-colors pointer-events-none"></td>
                  </tr>
               </tbody>
            </table>
         </div>

         <!-- Additional Info Cards -->
         <div class="grid md:grid-cols-2 gap-4">
             <div class="bg-slate-800/40 border border-slate-700/50 p-4 rounded-xl flex gap-4 items-start">
                <span class="bg-slate-700/50 p-2 rounded-lg text-slate-300 shrink-0">
                   <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="M12 5l7 7-7 7"/></svg>
                </span>
                <div>
                   <h5 class="text-slate-200 font-bold text-sm mb-1">Weight Calculation</h5>
                   <p class="text-xs text-slate-400">Weights listed are approximations for typical 64pcs (2 wheels, 32H) lengths (260mm-270mm).</p>
                </div>
             </div>
             <div class="bg-indigo-500/10 border border-indigo-500/20 p-4 rounded-xl flex gap-4 items-start">
                 <span class="bg-indigo-500/20 p-2 rounded-lg text-indigo-400 shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>
                 </span>
                <div>
                   <h5 class="text-indigo-300 font-bold text-sm mb-1">Custom Requests</h5>
                   <p class="text-xs text-indigo-300/70">Need a specific spoke not listed here? We can source Sapim CX-Ray or other premium options upon request.</p>
                </div>
             </div>
         </div>

         <!-- Need Help Section -->
         <div class="mt-8 rounded-2xl bg-gradient-to-r from-slate-900 to-slate-800 p-6 shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-700/50">
            <div class="flex flex-col md:flex-row items-center justify-between gap-6 text-center md:text-left">
               <div class="space-y-2">
                  <h4 class="text-lg font-bold text-slate-100">Need Spoke Length Help?</h4>
                  <p class="text-sm text-slate-400 max-w-xl leading-relaxed">
                     If you are looking to purchase individual replacement spokes or rebuild a previous TANZANITE wheelset, you may require assistance with spoke length calculations.
                  </p>
               </div>
               <div class="flex flex-col sm:flex-row gap-3 shrink-0">
                  <button
                     type="button"
                     class="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-full bg-emerald-500 hover:bg-emerald-400 text-white text-sm font-bold shadow-lg shadow-emerald-500/20 transition-all group"
                     @click="goToSpokeCalculator"
                  >
                     <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="group-hover:scale-110 transition-transform"><rect x="4" y="2" width="16" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12.01" y2="18"></line><line x1="8" y1="18" x2="8.01" y2="18"></line><line x1="16" y1="18" x2="16.01" y2="18"></line><line x1="12" y1="14" x2="12.01" y2="14"></line><line x1="8" y1="14" x2="8.01" y2="14"></line><line x1="16" y1="14" x2="16.01" y2="14"></line><line x1="12" y1="10" x2="12.01" y2="10"></line><line x1="8" y1="10" x2="8.01" y2="10"></line><line x1="16" y1="10" x2="16.01" y2="10"></line></svg>
                     Spoke Calculator
                  </button>

                  <button
                     type="button"
                     class="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-full bg-sky-500 hover:bg-sky-400 text-white text-sm font-bold shadow-lg shadow-sky-500/20 transition-all group"
                     @click="openSupportChat"
                  >
                     <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="group-hover:scale-110 transition-transform"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>
                     Live Chat
                  </button>

                  <a
                     href="mailto:support@tanzanite.site"
                     class="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-full bg-slate-700 hover:bg-slate-600 text-white text-sm font-bold shadow-lg border border-slate-600 transition-all hover:border-slate-500"
                  >
                     <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
                     Email Us
                  </a>
               </div>
            </div>
         </div>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useLocalePath, useRouter } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { useChatWidget } from '~/composables/useChatWidget'

const router = useRouter()
const localePath = useLocalePath()
const { openChat } = useChatWidget()
const activeTab = ref<'basic-concepts' | 'choose-spoke'>('basic-concepts')

const goToSpokeCalculator = async () => {
  await router.push(localePath('/spoke-calculator'))
}

const openSupportChat = () => {
  openChat({ showAgentList: true })
}
</script>
