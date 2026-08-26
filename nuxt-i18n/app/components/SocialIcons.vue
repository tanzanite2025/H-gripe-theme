<template>
  <ul v-if="itemsToShow.length" class="flex list-none justify-center gap-3 m-0 p-0">
    <li v-for="item in itemsToShow" :key="`${item.network}:${item.url}`">
      <a
        class="inline-flex size-8 items-center justify-center rounded-full tz-surface-subtle tz-text-primary transition-[background,transform] duration-200 ease-in-out hover:tz-surface-subtle hover:-translate-y-px focus-visible:tz-surface-subtle focus-visible:-translate-y-px"
        :href="item.url"
        :aria-label="item.label"
        target="_blank"
        rel="noopener"
      >
        <span class="inline-flex leading-[0]" aria-hidden="true">
          <slot name="icon" :item="item">
            <img
              :src="socialIconSrc(item.network)"
              alt=""
              width="24"
              height="24"
              aria-hidden="true"
              class="block size-6 object-contain"
              loading="lazy"
              decoding="async"
            />
          </slot>
        </span>
      </a>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSocialLinks } from '~/composables/useSocialLinks'
import { normalizeSocialLinkItems } from '~/utils/socialLinks'

export interface SocialItem {
  url: string
  label: string
  network?: string
}

const props = defineProps<{ items?: SocialItem[] }>()
const { socialLinks } = useSocialLinks()

const socialIconSrc = (name?: string) => `/icons/social/${String(name || '').toLowerCase().trim()}.svg`

const itemsToShow = computed(() => {
  const items = props.items && props.items.length ? props.items : socialLinks.value
  return normalizeSocialLinkItems(items)
})
</script>
