<template>
  <aside class="brand-gallery-lightbox__details">
    <p v-if="gallery.description" class="brand-gallery-lightbox__description">
      {{ gallery.description }}
    </p>

    <p
      v-if="gallery.galleryDetailsLoading"
      class="brand-gallery-lightbox__loading"
    >
      {{ labels.loadingDetails }}
    </p>

    <div class="brand-gallery-lightbox__products">
      <h3 class="brand-gallery-lightbox__section-title">
        {{ labels.relatedProducts }}
      </h3>

      <div v-if="productLinks.length" class="space-y-2">
        <NuxtLink
          v-for="product in productLinks"
          :key="product.product_id"
          :to="productLinkPath(product)"
          class="brand-gallery-lightbox__product"
        >
          <span class="block truncate font-semibold">
            {{ product.name || product.slug }}
          </span>
          <span class="mt-0.5 block truncate font-mono text-[10px] text-slate-500">
            {{ product.slug }}
          </span>
        </NuxtLink>
      </div>
      <p v-else class="brand-gallery-lightbox__muted">
        {{ labels.noRelatedProducts }}
      </p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath } from '#imports'
import type {
  BrandGalleryPhoto,
  PictureWarehouseProductLink,
} from '~/types/brandGalleryPhotos'
import type { GalleryLightboxLabels } from '~/types/brandGalleryLightbox'

const props = defineProps<{
  gallery: BrandGalleryPhoto
  labels: GalleryLightboxLabels
}>()

const localePath = useLocalePath()
const productLinks = computed<PictureWarehouseProductLink[]>(() =>
  (props.gallery.productLinks || []).filter((product) => Boolean(product.slug || product.name)),
)

const productLinkPath = (product: PictureWarehouseProductLink): string => {
  const slug = String(product.slug || '').trim()
  return slug ? localePath(`/shop/${slug}`) : localePath('/shop')
}
</script>

<style scoped>
.brand-gallery-lightbox__details {
  min-width: 0;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1rem;
}

.brand-gallery-lightbox__description {
  margin: 0;
  color: rgba(226, 232, 240, 0.72);
  font-size: 0.75rem;
  line-height: 1.55;
  overflow-wrap: anywhere;
}

.brand-gallery-lightbox__loading {
  margin: 0.75rem 0 0;
  color: rgba(181, 255, 109, 0.76);
  font-size: 0.6875rem;
}

.brand-gallery-lightbox__products {
  margin-top: 1.25rem;
}

.brand-gallery-lightbox__section-title {
  margin: 0 0 0.65rem;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.3;
}

.brand-gallery-lightbox__product {
  display: block;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.04);
  padding: 0.6rem 0.7rem;
  color: #e2e8f0;
  font-size: 0.75rem;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.brand-gallery-lightbox__product:hover,
.brand-gallery-lightbox__product:focus-visible {
  border-color: rgba(181, 255, 109, 0.45);
  background: rgba(181, 255, 109, 0.08);
  color: #b5ff6d;
}

.brand-gallery-lightbox__muted {
  margin: 0;
  color: rgba(226, 232, 240, 0.5);
  font-size: 0.75rem;
}

@media (max-width: 767px) {
  .brand-gallery-lightbox__details {
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    border-left: 0;
    padding: 0.75rem;
  }

  .brand-gallery-lightbox__products {
    margin-top: 1rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .brand-gallery-lightbox__product {
    transition: none;
  }
}
</style>
