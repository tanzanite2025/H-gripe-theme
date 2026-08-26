<template>
  <section v-if="product" class="product-page" :aria-label="metaTitle">
    <nav
      v-if="productBreadcrumbItems.length"
      class="product-breadcrumb"
      aria-label="Breadcrumb"
    >
      <ol class="product-breadcrumb__list">
        <li
          v-for="(item, index) in productBreadcrumbItems"
          :key="`${item.type}-${item.id || item.path || index}`"
          class="product-breadcrumb__item"
        >
          <span v-if="index > 0" class="product-breadcrumb__separator" aria-hidden="true">/</span>
          <NuxtLink
            v-if="item.path && index < productBreadcrumbItems.length - 1"
            :to="item.path"
            class="product-breadcrumb__link"
          >
            {{ item.name }}
          </NuxtLink>
          <span v-else class="product-breadcrumb__current" aria-current="page">
            {{ item.name }}
          </span>
        </li>
      </ol>
    </nav>

    <div class="product-hero">
      <ProductDetailMediaGallery
        :items="productGalleryItems"
        :slots="productMediaSlots"
        :selected-media-id="selectedMediaId"
        :preview-media="previewMedia"
        :slots-overflowing="productMediaSlotsOverflowing"
        @select="selectMedia"
        @previous="selectPreviousMedia"
        @next="selectNextMedia"
      />

      <div class="product-summary">
        <h1 class="product-title">{{ product.name }}</h1>
        <SafeRichText
          v-if="product.short_description"
          class="product-description"
          :html="product.short_description"
        />
        <p v-else-if="productSummaryDescription" class="product-description">
          {{ productSummaryDescription }}
        </p>

        <div class="product-meta" aria-live="polite" aria-atomic="true">
          <span v-if="formattedPrice" class="product-price">{{ formattedPrice }}</span>
          <span
            v-if="product.product_specification_template?.name"
            class="product-specification-template-pill"
          >
            {{ product.product_specification_template.name }}
          </span>
        </div>

        <ProductDetailVariantSelector
          :variant-choices="variantChoices"
          :variant-option-groups="variantOptionGroups"
          :selected-variant-id="selectedVariantId"
          :selected-variant-weight="selectedVariantWeight"
          @select-option="({ slug, value }) => selectVariantOption(slug, value)"
          @update:selected-variant-id="selectedVariantId = $event"
        />

        <ProductDetailQuantityActions
          :quantity="selectedQuantity"
          :max-quantity="maxProductQuantity"
          :can-add-to-cart="canAddToCart"
          :can-buy-now="canBuyNow"
          :buy-now-unavailable-label="productBuyNowUnavailableLabel"
          :quantity-label="t('products.detail.quantity', 'Quantity')"
          :decrease-quantity-label="t('products.detail.decreaseQuantity', 'Decrease quantity')"
          :increase-quantity-label="t('products.detail.increaseQuantity', 'Increase quantity')"
          :buy-now-label="t('checkout.product.buyNow', 'Buy now')"
          @decrease="decreaseSelectedQuantity"
          @increase="increaseSelectedQuantity"
          @input="setSelectedQuantity"
          @add-to-cart="addSelectedToCart"
          @buy-now="checkoutSelectedWithPayment"
        />

        <ProductDetailPaymentSelector
          :options="productPaymentOptions"
          :selected-method="selectedProductPaymentMethod"
          :loading="paymentMethodsLoading"
          :error="paymentMethodsError"
          @update:selected-method="selectProductPaymentMethod"
        />

        <div v-if="shouldPrepareStripeExpressCheckout">
          <ProductDetailExpressCheckout
            ref="stripeExpressCheckoutElementRef"
            :publishable-key="stripeExpressCheckoutPublishableKey"
            :amount="stripeExpressCheckoutAmount"
            :currency="cartCurrency"
            :line-items="stripeExpressCheckoutLineItems"
            :allowed-shipping-countries="stripeExpressCheckoutAllowedShippingCountries"
            :disabled="isStripeExpressCheckoutProcessing"
            :error="stripeExpressCheckoutError"
            @ready="handleStripeExpressCheckoutAvailability"
            @available-payment-methods-change="handleStripeExpressCheckoutAvailability"
            @confirm="handleStripeExpressCheckoutConfirm"
            @shipping-address-change="handleStripeExpressCheckoutShippingAddressChange"
            @shipping-rate-change="handleStripeExpressCheckoutShippingRateChange"
            @error="handleStripeExpressCheckoutError"
          />
        </div>
      </div>
    </div>

    <ProductDetailSpecifications :groups="specGroups" />

    <div key="product-information-tabs" class="product-tabs-anchor">
      <ProductDetailInformationTabs
        :key="product.id"
        :details-html="product.description"
        :after-sales-html="product.after_sales_template?.content"
        :packaging-html="product.packaging_template?.content"
      />
    </div>

    <ProductReviewsSection
      :product-id="product.id"
      :initial-summary="shopProduct?.reviewSummary"
    />

    <ProductRecommendations
      :key="`product-recommendations-${product.id}`"
      surface="product_detail_bottom"
      :title="t('recommendations.productDetailTitle', 'You may also like')"
      :product-id="product.id"
      :category-id="product.product_specification_template_id || product.product_specification_template?.id || null"
      :exclude-product-ids="[product.id]"
      :limit="6"
    />
  </section>
  <section v-else-if="pending" class="product-page product-page--pending">Loading...</section>
  <section v-else class="product-page product-page--error" role="alert">Product not found.</section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { useProductDetailData } from '~/composables/useProductDetailData'
