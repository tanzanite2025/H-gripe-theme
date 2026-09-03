<template>
  <article class="refund-cancellation-policy-content" :class="{ 'refund-cancellation-policy-content--compact': compact }">
    <section class="refund-cancellation-policy-content__cancellation">
      <header class="refund-cancellation-policy-content__cancellation-header">
        <h2 v-if="showTitle" class="refund-cancellation-policy-content__title">
          {{ cancellationPolicy.title }}
        </h2>
        <p class="refund-cancellation-policy-content__intro">
          {{ cancellationPolicy.intro }}
        </p>
      </header>

      <div class="refund-cancellation-policy-content__cancellation-windows">
        <section
          v-for="policyWindow in cancellationPolicy.windows"
          :key="policyWindow.id"
          class="refund-cancellation-policy-content__cancellation-window"
        >
          <p class="refund-cancellation-policy-content__window-label">{{ policyWindow.label }}</p>
          <h3 class="refund-cancellation-policy-content__window-title">{{ policyWindow.title }}</h3>
          <p class="refund-cancellation-policy-content__window-body">{{ policyWindow.body }}</p>
        </section>
      </div>
    </section>

    <header v-if="canShowPolicy && policy.intro" class="refund-cancellation-policy-content__header">
      <p v-if="policy.intro" class="refund-cancellation-policy-content__intro">
        {{ policy.intro }}
        <template v-if="contactEmail">
          (<a :href="contactEmailHref" class="refund-cancellation-policy-content__link">{{ contactEmail }}</a>)
        </template>
      </p>
    </header>

    <StorefrontDataNotice
      v-if="policyFallbackNotice"
      class="refund-cancellation-policy-content__notice"
      tone="fallback"
      :title="policyFallbackNotice.title"
      :description="policyFallbackNotice.description"
    />

    <div v-if="loadError" class="refund-cancellation-policy-content__state">
      <StorefrontDataNotice
        tone="error"
        role="alert"
        :title="t('storefrontDataNotice.refundCancellationPolicy.error.title')"
        :description="t('storefrontDataNotice.refundCancellationPolicy.error.description')"
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

    <div v-else-if="pending && !hasResolvedPolicy" class="refund-cancellation-policy-content__state">
      <StorefrontDataNotice
        tone="empty"
        :title="t('storefrontDataNotice.refundCancellationPolicy.loading.title')"
        :description="t('storefrontDataNotice.refundCancellationPolicy.loading.description')"
      />
    </div>

    <div v-else-if="!hasContent" class="refund-cancellation-policy-content__state">
      <StorefrontDataNotice
        tone="empty"
        :title="t('storefrontDataNotice.refundCancellationPolicy.empty.title')"
        :description="t('storefrontDataNotice.refundCancellationPolicy.empty.description')"
      />
    </div>

    <div v-else class="refund-cancellation-policy-content__sections">
      <section
        v-for="(section, index) in policy.sections"
        :id="sectionAnchor(section.id, index)"
        :key="`${section.id}-${index}`"
        class="refund-cancellation-policy-content__section"
      >
        <div class="refund-cancellation-policy-content__section-copy">
          <h3 class="refund-cancellation-policy-content__section-title">{{ section.title }}</h3>
          <p
            v-for="(paragraph, paragraphIndex) in paragraphs(section.body)"
            :key="`${section.id}-paragraph-${paragraphIndex}`"
            class="refund-cancellation-policy-content__body"
          >
            {{ paragraph }}
          </p>
          <ul v-if="section.bullets?.length" class="refund-cancellation-policy-content__list">
            <li v-for="(bullet, bulletIndex) in section.bullets" :key="`${section.id}-bullet-${bulletIndex}`">
              {{ bullet }}
            </li>
          </ul>
        </div>

        <figure v-if="section.image?.url" class="refund-cancellation-policy-content__figure">
          <StorefrontImage
            :src="section.image.url"
            :alt="section.image.alt || section.title"
            preset="content"
            width="1200"
            height="800"
            class="refund-cancellation-policy-content__image"
          />
          <figcaption v-if="section.image.caption" class="refund-cancellation-policy-content__caption">
            {{ section.image.caption }}
          </figcaption>
        </figure>
      </section>
    </div>

    <footer v-if="canShowPolicy && (policy.contact_label || policy.contact_url || policy.updated_at)" class="refund-cancellation-policy-content__footer">
      <p v-if="policy.contact_label" class="refund-cancellation-policy-content__contact">{{ policy.contact_label }}</p>
      <NuxtLink
        v-if="policy.contact_url && !isExternalURL(policy.contact_url)"
        :to="policy.contact_url"
        class="refund-cancellation-policy-content__link"
      >
        {{ t('storefrontDataNotice.refundCancellationPolicy.contactSupport') }}
      </NuxtLink>
      <a
        v-else-if="policy.contact_url"
        :href="policy.contact_url"
        class="refund-cancellation-policy-content__link"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ t('storefrontDataNotice.refundCancellationPolicy.contactSupport') }}
      </a>
      <p v-if="policy.updated_at" class="refund-cancellation-policy-content__updated">
        {{ t('storefrontDataNotice.refundCancellationPolicy.lastUpdated', { date: formatDate(policy.updated_at) }) }}
      </p>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import StorefrontDataNotice from '~/components/StorefrontDataNotice.vue'
