<template>
  <Teleport to="body">
    <div
      v-if="modelValue && kind === 'tire'"
      class="tire-rim-width-selection-dialog__picker-mask fixed inset-0 z-[14000] flex items-center justify-center p-2 md:p-5 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
      role="presentation"
      @click.self="closeTireRimWidthSelectionDialog"
      @keydown.esc.stop.prevent="closeTireRimWidthSelectionDialog"
    >
      <section
        ref="widthSelectionDialogElement"
        class="tire-rim-width-selection-dialog__picker tz-mobile-dialog-surface w-full max-w-2xl overflow-hidden rounded-2xl p-4"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="dialogTitleId"
        @click.stop
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 :id="dialogTitleId" class="text-base font-semibold tz-text-primary">
              {{ t('guidesTireRimHelper.tireWidthPickerTitle') }}
            </h2>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ description || t('guidesTireRimHelper.tireWidthPickerDescription') }}
            </p>
          </div>
          <button
            type="button"
            class="tz-global-close-btn shrink-0"
            :aria-label="t('guidesTireRimHelper.tireWidthPickerClose')"
            @click="closeTireRimWidthSelectionDialog"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>

        <div class="tire-rim-width-selection-dialog__picker-options mt-4 min-h-0 flex-1 overflow-y-auto pr-1">
          <p class="mb-2 text-xs font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.tireWidthPresetCSection') }}
          </p>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="preset in cMarkingTireWidthPresets"
              :key="preset.id"
              type="button"
              class="tire-rim-width-selection-dialog__preset rounded-lg px-3 py-2 text-left transition-colors"
              :class="{ 'tire-rim-width-selection-dialog__preset--active': selectedTireWidthPresetId === preset.id }"
              @click="selectTireWidthPresetAndCloseDialog(preset)"
            >
              <span class="block text-sm font-semibold tz-text-primary">{{ preset.label }}</span>
              <span class="mt-0.5 block tz-caption tz-text-secondary">
                {{ preset.millimeters }} mm · {{ formatTireWidthPresetInches(preset) }}″
              </span>
            </button>
          </div>

          <p class="mb-2 mt-5 text-xs font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.tireWidthPresetInchSection') }}
          </p>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="preset in inchMarkingTireWidthPresets"
              :key="preset.id"
              type="button"
              class="tire-rim-width-selection-dialog__preset rounded-lg px-3 py-2 text-left transition-colors"
              :class="{ 'tire-rim-width-selection-dialog__preset--active': selectedTireWidthPresetId === preset.id }"
              @click="selectTireWidthPresetAndCloseDialog(preset)"
            >
              <span class="block text-sm font-semibold tz-text-primary">{{ preset.label }}</span>
              <span class="mt-0.5 block tz-caption tz-text-secondary">
                {{ preset.millimeters }} mm · {{ formatTireWidthPresetInches(preset) }}″
              </span>
            </button>
          </div>

          <p class="mb-2 mt-5 text-xs font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.tireWidthPresetMillimeterSection') }}
          </p>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="tireWidthMm in additionalMillimeterTireWidthPresets"
              :key="tireWidthMm"
              type="button"
              class="tire-rim-width-selection-dialog__preset rounded-lg px-3 py-2 text-left transition-colors"
              :class="{ 'tire-rim-width-selection-dialog__preset--active': !selectedTireWidthPresetId && selectedTireWidthMm === tireWidthMm }"
              @click="selectTireWidthMillimetersAndCloseDialog(tireWidthMm)"
            >
              <span class="block text-sm font-semibold tz-text-primary">{{ tireWidthMm }} mm</span>
              <span class="mt-0.5 block tz-caption tz-text-secondary">
                {{ formatTireWidthMillimetersAsInches(tireWidthMm) }}″
              </span>
            </button>
          </div>

        </div>
      </section>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="modelValue && kind === 'rim'"
      class="tire-rim-width-selection-dialog__picker-mask fixed inset-0 z-[14000] flex items-center justify-center p-2 md:p-5 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
      role="presentation"
      @click.self="closeTireRimWidthSelectionDialog"
      @keydown.esc.stop.prevent="closeTireRimWidthSelectionDialog"
    >
      <section
        ref="widthSelectionDialogElement"
        class="tire-rim-width-selection-dialog__picker tz-mobile-dialog-surface w-full max-w-6xl overflow-hidden rounded-2xl p-5 md:p-6"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="dialogTitleId"
        @click.stop
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 :id="dialogTitleId" class="text-base font-semibold tz-text-primary">
              {{ t('guidesTireRimHelper.rimInnerWidthPickerTitle') }}
            </h2>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ description || t('guidesTireRimHelper.rimInnerWidthPickerDescription') }}
            </p>
          </div>
          <button
            type="button"
            class="tz-global-close-btn shrink-0"
            :aria-label="t('guidesTireRimHelper.rimInnerWidthPickerClose')"
            @click="closeTireRimWidthSelectionDialog"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>

        <div class="tire-rim-width-selection-dialog__rim-inner-width-options mt-5 grid grid-cols-4 gap-2 sm:grid-cols-6 lg:grid-cols-8">
          <button
            v-for="rimInnerWidth in supportedRimInnerWidthsMm"
            :key="rimInnerWidth"
            type="button"
            class="tire-rim-width-selection-dialog__preset rounded-lg px-3 py-3 text-center transition-colors"
            :class="{ 'tire-rim-width-selection-dialog__preset--active': selectedRimInnerWidthMm === rimInnerWidth }"
            @click="selectRimInnerWidthAndCloseDialog(rimInnerWidth)"
          >
            <span class="block text-sm font-semibold tz-text-primary">{{ rimInnerWidth }} mm</span>
          </button>
        </div>
      </section>
    </div>
  </Teleport>

