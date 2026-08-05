<template>
  <section class="popular-searches" :aria-label="t('search.popularAriaLabel')">
    <h3 class="popular-searches__title">{{ t('search.popularTitle') }}</h3>
    <div class="popular-searches__chips">
      <button
        v-for="keyword in keywords"
        :key="keyword"
        type="button"
        class="popular-searches__chip"
        :class="{ 'popular-searches__chip--active': isSelected(keyword) }"
        @click="toggle(keyword)"
      >
        <span class="popular-searches__chip-label">{{ keyword }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
const { t } = useI18n()

const props = defineProps<{
  keywords: string[]
  modelValue?: string[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const isSelected = (keyword: string) => {
  return Array.isArray(props.modelValue) && props.modelValue.includes(keyword)
}

const toggle = (keyword: string) => {
  const current = Array.isArray(props.modelValue) ? [...props.modelValue] : []
  const index = current.indexOf(keyword)
  if (index === -1) {
    current.push(keyword)
  } else {
    current.splice(index, 1)
  }
  emit('update:modelValue', current)
}
</script>

<style scoped>
.popular-searches {
  margin: 0;
}

.popular-searches__title {
  margin: 0 0 8px;
  font-size: 0.74rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--tz-text-secondary);
  text-align: left;
}

.popular-searches__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.popular-searches__chip {
  position: relative;
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  padding: 0.32rem 0.62rem;
  border-radius: 9999px;
  background: #070707;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: none;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  transition: background-color 0.2s, color 0.2s, border-color 0.2s, transform 0.08s ease;
}

.popular-searches__chip--active {
  background: #050505;
  border-color: rgba(181, 255, 109, 0.62);
  box-shadow: none;
  color: #ffffff;
}

@media (min-width: 769px) {
  .popular-searches {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    column-gap: 14px;
  }

  .popular-searches__title {
    margin: 0;
    font-size: 0.68rem;
    white-space: nowrap;
  }

  .popular-searches__chips {
    gap: 6px;
  }
}
</style>
