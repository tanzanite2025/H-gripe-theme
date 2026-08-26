<template>
  <figure
    class="home-hero-visual-showcase-figure"
    :class="`home-hero-visual-showcase-figure--${item.layoutVariant}`"
  >
    <StorefrontImage
      :src="item.src"
      :alt="item.altText"
      class="home-hero-visual-showcase-figure__image h-full w-full object-cover"
      :sizes="sizes"
      densities="1x"
      format="webp"
      quality="82"
      :loading="loading"
      :fetchpriority="fetchpriority"
      decoding="async"
    />
    <figcaption
      v-if="captionVisibility !== 'hidden'"
      class="home-hero-visual-showcase-figure__caption"
      :class="{ 'home-hero-visual-showcase-figure__caption--sr-only': captionVisibility === 'sr-only' }"
    >
      <span class="home-hero-visual-showcase-figure__title">{{ item.title }}</span>
      <span v-if="item.caption" class="home-hero-visual-showcase-figure__description">{{ item.caption }}</span>
    </figcaption>
  </figure>
</template>

<script setup lang="ts">
import StorefrontImage from '~/components/StorefrontImage.vue'
import type { HomeHeroVisualShowcaseItem } from '~/types/homeHeroVisualShowcase'

withDefaults(defineProps<{
  item: HomeHeroVisualShowcaseItem
  sizes?: string
  loading?: 'eager' | 'lazy'
  fetchpriority?: 'high' | 'low' | 'auto'
  captionVisibility?: 'inline' | 'sr-only' | 'hidden'
}>(), {
  sizes: 'xs:100vw sm:50vw lg:22vw',
  loading: 'lazy',
  fetchpriority: 'low',
  captionVisibility: 'inline',
})
</script>

<style scoped>
.home-hero-visual-showcase-figure {
  position: relative;
  aspect-ratio: 3 / 4;
  min-width: 0;
  overflow: hidden;
  margin: 0;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 1rem;
  background: var(--tz-image-loading-surface);
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.35);
}

.home-hero-visual-showcase-figure::after {
  position: absolute;
  inset: 0;
  content: '';
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.02) 45%, rgba(0, 0, 0, 0.28));
  pointer-events: none;
}

.home-hero-visual-showcase-figure__image {
  display: block;
}

.home-hero-visual-showcase-figure__caption {
  position: absolute;
  z-index: 1;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  min-height: 2.8rem;
  flex-direction: column;
  justify-content: center;
  gap: 0.1rem;
  padding: 0.45rem 0.65rem 0.5rem;
  background: rgba(255, 255, 255, 0.96);
  color: #111318;
}

.home-hero-visual-showcase-figure__caption--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.home-hero-visual-showcase-figure__title,
.home-hero-visual-showcase-figure__description {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-hero-visual-showcase-figure__title {
  font-size: 0.72rem;
  font-weight: 800;
  line-height: 1.1;
}

.home-hero-visual-showcase-figure__description {
  font-size: 0.6rem;
  line-height: 1.2;
  opacity: 0.72;
}

@media (max-width: 639px) {
  .home-hero-visual-showcase-figure {
    border-radius: 0.75rem;
  }

  .home-hero-visual-showcase-figure__caption {
    min-height: 3rem;
    padding-inline: 0.5rem;
  }

  .home-hero-visual-showcase-figure__title {
    font-size: 0.65rem;
  }

  .home-hero-visual-showcase-figure__description {
    font-size: 0.54rem;
  }
}
</style>
