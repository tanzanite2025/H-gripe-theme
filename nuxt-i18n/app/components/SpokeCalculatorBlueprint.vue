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

            <div class="spoke-calculator__blueprint-sheet">
              <div class="spoke-calculator__blueprint-sheet-header">
                <div>
                  <p class="spoke-calculator__blueprint-eyebrow">Numbered schematic</p>
                  <h4 class="spoke-calculator__blueprint-title">Front wheel geometry</h4>
                </div>
                <p class="spoke-calculator__blueprint-note">
                  The numbers in the drawing match the table below, so the wheel view stays readable on mobile.
                </p>
              </div>

              <div class="spoke-calculator__blueprint-grid">
                <div class="spoke-calculator__schematic-grid">
                  <SpokeWheelSchematic wheel="front" side="disc" />
                  <SpokeWheelSchematic wheel="front" side="non-disc" />
                </div>

                <div class="spoke-calculator__legend-table">
                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">1</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>ERD</strong>
                      <span>Effective rim diameter</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="front-blueprint-erd"
                          v-model.number="frontConfig.erd"
                          type="number"
                          min="400"
                          max="750"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">mm</span>
                      </div>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row spoke-calculator__legend-row--split">
                    <div class="spoke-calculator__legend-badge">2</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Left flange</strong>
                      <span>PCDL / center distance</span>
                    </div>
                    <div class="spoke-calculator__legend-fields">
                      <label for="front-blueprint-left-flange-pcd" class="spoke-calculator__legend-field">
                        <span>PCDL</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="front-blueprint-left-flange-pcd"
                            v-model.number="frontConfig.leftFlangePcd"
                            type="number"
                            min="30"
                            max="80"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                      <label for="front-blueprint-left-flange" class="spoke-calculator__legend-field">
                        <span>Center</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="front-blueprint-left-flange"
                            v-model.number="frontConfig.leftFlange"
                            type="number"
                            min="10"
                            max="60"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row spoke-calculator__legend-row--split">
                    <div class="spoke-calculator__legend-badge">3</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Right flange</strong>
                      <span>PCDR / center distance</span>
                    </div>
                    <div class="spoke-calculator__legend-fields">
                      <label for="front-blueprint-right-flange-pcd" class="spoke-calculator__legend-field">
                        <span>PCDR</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="front-blueprint-right-flange-pcd"
                            v-model.number="frontConfig.rightFlangePcd"
                            type="number"
                            min="30"
                            max="80"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                      <label for="front-blueprint-right-flange" class="spoke-calculator__legend-field">
                        <span>Center</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="front-blueprint-right-flange"
                            v-model.number="frontConfig.rightFlange"
                            type="number"
                            min="10"
                            max="60"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">4</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Rim offset</strong>
                      <span>Dish adjustment</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="front-blueprint-rim-offset"
                          v-model.number="frontConfig.rimOffsetMm"
                          type="number"
                          min="-20"
                          max="20"
                          step="0.1"
                          placeholder="0"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">mm</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="spoke-calculator__build-settings">
              <div class="spoke-calculator__build-settings-header">
                <p class="spoke-calculator__build-settings-title">Build settings</p>
              </div>

              <div class="spoke-calculator__build-settings-grid">
                <div class="spoke-calculator__setting-field">
                  <label for="front-spoke-count" class="block text-xs font-medium tz-text-secondary">Spoke count</label>
                  <SpokeCalculatorSelect
                    id="front-spoke-count"
                    v-model="frontConfig.spokeCount"
                    :options="spokeCountOptions"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="front-lacing" class="block text-xs font-medium tz-text-secondary">Lacing pattern</label>
                  <SpokeCalculatorSelect
                    id="front-lacing"
                    v-model="frontConfig.crossing"
                    :options="lacingOptions"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="front-nipple" class="block text-xs font-medium tz-text-secondary">Nipple type</label>
                  <SpokeCalculatorSelect
                    id="front-nipple"
                    v-model="frontConfig.nippleType"
                    :options="nippleTypeOptions"
                  />
                </div>

                <div v-if="frontConfig.nippleType === 'hidden'" class="spoke-calculator__setting-field">
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

                <div class="spoke-calculator__setting-field">
                  <label for="front-rim-brand" class="block text-xs font-medium tz-text-secondary">Rim Brand</label>
                  <SpokeCalculatorSelect
                    id="front-rim-brand"
                    v-model="frontConfig.rimBrandId"
                    :options="rimBrandOptions"
                    placeholder="Select Brand"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="front-rim-model" class="block text-xs font-medium tz-text-secondary">Rim Model</label>
                  <SpokeCalculatorSelect
                    id="front-rim-model"
                    v-model="frontConfig.rimModelId"
                    :disabled="!frontRimModels.length"
                    :options="frontRimModelOptions"
                    placeholder="Select Model"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="front-hub-brand" class="block text-xs font-medium tz-text-secondary">Hub Brand</label>
                  <SpokeCalculatorSelect
                    id="front-hub-brand"
                    v-model="frontConfig.hubBrandId"
                    :options="hubBrandOptions"
                    placeholder="Select Brand"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="front-hub-model" class="block text-xs font-medium tz-text-secondary">Hub Model</label>
                  <SpokeCalculatorSelect
                    id="front-hub-model"
                    v-model="frontConfig.hubModelId"
                    :disabled="!frontHubModels.length"
                    :options="frontHubModelOptions"
                    placeholder="Select Model"
                  />
                </div>
              </div>
            </div>

          </div>

          <!-- ========== REAR WHEEL COLUMN ========== -->
          <div class="spoke-calculator__panel space-y-4">
            <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">Rear Wheel</h3>

            <div class="spoke-calculator__blueprint-sheet">
              <div class="spoke-calculator__blueprint-sheet-header">
                <div>
                  <p class="spoke-calculator__blueprint-eyebrow">Numbered schematic</p>
                  <h4 class="spoke-calculator__blueprint-title">Rear wheel geometry</h4>
                </div>
                <p class="spoke-calculator__blueprint-note">
                  The rear view uses the same numbered layout, so disc and non-disc sides stay easy to compare.
                </p>
              </div>

              <div class="spoke-calculator__blueprint-grid">
                <div class="spoke-calculator__schematic-grid">
                  <SpokeWheelSchematic wheel="rear" side="disc" />
                  <SpokeWheelSchematic wheel="rear" side="non-disc" />
                </div>

                <div class="spoke-calculator__legend-table">
                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">1</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>ERD</strong>
                      <span>Effective rim diameter</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="rear-blueprint-erd"
                          v-model.number="rearConfig.erd"
                          type="number"
                          min="400"
                          max="750"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">mm</span>
                      </div>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row spoke-calculator__legend-row--split">
                    <div class="spoke-calculator__legend-badge">2</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Left flange</strong>
                      <span>PCDL / center distance</span>
                    </div>
                    <div class="spoke-calculator__legend-fields">
                      <label for="rear-blueprint-left-flange-pcd" class="spoke-calculator__legend-field">
                        <span>PCDL</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="rear-blueprint-left-flange-pcd"
                            v-model.number="rearConfig.leftFlangePcd"
                            type="number"
                            min="30"
                            max="80"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                      <label for="rear-blueprint-left-flange" class="spoke-calculator__legend-field">
                        <span>Center</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="rear-blueprint-left-flange"
                            v-model.number="rearConfig.leftFlange"
                            type="number"
                            min="10"
                            max="60"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row spoke-calculator__legend-row--split">
                    <div class="spoke-calculator__legend-badge">3</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Right flange</strong>
                      <span>PCDR / center distance</span>
                    </div>
                    <div class="spoke-calculator__legend-fields">
                      <label for="rear-blueprint-right-flange-pcd" class="spoke-calculator__legend-field">
                        <span>PCDR</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="rear-blueprint-right-flange-pcd"
                            v-model.number="rearConfig.rightFlangePcd"
                            type="number"
                            min="30"
                            max="80"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                      <label for="rear-blueprint-right-flange" class="spoke-calculator__legend-field">
                        <span>Center</span>
                        <div class="spoke-calculator__unit-field">
                          <input
                            id="rear-blueprint-right-flange"
                            v-model.number="rearConfig.rightFlange"
                            type="number"
                            min="10"
                            max="60"
                            class="spoke-calculator__control spoke-calculator__control--with-unit"
                          />
                          <span class="spoke-calculator__unit">mm</span>
                        </div>
                      </label>
                    </div>
                  </div>

                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">4</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>Rim offset</strong>
                      <span>Dish adjustment</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="rear-blueprint-rim-offset"
                          v-model.number="rearConfig.rimOffsetMm"
                          type="number"
                          min="-20"
                          max="20"
                          step="0.1"
                          placeholder="0"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">mm</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="spoke-calculator__build-settings">
              <div class="spoke-calculator__build-settings-header">
                <p class="spoke-calculator__build-settings-title">Build settings</p>
              </div>

              <div class="spoke-calculator__build-settings-grid">
                <div class="spoke-calculator__setting-field">
                  <label for="rear-spoke-count" class="block text-xs font-medium tz-text-secondary">Spoke count</label>
                  <SpokeCalculatorSelect
                    id="rear-spoke-count"
                    v-model="rearConfig.spokeCount"
                    :options="spokeCountOptions"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="rear-lacing" class="block text-xs font-medium tz-text-secondary">Lacing pattern</label>
                  <SpokeCalculatorSelect
                    id="rear-lacing"
                    v-model="rearConfig.crossing"
                    :options="lacingOptions"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="rear-nipple" class="block text-xs font-medium tz-text-secondary">Nipple type</label>
                  <SpokeCalculatorSelect
                    id="rear-nipple"
                    v-model="rearConfig.nippleType"
                    :options="nippleTypeOptions"
                  />
                </div>

                <div v-if="rearConfig.nippleType === 'hidden'" class="spoke-calculator__setting-field">
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

                <div class="spoke-calculator__setting-field">
                  <label for="rear-rim-brand" class="block text-xs font-medium tz-text-secondary">Rim Brand</label>
                  <SpokeCalculatorSelect
                    id="rear-rim-brand"
                    v-model="rearConfig.rimBrandId"
                    :options="rimBrandOptions"
                    placeholder="Select Brand"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="rear-rim-model" class="block text-xs font-medium tz-text-secondary">Rim Model</label>
                  <SpokeCalculatorSelect
                    id="rear-rim-model"
                    v-model="rearConfig.rimModelId"
                    :disabled="!rearRimModels.length"
                    :options="rearRimModelOptions"
                    placeholder="Select Model"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="rear-hub-brand" class="block text-xs font-medium tz-text-secondary">Hub Brand</label>
                  <SpokeCalculatorSelect
                    id="rear-hub-brand"
                    v-model="rearConfig.hubBrandId"
                    :options="hubBrandOptions"
                    placeholder="Select Brand"
                  />
                </div>

                <div class="spoke-calculator__setting-field">
                  <label for="rear-hub-model" class="block text-xs font-medium tz-text-secondary">Hub Model</label>
                  <SpokeCalculatorSelect
                    id="rear-hub-model"
                    v-model="rearConfig.hubModelId"
                    :disabled="!rearHubModels.length"
                    :options="rearHubModelOptions"
                    placeholder="Select Model"
                  />
                </div>
              </div>
            </div>
        </div>

        </div>

        <!-- Action row -->
        <div class="mt-6 flex flex-col gap-3 md:flex-row md:items-center md:justify-between border-t tz-border-subtle pt-4">
          <p class="tz-description tz-text-muted max-w-md">
            Spoke lengths and the predicted per-spoke tension ratio are derived from the selected wheel geometry. Use measured rim and hub dimensions for a production build.
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
              <div class="spoke-calculator__tension-summary">
                <div class="flex items-baseline justify-between gap-3">
                  <span class="tz-compact-label tz-text-muted">Predicted tension ratio</span>
                  <span class="text-lg font-semibold text-[var(--tz-site-accent)]">{{ formatTensionRatio(frontResult?.tensionRatio) }}</span>
                </div>
                <div class="mt-1 tz-caption tz-text-muted">
                  Left : Right {{ formatDirectionalTensionRatio(frontResult?.tensionRatio) }} · lower side {{ lowerTensionSideLabel(frontResult?.tensionRatio) }}
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
              <div class="spoke-calculator__tension-summary">
                <div class="flex items-baseline justify-between gap-3">
                  <span class="tz-compact-label tz-text-muted">Predicted tension ratio</span>
                  <span class="text-lg font-semibold text-[var(--tz-site-accent)]">{{ formatTensionRatio(rearResult?.tensionRatio) }}</span>
                </div>
                <div class="mt-1 tz-caption tz-text-muted">
                  Left : Right {{ formatDirectionalTensionRatio(rearResult?.tensionRatio) }} · lower side {{ lowerTensionSideLabel(rearResult?.tensionRatio) }}
                </div>
              </div>
            </div>
          </div>

          <div class="spoke-calculator__results-note mt-6">
            Results use verified catalog lengths first. If no verified build matches, the calculator falls back to the dimensions entered above.
          </div>
        </section>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeWheelSchematic from '~/components/SpokeWheelSchematic.vue'
