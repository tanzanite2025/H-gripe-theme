<template>
  <section id="home-main-product-categories" class="bg-transparent py-8 text-white sm:py-10 lg:py-14">
    <div class="page-content-shell px-0 md:px-6">
      <div class="flex flex-col gap-4 sm:gap-5 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-3xl">
          <span class="home-main-product-categories__eyebrow">
            <Icon name="lucide:star" class="h-3.5 w-3.5" aria-hidden="true" />
            {{ mainProductsCopy.eyebrow }}
          </span>
          <h2 class="mt-3 text-2xl font-semibold leading-tight text-white sm:text-3xl">
            {{ mainProductsCopy.heading }}
          </h2>
        </div>

        <NuxtLink
          :to="localePath('/shop')"
          class="premium-button premium-button--active w-full justify-center sm:w-auto"
        >
          <Icon name="lucide:shopping-bag" class="mr-2 h-4 w-4" aria-hidden="true" />
          {{ mainProductsCopy.shopAll }}
        </NuxtLink>
      </div>

      <StorefrontDataNotice
        v-if="homeMainProductCategoriesIsLocaleFallback && items.length"
        class="mt-5"
        :title="mainProductsCopy.localeFallbackTitle"
        :description="mainProductsCopy.localeFallbackDescription"
      />

      <StorefrontDataNotice
        v-if="homeMainProductCategoriesPending && !items.length"
        class="mt-5"
        tone="empty"
        :title="mainProductsCopy.loadingTitle"
        :description="mainProductsCopy.loadingDescription"
      />

      <StorefrontDataNotice
        v-else-if="homeMainProductCategoriesError"
        class="mt-5"
        tone="error"
        role="alert"
        :title="mainProductsCopy.errorTitle"
        :description="mainProductsCopy.errorDescription"
      >
        <template #actions>
          <button
            type="button"
            class="storefront-data-notice-action"
            :disabled="homeMainProductCategoriesPending"
            @click="retryHomeMainProductCategories"
          >
            <Icon name="lucide:refresh-cw" aria-hidden="true" />
            {{ t('common.retry') }}
          </button>
        </template>
      </StorefrontDataNotice>

      <div v-else-if="items.length" class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <NuxtLink
          v-for="item in items"
          :key="item.id"
          :to="targetFor(item)"
          :external="isExternalTarget(item.targetUrl)"
          :target="isExternalTarget(item.targetUrl) ? '_blank' : undefined"
          :rel="isExternalTarget(item.targetUrl) ? 'noopener noreferrer' : undefined"
          class="home-main-product-categories__card premium-card group"
        >
          <div class="home-main-product-categories__image-wrap">
            <StorefrontImage
              :src="item.src"
              :alt="item.altText"
              :width="item.width"
              :height="item.height"
              preset="card"
              sizes="xs:100vw sm:50vw lg:33vw"
              class="home-main-product-categories__image"
            />
            <span class="home-main-product-categories__order" aria-hidden="true">
              {{ String(item.desktopOrder).padStart(2, '0') }}
            </span>
          </div>
          <div class="home-main-product-categories__content">
            <div class="min-w-0">
              <h3 class="home-main-product-categories__title">{{ item.title }}</h3>
              <p v-if="item.caption" class="home-main-product-categories__caption">{{ item.caption }}</p>
            </div>
            <span class="home-main-product-categories__action">
              {{ item.targetLabel || mainProductsCopy.cardAction }}
              <Icon name="lucide:arrow-up-right" class="h-4 w-4" aria-hidden="true" />
            </span>
          </div>
        </NuxtLink>
      </div>

      <StorefrontDataNotice
        v-else
        class="mt-5"
        :title="emptyStateTitle"
        :description="emptyStateCopy"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import StorefrontDataNotice from '~/components/StorefrontDataNotice.vue'
import type { HomeMainProductCategoryItem } from '~/types/homeMainProductCategories'
import { useHomeMainProductCategories } from '~/composables/useHomeMainProductCategories'

const { t } = useI18n()
const localePath = useLocalePath()
const {
  homeMainProductCategoryItems: items,
  homeMainProductCategoriesIsConfigured,
  homeMainProductCategoriesIsLocaleFallback,
  homeMainProductCategoriesPending,
  homeMainProductCategoriesError,
  refreshHomeMainProductCategories,
} = await useHomeMainProductCategories()

