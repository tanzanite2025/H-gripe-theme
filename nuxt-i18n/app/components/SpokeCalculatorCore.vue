<template>
  <div class="spoke-calculator">
    <div class="grid gap-6 items-start">
      <section class="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 shadow-sm">
        <h2 class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400 mb-4">
          Wheel setup
        </h2>

        <!-- Two-column layout: Front Wheel | Rear Wheel -->
        <div class="grid gap-6 md:grid-cols-2">
          <!-- ========== FRONT WHEEL COLUMN ========== -->
          <div class="space-y-4 p-4 rounded-xl border border-slate-800 bg-slate-950/40">
            <h3 class="text-sm font-semibold text-sky-400 uppercase tracking-wide">Front Wheel</h3>

            <!-- Spoke count -->
            <div class="space-y-1.5">
              <label for="front-spoke-count" class="block text-xs font-medium text-slate-200">Spoke count</label>
              <select id="front-spoke-count" v-model.number="frontConfig.spokeCount" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option :value="24">24</option>
                <option :value="28">28</option>
                <option :value="32">32</option>
                <option :value="36">36</option>
              </select>
            </div>

            <!-- Lacing pattern -->
            <div class="space-y-1.5">
              <label for="front-lacing" class="block text-xs font-medium text-slate-200">Lacing pattern</label>
              <select id="front-lacing" v-model.number="frontConfig.crossing" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option :value="2">2-cross</option>
                <option :value="3">3-cross</option>
                <option :value="4">4-cross</option>
              </select>
            </div>

            <!-- Nipple type -->
            <div class="space-y-1.5">
              <label for="front-nipple" class="block text-xs font-medium text-slate-200">Nipple type</label>
              <select id="front-nipple" v-model="frontConfig.nippleType" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option value="standard">Standard external</option>
                <option value="hidden">Hidden / aero</option>
              </select>
            </div>

            <!-- Rim -->
            <div class="space-y-1.5">
              <label for="front-rim" class="block text-xs font-medium text-slate-200">
                Rim <span class="ml-1 text-[10px] font-normal text-slate-400">(select from shop rims)</span>
              </label>
              <select id="front-rim" v-model="frontConfig.rimId" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option v-if="!rimOptions.length" value="" disabled>No rim products configured yet</option>
                <option v-for="rim in rimOptions" :key="rim.id" :value="rim.id">{{ rim.label }}</option>
              </select>
            </div>

            <!-- Hub -->
            <div class="space-y-1.5">
              <label for="front-hub" class="block text-xs font-medium text-slate-200">
                Hub <span class="ml-1 text-[10px] font-normal text-slate-400">(select from shop hubs)</span>
              </label>
              <select id="front-hub" v-model="frontConfig.hubId" :disabled="!frontHubOptions.length" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500 disabled:opacity-50 disabled:cursor-not-allowed">
                <option v-if="!frontHubOptions.length" value="" disabled>No hubs available</option>
                <option v-for="hub in frontHubOptions" :key="hub.id" :value="hub.id">{{ hub.label }}</option>
              </select>
            </div>

            <!-- ERD -->
            <div class="space-y-1.5">
              <label for="front-erd" class="block text-xs font-medium text-slate-200">ERD (effective rim diameter)</label>
              <div class="flex items-center gap-2">
                <input id="front-erd" v-model.number="frontConfig.erd" type="number" min="400" max="750" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange distance -->
            <div class="space-y-1.5">
              <label for="front-left-flange" class="block text-xs font-medium text-slate-200">Left flange distance</label>
              <div class="flex items-center gap-2">
                <input id="front-left-flange" v-model.number="frontConfig.leftFlange" type="number" min="10" max="60" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange distance -->
            <div class="space-y-1.5">
              <label for="front-right-flange" class="block text-xs font-medium text-slate-200">Right flange distance</label>
              <div class="flex items-center gap-2">
                <input id="front-right-flange" v-model.number="frontConfig.rightFlange" type="number" min="10" max="60" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>
          </div>

          <!-- ========== REAR WHEEL COLUMN ========== -->
          <div class="space-y-4 p-4 rounded-xl border border-slate-800 bg-slate-950/40">
            <h3 class="text-sm font-semibold text-emerald-400 uppercase tracking-wide">Rear Wheel</h3>

            <!-- Spoke count -->
            <div class="space-y-1.5">
              <label for="rear-spoke-count" class="block text-xs font-medium text-slate-200">Spoke count</label>
              <select id="rear-spoke-count" v-model.number="rearConfig.spokeCount" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option :value="24">24</option>
                <option :value="28">28</option>
                <option :value="32">32</option>
                <option :value="36">36</option>
              </select>
            </div>

            <!-- Lacing pattern -->
            <div class="space-y-1.5">
              <label for="rear-lacing" class="block text-xs font-medium text-slate-200">Lacing pattern</label>
              <select id="rear-lacing" v-model.number="rearConfig.crossing" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option :value="2">2-cross</option>
                <option :value="3">3-cross</option>
                <option :value="4">4-cross</option>
              </select>
            </div>

            <!-- Nipple type -->
            <div class="space-y-1.5">
              <label for="rear-nipple" class="block text-xs font-medium text-slate-200">Nipple type</label>
              <select id="rear-nipple" v-model="rearConfig.nippleType" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option value="standard">Standard external</option>
                <option value="hidden">Hidden / aero</option>
              </select>
            </div>

            <!-- Rim -->
            <div class="space-y-1.5">
              <label for="rear-rim" class="block text-xs font-medium text-slate-200">
                Rim <span class="ml-1 text-[10px] font-normal text-slate-400">(select from shop rims)</span>
              </label>
              <select id="rear-rim" v-model="rearConfig.rimId" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
                <option v-if="!rimOptions.length" value="" disabled>No rim products configured yet</option>
                <option v-for="rim in rimOptions" :key="rim.id" :value="rim.id">{{ rim.label }}</option>
              </select>
            </div>

            <!-- Hub -->
            <div class="space-y-1.5">
              <label for="rear-hub" class="block text-xs font-medium text-slate-200">
                Hub <span class="ml-1 text-[10px] font-normal text-slate-400">(select from shop hubs)</span>
              </label>
              <select id="rear-hub" v-model="rearConfig.hubId" :disabled="!rearHubOptions.length" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500 disabled:opacity-50 disabled:cursor-not-allowed">
                <option v-if="!rearHubOptions.length" value="" disabled>No hubs available</option>
                <option v-for="hub in rearHubOptions" :key="hub.id" :value="hub.id">{{ hub.label }}</option>
              </select>
            </div>

            <!-- ERD -->
            <div class="space-y-1.5">
              <label for="rear-erd" class="block text-xs font-medium text-slate-200">ERD (effective rim diameter)</label>
              <div class="flex items-center gap-2">
                <input id="rear-erd" v-model.number="rearConfig.erd" type="number" min="400" max="750" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange distance -->
            <div class="space-y-1.5">
              <label for="rear-left-flange" class="block text-xs font-medium text-slate-200">Left flange distance</label>
              <div class="flex items-center gap-2">
                <input id="rear-left-flange" v-model.number="rearConfig.leftFlange" type="number" min="10" max="60" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange distance -->
            <div class="space-y-1.5">
              <label for="rear-right-flange" class="block text-xs font-medium text-slate-200">Right flange distance</label>
              <div class="flex items-center gap-2">
                <input id="rear-right-flange" v-model.number="rearConfig.rightFlange" type="number" min="10" max="60" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500" />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Tools row (spans full width) -->
        <div class="mt-6 space-y-1.5">
          <label for="tools" class="block text-xs font-medium text-slate-200">Tools</label>
          <select id="tools" v-model="selectedTool" class="block w-full rounded-lg border border-slate-700 bg-slate-950/80 px-3 py-2.5 text-sm text-slate-50 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500">
            <option value="spoke-wrench">Spoke wrench</option>
            <option value="truing-stand">Truing stand</option>
            <option value="tension-meter">Tension meter</option>
            <option value="dishing-tool">Dishing tool</option>
          </select>
        </div>

        <!-- Action row -->
        <div class="mt-6 flex flex-col gap-3 md:flex-row md:items-center md:justify-between border-t border-slate-800 pt-4">
          <p class="text-[11px] text-slate-500 max-w-md">
            This is only a visual prototype. Replace the mock formula in the script section with your own calculation logic.
          </p>
          <div class="flex items-center gap-3">
            <button
              type="button"
              class="inline-flex items-center rounded-lg bg-sky-500 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-sky-600 focus:outline-none focus:ring-2 focus:ring-sky-400 focus:ring-offset-2 focus:ring-offset-slate-950 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
              @click="onCalculate"
            >
              <span v-if="loading">Calculating...</span>
              <span v-else>Recalculate</span>
            </button>
            <p v-if="error" class="text-[11px] text-rose-400">{{ error }}</p>
          </div>
        </div>

        <!-- Estimated Lengths (4 result boxes aligned with columns) -->
        <section class="mt-6 rounded-2xl border border-slate-800 bg-slate-900/80 p-6 shadow-sm">
          <h2 class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400 mb-3">
            Estimated lengths
          </h2>

          <p class="mb-4 text-xs text-slate-400">
            These values are placeholders so that you can validate the layout and UX. Once the API is ready, you can return precise lengths from your backend or a Nuxt server route.
          </p>

          <div class="grid gap-4 md:grid-cols-2">
            <!-- Front Wheel Results -->
            <div class="space-y-3">
              <div class="text-xs font-semibold text-sky-400 uppercase tracking-wide mb-2">Front Wheel</div>
              <div class="grid gap-3 grid-cols-2">
                <div class="rounded-xl border border-slate-800 bg-slate-950/70 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Left side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ frontLeftDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
                <div class="rounded-xl border border-slate-800 bg-slate-950/70 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Right side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ frontRightDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Rear Wheel Results -->
            <div class="space-y-3">
              <div class="text-xs font-semibold text-emerald-400 uppercase tracking-wide mb-2">Rear Wheel</div>
              <div class="grid gap-3 grid-cols-2">
                <div class="rounded-xl border border-slate-800 bg-slate-950/70 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Left side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ rearLeftDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
                <div class="rounded-xl border border-slate-800 bg-slate-950/70 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Right side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ rearRightDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="mt-6 text-[11px] text-slate-400 space-y-3 leading-relaxed">
            <div>
              <strong class="block text-slate-300 mb-1">Disclaimer: Guide to Using Spoke Length Calculation Results</strong>
              <p>
                The spoke length calculator provided on this page generates theoretical recommendations based on standard mathematical models and the data you input. We wish to remind you that the calculation results serve only as a starting point for your spoke procurement and wheel assembly, and are not an absolute standard.
              </p>
            </div>

            <div>
              <strong class="block text-slate-300 mb-1">Reasons for Minor Adjustments:</strong>
              <p class="mb-2">Bicycle wheel components are not perfectly uniform, and minor deviations may cause theoretical values to differ from ideal actual values:</p>
              <ul class="list-disc list-outside ml-4 space-y-1">
                <li>
                  <strong class="text-slate-300">Variation in Effective Rim Diameter (ERD):</strong> The ERD provided by the manufacturer may slightly differ from your actual measurement. We strongly recommend measuring the ERD yourself before proceeding.
                </li>
                <li>
                  <strong class="text-slate-300">Hub Geometry Dimensions:</strong> Slight differences in left/right flange distances and flange diameters.
                </li>
                <li>
                  <strong class="text-slate-300">Actual Operation Tolerances:</strong> When actually lacing the wheel, different tension controls and requirements for thread engagement depth may necessitate adjusting the length up or down by 0.5mm to 2mm.
                </li>
              </ul>
            </div>

            <div>
              <strong class="block text-slate-300 mb-1">Our Recommendation:</strong>
              <p>
                Please make minor adjustments based on your specific situation. Generally, lengths calculated within plus or minus 2mm are considered acceptable. If you are pursuing a perfect fit, please be sure to double-check measurements or consult with a professional. This tool is not responsible for any losses caused by data errors.
              </p>
            </div>
          </div>
        </section>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useSpokeCalculator } from '~/composables/useSpokeCalculator'

