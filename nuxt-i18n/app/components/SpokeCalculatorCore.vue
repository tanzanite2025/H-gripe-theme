<template>
  <div class="spoke-calculator">
    <div class="grid gap-6 items-start">
      <section
        class="rounded-2xl p-6 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]"
      >
        <h2 class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400 mb-4">
          Wheel setup
        </h2>

        <!-- Two-column layout: Front Wheel | Rear Wheel -->
        <div class="grid gap-6 md:grid-cols-2">
          <!-- ========== FRONT WHEEL COLUMN ========== -->
          <div class="space-y-4 p-4 rounded-xl bg-slate-950/70 shadow-[0_8px_22px_-14px_rgba(0,0,0,0.95)]">
            <h3 class="text-sm font-semibold text-sky-400 uppercase tracking-wide">Front Wheel</h3>

            <!-- Spoke count -->
            <div class="space-y-1.5">
              <label for="front-spoke-count" class="block text-xs font-medium text-slate-200">Spoke count</label>
              <select
                id="front-spoke-count"
                v-model.number="frontConfig.spokeCount"
                class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
              >
                <option :value="16" class="bg-slate-900 text-slate-50">16</option>
                <option :value="18" class="bg-slate-900 text-slate-50">18</option>
                <option :value="20" class="bg-slate-900 text-slate-50">20</option>
                <option :value="24" class="bg-slate-900 text-slate-50">24</option>
                <option :value="28" class="bg-slate-900 text-slate-50">28</option>
                <option :value="32" class="bg-slate-900 text-slate-50">32</option>
                <option :value="36" class="bg-slate-900 text-slate-50">36</option>
              </select>
            </div>

            <!-- Lacing pattern -->
            <div class="space-y-1.5">
              <label for="front-lacing" class="block text-xs font-medium text-slate-200">Lacing pattern</label>
              <select
                id="front-lacing"
                v-model.number="frontConfig.crossing"
                class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
              >
                <option :value="0" class="bg-slate-900 text-slate-50">0-cross (Radial)</option>
                <option :value="1" class="bg-slate-900 text-slate-50">1-cross</option>
                <option :value="2" class="bg-slate-900 text-slate-50">2-cross</option>
                <option :value="3" class="bg-slate-900 text-slate-50">3-cross</option>
                <option :value="4" class="bg-slate-900 text-slate-50">4-cross</option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <!-- Nipple type -->
              <div class="space-y-1.5">
                <label for="front-nipple" class="block text-xs font-medium text-slate-200">Nipple type</label>
                <select
                  id="front-nipple"
                  v-model="frontConfig.nippleType"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="standard" class="bg-slate-900 text-slate-50">Standard external</option>
                  <option value="hidden" class="bg-slate-900 text-slate-50">Hidden / aero</option>
                </select>
              </div>

              <!-- Nipple length (hidden nipples only) -->
              <div v-if="frontConfig.nippleType === 'hidden'" class="space-y-1.5">
                <label for="front-nipple-length" class="block text-xs font-medium text-slate-200">Nipple length</label>
                <div class="flex items-center gap-2">
                  <input
                    id="front-nipple-length"
                    v-model.number="frontConfig.nippleLength"
                    type="number"
                    min="0"
                    max="30"
                    placeholder="e.g. 12"
                    class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                  />
                  <span class="text-[11px] text-slate-400">mm</span>
                </div>
              </div>
            </div>

            <!-- Rim Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="front-rim-brand" class="block text-xs font-medium text-slate-200">
                  Rim Brand
                </label>
                <select
                  id="front-rim-brand"
                  v-model="frontConfig.rimBrandId"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="" disabled>Select Brand</option>
                  <option v-for="brand in RIM_DATABASE" :key="brand.id" :value="brand.id" class="bg-slate-900 text-slate-50">{{ brand.name }}</option>
                </select>
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="front-rim-model" class="block text-xs font-medium text-slate-200">
                  Rim Model
                </label>
                <select
                  id="front-rim-model"
                  v-model="frontConfig.rimModelId"
                   :disabled="!frontRimModels.length"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="" disabled>Select Model</option>
                  <option v-for="rim in frontRimModels" :key="rim.id" :value="rim.id" class="bg-slate-900 text-slate-50">{{ rim.name }}</option>
                </select>
               </div>
            </div>

            <!-- Hub Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="front-hub-brand" class="block text-xs font-medium text-slate-200">
                  Hub Brand
                </label>
                <select
                  id="front-hub-brand"
                  v-model="frontConfig.hubBrandId"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="" disabled>Select Brand</option>
                  <option v-for="brand in HUB_DATABASE" :key="brand.id" :value="brand.id" class="bg-slate-900 text-slate-50">{{ brand.name }}</option>
                </select>
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="front-hub-model" class="block text-xs font-medium text-slate-200">
                  Hub Model
                </label>
                <select
                  id="front-hub-model"
                  v-model="frontConfig.hubModelId"
                   :disabled="!frontHubModels.length"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="" disabled>Select Model</option>
                  <option v-for="hub in frontHubModels" :key="hub.id" :value="hub.id" class="bg-slate-900 text-slate-50">{{ hub.name }}</option>
                </select>
               </div>
            </div>

            <!-- ERD -->
            <div class="space-y-1.5">
              <label for="front-erd" class="block text-xs font-medium text-slate-200">ERD (effective rim diameter)</label>
              <div class="flex items-center gap-2">
                <input
                  id="front-erd"
                  v-model.number="frontConfig.erd"
                  type="number"
                  min="400"
                  max="750"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange distance -->
            <div class="space-y-1.5">
              <label for="front-left-flange" class="block text-xs font-medium text-slate-200">Left flange distance</label>
              <div class="flex items-center gap-2">
                <input
                  id="front-left-flange"
                  v-model.number="frontConfig.leftFlange"
                  type="number"
                  min="10"
                  max="60"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange distance -->
            <div class="space-y-1.5">
              <label for="front-right-flange" class="block text-xs font-medium text-slate-200">Right flange distance</label>
              <div class="flex items-center gap-2">
                <input
                  id="front-right-flange"
                  v-model.number="frontConfig.rightFlange"
                  type="number"
                  min="10"
                  max="60"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange PCD -->
            <div class="space-y-1.5">
              <label for="front-left-flange-pcd" class="block text-xs font-medium text-slate-200">Left flange PCD</label>
              <div class="flex items-center gap-2">
                <input
                  id="front-left-flange-pcd"
                  v-model.number="frontConfig.leftFlangePcd"
                  type="number"
                  min="30"
                  max="80"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange PCD -->
            <div class="space-y-1.5">
              <label for="front-right-flange-pcd" class="block text-xs font-medium text-slate-200">Right flange PCD</label>
              <div class="flex items-center gap-2">
                <input
                  id="front-right-flange-pcd"
                  v-model.number="frontConfig.rightFlangePcd"
                  type="number"
                  min="30"
                  max="80"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>
          </div>

          <!-- ========== REAR WHEEL COLUMN ========== -->
          <div class="space-y-4 p-4 rounded-xl bg-slate-950/70 shadow-[0_8px_22px_-14px_rgba(0,0,0,0.95)]">
            <h3 class="text-sm font-semibold text-emerald-400 uppercase tracking-wide">Rear Wheel</h3>

            <!-- Spoke count -->
            <div class="space-y-1.5">
              <label for="rear-spoke-count" class="block text-xs font-medium text-slate-200">Spoke count</label>
              <select
                id="rear-spoke-count"
                v-model.number="rearConfig.spokeCount"
                class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
              >
                <option :value="16" class="bg-slate-900 text-slate-50">16</option>
                <option :value="18" class="bg-slate-900 text-slate-50">18</option>
                <option :value="20" class="bg-slate-900 text-slate-50">20</option>
                <option :value="24" class="bg-slate-900 text-slate-50">24</option>
                <option :value="28" class="bg-slate-900 text-slate-50">28</option>
                <option :value="32" class="bg-slate-900 text-slate-50">32</option>
                <option :value="36" class="bg-slate-900 text-slate-50">36</option>
              </select>
            </div>

            <!-- Lacing pattern -->
            <div class="space-y-1.5">
              <label for="rear-lacing" class="block text-xs font-medium text-slate-200">Lacing pattern</label>
              <select
                id="rear-lacing"
                v-model.number="rearConfig.crossing"
                class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
              >
                <option :value="0" class="bg-slate-900 text-slate-50">0-cross (Radial)</option>
                <option :value="1" class="bg-slate-900 text-slate-50">1-cross</option>
                <option :value="2" class="bg-slate-900 text-slate-50">2-cross</option>
                <option :value="3" class="bg-slate-900 text-slate-50">3-cross</option>
                <option :value="4" class="bg-slate-900 text-slate-50">4-cross</option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <!-- Nipple type -->
              <div class="space-y-1.5">
                <label for="rear-nipple" class="block text-xs font-medium text-slate-200">Nipple type</label>
                <select
                  id="rear-nipple"
                  v-model="rearConfig.nippleType"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="standard" class="bg-slate-900 text-slate-50">Standard external</option>
                  <option value="hidden" class="bg-slate-900 text-slate-50">Hidden / aero</option>
                </select>
              </div>

              <!-- Nipple length (hidden nipples only) -->
              <div v-if="rearConfig.nippleType === 'hidden'" class="space-y-1.5">
                <label for="rear-nipple-length" class="block text-xs font-medium text-slate-200">Nipple length</label>
                <div class="flex items-center gap-2">
                  <input
                    id="rear-nipple-length"
                    v-model.number="rearConfig.nippleLength"
                    type="number"
                    min="0"
                    max="30"
                    placeholder="e.g. 12"
                    class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                  />
                  <span class="text-[11px] text-slate-400">mm</span>
                </div>
              </div>
            </div>

            <!-- Rim Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="rear-rim-brand" class="block text-xs font-medium text-slate-200">
                  Rim Brand
                </label>
                <select
                  id="rear-rim-brand"
                  v-model="rearConfig.rimBrandId"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="" disabled>Select Brand</option>
                  <option v-for="brand in RIM_DATABASE" :key="brand.id" :value="brand.id" class="bg-slate-900 text-slate-50">{{ brand.name }}</option>
                </select>
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="rear-rim-model" class="block text-xs font-medium text-slate-200">
                  Rim Model
                </label>
                <select
                  id="rear-rim-model"
                  v-model="rearConfig.rimModelId"
                   :disabled="!rearRimModels.length"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="" disabled>Select Model</option>
                  <option v-for="rim in rearRimModels" :key="rim.id" :value="rim.id" class="bg-slate-900 text-slate-50">{{ rim.name }}</option>
                </select>
               </div>
            </div>

            <!-- Hub Selection -->
            <div class="space-y-3">
               <!-- Brand -->
               <div class="space-y-1.5">
                <label for="rear-hub-brand" class="block text-xs font-medium text-slate-200">
                  Hub Brand
                </label>
                <select
                  id="rear-hub-brand"
                  v-model="rearConfig.hubBrandId"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                >
                  <option value="" disabled>Select Brand</option>
                  <option v-for="brand in HUB_DATABASE" :key="brand.id" :value="brand.id" class="bg-slate-900 text-slate-50">{{ brand.name }}</option>
                </select>
              </div>

               <!-- Model -->
              <div class="space-y-1.5">
                <label for="rear-hub-model" class="block text-xs font-medium text-slate-200">
                  Hub Model
                </label>
                <select
                  id="rear-hub-model"
                  v-model="rearConfig.hubModelId"
                   :disabled="!rearHubModels.length"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="" disabled>Select Model</option>
                  <option v-for="hub in rearHubModels" :key="hub.id" :value="hub.id" class="bg-slate-900 text-slate-50">{{ hub.name }}</option>
                </select>
               </div>
            </div>

            <!-- ERD -->
            <div class="space-y-1.5">
              <label for="rear-erd" class="block text-xs font-medium text-slate-200">ERD (effective rim diameter)</label>
              <div class="flex items-center gap-2">
                <input
                  id="rear-erd"
                  v-model.number="rearConfig.erd"
                  type="number"
                  min="400"
                  max="750"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange distance -->
            <div class="space-y-1.5">
              <label for="rear-left-flange" class="block text-xs font-medium text-slate-200">Left flange distance</label>
              <div class="flex items-center gap-2">
                <input
                  id="rear-left-flange"
                  v-model.number="rearConfig.leftFlange"
                  type="number"
                  min="10"
                  max="60"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange distance -->
            <div class="space-y-1.5">
              <label for="rear-right-flange" class="block text-xs font-medium text-slate-200">Right flange distance</label>
              <div class="flex items-center gap-2">
                <input
                  id="rear-right-flange"
                  v-model.number="rearConfig.rightFlange"
                  type="number"
                  min="10"
                  max="60"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Left flange PCD -->
            <div class="space-y-1.5">
              <label for="rear-left-flange-pcd" class="block text-xs font-medium text-slate-200">Left flange PCD</label>
              <div class="flex items-center gap-2">
                <input
                  id="rear-left-flange-pcd"
                  v-model.number="rearConfig.leftFlangePcd"
                  type="number"
                  min="30"
                  max="80"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>

            <!-- Right flange PCD -->
            <div class="space-y-1.5">
              <label for="rear-right-flange-pcd" class="block text-xs font-medium text-slate-200">Right flange PCD</label>
              <div class="flex items-center gap-2">
                <input
                  id="rear-right-flange-pcd"
                  v-model.number="rearConfig.rightFlangePcd"
                  type="number"
                  min="30"
                  max="80"
                  class="block w-full rounded-lg border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.8),0_0_14px_rgba(56,189,248,0.35)]"
                />
                <span class="text-[11px] text-slate-400">mm</span>
              </div>
            </div>
          </div>
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
        <section
          class="mt-6 rounded-2xl p-6 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]"
        >
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
                <div class="rounded-xl bg-slate-950/80 px-4 py-3 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Left side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ frontLeftDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
                <div class="rounded-xl bg-slate-950/80 px-4 py-3 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]">
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
                <div class="rounded-xl bg-slate-950/80 px-4 py-3 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]">
                  <div class="text-[11px] uppercase tracking-[0.16em] text-slate-500 mb-1">Left side</div>
                  <div class="flex items-baseline gap-1">
                    <span class="text-2xl font-semibold text-slate-50">{{ rearLeftDisplay }}</span>
                    <span class="text-xs text-slate-400">mm</span>
                  </div>
                </div>
                <div class="rounded-xl bg-slate-950/80 px-4 py-3 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]">
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
import { computed, reactive, watch } from 'vue'
import { RIM_DATABASE, HUB_DATABASE, type RimModel, type HubModel } from '~/data/spoke-calculator/database'

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