const mainProductsCopy = computed(() => ({
  eyebrow: t('home.mainProductCategories.eyebrow'),
  heading: t('home.mainProductCategories.heading'),
  shopAll: t('home.mainProductCategories.shopAll'),
  cardAction: t('home.mainProductCategories.cardAction'),
  localeFallbackTitle: t('home.mainProductCategories.localeFallbackTitle'),
  localeFallbackDescription: t('home.mainProductCategories.localeFallbackDescription'),
  loadingTitle: t('home.mainProductCategories.loadingTitle'),
  loadingDescription: t('home.mainProductCategories.loadingDescription'),
  errorTitle: t('home.mainProductCategories.errorTitle'),
  errorDescription: t('home.mainProductCategories.errorDescription'),
  emptyConfiguredTitle: t('home.mainProductCategories.emptyConfiguredTitle'),
  emptyConfiguredDescription: t('home.mainProductCategories.emptyConfiguredDescription'),
  emptyUnconfiguredTitle: t('home.mainProductCategories.emptyUnconfiguredTitle'),
  emptyUnconfiguredDescription: t('home.mainProductCategories.emptyUnconfiguredDescription'),
}))

const emptyStateTitle = computed(() => (
  homeMainProductCategoriesIsConfigured.value
    ? mainProductsCopy.value.emptyConfiguredTitle
    : mainProductsCopy.value.emptyUnconfiguredTitle
))

const emptyStateCopy = computed(() => (
  homeMainProductCategoriesIsConfigured.value
    ? mainProductsCopy.value.emptyConfiguredDescription
    : mainProductsCopy.value.emptyUnconfiguredDescription
))

const isExternalTarget = (value?: string): boolean => Boolean(value && /^(?:https?:)?\/\//i.test(value))

const targetFor = (item: HomeMainProductCategoryItem): string => {
  const target = String(item.targetUrl || '/shop').trim() || '/shop'
  if (isExternalTarget(target)) return target
  return localePath(target.startsWith('/') ? target : `/${target}`)
}

const retryHomeMainProductCategories = () => {
  void refreshHomeMainProductCategories()
}
</script>

<style scoped>
#home-main-product-categories {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.home-main-product-categories__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgba(181, 255, 109, 0.3);
  border-radius: 999px;
  padding: 0.25rem 0.75rem;
  background: rgba(181, 255, 109, 0.1);
  color: var(--tz-brand-primary);
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.16em;
  line-height: 1.25;
  text-transform: uppercase;
}

.home-main-product-categories__card {
  display: grid;
  min-width: 0;
  overflow: hidden;
  border-radius: 1rem;
  color: inherit;
  text-decoration: none;
  transition: transform 180ms ease;
}

.home-main-product-categories__card:hover {
  transform: translateY(-2px);
}

.home-main-product-categories__card:focus-visible {
  outline: 2px solid rgba(181, 255, 109, 0.92);
  outline-offset: 3px;
}

.home-main-product-categories__image-wrap {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: var(--tz-card-surface);
}

.home-main-product-categories__image {
  --tz-image-loading-surface: var(--tz-card-surface);
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 300ms ease;
}

.home-main-product-categories__card:hover .home-main-product-categories__image {
  transform: scale(1.035);
}

.home-main-product-categories__image-wrap::after {
  position: absolute;
  inset: 0;
  content: '';
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.02) 42%, rgba(0, 0, 0, 0.46));
  pointer-events: none;
}

.home-main-product-categories__order {
  position: absolute;
  z-index: 1;
  top: 0.75rem;
  left: 0.75rem;
  color: rgba(255, 255, 255, 0.82);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.16em;
}

.home-main-product-categories__content {
  display: flex;
  min-width: 0;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
}

.home-main-product-categories__title {
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-card-title);
  font-weight: 700;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-main-product-categories__caption {
  display: -webkit-box;
  margin-top: 0.35rem;
  overflow: hidden;
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-caption);
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.home-main-product-categories__action {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.3rem;
  color: var(--tz-brand-primary);
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

@media (max-width: 420px) {
  .home-main-product-categories__content {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.65rem;
  }

}

@media (prefers-reduced-motion: reduce) {
  .home-main-product-categories__card,
  .home-main-product-categories__image {
    transition: none;
  }
}
</style>
