<template>
  <div class="spoke-calculator">
    <div class="grid gap-6 items-start">
      <section class="spoke-calculator__shell">
        <h2 class="text-xs font-semibold uppercase tracking-[0.18em] tz-text-secondary mb-4">
          Wheel setup
        </h2>

        <!-- Two-column layout: Front Wheel | Rear Wheel -->
        <div class="grid gap-6 md:grid-cols-2">
          <!-- ========== FRONT WHEEL COLUMN ========== -->
          <div class="spoke-calculator__panel space-y-4">
            <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">Front Wheel</h3>

            <!-- Spoke count -->
              <div class="space-y-1.5">
                <label for="front-spoke-count" class="block text-xs font-medium tz-text-secondary">Spoke count</label>
                <SpokeCalculatorSelect
                  id="front-spoke-count"
                  v-model="frontConfig.spokeCount"
                  :options="spokeCountOptions"
                />
            </div>

            <!-- Lacing pattern -->
              <div class="space-y-1.5">
                <label for="front-lacing" class="block text-xs font-medium tz-text-secondary">Lacing pattern</label>
                <SpokeCalculatorSelect
                  id="front-lacing"
                  v-model="frontConfig.crossing"
                  :options="lacingOptions"
                />
            </div>

            <div class="space-y-3">
              <!-- Nipple type -->
              <div class="space-y-1.5">
                <label for="front-nipple" class="block text-xs font-medium tz-text-secondary">Nipple type</label>
                <SpokeCalculatorSelect
                  id="front-nipple"
                  v-model="frontConfig.nippleType"
                  :options="nippleTypeOptions"
                />
              </div>

              <!-- Nipple length (hidden nipples only) -->
              <div v-if="frontConfig.nippleType === 'hidden'" class="space-y-1.5">
                <label for="front-nipple-length" class="block text-xs font-medium tz-text-secondary">Nipple length</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-nipple-length"
                    v-model.number="frontConfig.nippleLength"
                    type="number"
                    min="0"
                    max="30"
                    placeholder="e.g. 12"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>
            </div>

            <!-- Rim Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="front-rim-brand" class="block text-xs font-medium tz-text-secondary">
                  Rim Brand
                </label>
                <SpokeCalculatorSelect
                  id="front-rim-brand"
                  v-model="frontConfig.rimBrandId"
                  :options="rimBrandOptions"
                  placeholder="Select Brand"
                />
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="front-rim-model" class="block text-xs font-medium tz-text-secondary">
                  Rim Model
                </label>
                <SpokeCalculatorSelect
                  id="front-rim-model"
                  v-model="frontConfig.rimModelId"
                  :disabled="!frontRimModels.length"
                  :options="frontRimModelOptions"
                  placeholder="Select Model"
                />
               </div>
            </div>

            <!-- Hub Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="front-hub-brand" class="block text-xs font-medium tz-text-secondary">
                  Hub Brand
                </label>
                <SpokeCalculatorSelect
                  id="front-hub-brand"
                  v-model="frontConfig.hubBrandId"
                  :options="hubBrandOptions"
                  placeholder="Select Brand"
                />
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="front-hub-model" class="block text-xs font-medium tz-text-secondary">
                  Hub Model
                </label>
                <SpokeCalculatorSelect
                  id="front-hub-model"
                  v-model="frontConfig.hubModelId"
                  :disabled="!frontHubModels.length"
                  :options="frontHubModelOptions"
                  placeholder="Select Model"
                />
               </div>
            </div>

            <!-- ERD -->
              <div class="space-y-1.5">
                <label for="front-erd" class="block text-xs font-medium tz-text-secondary">ERD (effective rim diameter)</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-erd"
                    v-model.number="frontConfig.erd"
                    type="number"
                    min="400"
                    max="750"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Left flange distance -->
              <div class="space-y-1.5">
                <label for="front-left-flange" class="block text-xs font-medium tz-text-secondary">Left flange distance</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-left-flange"
                    v-model.number="frontConfig.leftFlange"
                    type="number"
                    min="10"
                    max="60"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Right flange distance -->
              <div class="space-y-1.5">
                <label for="front-right-flange" class="block text-xs font-medium tz-text-secondary">Right flange distance</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-right-flange"
                    v-model.number="frontConfig.rightFlange"
                    type="number"
                    min="10"
                    max="60"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Left flange PCD -->
              <div class="space-y-1.5">
                <label for="front-left-flange-pcd" class="block text-xs font-medium tz-text-secondary">Left flange PCD</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-left-flange-pcd"
                    v-model.number="frontConfig.leftFlangePcd"
                    type="number"
                    min="30"
                    max="80"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Right flange PCD -->
              <div class="space-y-1.5">
                <label for="front-right-flange-pcd" class="block text-xs font-medium tz-text-secondary">Right flange PCD</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="front-right-flange-pcd"
                    v-model.number="frontConfig.rightFlangePcd"
                    type="number"
                    min="30"
                    max="80"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>
            </div>

          <!-- ========== REAR WHEEL COLUMN ========== -->
          <div class="spoke-calculator__panel space-y-4">
            <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">Rear Wheel</h3>

            <!-- Spoke count -->
              <div class="space-y-1.5">
                <label for="rear-spoke-count" class="block text-xs font-medium tz-text-secondary">Spoke count</label>
                <SpokeCalculatorSelect
                  id="rear-spoke-count"
                  v-model="rearConfig.spokeCount"
                  :options="spokeCountOptions"
                />
            </div>

            <!-- Lacing pattern -->
              <div class="space-y-1.5">
                <label for="rear-lacing" class="block text-xs font-medium tz-text-secondary">Lacing pattern</label>
                <SpokeCalculatorSelect
                  id="rear-lacing"
                  v-model="rearConfig.crossing"
                  :options="lacingOptions"
                />
            </div>

            <div class="space-y-3">
              <!-- Nipple type -->
              <div class="space-y-1.5">
                <label for="rear-nipple" class="block text-xs font-medium tz-text-secondary">Nipple type</label>
                <SpokeCalculatorSelect
                  id="rear-nipple"
                  v-model="rearConfig.nippleType"
                  :options="nippleTypeOptions"
                />
              </div>

              <!-- Nipple length (hidden nipples only) -->
              <div v-if="rearConfig.nippleType === 'hidden'" class="space-y-1.5">
                <label for="rear-nipple-length" class="block text-xs font-medium tz-text-secondary">Nipple length</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-nipple-length"
                    v-model.number="rearConfig.nippleLength"
                    type="number"
                    min="0"
                    max="30"
                    placeholder="e.g. 12"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>
            </div>

            <!-- Rim Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="rear-rim-brand" class="block text-xs font-medium tz-text-secondary">
                  Rim Brand
                </label>
                <SpokeCalculatorSelect
                  id="rear-rim-brand"
                  v-model="rearConfig.rimBrandId"
                  :options="rimBrandOptions"
                  placeholder="Select Brand"
                />
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="rear-rim-model" class="block text-xs font-medium tz-text-secondary">
                  Rim Model
                </label>
                <SpokeCalculatorSelect
                  id="rear-rim-model"
                  v-model="rearConfig.rimModelId"
                  :disabled="!rearRimModels.length"
                  :options="rearRimModelOptions"
                  placeholder="Select Model"
                />
               </div>
            </div>

            <!-- Hub Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="rear-hub-brand" class="block text-xs font-medium tz-text-secondary">
                  Hub Brand
                </label>
                <SpokeCalculatorSelect
                  id="rear-hub-brand"
                  v-model="rearConfig.hubBrandId"
                  :options="hubBrandOptions"
                  placeholder="Select Brand"
                />
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="rear-hub-model" class="block text-xs font-medium tz-text-secondary">
                  Hub Model
                </label>
                <SpokeCalculatorSelect
                  id="rear-hub-model"
                  v-model="rearConfig.hubModelId"
                  :disabled="!rearHubModels.length"
                  :options="rearHubModelOptions"
                  placeholder="Select Model"
                />
               </div>
            </div>

            <!-- ERD -->
              <div class="space-y-1.5">
                <label for="rear-erd" class="block text-xs font-medium tz-text-secondary">ERD (effective rim diameter)</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-erd"
                    v-model.number="rearConfig.erd"
                    type="number"
                    min="400"
                    max="750"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Left flange distance -->
              <div class="space-y-1.5">
                <label for="rear-left-flange" class="block text-xs font-medium tz-text-secondary">Left flange distance</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-left-flange"
                    v-model.number="rearConfig.leftFlange"
                    type="number"
                    min="10"
                    max="60"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Right flange distance -->
              <div class="space-y-1.5">
                <label for="rear-right-flange" class="block text-xs font-medium tz-text-secondary">Right flange distance</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-right-flange"
                    v-model.number="rearConfig.rightFlange"
                    type="number"
                    min="10"
                    max="60"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Left flange PCD -->
              <div class="space-y-1.5">
                <label for="rear-left-flange-pcd" class="block text-xs font-medium tz-text-secondary">Left flange PCD</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-left-flange-pcd"
                    v-model.number="rearConfig.leftFlangePcd"
                    type="number"
                    min="30"
                    max="80"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>

            <!-- Right flange PCD -->
              <div class="space-y-1.5">
                <label for="rear-right-flange-pcd" class="block text-xs font-medium tz-text-secondary">Right flange PCD</label>
                <div class="spoke-calculator__unit-field">
                  <input
                    id="rear-right-flange-pcd"
                    v-model.number="rearConfig.rightFlangePcd"
                    type="number"
                    min="30"
                    max="80"
                    class="spoke-calculator__control spoke-calculator__control--with-unit"
                  />
                  <span class="spoke-calculator__unit">mm</span>
                </div>
              </div>
            </div>
        </div>

        <!-- Action row -->
        <div class="mt-6 flex flex-col gap-3 md:flex-row md:items-center md:justify-between border-t tz-border-subtle pt-4">
          <p class="tz-description tz-text-muted max-w-md">
            This is only a visual prototype. Replace the mock formula in the script section with your own calculation logic.
          </p>
          <div class="flex items-center gap-3">
            <button
              type="button"
              class="inline-flex items-center rounded-lg bg-[var(--tz-action-primary)] px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-[var(--tz-action-primary-hover)] focus:outline-none focus:ring-2 focus:ring-[color:var(--tz-site-accent)] focus:ring-offset-2 focus:ring-offset-[var(--tz-card-surface)] disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
              @click="onCalculate"
            >
              <span v-if="loading">Calculating...</span>
              <span v-else>Recalculate</span>
            </button>
            <p v-if="error" class="tz-caption text-rose-400">{{ error }}</p>
          </div>
        </div>

        <!-- Spoke lengths (4 result boxes aligned with columns) -->
        <section class="spoke-calculator__results-shell mt-6">
          <h2 class="text-xs font-semibold uppercase tracking-[0.18em] tz-text-secondary mb-3">
            Spoke lengths
          </h2>

          <p class="tz-description mb-4 tz-text-secondary">
            Verified catalog lengths are used first. If no verified result matches the current selection, the calculator uses the dimensions above.
          </p>

          <div class="grid gap-4 md:grid-cols-2">
            <!-- Front Wheel Results -->
            <div class="space-y-3">
              <div class="text-xs font-semibold tz-text-accent uppercase tracking-wide mb-2">Front Wheel</div>
              <div class="grid gap-3 grid-cols-2">
                <div class="spoke-calculator__result-card px-4 py-3">
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <span class="tz-compact-label tz-text-muted">Left side</span>
                    <span v-if="frontLeftSourceLabel" class="spoke-calculator__source-badge">{{ frontLeftSourceLabel }}</span>
                  </div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ frontLeftDisplay }}</span>
                    <span v-if="frontLeftDisplay !== '--'" class="text-xs tz-text-muted">mm</span>
                  </div>
                </div>
                <div class="spoke-calculator__result-card px-4 py-3">
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <span class="tz-compact-label tz-text-muted">Right side</span>
                    <span v-if="frontRightSourceLabel" class="spoke-calculator__source-badge">{{ frontRightSourceLabel }}</span>
                  </div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ frontRightDisplay }}</span>
                    <span v-if="frontRightDisplay !== '--'" class="text-xs tz-text-muted">mm</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Rear Wheel Results -->
            <div class="space-y-3">
              <div class="text-xs font-semibold tz-text-accent uppercase tracking-wide mb-2">Rear Wheel</div>
              <div class="grid gap-3 grid-cols-2">
                <div class="spoke-calculator__result-card px-4 py-3">
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <span class="tz-compact-label tz-text-muted">Left side</span>
                    <span v-if="rearLeftSourceLabel" class="spoke-calculator__source-badge">{{ rearLeftSourceLabel }}</span>
                  </div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ rearLeftDisplay }}</span>
                    <span v-if="rearLeftDisplay !== '--'" class="text-xs tz-text-muted">mm</span>
                  </div>
                </div>
                <div class="spoke-calculator__result-card px-4 py-3">
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <span class="tz-compact-label tz-text-muted">Right side</span>
                    <span v-if="rearRightSourceLabel" class="spoke-calculator__source-badge">{{ rearRightSourceLabel }}</span>
                  </div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ rearRightDisplay }}</span>
                    <span v-if="rearRightDisplay !== '--'" class="text-xs tz-text-muted">mm</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="tz-description mt-6 tz-text-secondary space-y-3">
            <div>
              <strong class="block tz-text-primary mb-1">Disclaimer: Guide to Using Spoke Length Calculation Results</strong>
              <p>
                The spoke length calculator provided on this page generates theoretical recommendations based on standard mathematical models and the data you input. We wish to remind you that the calculation results serve only as a starting point for your spoke procurement and wheel assembly, and are not an absolute standard.
              </p>
            </div>

            <div>
              <strong class="block tz-text-primary mb-1">Reasons for Minor Adjustments:</strong>
              <p class="mb-2">Bicycle wheel components are not perfectly uniform, and minor deviations may cause theoretical values to differ from ideal actual values:</p>
              <ul class="list-disc list-outside ml-4 space-y-1">
                <li>
                  <strong class="tz-text-primary">Variation in Effective Rim Diameter (ERD):</strong> The ERD provided by the manufacturer may slightly differ from your actual measurement. We strongly recommend measuring the ERD yourself before proceeding.
                </li>
                <li>
                  <strong class="tz-text-primary">Hub Geometry Dimensions:</strong> Slight differences in left/right flange distances and flange diameters.
                </li>
                <li>
                  <strong class="tz-text-primary">Actual Operation Tolerances:</strong> When actually lacing the wheel, different tension controls and requirements for thread engagement depth may necessitate adjusting the length up or down by 0.5mm to 2mm.
                </li>
              </ul>
            </div>

            <div>
              <strong class="block tz-text-primary mb-1">Our Recommendation:</strong>
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
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import type { HubGeometry, HubModel, RimModel, WheelBuildPreset } from '~/data/spoke-calculator/database'
import { computeSpokeLength } from '~/utils/spokeMath'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'