import { useProductDetailMedia } from '~/composables/useProductDetailMedia'
import { useProductDetailPurchase } from '~/composables/useProductDetailPurchase'
import { useProductDetailPresentation } from '~/composables/useProductDetailPresentation'
import { useProductDetailTracking } from '~/composables/useProductDetailTracking'
import { useProductDetailVariants } from '~/composables/useProductDetailVariants'
import { useProductDetailSeo } from '~/composables/seo/useProductDetailSeo'
import ProductDetailExpressCheckout from '~/components/shop/product-detail/ProductDetailExpressCheckout.vue'
import ProductDetailMediaGallery from '~/components/shop/product-detail/ProductDetailMediaGallery.vue'
import ProductDetailPaymentSelector from '~/components/shop/product-detail/ProductDetailPaymentSelector.vue'
import ProductDetailQuantityActions from '~/components/shop/product-detail/ProductDetailQuantityActions.vue'
import ProductDetailSpecifications from '~/components/shop/product-detail/ProductDetailSpecifications.vue'
import ProductDetailVariantSelector from '~/components/shop/product-detail/ProductDetailVariantSelector.vue'
import ProductRecommendations from '~/components/shop/ProductRecommendations.vue'
import ProductReviewsSection from '~/components/shop/ProductReviewsSection.vue'

const { t } = useI18n()
const {
  slug,
  product,
  pending,
  shopProduct,
  metaTitle,
  productSummaryDescription,
  productImages,
  displayCurrency,
} = await useProductDetailData()

const {
  selectedVariantId,
  activeVariants,
  selectedVariant,
  selectedVariantWeight,
  selectedCartTitle,
  variantOptionDefinitions,
  variantOptionGroups,
  variantChoices,
  variantLabel,
  selectVariantOption,
  effectivePrice,
  currentCurrency,
  currentDisplayPrice,
  selectedAvailability,
  formattedPrice,
} = useProductDetailVariants(product, displayCurrency)

const {
  selectedMediaId,
  productGalleryItems,
  productMediaSlots,
  productMediaSlotsOverflowing,
  previewMedia,
  primaryMediaThumbnail,
  selectMedia,
  selectPreviousMedia,
  selectNextMedia,
} = useProductDetailMedia({
  product,
  selectedVariantId,
  metaTitle,
})

