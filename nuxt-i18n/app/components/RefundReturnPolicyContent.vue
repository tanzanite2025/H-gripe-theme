<template>
  <article class="refund-return-policy-content" :class="{ 'refund-return-policy-content--compact': compact }">
    <header v-if="canShowPolicy && (showTitle || policy.intro)" class="refund-return-policy-content__header">
      <h2 v-if="showTitle" class="refund-return-policy-content__title">{{ policy.title }}</h2>
      <p v-if="policy.intro" class="refund-return-policy-content__intro">
        {{ policy.intro }}
        <template v-if="contactEmail">
          (<a :href="contactEmailHref" class="refund-return-policy-content__link">{{ contactEmail }}</a>)
        </template>
      </p>
    </header>

    <StorefrontDataNotice
      v-if="policyFallbackNotice"
      class="refund-return-policy-content__notice"
      tone="fallback"
      :title="policyFallbackNotice.title"
      :description="policyFallbackNotice.description"
    />

    <div v-if="loadError" class="refund-return-policy-content__state">
      <StorefrontDataNotice
        tone="error"
        role="alert"
        :title="t('storefrontDataNotice.refundReturnPolicy.error.title')"
        :description="t('storefrontDataNotice.refundReturnPolicy.error.description')"
      >
        <template #actions>
          <button
            type="button"
            class="storefront-data-notice-action"
            :disabled="pending"
            @click="retryPolicy"
          >
            <Icon name="lucide:refresh-cw" aria-hidden="true" />
            {{ t('common.retry') }}
          </button>
        </template>
      </StorefrontDataNotice>
    </div>

    <div v-else-if="pending && !hasResolvedPolicy" class="refund-return-policy-content__state">
      <StorefrontDataNotice
        tone="empty"
        :title="t('storefrontDataNotice.refundReturnPolicy.loading.title')"
        :description="t('storefrontDataNotice.refundReturnPolicy.loading.description')"
      />
    </div>

    <div v-else-if="!hasContent" class="refund-return-policy-content__state">
      <StorefrontDataNotice
        tone="empty"
        :title="t('storefrontDataNotice.refundReturnPolicy.empty.title')"
        :description="t('storefrontDataNotice.refundReturnPolicy.empty.description')"
      />
    </div>

    <div v-else class="refund-return-policy-content__sections">
      <section
        v-for="(section, index) in policy.sections"
        :id="sectionAnchor(section.id, index)"
        :key="`${section.id}-${index}`"
        class="refund-return-policy-content__section"
      >
        <div class="refund-return-policy-content__section-copy">
          <h3 class="refund-return-policy-content__section-title">{{ section.title }}</h3>
          <p
            v-for="(paragraph, paragraphIndex) in paragraphs(section.body)"
            :key="`${section.id}-paragraph-${paragraphIndex}`"
            class="refund-return-policy-content__body"
          >
            {{ paragraph }}
          </p>
          <ul v-if="section.bullets?.length" class="refund-return-policy-content__list">
            <li v-for="(bullet, bulletIndex) in section.bullets" :key="`${section.id}-bullet-${bulletIndex}`">
              {{ bullet }}
            </li>
          </ul>
        </div>

        <figure v-if="section.image?.url" class="refund-return-policy-content__figure">
          <StorefrontImage
            :src="section.image.url"
            :alt="section.image.alt || section.title"
            preset="content"
            width="1200"
            height="800"
            class="refund-return-policy-content__image"
          />
          <figcaption v-if="section.image.caption" class="refund-return-policy-content__caption">
            {{ section.image.caption }}
          </figcaption>
        </figure>
      </section>
    </div>

    <footer v-if="canShowPolicy && (policy.contact_label || policy.contact_url || policy.updated_at)" class="refund-return-policy-content__footer">
      <p v-if="policy.contact_label" class="refund-return-policy-content__contact">{{ policy.contact_label }}</p>
      <NuxtLink
        v-if="policy.contact_url && !isExternalURL(policy.contact_url)"
        :to="policy.contact_url"
        class="refund-return-policy-content__link"
      >
        {{ t('storefrontDataNotice.refundReturnPolicy.contactSupport') }}
      </NuxtLink>
      <a
        v-else-if="policy.contact_url"
        :href="policy.contact_url"
        class="refund-return-policy-content__link"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ t('storefrontDataNotice.refundReturnPolicy.contactSupport') }}
      </a>
      <p v-if="policy.updated_at" class="refund-return-policy-content__updated">
        {{ t('storefrontDataNotice.refundReturnPolicy.lastUpdated', { date: formatDate(policy.updated_at) }) }}
      </p>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import StorefrontDataNotice from '~/components/StorefrontDataNotice.vue'