interface WheelConfig {
  spokeCount: number
  crossing: number
  nippleType: 'standard' | 'hidden'
  nippleLength: number | null
  
  // Selection State
  rimBrandId: string | null
  rimModelId: string | null
  hubBrandId: string | null
  hubModelId: string | null

  // Geometry Data
  erd: number | null
  leftFlange: number | null
  rightFlange: number | null
  leftFlangePcd: number | null
  rightFlangePcd: number | null
}

// Front wheel configuration
const frontConfig = reactive<WheelConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  nippleLength: 12,
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
  erd: 622,
  leftFlange: 35,
  rightFlange: 35,
  leftFlangePcd: 50,
  rightFlangePcd: 50,
})

// Rear wheel configuration
const rearConfig = reactive<WheelConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  nippleLength: 12,
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
  erd: 622,
  leftFlange: 35,
  rightFlange: 20,
  leftFlangePcd: 55,
  rightFlangePcd: 55,
})

const { rims, hubs, presets, options: catalogOptions } = useSpokeCalculatorCatalog()

const spokeCountOptions = computed(() => catalogOptions.value.spokeCounts)
const lacingOptions = computed(() => catalogOptions.value.crossings)
const nippleTypeOptions = computed(() => catalogOptions.value.nippleTypes)

const rimBrandOptions = computed(() => rims.value.map(brand => ({
  label: brand.name,
  value: brand.id,
})))

