<template>
  <div class="tubeless-guide space-y-8">
    <!-- 1. Intro Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold text-slate-100 mb-6 flex items-center justify-center gap-2">
        Tubeless Tires Overview
      </h2>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-5xl mx-auto">
         <GuideImage
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
            src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires1.webp"
            alt="Comparison of tubeless, inner tube and tubular road bike tires"
            :zoomOnClick="true"
            caption="Construction Differences: Tube vs Tubeless vs Tubular"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
            src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires2.webp"
            alt="Pros and cons chart comparing tubeless, inner tube and tubular tires"
            :zoomOnClick="true"
            caption="Performance Matrix: Resistance, Protection & Comfort"
          />
      </div>
    </div>

    <!-- 2. System Comparison (Definitions) -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-3 mb-6">
         <span class="h-px w-8 bg-slate-700"></span>
         <h3 class="text-lg font-bold text-slate-100 uppercase tracking-wider">System Definitions</h3>
         <span class="h-px w-8 bg-slate-700"></span>
       </div>

       <div class="grid md:grid-cols-2 gap-8 items-stretch">
          <!-- Inner Tube Card -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-5 shadow-[0_4px_16px_rgba(0,0,0,0.45)] flex flex-col items-center border border-[rgba(255,255,255,0.08)]">
             <strong class="block tz-text-primary text-lg font-bold mb-3">Inner Tube System</strong>
          <p class="tz-text-secondary text-sm leading-relaxed mb-4">
               The classic standard. Tires composed of "inner tube + outer tire". The pneumatic tire resides in the inner cavity.
             </p>
             <div class="mt-auto pt-4 border-t border-[rgba(255,255,255,0.08)] w-full text-center">
                <span class="text-xs font-mono tz-text-muted uppercase">Universal Compatibility</span>
             </div>
          </div>

           <!-- Tubeless Card -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-5 shadow-[0_4px_16px_rgba(0,0,0,0.45)] flex flex-col items-center border border-[rgba(255,255,255,0.08)] ring-1 ring-[rgba(255,255,255,0.04)]">
             <strong class="block tz-text-primary text-lg font-bold mb-3">Tubeless System</strong>
          <p class="tz-text-secondary text-sm leading-relaxed mb-4">
               No inner tube required. The tire, specialized rim, and valve create a closed air-tight space, similar to automotive tires.
             </p>
             <div class="mt-auto pt-4 border-t border-[rgba(255,255,255,0.08)] w-full text-center">
                <span class="text-xs font-mono tz-text-muted uppercase">Performance Standard</span>
             </div>
          </div>
       </div>

       <div class="mt-8">
          <button
            type="button"
              class="inline-flex items-center justify-center rounded-full bg-[var(--tz-card-surface)] border border-[rgba(255,255,255,0.1)] px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-[0_4px_16px_rgba(0,0,0,0.35)] transition-colors hover:border-[rgba(181,255,109,0.28)] hover:bg-white/[0.03] hover:text-[var(--tz-brand-primary)]"
            @click="$emit('changeTab', 'choose')"
          >
            Check Specifications
          </button>
       </div>
    </div>

    <!-- 3. Suitability Card -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center border border-[rgba(255,255,255,0.08)]">
        <h3 class="text-xl font-bold text-slate-100 mb-4">Is Tubeless For Me?</h3>
      <p class="tz-text-secondary text-sm leading-relaxed max-w-3xl mx-auto mb-6">
           If you ride daily commutes or casual leisure, <strong>Inner Tubes</strong> are sufficient and low maintenance.
           <br class="mb-2 block content-['']" />
           If you ride on <strong>rough terrain</strong>, desire <strong>lower rolling resistance</strong>, or want <strong>advanced puncture protection</strong>, then <strong>Tubeless</strong> is the superior choice.
        </p>

        <!-- Compatibility Note -->
        <div class="mt-6 p-4 rounded-xl bg-[var(--tz-card-surface)] border border-amber-500/20 text-left max-w-3xl mx-auto shadow-[0_4px_16px_rgba(0,0,0,0.35)]">
          <strong class="block text-amber-400 text-sm mb-2">Note:</strong>
              <p class="tz-text-secondary text-xs leading-relaxed">
            If you previously used an aluminum rim with a traditional inner tube and tire setup, and now plan to upgrade to a tubeless system, please first refer to the rim cross-section diagram above.
            <br/><br/>
            Rims designed for tubeless compatibility can support both inner tube setups and tubeless tires. However, if your rim matches the first example shown above—designed only for standard inner tube use—it is not compatible with tubeless tires and must be replaced.
          </p>
        </div>
     </div>

    <!-- 4. Key Differences Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold text-slate-100 mb-6 uppercase tracking-wider">Key Advantages Overview</h3>
       
       <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-[0_4px_16px_rgba(0,0,0,0.45)] border border-[rgba(255,255,255,0.08)]">
              <strong class="block tz-text-primary text-sm mb-2">Self-Sealing</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">Sealant fills small punctures automatically instantly. No snake bites.</p>
          </div>
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-[0_4px_16px_rgba(0,0,0,0.45)] border border-[rgba(255,255,255,0.08)]">
              <strong class="block tz-text-primary text-sm mb-2">Lower Pressure</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">Can run ~80psi (Road) or lower for better grip and comfort without pinch flats.</p>
          </div>
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-[0_4px_16px_rgba(0,0,0,0.45)] border border-[rgba(255,255,255,0.08)]">
              <strong class="block tz-text-primary text-sm mb-2">Rolling Resistance</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">No internal friction between tube and tire means faster rolling speed.</p>
          </div>
       </div>
    </div>

    <!-- Navigation CTAs -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-6 text-center flex flex-col justify-center items-center hover:-translate-y-1 transition-transform">
           <h4 class="text-slate-100 font-bold mb-2">Installation Guide</h4>
           <p class="text-xs tz-text-muted mb-4">Step-by-step visual instructions</p>
           <button
             type="button"
              class="rounded-full bg-[var(--tz-card-surface)] border border-[rgba(255,255,255,0.1)] px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-[0_4px_16px_rgba(0,0,0,0.35)] transition-colors hover:border-[rgba(181,255,109,0.28)] hover:bg-white/[0.03] hover:text-[var(--tz-brand-primary)]"
             @click="$emit('changeTab', 'installation')"
           >
             Go to Installation
           </button>
        </div>

        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-6 text-center flex flex-col justify-center items-center hover:-translate-y-1 transition-transform">
           <h4 class="text-slate-100 font-bold mb-2">Pressure Check</h4>
           <p class="text-xs tz-text-muted mb-4">Find optimal PSI for Tubeless</p>
           <button
             type="button"
              class="rounded-full bg-[var(--tz-card-surface)] border border-[rgba(255,255,255,0.1)] px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-[0_4px_16px_rgba(0,0,0,0.35)] transition-colors hover:border-[rgba(181,255,109,0.28)] hover:bg-white/[0.03] hover:text-[var(--tz-brand-primary)]"
             @click="$emit('changeTab', 'choose')"
           >
             Check Pressure
           </button>
        </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

// Match SizeChartsTabId from parent
type GuideTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

defineEmits<{
  (e: 'changeTab', tab: GuideTabId): void
}>()
</script>