interface RimOption {
  id: string
  label: string
}

interface HubOption {
  id: string
  label: string
  position: 'front' | 'rear' | 'front-rear-compatible'
}

interface SpokeProductsResponse {
  rims: RimOption[]
  hubs: HubOption[]
  nipples: { id: string; label: string }[]
}

interface WheelConfig {
  spokeCount: number
  crossing: number
  nippleType: 'standard' | 'hidden'
  rimId: string | null
  hubId: string | null
  erd: number | null
  leftFlange: number | null
  rightFlange: number | null
}

// Front wheel configuration
const frontConfig = reactive<WheelConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  rimId: null,
  hubId: null,
  erd: 622,
  leftFlange: 35,
  rightFlange: 35,
})

// Rear wheel configuration
const rearConfig = reactive<WheelConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  rimId: null,
  hubId: null,
  erd: 622,
  leftFlange: 35,
  rightFlange: 35,
})

const selectedTool = ref<string>('spoke-wrench')

// Load rim / hub / nipple options from backend JSON API.
const {
  data: productsData,
  pending: productsLoading,
  error: productsError,
} = await useFetch<SpokeProductsResponse>('/api/spoke-products')

const rimOptions = computed<RimOption[]>(() => productsData.value?.rims ?? [])
const hubOptions = computed<HubOption[]>(() => productsData.value?.hubs ?? [])