const hubBrandOptions = computed(() => hubs.value.map(brand => ({
  label: brand.name,
  value: brand.id,
})))

// --- Computed Models based on Brand Selection ---

// Front Rim Models
const frontRimModels = computed<RimModel[]>(() => {
  if (!frontConfig.rimBrandId) return []
  const brand = rims.value.find(b => b.id === frontConfig.rimBrandId)
  return brand ? brand.items : []
})

const frontRimModelOptions = computed(() => frontRimModels.value.map(rim => ({
  label: rim.name,
  value: rim.id,
})))

// Front Hub Models
const frontHubModels = computed<HubModel[]>(() => {
  if (!frontConfig.hubBrandId) return []
  const brand = hubs.value.find(b => b.id === frontConfig.hubBrandId)
  return brand ? brand.items : []
})

const frontHubModelOptions = computed(() => frontHubModels.value.map(hub => ({
  label: hub.name,
  value: hub.id,
})))

// Rear Rim Models
const rearRimModels = computed<RimModel[]>(() => {
  if (!rearConfig.rimBrandId) return []
  const brand = rims.value.find(b => b.id === rearConfig.rimBrandId)
  return brand ? brand.items : []
})

const rearRimModelOptions = computed(() => rearRimModels.value.map(rim => ({
  label: rim.name,
  value: rim.id,
})))