import { computed, reactive, ref, watch } from 'vue'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import type { HubGeometry, HubModel, RimModel, WheelBuildPreset } from '~/data/spoke-calculator/database'
import {
  computeSpokeLength,
  computeSpokeTensionRatio,
  effectiveSpokeFlangeDistance,
  type SpokeTensionRatio,
} from '~/utils/spokeMath'
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
  rimOffsetMm: number
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
  rimOffsetMm: 0,
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
  rimOffsetMm: 0,
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
  tensionRatio: SpokeTensionRatio | null
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

const formatTensionRatio = (value: SpokeTensionRatio | null | undefined) => (
  value == null ? '--' : `${Math.round(value.lowerToHigher * 100)}%`
)

const formatDirectionalTensionRatio = (value: SpokeTensionRatio | null | undefined) => {
  if (!value) return '--'
  if (value.leftToRight <= 1) {
    return `${Math.round(value.leftToRight * 100)}% : 100%`
  }
  return `100% : ${Math.round(value.rightToLeft * 100)}%`
}

const lowerTensionSideLabel = (value: SpokeTensionRatio | null | undefined) => {
  if (!value || value.lowerSide === 'balanced') return 'balanced'
  return value.lowerSide
}

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
            front_rim_offset_mm: frontConfig.rimOffsetMm,
            rear_rim_offset_mm: rearConfig.rimOffsetMm,
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
    tensionRatio: computeSpokeTensionRatio(
      effectiveSpokeFlangeDistance(config.leftFlange ?? 0, config.rimOffsetMm, 'left'),
      effectiveSpokeFlangeDistance(config.rightFlange ?? 0, config.rimOffsetMm, 'right'),
      leftLengthMm ?? 0,
      rightLengthMm ?? 0
    ),
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
  if (Math.abs(config.rimOffsetMm) > 0.001) return false
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

  const leftFlange = effectiveSpokeFlangeDistance(config.leftFlange, config.rimOffsetMm, 'left')
  const rightFlange = effectiveSpokeFlangeDistance(config.rightFlange, config.rimOffsetMm, 'right')
  if (leftFlange <= 0 || rightFlange <= 0) return null

  return {
    leftLengthMm: computeSpokeLength(
      config.erd,
      config.leftFlangePcd,
      leftFlange,
      config.spokeCount,
      config.crossing,
      config.nippleType,
      config.nippleLength
    ),
    rightLengthMm: computeSpokeLength(
      config.erd,
      config.rightFlangePcd,
      rightFlange,
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

.spoke-calculator__tension-summary {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
  padding: 0.75rem 1rem;
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

.spoke-calculator__build-settings {
  display: grid;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background: rgba(255, 255, 255, 0.72);
}

.spoke-calculator__build-settings-header {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}

.spoke-calculator__build-settings-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.spoke-calculator__build-settings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  min-width: 0;
}

.spoke-calculator__setting-field {
  display: grid;
  gap: 0.3rem;
  min-width: 0;
}

.spoke-calculator__results-note {
  padding: 0.95rem 1rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  line-height: 1.55;
}

.spoke-calculator__blueprint-sheet {
  display: grid;
  gap: 0.75rem;
  padding: 0.875rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(248, 250, 252, 0.96)),
    var(--spoke-shell-surface);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    0 12px 28px rgba(20, 32, 43, 0.08);
}

.spoke-calculator__blueprint-sheet-header {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.spoke-calculator__blueprint-eyebrow {
  margin: 0 0 0.1rem;
  color: rgba(5, 150, 105, 0.84);
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.spoke-calculator__blueprint-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1.2;
}

.spoke-calculator__blueprint-note {
  max-width: 18rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  line-height: 1.4;
  text-align: right;
}

.spoke-calculator__blueprint-grid {
  display: grid;
  gap: 0.85rem;
}

.spoke-calculator__schematic-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.spoke-calculator__legend-table {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background: rgba(255, 255, 255, 0.7);
}

.spoke-calculator__legend-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(0, 1.55fr);
  gap: 0.75rem;
  align-items: start;
  padding: 0.65rem 0.7rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
}

.spoke-calculator__legend-row--split {
  align-items: center;
}

.spoke-calculator__legend-badge {
  display: inline-flex;
  width: 1.8rem;
  height: 1.8rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.28);
  border-radius: 50%;
  background: rgba(5, 150, 105, 0.08);
  color: var(--tz-text-accent);
  font-size: 0.8rem;
  font-weight: 800;
  line-height: 1;
}