import StorefrontImage from '~/components/StorefrontImage.vue'
import { useRefundReturnPolicy } from '~/composables/useRefundReturnPolicy'
import type { RefundReturnPolicy } from '~/types/refundReturnPolicy'

const { t } = useI18n()
const props = withDefaults(defineProps<{
  policy?: RefundReturnPolicy | null
  contactEmail?: string
  compact?: boolean
  showTitle?: boolean
}>(), {
  policy: null,
  contactEmail: '',
  compact: false,
  showTitle: true,
})

const {
  policy: fetchedPolicy,
  pending,
  error,
  refresh,
  hasRemotePolicy,
  isFallbackPolicy,
} = useRefundReturnPolicy()
const policy = computed(() => props.policy || fetchedPolicy.value)
const hasProvidedPolicy = computed(() => Boolean(props.policy))
const hasResolvedPolicy = computed(() => hasProvidedPolicy.value || hasRemotePolicy.value)
const loadError = computed(() => !hasProvidedPolicy.value && Boolean(error.value))
const canShowPolicy = computed(() => !loadError.value && hasResolvedPolicy.value)
const contactEmail = computed(() => props.contactEmail.trim())
const contactEmailHref = computed(() => `mailto:${contactEmail.value}`)
const hasContent = computed(() => policy.value.sections.length > 0 || Boolean(policy.value.intro))
const policyFallbackNotice = computed(() => {
  if (hasProvidedPolicy.value || !isFallbackPolicy.value || !hasResolvedPolicy.value) return null

  return {
    title: t('storefrontDataNotice.refundReturnPolicy.fallbackNotice.title'),
    description: t('storefrontDataNotice.refundReturnPolicy.fallbackNotice.description'),
  }
})

const paragraphs = (body?: string): string[] => String(body || '')
  .split(/\r?\n+/)
  .map((item) => item.trim())
  .filter(Boolean)

const sectionAnchor = (id: string, index: number): string => {
  const normalized = String(id || '')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return normalized ? `refund-return-${normalized}` : `refund-return-section-${index + 1}`
}

const isExternalURL = (value: string): boolean => /^(?:https?:|mailto:)/i.test(value)

const formatDate = (value: string): string => {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'long', day: 'numeric' }).format(date)
}

const retryPolicy = () => {
  void refresh()
}
</script>

<style scoped>
.refund-return-policy-content {
  width: 100%;
  color: var(--tz-text-secondary);
}

.refund-return-policy-content__header {
  margin-bottom: 2rem;
}

.refund-return-policy-content__title {
  margin: 0 0 0.8rem;
  color: #ffffff;
  font-size: clamp(1.35rem, 2vw, 2rem);
  font-weight: 700;
  line-height: 1.2;
}

.refund-return-policy-content__intro,
.refund-return-policy-content__body,
.refund-return-policy-content__contact {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.8;
}

.refund-return-policy-content__sections {
  display: grid;
  gap: 2rem;
}

.refund-return-policy-content__section {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.25rem;
}

.refund-return-policy-content__section-title {
  margin: 0 0 0.75rem;
  color: #ffffff;
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.3;
}

.refund-return-policy-content__body + .refund-return-policy-content__body {
  margin-top: 0.75rem;
}

.refund-return-policy-content__list {
  display: grid;
  gap: 0.6rem;
  margin: 1rem 0 0;
  padding-inline-start: 1.2rem;
  font-size: 0.95rem;
  line-height: 1.7;
}

.refund-return-policy-content__figure {
  margin: 0;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.45);
}

.refund-return-policy-content__image {
  display: block;
  width: 100%;
  max-height: 28rem;
  object-fit: contain;
}

.refund-return-policy-content__caption,
.refund-return-policy-content__updated {
  padding: 0.7rem 0.9rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.5;
}

.refund-return-policy-content__footer {
  display: grid;
  gap: 0.65rem;
  margin-top: 2rem;
  padding-top: 1.25rem;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
}

.refund-return-policy-content__notice {
  margin-bottom: 1.25rem;
}

.refund-return-policy-content__state {
  display: grid;
  min-height: 8rem;
  align-items: center;
}

.refund-return-policy-content__link {
  color: #5eead4;
  text-decoration: underline;
  text-underline-offset: 0.18em;
}

.refund-return-policy-content--compact .refund-return-policy-content__header {
  margin-bottom: 1.25rem;
}

.refund-return-policy-content--compact .refund-return-policy-content__sections {
  gap: 1.25rem;
}

@media (min-width: 768px) {
  .refund-return-policy-content__section:has(.refund-return-policy-content__figure) {
    grid-template-columns: minmax(0, 1fr) minmax(16rem, 0.8fr);
    align-items: start;
  }
}
</style>