// --- Computed Models based on Brand Selection ---

// Front Rim Models
const frontRimModels = computed<RimModel[]>(() => {
  if (!frontConfig.rimBrandId) return []
  const brand = RIM_DATABASE.find(b => b.id === frontConfig.rimBrandId)
  return brand ? brand.items : []
})

// Front Hub Models
const frontHubModels = computed<HubModel[]>(() => {
  if (!frontConfig.hubBrandId) return []
  const brand = HUB_DATABASE.find(b => b.id === frontConfig.hubBrandId)
  return brand ? brand.items : []
})

// Rear Rim Models
const rearRimModels = computed<RimModel[]>(() => {
  if (!rearConfig.rimBrandId) return []
  const brand = RIM_DATABASE.find(b => b.id === rearConfig.rimBrandId)
  return brand ? brand.items : []
})

// Rear Hub Models
const rearHubModels = computed<HubModel[]>(() => {
  if (!rearConfig.hubBrandId) return []
  const brand = HUB_DATABASE.find(b => b.id === rearConfig.hubBrandId)
  return brand ? brand.items : []
})

// --- Watchers for Auto-Population ---

// Front Rim Change
watch(
  () => frontConfig.rimModelId,
  (newId) => {
    if (!newId) return
    const model = frontRimModels.value.find(m => m.id === newId)
    if (model) {
      frontConfig.erd = model.erd
    }
  }
)