// Rear Hub Models
const rearHubModels = computed<HubModel[]>(() => {
  if (!rearConfig.hubBrandId) return []
  const brand = hubs.value.find(b => b.id === rearConfig.hubBrandId)
  return brand ? brand.items : []
})

const rearHubModelOptions = computed(() => rearHubModels.value.map(hub => ({
  label: hub.name,
  value: hub.id,
})))

const applyHubGeometry = (config: WheelConfig, geometry?: HubGeometry | null) => {
  config.leftFlange = geometry?.leftFlange ?? null
  config.rightFlange = geometry?.rightFlange ?? null
  config.leftFlangePcd = geometry?.leftFlangePcd ?? null
  config.rightFlangePcd = geometry?.rightFlangePcd ?? null
}

// --- Watchers for Auto-Population ---

// Front Rim Change
watch(
  () => frontConfig.rimModelId,
  (newId) => {
    if (!newId) {
      frontConfig.erd = null
      return
    }
    const model = frontRimModels.value.find(m => m.id === newId)
    if (model && model.erd != null) {
      frontConfig.erd = model.erd
    } else {
      frontConfig.erd = null
    }
  }
)

// Front Hub Change
watch(
  () => frontConfig.hubModelId,
  (newId) => {
    if (!newId) {
      applyHubGeometry(frontConfig, null)
      return
    }
    const model = frontHubModels.value.find(m => m.id === newId)
    applyHubGeometry(frontConfig, model?.front ?? null)
  }
)

