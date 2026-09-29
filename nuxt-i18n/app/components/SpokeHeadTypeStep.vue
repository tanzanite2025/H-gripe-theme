<template>
  <section class="spoke-head-step" aria-labelledby="spoke-head-step-title">
    <SpokeStepNavigation
      :current-step="currentStep"
      @select="emit('select-step', $event)"
      @next="emit('next')"
    />

    <div class="spoke-head-step__intro">
      <div>
        <h2 id="spoke-head-step-title" class="spoke-head-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepOnePrompt', '先选择与你的花鼓匹配的辐条类型') }}
        </h2>
      </div>
    </div>

    <div class="spoke-head-step__options" role="group" :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.spokeHeadType', '花鼓辐条头类型')">
      <button
        type="button"
        class="spoke-head-step__option"
        :class="{ 'spoke-head-step__option--selected': selectedType === 'j_bend' }"
        :aria-pressed="selectedType === 'j_bend'"
        @click="selectedType = 'j_bend'"
      >
        <span class="spoke-head-step__option-head">
          <span class="spoke-head-step__number">A</span>
          <span class="spoke-head-step__option-title">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.jBend', 'J 弯头') }}
          </span>
          <span v-if="selectedType === 'j_bend'" class="spoke-head-step__selected-mark" aria-hidden="true">✓</span>
        </span>

        <span class="spoke-head-step__option-content">
          <span
            class="spoke-head-step__diagram spoke-head-step__diagram--j-bend"
            aria-hidden="true"
          >
            <img
              src="/technical/spoke-calculator/head-types/j-bend-spoke.webp"
              :alt="t('resourcesSpokeCalculator.calculator.physicalCorrections.jBendHint', 'The spoke forms a hook at the flange hole.')"
            />
          </span>
          <span class="spoke-head-step__explanation">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.jBendHint', 'The spoke forms a hook at the flange hole.') }}
          </span>
        </span>
      </button>

      <button
        type="button"
        class="spoke-head-step__option"
        :class="{ 'spoke-head-step__option--selected': selectedType === 'straight_pull' }"
        :aria-pressed="selectedType === 'straight_pull'"
        @click="selectedType = 'straight_pull'"
      >
        <span class="spoke-head-step__option-head">
          <span class="spoke-head-step__number">B</span>
          <span class="spoke-head-step__option-title">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPull', '直拉式') }}
          </span>
          <span v-if="selectedType === 'straight_pull'" class="spoke-head-step__selected-mark" aria-hidden="true">✓</span>
        </span>

        <span class="spoke-head-step__option-content">
          <span
            class="spoke-head-step__diagram spoke-head-step__diagram--straight-pull"
            aria-hidden="true"
          >
            <img
              src="/technical/spoke-calculator/head-types/straight-pull-spoke.webp"
              :alt="t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPullHint', 'The spoke exits the hub slot in a straight line with no J-shaped bend.')"
            />
          </span>
          <span class="spoke-head-step__explanation">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPullHint', 'The spoke exits the hub slot in a straight line with no J-shaped bend.') }}
          </span>
        </span>
      </button>
    </div>

    <SpokeHeadMeasurementGuide />

  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import SpokeHeadMeasurementGuide from '~/components/SpokeHeadMeasurementGuide.vue'
import SpokeStepNavigation from '~/components/SpokeStepNavigation.vue'

const { t } = useI18n()
const props = withDefaults(defineProps<{
  currentStep?: number
  modelValue?: 'j_bend' | 'straight_pull'
}>(), {
  currentStep: 1,
  modelValue: 'j_bend',
})

const emit = defineEmits<{
  'update:modelValue': [value: 'j_bend' | 'straight_pull']
  'select-step': [step: number]
  next: []
}>()

const selectedType = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
</script>

<style scoped>
.spoke-head-step {
  --head-step-border: rgba(15, 23, 42, 0.12);
  --head-step-text: #0f172a;
  --head-step-muted: #475569;
  --head-step-accent: #059669;
  width: 100%;
  padding: 20px;
  border: 1px dashed rgba(15, 23, 42, 0.16);
  border-radius: 24px;
  background:
    linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    #f8fafc;
  background-size: 28px 28px;
  color: var(--head-step-text);
}

.spoke-head-step__intro {
  margin-bottom: 14px;
  text-align: center;
}

.spoke-head-step__title {
  margin: 0;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-head-step__options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.spoke-head-step__option {
  display: grid;
  gap: 10px;
  min-width: 0;
  padding: 12px;
  border: 1px solid var(--head-step-border);
  border-radius: 20px;
  background: #ffffff;
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.spoke-head-step__option:hover {
  border-color: rgba(5, 150, 105, 0.48);
  box-shadow: 0 10px 22px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

.spoke-head-step__option:focus-visible {
  outline: 2px solid var(--head-step-accent);
  outline-offset: 3px;
}

.spoke-head-step__option--selected {
  border-color: var(--head-step-accent);
  box-shadow: inset 0 0 0 1px var(--head-step-accent), 0 10px 22px rgba(5, 150, 105, 0.1);
}

.spoke-head-step__option-head {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.spoke-head-step__number,
.spoke-head-step__selected-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  flex: 0 0 auto;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 900;
}

.spoke-head-step__number {
  background: rgba(15, 23, 42, 0.06);
  color: #334155;
}

.spoke-head-step__option--selected .spoke-head-step__number {
  background: var(--head-step-accent);
  color: #ffffff;
}

.spoke-head-step__option-title {
  min-width: 0;
  color: var(--head-step-text);
  font-size: 14px;
  font-weight: 800;
}

.spoke-head-step__selected-mark {
  margin-left: auto;
  background: var(--head-step-accent);
  color: #ffffff;
}

.spoke-head-step__option-content {
  display: grid;
  grid-template-columns: minmax(0, 1.55fr) minmax(120px, 0.7fr);
  gap: 12px;
  align-items: center;
  min-width: 0;
}

.spoke-head-step__diagram {
  display: block;
  overflow: hidden;
  max-width: 600px;
  margin-inline: auto;
  aspect-ratio: 2 / 1;
  border: 1px solid rgba(15, 23, 42, 0.1);
  border-radius: 16px;
  background-color: #ffffff;
}

.spoke-head-step__diagram img {
  display: block;
  width: 100%;
  height: 100%;
  aspect-ratio: 2 / 1;
  object-fit: contain;
}

.spoke-head-step__explanation {
  align-self: stretch;
  display: flex;
  align-items: center;
  padding: 12px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 14px;
  background: #f8fafc;
  color: var(--head-step-muted);
  font-size: 12px;
  line-height: 1.55;
}

@media (max-width: 767px) {
  .spoke-head-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-head-step__intro {
    margin-bottom: 12px;
  }

  .spoke-head-step__options {
    grid-template-columns: 1fr;
  }

  .spoke-head-step__option-content {
    grid-template-columns: 1fr;
  }
}
</style>