// Front Hub Change
watch(
  () => frontConfig.hubModelId,
  (newId) => {
    if (!newId) return
    const model = frontHubModels.value.find(m => m.id === newId)
    if (model && model.front) {
      frontConfig.leftFlange = model.front.leftFlange
      frontConfig.rightFlange = model.front.rightFlange
      frontConfig.leftFlangePcd = model.front.leftFlangePcd
      frontConfig.rightFlangePcd = model.front.rightFlangePcd
    } else if (model && model.rear) {
       // Fallback if user selects a rear hub for front (unlikely but possible logic)
       // Or handle front/rear compatible logic here
    }
  }
)

// Rear Rim Change
watch(
  () => rearConfig.rimModelId,
  (newId) => {
    if (!newId) return
    const model = rearRimModels.value.find(m => m.id === newId)
    if (model) {
      rearConfig.erd = model.erd
    }
  }
)

// Rear Hub Change
watch(
  () => rearConfig.hubModelId,
  (newId) => {
    if (!newId) return
    const model = rearHubModels.value.find(m => m.id === newId)
    if (model && model.rear) {
      rearConfig.leftFlange = model.rear.leftFlange
      rearConfig.rightFlange = model.rear.rightFlange
      rearConfig.leftFlangePcd = model.rear.leftFlangePcd
      rearConfig.rightFlangePcd = model.rear.rightFlangePcd
    } else if (model && model.front) {
        // Fallback
    }
  }
)