.spoke-calculator__legend-heading {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.spoke-calculator__legend-heading strong {
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 750;
  line-height: 1.2;
}

.spoke-calculator__legend-heading span {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.35;
}

.spoke-calculator__legend-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
  min-width: 0;
}

.spoke-calculator__legend-field {
  display: grid;
  gap: 0.28rem;
  color: var(--tz-text-secondary);
  font-size: 0.65rem;
  font-weight: 700;
}

.spoke-calculator__legend-field > span {
  color: var(--tz-text-secondary);
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-calculator__legend-control {
  min-width: 0;
}

@media (max-width: 1100px) {
  .spoke-calculator__schematic-grid {
    grid-template-columns: 1fr;
  }

  .spoke-calculator__blueprint-note {
    max-width: none;
  }
}

@media (max-width: 767px) {
  .spoke-calculator__panel {
    padding: 0.5rem 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__panel + .spoke-calculator__panel {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--spoke-border);
  }

  .spoke-calculator__build-settings-grid {
    grid-template-columns: 1fr;
  }

  .spoke-calculator__build-settings-header {
    flex-direction: column;
  }

  .spoke-calculator__blueprint-sheet {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__blueprint-sheet-header {
    flex-direction: column;
    padding-bottom: 0.35rem;
    border-bottom: 1px solid var(--spoke-border);
  }

  .spoke-calculator__blueprint-note {
    text-align: left;
  }

  .spoke-calculator__legend-table,
  .spoke-calculator__build-settings,
  .spoke-calculator__results-shell {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__legend-row {
    grid-template-columns: 1fr;
    padding: 0.4rem;
    gap: 0.4rem;
  }

  .spoke-calculator__legend-table {
    gap: 0.45rem;
  }

  .spoke-calculator__legend-fields {
    grid-template-columns: 1fr;
    gap: 0.4rem;
  }

  .spoke-calculator__build-settings-header {
    padding-bottom: 0.35rem;
    border-bottom: 1px solid var(--spoke-border);
  }

  .spoke-wheel {
    gap: 0.25rem;
  }

  .spoke-wheel__header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
  }

  .spoke-wheel__title {
    text-align: left;
  }
}
</style>