const {
  maxProductQuantity,
  selectedQuantity,
  setSelectedQuantity,
  decreaseSelectedQuantity,
  increaseSelectedQuantity,
  canAddToCart,
  canBuyNow,
  productBuyNowUnavailableLabel,
  selectedProductPaymentMethod,
  productPaymentOptions,
  paymentMethodsLoading,
  paymentMethodsError,
  selectProductPaymentMethod,
  stripeExpressCheckoutPublishableKey,
  stripeExpressCheckoutError,
  isStripeExpressCheckoutProcessing,
  stripeExpressCheckoutElementRef,
  stripeExpressCheckoutAmount,
  stripeExpressCheckoutLineItems,
  stripeExpressCheckoutAllowedShippingCountries,
  shouldPrepareStripeExpressCheckout,
  addSelectedToCart,
  checkoutSelectedWithPayment,
  handleStripeExpressCheckoutAvailability,
  handleStripeExpressCheckoutShippingAddressChange,
  handleStripeExpressCheckoutShippingRateChange,
  handleStripeExpressCheckoutConfirm,
  handleStripeExpressCheckoutError,
  cartCurrency,
} = useProductDetailPurchase({
  product,
  shopProduct,
  selectedVariant,
  selectedVariantWeight,
  selectedCartTitle,
  effectivePrice,
  currentCurrency,
  selectedAvailability,
  primaryMediaThumbnail,
})

const {
  specGroups,
  productBreadcrumbItems,
} = useProductDetailPresentation({ product })

useProductDetailSeo({
  product,
  slug,
  productImages,
  activeVariants,
  selectedVariant,
  selectedAvailability,
  currentDisplayPrice,
  variantOptionDefinitions,
  variantLabel,
  displayCurrency,
  breadcrumbItems: productBreadcrumbItems,
})

useProductDetailTracking({
  product,
  formattedPrice,
  primaryMediaThumbnail,
})
</script>

<style scoped>
.product-page {
  --product-control-pill-height: 2.125rem;
  --product-control-pill-radius: 999px;
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
  color: var(--tz-text-primary);
  padding: 2rem 1rem 4rem;
}

.product-breadcrumb {
  overflow-x: auto;
}

.product-breadcrumb__list {
  display: flex;
  min-width: max-content;
  align-items: center;
  gap: 0.45rem;
  margin: 0;
  padding: 0;
  list-style: none;
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
}

.product-breadcrumb__item {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.product-breadcrumb__separator {
  color: var(--tz-text-muted);
}

.product-breadcrumb__link {
  color: inherit;
  text-decoration: none;
}

.product-breadcrumb__link:hover,
.product-breadcrumb__link:focus-visible {
  color: var(--tz-site-accent);
}

.product-breadcrumb__current {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.product-page--pending,
.product-page--error {
  padding: 4rem 1rem;
  color: var(--tz-text-secondary);
  text-align: center;
  font-size: 1.1rem;
}

.product-hero {
  display: grid;
  gap: 2rem;
  align-items: start;
}

.product-hero > * {
  min-width: 0;
}

@media (min-width: 900px) {
  .product-hero {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: clamp(2rem, 4vw, 4rem);
  }
}

.product-summary {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  min-width: 0;
  width: 100%;
}

.product-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.6rem, 2vw + 0.9rem, 2.4rem);
  font-weight: 600;
}

.product-description {
  color: var(--tz-text-secondary);
  line-height: 1.65;
}

.product-description :deep(p) {
  margin-bottom: 0.5rem;
}

.product-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  color: var(--tz-text-secondary);
  font-size: 1rem;
}

.product-price {
  color: #059669;
  font-weight: 600;
  font-size: 1.15rem;
}

.product-specification-template-pill {
  display: inline-flex;
  height: var(--product-control-pill-height);
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: var(--product-control-pill-radius);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 0.72rem;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .product-page {
    padding-inline: 1rem;
  }
}
</style>
