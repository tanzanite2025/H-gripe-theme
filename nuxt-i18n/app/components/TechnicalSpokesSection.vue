<template>
  <div>
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
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
</script>