// Filter hubs by wheel position
const frontHubOptions = computed(() =>
  hubOptions.value.filter((h) => h.position === 'front' || h.position === 'front-rear-compatible'),
)
const rearHubOptions = computed(() =>
  hubOptions.value.filter((h) => h.position === 'rear' || h.position === 'front-rear-compatible'),
)

// Sync rim selection when options load
watch(
  rimOptions,
  (list) => {
    const safe = list ?? []
    if (!safe.length) {
      frontConfig.rimId = null
      rearConfig.rimId = null
      return
    }
    if (!safe.some((r) => r.id === frontConfig.rimId)) {
      frontConfig.rimId = safe[0].id
    }
    if (!safe.some((r) => r.id === rearConfig.rimId)) {
      rearConfig.rimId = safe[0].id
    }
  },
  { immediate: true },
)

// Sync hub selection when options load
watch(
  frontHubOptions,
  (list) => {
    const safe = list ?? []
    if (!safe.length) {
      frontConfig.hubId = null
      return
    }
    if (!safe.some((h) => h.id === frontConfig.hubId)) {
      frontConfig.hubId = safe[0].id
    }
  },
  { immediate: true },
)

watch(
  rearHubOptions,
  (list) => {
    const safe = list ?? []
    if (!safe.length) {
      rearConfig.hubId = null
      return
    }
    if (!safe.some((h) => h.id === rearConfig.hubId)) {
      rearConfig.hubId = safe[0].id
    }
  },
  { immediate: true },
)