// Rear Rim Change
watch(
  () => rearConfig.rimModelId,
  (newId) => {
    if (!newId) {
      rearConfig.erd = null
      return
    }
    const model = rearRimModels.value.find(m => m.id === newId)
    if (model && model.erd != null) {
      rearConfig.erd = model.erd
    } else {
      rearConfig.erd = null
    }
  }
)

// Rear Hub Change
watch(
  () => rearConfig.hubModelId,
  (newId) => {
    if (!newId) {
      applyHubGeometry(rearConfig, null)
      return
    }
    const model = rearHubModels.value.find(m => m.id === newId)
    applyHubGeometry(rearConfig, model?.rear ?? null)
  }
)

const loading = ref(false)
const error = ref<string | null>(null)

type WheelPosition = 'front' | 'rear'
type ResultSource = 'verified' | 'calculated'

interface SpokeResult {
  leftLengthMm: number | null
  rightLengthMm: number | null
  leftSource: ResultSource | null
  rightSource: ResultSource | null
}

interface CalculatedWheelResult {
  leftLengthMm: number | null
  rightLengthMm: number | null
}

const frontResult = ref<SpokeResult | null>(null)
const rearResult = ref<SpokeResult | null>(null)
const lastTrackedCalculation = ref('')
const { track: trackBehaviorEvent } = useBehaviorEvents()

const formatResultLength = (value: number | null | undefined) => (
  value == null ? '--' : value.toFixed(1)
)

const resultSourceLabel = (source: ResultSource | null | undefined) => {
  if (source === 'verified') return 'Verified'
  if (source === 'calculated') return 'Calculated'
  return ''
}

const frontLeftDisplay = computed(() => formatResultLength(frontResult.value?.leftLengthMm))
const frontRightDisplay = computed(() => formatResultLength(frontResult.value?.rightLengthMm))
const rearLeftDisplay = computed(() => formatResultLength(rearResult.value?.leftLengthMm))
const rearRightDisplay = computed(() => formatResultLength(rearResult.value?.rightLengthMm))

