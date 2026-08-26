<template>
  <section
    v-if="groups.length"
    class="product-specs"
    aria-label="Product specifications"
  >
    <h2>Specifications</h2>
    <div v-for="group in groups" :key="group.name" class="spec-group">
      <h3 v-if="shouldShowSpecGroupHeading(group.name, groups.length)">{{ group.name }}</h3>
      <dl>
        <div v-for="item in group.items" :key="item.slug" class="spec-pill">
          <dt>{{ item.name }}</dt>
          <dd>{{ item.displayValue }}</dd>
        </div>
      </dl>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ProductSpecificationGroup } from '~/types/productDetail'

defineProps<{
  groups: ProductSpecificationGroup[]
}>()

const genericSpecGroupNames = new Set(['specification', 'specifications', 'specs', '规格', '規格'])

const shouldShowSpecGroupHeading = (name: string, groupCount: number) => {
  const normalizedName = String(name || '').trim().toLowerCase()
  if (!normalizedName || genericSpecGroupNames.has(normalizedName)) return false
  return groupCount > 1
}
</script>

<style scoped>
.product-specs h2 {
  margin-bottom: 0.75rem;
  color: var(--tz-text-primary);
  font-size: 1.5rem;
}

.product-specs {
  border-radius: 1.25rem;
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-card);
  padding: 1.25rem;
}

.spec-group + .spec-group {
  margin-top: 1.25rem;
}

.spec-group h3 {
  margin-bottom: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0;
}

.spec-group dl {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.55rem;
}

.spec-pill {
  display: flex;
  height: var(--product-control-pill-height);
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: var(--product-control-pill-radius);
  background: var(--tz-surface-subtle);
  padding: 0 0.78rem;
}

.spec-pill dt {
  min-width: 0;
  overflow: hidden;
  color: var(--tz-text-muted);
  font-size: 0.74rem;
  font-weight: 700;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spec-pill dd {
  min-width: 0;
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 600;
  text-align: right;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
