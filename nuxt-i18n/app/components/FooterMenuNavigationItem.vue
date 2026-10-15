<template>
  <li class="footer-menu-navigation-item">
    <NuxtLink
      v-if="item.to && !item.external"
      class="footer-menu-navigation-item__link"
      :to="localePath(item.to)"
    >
      {{ footerNavigationItemDisplayLabel }}
    </NuxtLink>
    <a
      v-else-if="item.to && item.external"
      class="footer-menu-navigation-item__link"
      :href="item.to"
      target="_blank"
      rel="noopener noreferrer"
    >
      {{ footerNavigationItemDisplayLabel }}
    </a>
    <span
      v-else
      class="footer-menu-navigation-item__group-label"
    >
      {{ footerNavigationItemDisplayLabel }}
    </span>

    <ul
      v-if="item.children?.length"
      class="footer-menu-navigation-item__children"
    >
      <FooterMenuNavigationItem
        v-for="(child, childIndex) in item.children"
        :key="getFooterMenuNavigationItemKey(child, childIndex)"
        :item="child"
      />
    </ul>
  </li>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import type { FooterNavigationItem } from '~/utils/footerMenus'

defineOptions({ name: 'FooterMenuNavigationItem' })

const props = defineProps<{
  item: FooterNavigationItem
}>()

const localePath = useLocalePath()
const { t, te } = useI18n()
const footerNavigationItemDisplayLabel = computed(() => {
  const { labelKey, fallback } = props.item
  return labelKey && te(labelKey) ? t(labelKey) : fallback || labelKey || ''
})

const getFooterMenuNavigationItemKey = (
  item: FooterNavigationItem,
  index: number,
) => `${item.to || item.labelKey || item.fallback || 'group'}-${index}`
</script>

<style scoped>
.footer-menu-navigation-item {
  min-width: 0;
  margin: 0;
  padding: 0;
  list-style: none;
}

.footer-menu-navigation-item__link,
.footer-menu-navigation-item__group-label {
  font-size: 0.9rem;
  font-weight: 500;
  line-height: 1.45;
  color: var(--tz-text-secondary);
  text-decoration: none;
  display: inline-block;
  max-width: 100%;
  transition: color 0.2s ease, transform 0.2s ease;
}

.footer-menu-navigation-item__group-label {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.footer-menu-navigation-item__link:hover,
.footer-menu-navigation-item__link:focus-visible {
  color: var(--tz-text-primary);
  transform: translateX(4px);
  text-shadow: none;
}

.footer-menu-navigation-item__children {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  margin: 0.55rem 0 0;
  padding: 0 0 0 0.85rem;
  border-left: 1px solid var(--tz-border-subtle);
  list-style: none;
}

@media (max-width: 768px) {
  .footer-menu-navigation-item__link,
  .footer-menu-navigation-item__group-label {
    display: inline-flex;
    align-items: center;
    min-height: 2.25rem;
    padding-block: 0.2rem;
    line-height: 1.45;
  }

  .footer-menu-navigation-item__children {
    gap: 0.25rem;
    margin-top: 0.3rem;
    padding-left: 0.75rem;
  }

}
</style>
