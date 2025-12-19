<template>
  <div class="company-page">
    <h1 class="sr-only">Contact Us</h1>

    <section class="company-section">
      <h2 class="company-section__title">Contact Tanzanite</h2>
      <p class="company-section__body">
        Contact channels and support links will be published on this page. You can keep using this
        entry to reach customer service, send product enquiries, or follow future social links.
      </p>
      <p class="company-section__body">
        For urgent support, please email <a class="company-section__link" href="mailto:support@tanzanite.site">support@tanzanite.site</a>.
      </p>

      <div class="mt-6">
        <ContactLocationMap />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead } from '#imports'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
import { contactLocation } from '~/utils/contactLocation'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Contact Us',
  script: [
    {
      type: 'application/ld+json',
      children: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'LocalBusiness',
        name: contactLocation.name,
        email: 'support@tanzanite.site',
        address: {
          '@type': 'PostalAddress',
          streetAddress: contactLocation.addressText,
          addressLocality: 'Xiamen',
          addressRegion: 'Fujian',
          addressCountry: 'CN',
        },
        ...(contactLocation.lat && contactLocation.lng
          ? {
              geo: {
                '@type': 'GeoCoordinates',
                latitude: contactLocation.lat,
                longitude: contactLocation.lng,
              },
            }
          : {}),
      }),
    } as any,
  ],
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.company-section {
  background: rgba(15, 23, 42, 0.9);
  border-radius: 1rem;
  padding: 1.25rem 1.35rem;
  box-shadow:
    0 14px 40px -24px rgba(0, 0, 0, 1),
    0 0 18px rgba(0, 0, 0, 0.9);
}

.company-section__title {
  margin: 0 0 0.75rem;
  font-size: 1.1rem;
  font-weight: 700;
  color: #ffffff;
}

.company-section__body {
  margin: 0 0 0.6rem;
  font-size: 0.95rem;
  line-height: 1.6;
  color: rgba(226, 232, 240, 0.9);
}

.company-section__link {
  color: #38bdf8;
  text-decoration: underline;
  text-decoration-color: rgba(56, 189, 248, 0.6);
}

.company-section__link:hover {
  color: #7dd3fc;
}
</style>