</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { useFocusTrap } from '@vueuse/integrations/useFocusTrap'
import { useI18n } from '#imports'
import {
  TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM,
  TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM,
} from '~/data/tireguides/tireRimEngineeringReferenceUiOptions'
import { commonTireWidthInputPresets, type TireWidthInputPreset } from '~/data/tireguides/tireRimWidthInputConversion'

const props = defineProps<{
  modelValue: boolean
  kind: 'tire' | 'rim'
  selectedTireWidthPresetId?: string
  selectedTireWidthMm?: number | null
  selectedRimInnerWidthMm?: number
  description?: string
}>()
const emit = defineEmits<{
  'update:modelValue': [open: boolean]
  selectTireWidthPreset: [preset: TireWidthInputPreset]
  selectTireWidthMillimeters: [widthMm: number]
  selectRimInnerWidth: [widthMm: number]
}>()
const { t } = useI18n()
const dialogTitleId = useId()
const widthSelectionDialogElement = ref<HTMLElement | null>(null)
const {
  activate: activateTireRimWidthSelectionFocusTrap,
  deactivate: deactivateTireRimWidthSelectionFocusTrap,
} = useFocusTrap(widthSelectionDialogElement, { escapeDeactivates: false, allowOutsideClick: true })

const cMarkingTireWidthPresets = commonTireWidthInputPresets.filter(preset => preset.category === 'c-marking')
const inchMarkingTireWidthPresets = commonTireWidthInputPresets.filter(preset => preset.category === 'inch-marking')
const supportedRimInnerWidthsMm = TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM
// Keep every size previously offered by the simulator in the shared picker.
const additionalMillimeterTireWidthPresets = TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM.filter(
  width => !commonTireWidthInputPresets.some(preset => preset.millimeters === width),
)

const formatTireWidthPresetInches = (preset: TireWidthInputPreset): string =>
  String(Number(preset.inches.toFixed(2)))
const formatTireWidthMillimetersAsInches = (widthMm: number): string =>
  String(Number((widthMm / 25.4).toFixed(2)))

const closeTireRimWidthSelectionDialog = (): void => {
  emit('update:modelValue', false)
}
const selectTireWidthPresetAndCloseDialog = (preset: TireWidthInputPreset): void => {
  emit('selectTireWidthPreset', preset)
  closeTireRimWidthSelectionDialog()
}
const selectTireWidthMillimetersAndCloseDialog = (widthMm: number): void => {
  emit('selectTireWidthMillimeters', widthMm)
  closeTireRimWidthSelectionDialog()
}
const selectRimInnerWidthAndCloseDialog = (widthMm: number): void => {
  emit('selectRimInnerWidth', widthMm)
  closeTireRimWidthSelectionDialog()
}

const handleTireRimWidthSelectionDialogEscapeKey = (event: KeyboardEvent): void => {
  if (event.key !== 'Escape' || !props.modelValue) return
  event.preventDefault()
  closeTireRimWidthSelectionDialog()
}

watch(() => props.modelValue, async (open) => {
  if (!open) {
    deactivateTireRimWidthSelectionFocusTrap()
    return
  }
  await nextTick()
  if (props.modelValue) activateTireRimWidthSelectionFocusTrap()
}, { flush: 'post' })

onMounted(() => {
  window.addEventListener('keydown', handleTireRimWidthSelectionDialogEscapeKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleTireRimWidthSelectionDialogEscapeKey)
  deactivateTireRimWidthSelectionFocusTrap()
})
</script>

<style scoped>
.tire-rim-width-selection-dialog__picker-mask {
  background: rgba(15, 23, 42, 0.35);
  backdrop-filter: blur(3px);
}

.tire-rim-width-selection-dialog__picker {
  display: flex;
  flex-direction: column;
  max-height: min(860px, 96vh);
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-panel-surface);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.24);
}

.tire-rim-width-selection-dialog__picker-options {
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.tire-rim-width-selection-dialog__rim-inner-width-options {
  min-height: 0;
  max-height: min(42rem, 68vh);
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}


.tire-rim-width-selection-dialog__preset {
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-control-surface);
}

.tire-rim-width-selection-dialog__preset:hover,
.tire-rim-width-selection-dialog__preset:focus-visible,
.tire-rim-width-selection-dialog__preset--active {
  border-color: var(--tz-site-accent);
  background: color-mix(in srgb, var(--tz-site-accent) 10%, var(--tz-form-control-surface));
  outline: none;
}


@media (max-width: 767px) {
  .tire-rim-width-selection-dialog__picker {
    max-height: min(
      860px,
      calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2)
    );
  }
}
</style>