const frontLeftSourceLabel = computed(() => resultSourceLabel(frontResult.value?.leftSource))
const frontRightSourceLabel = computed(() => resultSourceLabel(frontResult.value?.rightSource))
const rearLeftSourceLabel = computed(() => resultSourceLabel(rearResult.value?.leftSource))
const rearRightSourceLabel = computed(() => resultSourceLabel(rearResult.value?.rightSource))

const onCalculate = () => {
  error.value = null
  loading.value = true

  try {
    const completedWheelCount = updateResults()

    if (completedWheelCount > 0) {
      const fingerprint = JSON.stringify({
        front: frontConfig,
        rear: rearConfig,
      })

      if (fingerprint !== lastTrackedCalculation.value) {
        lastTrackedCalculation.value = fingerprint
        trackBehaviorEvent({
          eventType: 'calculator_use',
          metadata: {
            source: 'spoke_calculator',
            wheel_count: completedWheelCount,
            front_spoke_count: frontConfig.spokeCount,
            rear_spoke_count: rearConfig.spokeCount,
            front_crossing: frontConfig.crossing,
            rear_crossing: rearConfig.crossing,
            front_rim_selected: Boolean(frontConfig.rimModelId),
            rear_rim_selected: Boolean(rearConfig.rimModelId),
            front_hub_selected: Boolean(frontConfig.hubModelId),
            rear_hub_selected: Boolean(rearConfig.hubModelId),
          },
        })
      }
    }
  } catch (e: any) {
    error.value = e?.message || 'Calculation failed'
  } finally {
    loading.value = false
  }
}

const updateResults = () => {
  frontResult.value = buildWheelResult(frontConfig, 'front')
  rearResult.value = buildWheelResult(rearConfig, 'rear')

  return [frontResult.value, rearResult.value].filter(result => (
    result && (result.leftLengthMm != null || result.rightLengthMm != null)
  )).length
}

const buildWheelResult = (config: WheelConfig, wheel: WheelPosition): SpokeResult | null => {
  const verified = findVerifiedWheelLengths(config, wheel)
  const calculated = calculateWheel(config)

  const leftLengthMm = verified?.leftLengthMm ?? calculated?.leftLengthMm ?? null
  const rightLengthMm = verified?.rightLengthMm ?? calculated?.rightLengthMm ?? null

  if (leftLengthMm == null && rightLengthMm == null) return null

  return {
    leftLengthMm,
    rightLengthMm,
    leftSource: verified?.leftLengthMm != null ? 'verified' : calculated?.leftLengthMm != null ? 'calculated' : null,
    rightSource: verified?.rightLengthMm != null ? 'verified' : calculated?.rightLengthMm != null ? 'calculated' : null,
  }
}

const findVerifiedWheelLengths = (config: WheelConfig, wheel: WheelPosition): CalculatedWheelResult | null => {
  const preset = presets.value.find(item => presetMatchesWheelConfig(item, config, wheel))
  const actual = preset?.actualLengths
  if (!actual) return null

  const leftLengthMm = wheel === 'front' ? actual.frontLeft : actual.rearLeft
  const rightLengthMm = wheel === 'front' ? actual.frontRight : actual.rearRight
  if (leftLengthMm == null && rightLengthMm == null) return null

  return {
    leftLengthMm: leftLengthMm ?? null,
    rightLengthMm: rightLengthMm ?? null,
  }
}

const presetMatchesWheelConfig = (preset: WheelBuildPreset, config: WheelConfig, wheel: WheelPosition) => {
  if (!config.rimBrandId || !config.rimModelId || !config.hubBrandId || !config.hubModelId) return false
  if (preset.wheelPosition && preset.wheelPosition !== 'auto' && preset.wheelPosition !== wheel) return false
  if (!preset.actualLengths) return false
  if (wheel === 'front' && preset.actualLengths.frontLeft == null && preset.actualLengths.frontRight == null) return false
  if (wheel === 'rear' && preset.actualLengths.rearLeft == null && preset.actualLengths.rearRight == null) return false

  return preset.rimBrandId === config.rimBrandId &&
    preset.rimModelId === config.rimModelId &&
    preset.hubBrandId === config.hubBrandId &&
    preset.hubModelId === config.hubModelId &&
    preset.spokeCount === config.spokeCount &&
    preset.crossing === config.crossing &&
    preset.nippleType === config.nippleType &&
    compatibleNippleLength(preset, config)
}

