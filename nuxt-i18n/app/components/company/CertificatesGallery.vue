<template>
  <div class="flex flex-col gap-8 md:gap-12">
    
    <!-- Hero Section -->
    <section class="relative text-center">
      <h2 class="text-2xl md:text-3xl font-bold text-white mb-4 tracking-tight">
        Certified <span class="text-teal-400">Quality & Safety</span>
      </h2>
      <p class="text-slate-300 max-w-3xl mx-auto text-base md:text-lg leading-relaxed font-light">
        TANZANITE products are UCI approved and exceed ISO 4210 international standards, ensuring ultimate safety for every ride.
      </p>
    </section>

    <!-- UCI Accreditation (Highlight) -->
    <section class="max-w-6xl mx-auto w-full">
      <div class="relative bg-gradient-to-br from-[#1a1f2e] to-[#0f1219] rounded-3xl overflow-hidden border border-slate-700/50 shadow-2xl">
        <div class="grid grid-cols-1 md:grid-cols-2">
            <!-- Content -->
            <div class="p-4 md:p-12 flex flex-col justify-center">
                <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border border-white/10 w-fit mb-6">
                   <div class="w-2 h-2 rounded-full bg-green-400 animate-pulse"></div>
                   <span class="text-xs font-semibold text-slate-300 tracking-wider uppercase">Official Certification</span>
                </div>
                <h3 class="text-2xl md:text-3xl font-bold text-white mb-4">UCI Approved Wheels</h3>
                <p class="text-slate-300 mb-6 leading-relaxed text-sm md:text-base">
                    TANZANITE wheels have passed the stringent impact tests required by the Union Cycliste Internationale (UCI). They are fully approved for use in all UCI-sanctioned road, track, and cyclo-cross races.
                </p>
                 <ul class="space-y-3 mb-8">
                    <li v-for="item in uciModels" :key="item" class="flex items-center gap-3">
                        <svg class="w-5 h-5 text-teal-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                        <span class="text-slate-200 text-sm">{{ item }}</span>
                    </li>
                 </ul>
                 <a href="https://www.uci.org" target="_blank" rel="noopener" class="inline-flex items-center gap-2 text-teal-400 hover:text-teal-300 font-semibold transition-colors text-sm group w-fit">
                    <span>View on UCI Website</span>
                    <svg class="w-4 h-4 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" /></svg>
                 </a>
            </div>
            <!-- Visual -->
            <div class="relative min-h-[300px] md:min-h-full bg-slate-800 flex items-center justify-center p-8 overflow-hidden group">
               <!-- Decorative elements -->
               <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-teal-500/10 via-transparent to-transparent opacity-50"></div>
               
               <!-- Placeholder for UCI Logo/Wheel Image -->
               <div class="relative z-10 text-center">
                   <!-- UCI Logo Simulation -->
                    <div class="w-48 h-16 mx-auto mb-6 bg-white/5 backdrop-blur rounded-lg border border-white/10 flex items-center justify-center">
                        <span class="text-2xl font-black tracking-widest text-white/50">UCI</span>
                    </div>
                   <div class="w-48 h-48 mx-auto rounded-full border-4 border-slate-700/50 bg-slate-900 shadow-2xl relative group-hover:scale-105 transition-transform duration-700 flex items-center justify-center">
                        <div class="text-slate-600 text-[10px] uppercase tracking-widest">
                           UCI Approved<br>Sticker
                        </div>
                   </div>
               </div>
            </div>
        </div>
      </div>
    </section>

    <!-- Standards & Certificates Grid -->
    <section class="max-w-6xl mx-auto w-full">
        <div class="text-center mb-6">
            <h3 class="text-xl md:text-2xl font-bold text-white mb-3">Global Standards Compliance</h3>
            <p class="text-slate-400 text-sm max-w-2xl mx-auto">Verified by leading third-party testing laboratories including SGS and Intertek.</p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div v-for="(cert, index) in certificates" :key="index" class="group relative bg-[#11151e] rounded-2xl p-4 md:p-6 border border-slate-700/50 hover:border-teal-500/30 transition-all duration-300 hover:shadow-xl hover:shadow-teal-500/5 hover:-translate-y-1">
                
                <h4 class="text-lg font-bold text-white mb-2">{{ cert.title }}</h4>
                <p class="text-xs text-slate-400 mb-4 h-10 line-clamp-2">{{ cert.description }}</p>
                
                <div class="flex items-center justify-between mt-auto pt-4 border-t border-slate-800">
                    <span class="text-xs font-mono text-slate-500">{{ cert.code }}</span>
                    <button class="text-xs font-semibold text-white bg-slate-800 hover:bg-slate-700 px-3 py-1.5 rounded-lg transition-colors cursor-not-allowed opacity-50">
                        View Report
                    </button>
                </div>
            </div>
        </div>
    </section>

    <!-- Internal Standards -->
    <section class="max-w-4xl mx-auto w-full">
        <div class="bg-slate-900/80 backdrop-blur rounded-2xl border border-slate-800 p-4 md:p-8 text-center relative overflow-hidden">
             <!-- Background Glow -->
             <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full max-w-lg bg-teal-500/5 blur-[80px] pointer-events-none"></div>
             
             <div class="relative z-10">
                 <h3 class="text-xl font-bold text-white mb-4">TANZANITE In-House Standard</h3>
                 <p class="text-slate-300 text-sm leading-relaxed mb-6">
                     We don't just meet standards; we set them. Our internal testing protocols are rigorous, subjecting simulated road conditions to forces <span class="text-teal-400 font-bold">120% - 150%</span> higher than ISO 4210 requirements.
                 </p>
                 <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                     <div class="p-3 rounded-xl bg-black/20 border border-white/5">
                         <div class="text-2xl font-bold text-white mb-1">100+</div>
                         <div class="text-[10px] text-slate-400 uppercase tracking-wider">Test Cycles</div>
                     </div>
                     <div class="p-3 rounded-xl bg-black/20 border border-white/5">
                         <div class="text-2xl font-bold text-white mb-1">100%</div>
                         <div class="text-[10px] text-slate-400 uppercase tracking-wider">QC Inspection</div>
                     </div>
                      <div class="p-3 rounded-xl bg-black/20 border border-white/5">
                         <div class="text-2xl font-bold text-white mb-1">3 Year</div>
                         <div class="text-[10px] text-slate-400 uppercase tracking-wider">Warranty</div>
                     </div>
                      <div class="p-3 rounded-xl bg-black/20 border border-white/5">
                         <div class="text-2xl font-bold text-white mb-1">24/7</div>
                         <div class="text-[10px] text-slate-400 uppercase tracking-wider">Support</div>
                     </div>
                 </div>
             </div>
        </div>
    </section>

  </div>
</template>

<script setup lang="ts">
const uciModels = [
    'TANZANITE PRO 50 Disc',
    'TANZANITE ULTRA 45 Rim',
    'TANZANITE GRAVEL X',
    'TANZANITE TRACK AERO'
]

const certificates = [
    {
        title: 'ISO 4210-7:2014',
        description: 'Safety requirements for bicycles - Part 7: Wheels and rims test methods.',
        code: 'ISO-4210',
    },
    {
        title: 'SGS Impact Test',
        description: 'Vertical impact test with energy levels exceeding standard requirements.',
        code: 'SGS-IMP-2024',
    },
    {
        title: 'Intertek Fatigue Test',
        description: 'Dynamic fatigue testing ensuring long-term durability and reliability.',
        code: 'ITK-FAT-2024',
    }
]
</script>