// --- Calculation Logic (Existing) ---
// Frontend-only spoke length calculation
// Formula: L = sqrt((ERD/2)^2 + (PCD/2)^2 + flange^2 - ERD * PCD/2 * cos(cross_angle))
// where cross_angle = 4 * PI * crossing / spokeCount
function computeSpokeLength(
  erd: number,
  flangePcd: number,
  flangeDistance: number,
  spokeCount: number,
  crossing: number,
  nippleType: 'standard' | 'hidden' = 'standard',
  nippleLength: number | null = null
): number {
  const erdRadius = erd / 2
  const pcdRadius = flangePcd / 2
  const crossAngle = (4 * Math.PI * crossing) / spokeCount

  // Standard spoke length formula based on triangle geometry
  const lengthSquared =
    erdRadius * erdRadius +
    pcdRadius * pcdRadius +
    flangeDistance * flangeDistance -
    2 * erdRadius * pcdRadius * Math.cos(crossAngle)

  let length = Math.sqrt(lengthSquared)

  // Hidden nipple correction: ADD length based on nipple depth
  // 9mm nipple → +6mm, 12mm nipple → +9mm (nipple length - 3)
  if (nippleType === 'hidden' && nippleLength) {
    const correction = nippleLength - 3
    length += correction
  }

  return Number(length.toFixed(1))
}

