<template>
  <div v-if="variantChoices.length" class="product-purchase-panel">
    <div v-if="variantOptionGroups.length" class="variant-option-groups">
      <fieldset
        v-for="group in variantOptionGroups"
        :key="group.slug"
        class="variant-option-group"
      >
        <legend>{{ group.name }}</legend>
        <div class="variant-option-buttons">
          <button
            v-for="option in group.options"
            :key="`${group.slug}-${option.value}`"
            type="button"
            class="variant-option-button"
            :class="{
              'variant-option-button--selected': option.selected,
              'variant-option-button--out': !option.available,
              'variant-option-button--visual': group.presentation === 'color' || group.presentation === 'image',
            }"
            :aria-label="`${group.name}: ${option.label}`"
            :aria-pressed="option.selected"
            @click="emit('select-option', { slug: group.slug, value: option.value })"
          >
            <span
              v-if="group.presentation === 'color' || group.presentation === 'image'"
              class="variant-option-swatch"
              :class="{ 'variant-option-swatch--image': Boolean(option.swatchUrl) || group.presentation === 'image' }"
              :style="option.swatchUrl ? undefined : option.colorHex ? { backgroundColor: option.colorHex } : undefined"
              aria-hidden="true"
            >
              <StorefrontImage v-if="option.swatchUrl" :src="option.swatchUrl" :alt="option.label" preset="swatch" />
            </span>
            <span class="variant-option-button__label">{{ option.label }}</span>
            <small v-if="!option.available" class="variant-option-button__status">Out</small>
          </button>
        </div>
      </fieldset>
    </div>
    <div v-else-if="variantChoices.length > 1" class="product-variants">
      <label for="variant-select">Choose option</label>
      <select
        id="variant-select"
        :value="selectedVariantId ?? undefined"
        @change="handleVariantChange"
      >
        <option
          v-for="variant in variantChoices"
          :key="variant.id"
          :value="variant.id"
        >
          {{ variant.label }}
        </option>
      </select>
    </div>

    <dl v-if="selectedVariantWeight" class="selected-sku-facts" aria-live="polite" aria-atomic="true">
      <div class="selected-sku-fact-pill">
        <dt>Weight</dt>
        <dd>{{ selectedVariantWeight }}g</dd>
      </div>
    </dl>
  </div>
</template>

<script setup lang="ts">
import type { ProductVariantOptionGroup } from '~/types/productDetail'

interface VariantChoice {
  id: number
  label: string
}

defineProps<{
  variantChoices: VariantChoice[]
  variantOptionGroups: ProductVariantOptionGroup[]
  selectedVariantId: number | null
  selectedVariantWeight: number | null
}>()

const emit = defineEmits<{
  (event: 'select-option', payload: { slug: string; value: string }): void
  (event: 'update:selectedVariantId', value: number | null): void
}>()

const handleVariantChange = (event: Event) => {
  const value = Number((event.target as HTMLSelectElement).value)
  emit('update:selectedVariantId', Number.isFinite(value) && value > 0 ? value : null)
}
</script>

<style scoped>
.product-purchase-panel {
  display: grid;
  gap: 1.15rem;
  max-width: none;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-surface-card);
  padding: 1.15rem;
}

.variant-option-groups {
  display: grid;
  gap: 1rem;
}

.variant-option-group {
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.product-variants {
  display: grid;
  gap: 0.5rem;
  max-width: 100%;
}

.variant-option-group legend {
  margin-bottom: 0.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

.variant-option-buttons {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(6.5rem, 6.5rem));
  gap: 0.45rem;
}

.variant-option-button {
  display: inline-flex;
  width: 100%;
  height: var(--product-control-pill-height);
  min-height: 0;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: var(--product-control-pill-radius);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
  cursor: pointer;
  font: inherit;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.2;
  box-sizing: border-box;
  padding: 0 0.7rem;
  text-align: center;
  transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
  overflow-wrap: anywhere;
}

.variant-option-button__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.variant-option-button__status {
  display: inline-flex;
  height: 1rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: var(--tz-status-danger-text);
  font-size: 0.66rem;
  font-weight: 800;
  line-height: 1;
}

.variant-option-button--visual {
  min-width: 0;
}

.variant-option-swatch {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.35rem;
  background: var(--tz-surface-muted);
}

.variant-option-swatch--image {
  background: var(--tz-surface-subtle);
}

.variant-option-swatch img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.variant-option-button:hover {
  border-color: rgba(5, 150, 105, 0.65);
  background: rgba(5, 150, 105, 0.08);
  transform: translateY(-1px);
}

.variant-option-button--selected {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: #ffffff;
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.16);
}

.variant-option-button--out:not(.variant-option-button--selected) {
  color: var(--tz-text-muted);
}

.product-variants label {
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0;
}

.product-variants select {
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.8rem;
  background: var(--tz-form-control-surface);
  color-scheme: light;
  color: var(--tz-text-primary);
  padding: 0.7rem 0.9rem;
}

.product-variants option {
  background: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
}

.selected-sku-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin: 0;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 0.75rem;
}

.selected-sku-fact-pill {
  display: inline-flex;
  height: var(--product-control-pill-height);
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: var(--product-control-pill-radius);
  background: var(--tz-surface-subtle);
  padding: 0 0.78rem;
}

.selected-sku-facts dt {
  color: var(--tz-text-muted);
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.selected-sku-facts dd {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.86rem;
  font-weight: 700;
  line-height: 1;
  overflow-wrap: anywhere;
}

.variant-option-button:focus-visible,
.product-variants select:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 3px;
}
</style>