import StorefrontImage from '~/components/StorefrontImage.vue'
import { useRefundCancellationPolicy } from '~/composables/useRefundCancellationPolicy'
import { getRefundCancellationPolicyContent } from '~/data/refundCancellationPolicy'
import type { RefundCancellationPolicy } from '~/types/refundCancellationPolicy'

const { locale, t } = useI18n()
const props = withDefaults(defineProps<{
  policy?: RefundCancellationPolicy | null
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
} = useRefundCancellationPolicy()
const policy = computed(() => props.policy || fetchedPolicy.value)
const cancellationPolicy = computed(() => getRefundCancellationPolicyContent(locale.value))
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
    title: t('storefrontDataNotice.refundCancellationPolicy.fallbackNotice.title'),
    description: t('storefrontDataNotice.refundCancellationPolicy.fallbackNotice.description'),
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
  return normalized ? `refund-cancellation-${normalized}` : `refund-cancellation-section-${index + 1}`
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
.refund-cancellation-policy-content {
  width: 100%;
  color: var(--tz-text-secondary);
}

.refund-cancellation-policy-content__header {
  margin-bottom: 2rem;
}

.refund-cancellation-policy-content__cancellation {
  margin-bottom: 2.25rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid rgba(148, 163, 184, 0.18);
}

.refund-cancellation-policy-content__cancellation-header {
  margin-bottom: 1.25rem;
}

.refund-cancellation-policy-content__title {
  margin: 0 0 0.8rem;
  color: var(--tz-text-primary);
  font-size: clamp(1.35rem, 2vw, 2rem);
  font-weight: 700;
  line-height: 1.2;
}

.refund-cancellation-policy-content__intro,
.refund-cancellation-policy-content__body,
.refund-cancellation-policy-content__contact {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.8;
}

.refund-cancellation-policy-content__cancellation-windows {
  display: grid;
  gap: 1rem;
}

.refund-cancellation-policy-content__cancellation-window {
  display: grid;
  gap: 0.55rem;
  align-content: start;
  padding: 1rem;
  overflow-wrap: anywhere;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-left: 3px solid var(--tz-site-accent);
  border-radius: 0.5rem;
  background: var(--tz-surface-card);
}

.refund-cancellation-policy-content__window-label {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.4;
  text-transform: uppercase;
}

.refund-cancellation-policy-content__window-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
}

.refund-cancellation-policy-content__window-body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.7;
}

.refund-cancellation-policy-content__sections {
  display: grid;
  gap: 2rem;
}

.refund-cancellation-policy-content__section {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.25rem;
}

.refund-cancellation-policy-content__section-title {
  margin: 0 0 0.75rem;
  color: var(--tz-text-primary);
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.3;
}

.refund-cancellation-policy-content__body + .refund-cancellation-policy-content__body {
  margin-top: 0.75rem;
}

.refund-cancellation-policy-content__list {
  display: grid;
  gap: 0.6rem;
  margin: 1rem 0 0;
  padding-inline-start: 1.2rem;
  font-size: 0.95rem;
  line-height: 1.7;
}

.refund-cancellation-policy-content__figure {
  margin: 0;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 0.75rem;
  background: var(--tz-image-loading-surface);
}

.refund-cancellation-policy-content__image {
  display: block;
  width: 100%;
  max-height: 28rem;
  object-fit: contain;
}

.refund-cancellation-policy-content__caption,
.refund-cancellation-policy-content__updated {
  padding: 0.7rem 0.9rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.5;
}

.refund-cancellation-policy-content__footer {
  display: grid;
  gap: 0.65rem;
  margin-top: 2rem;
  padding-top: 1.25rem;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
}

.refund-cancellation-policy-content__notice {
  margin-bottom: 1.25rem;
}

.refund-cancellation-policy-content__state {
  display: grid;
  min-height: 8rem;
  align-items: center;
}

.refund-cancellation-policy-content__link {
  color: var(--tz-site-accent);
  text-decoration: underline;
  text-underline-offset: 0.18em;
}

.refund-cancellation-policy-content--compact .refund-cancellation-policy-content__header {
  margin-bottom: 1.25rem;
}

.refund-cancellation-policy-content--compact .refund-cancellation-policy-content__cancellation {
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
}

.refund-cancellation-policy-content--compact .refund-cancellation-policy-content__sections {
  gap: 1.25rem;
}

@media (min-width: 768px) {
  .refund-cancellation-policy-content__cancellation-windows {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .refund-cancellation-policy-content__section:has(.refund-cancellation-policy-content__figure) {
    grid-template-columns: minmax(0, 1fr) minmax(16rem, 0.8fr);
    align-items: start;
  }
}
</style>