// Use composable for calculations
const { loading, error, result: frontResult, calculate: calculateFront } = useSpokeCalculator()
const { result: rearResult, calculate: calculateRear } = useSpokeCalculator()

// Display values for front wheel
const frontLeftDisplay = computed(() => (frontResult.value?.leftLengthMm ?? 0).toFixed(1))
const frontRightDisplay = computed(() => (frontResult.value?.rightLengthMm ?? 0).toFixed(1))

// Display values for rear wheel
const rearLeftDisplay = computed(() => (rearResult.value?.leftLengthMm ?? 0).toFixed(1))
const rearRightDisplay = computed(() => (rearResult.value?.rightLengthMm ?? 0).toFixed(1))

const onCalculate = async () => {
  // Calculate both wheels
  await Promise.all([
    frontConfig.rimId && frontConfig.hubId
      ? calculateFront({
          rimId: frontConfig.rimId,
          hubId: frontConfig.hubId,
          wheelPosition: 'front',
          spokeCount: frontConfig.spokeCount,
          crossing: frontConfig.crossing,
        })
      : Promise.resolve(),
    rearConfig.rimId && rearConfig.hubId
      ? calculateRear({
          rimId: rearConfig.rimId,
          hubId: rearConfig.hubId,
          wheelPosition: 'rear',
          spokeCount: rearConfig.spokeCount,
          crossing: rearConfig.crossing,
        })
      : Promise.resolve(),
  ])
}
</script>
