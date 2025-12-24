<template>
  <section class="popular-searches" aria-label="Popular searches">
    <h3 class="popular-searches__title">Popular searches</h3>
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
  margin-top: 8px;
  margin-bottom: 8px;
}

.popular-searches__title {
  margin: 0 0 6px;
  font-size: 0.78rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 0.9);
  text-align: left;
}

.popular-searches__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.popular-searches__chip {
  position: relative;
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  padding: 0.35rem 0.65rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, rgba(15,23,42,0.92), rgba(15,23,42,0.72));
  border: none;
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.55);
  font-size: 0.875rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  transition: background-color 0.2s, color 0.2s, border-color 0.2s, transform 0.08s ease;
}

.popular-searches__chip--active {
  background: rgba(255, 255, 255, 0.86);
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.7);
  color: rgba(0, 0, 0, 0.92);
}
</style>