const compatibleNippleLength = (preset: WheelBuildPreset, config: WheelConfig) => {
  if (config.nippleType !== 'hidden') return true
  if (preset.nippleLength == null || config.nippleLength == null) return true
  return Math.abs(preset.nippleLength - config.nippleLength) < 0.01
}

const calculateWheel = (config: WheelConfig): CalculatedWheelResult | null => {
  if (!config.erd || !config.leftFlangePcd || !config.rightFlangePcd ||
      config.leftFlange == null || config.rightFlange == null) {
    return null
  }

  return {
    leftLengthMm: computeSpokeLength(
      config.erd,
      config.leftFlangePcd,
      config.leftFlange,
      config.spokeCount,
      config.crossing,
      config.nippleType,
      config.nippleLength
    ),
    rightLengthMm: computeSpokeLength(
      config.erd,
      config.rightFlangePcd,
      config.rightFlange,
      config.spokeCount,
      config.crossing,
      config.nippleType,
      config.nippleLength
    ),
  }
}

watch(
  () => ({
    front: { ...frontConfig },
    rear: { ...rearConfig },
    presets: presets.value,
  }),
  () => {
    try {
      updateResults()
    } catch (e: any) {
      error.value = e?.message || 'Calculation failed'
    }
  },
  { deep: true, immediate: true }
)
</script>

<style scoped>
.spoke-calculator {
  --spoke-shell-surface: var(--tz-card-surface);
  --spoke-panel-surface: var(--tz-form-panel-surface);
  --spoke-control-surface: var(--tz-input-surface);
  --spoke-result-surface: var(--tz-surface-subtle);
  --spoke-border: var(--tz-border-subtle);
  --spoke-border-strong: var(--tz-border-strong);
  --spoke-focus-ring: var(--tz-form-control-focus-ring);
  color: var(--tz-text-primary);
}

.spoke-calculator__shell,
.spoke-calculator__results-shell {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-shell-surface);
  box-shadow: 0 10px 26px -14px rgba(20, 32, 43, 0.12);
}

.spoke-calculator__shell {
  padding: 1.25rem;
}

.spoke-calculator__panel {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
  padding: 1rem;
}

.spoke-calculator__control {
  display: block;
  width: 100%;
  min-width: 0;
  border: 1px solid var(--spoke-border) !important;
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface) !important;
  background-image: none !important;
  color: var(--tz-text-primary) !important;
  padding: 0.75rem 0.875rem;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__control:focus,
.spoke-calculator__control:focus-visible {
  outline: none;
  border-color: var(--spoke-border-strong) !important;
  box-shadow: 0 0 0 1px var(--spoke-focus-ring) !important;
}

.spoke-calculator__control:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.spoke-calculator__control--with-unit {
  flex: 1 1 auto;
  border: 0 !important;
  border-radius: 0;
  background: transparent !important;
  box-shadow: none !important;
  padding-right: 0.75rem;
}

.spoke-calculator__unit-field {
  display: flex;
  width: 100%;
  align-items: stretch;
  gap: 0;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__unit-field:focus-within {
  border-color: var(--spoke-border-strong);
  box-shadow: 0 0 0 1px var(--spoke-focus-ring);
}

.spoke-calculator__unit-field .spoke-calculator__control:focus,
.spoke-calculator__unit-field .spoke-calculator__control:focus-visible {
  box-shadow: none !important;
}

.spoke-calculator__unit {
  display: inline-flex;
  flex: 0 0 auto;
  min-width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-left: 1px solid var(--spoke-border);
  padding: 0 0.75rem;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-caption);
  line-height: 1;
  white-space: nowrap;
}

.spoke-calculator__results-shell {
  padding: 1.25rem;
}

.spoke-calculator__result-card {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
}

.spoke-calculator__source-badge {
  flex: 0 0 auto;
  border: 1px solid var(--spoke-border);
  border-radius: 999px;
  padding: 0.125rem 0.375rem;
  color: var(--tz-text-muted);
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.spoke-calculator label {
  color: var(--tz-text-secondary) !important;
}
</style>