// Local state for calculation results (no API)
const loading = ref(false)
const error = ref<string | null>(null)

interface SpokeResult {
  leftLengthMm: number
  rightLengthMm: number
}

const frontResult = ref<SpokeResult | null>(null)
const rearResult = ref<SpokeResult | null>(null)

// Display values for front wheel
const frontLeftDisplay = computed(() => (frontResult.value?.leftLengthMm ?? 0).toFixed(1))
const frontRightDisplay = computed(() => (frontResult.value?.rightLengthMm ?? 0).toFixed(1))

// Display values for rear wheel
const rearLeftDisplay = computed(() => (rearResult.value?.leftLengthMm ?? 0).toFixed(1))
const rearRightDisplay = computed(() => (rearResult.value?.rightLengthMm ?? 0).toFixed(1))

const onCalculate = () => {
  error.value = null
  loading.value = true

  try {
    // Validate front wheel inputs
    if (frontConfig.erd && frontConfig.leftFlangePcd && frontConfig.rightFlangePcd &&
        frontConfig.leftFlange != null && frontConfig.rightFlange != null) {
      frontResult.value = {
        leftLengthMm: computeSpokeLength(
          frontConfig.erd,
          frontConfig.leftFlangePcd,
          frontConfig.leftFlange,
          frontConfig.spokeCount,
          frontConfig.crossing,
          frontConfig.nippleType,
          frontConfig.nippleLength
        ),
        rightLengthMm: computeSpokeLength(
          frontConfig.erd,
          frontConfig.rightFlangePcd,
          frontConfig.rightFlange,
          frontConfig.spokeCount,
          frontConfig.crossing,
          frontConfig.nippleType,
          frontConfig.nippleLength
        ),
      }
    }

    // Validate rear wheel inputs
    if (rearConfig.erd && rearConfig.leftFlangePcd && rearConfig.rightFlangePcd &&
        rearConfig.leftFlange != null && rearConfig.rightFlange != null) {
      rearResult.value = {
        leftLengthMm: computeSpokeLength(
          rearConfig.erd,
          rearConfig.leftFlangePcd,
          rearConfig.leftFlange,
          rearConfig.spokeCount,
          rearConfig.crossing,
          rearConfig.nippleType,
          rearConfig.nippleLength
        ),
        rightLengthMm: computeSpokeLength(
          rearConfig.erd,
          rearConfig.rightFlangePcd,
          rearConfig.rightFlange,
          rearConfig.spokeCount,
          rearConfig.crossing,
          rearConfig.nippleType,
          rearConfig.nippleLength
        ),
      }
    }
  } catch (e: any) {
    error.value = e?.message || 'Calculation failed'
  } finally {
    loading.value = false
  }
}
</script>
