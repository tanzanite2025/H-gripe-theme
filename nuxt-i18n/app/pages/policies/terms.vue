<template>
  <div class="company-page">
    <h1 class="sr-only">Terms of Service</h1>

    <div class="policies-content mt-4 space-y-8 tz-text-secondary">
      <p class="text-sm tz-text-secondary">
        Please read these terms carefully before using our website or making a purchase.
      </p>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">1. Agreement to Terms</h2>
        <p class="text-sm leading-relaxed">
          By accessing or using the Tanzanite website, you agree to be bound by these Terms of Service and all applicable laws and regulations. If you do not agree, please do not use this site.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">2. Use License</h2>
        <p class="text-sm leading-relaxed mb-3">
          Permission is granted to temporarily view the materials on this website for personal, non-commercial use. This license does not permit modification, copying, or reverse engineering.
        </p>
        <ul class="list-disc list-inside text-sm space-y-2 tz-text-secondary">
          <li>Using materials for commercial purposes</li>
          <li>Removing any copyright or proprietary notices</li>
          <li>Transferring materials to another person</li>
        </ul>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">3. Products and Pricing</h2>
        <p class="text-sm leading-relaxed">
          Products are subject to availability and price changes. We may correct errors and refuse or cancel orders where pricing or descriptions are inaccurate.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">4. Orders and Payment</h2>
        <p class="text-sm leading-relaxed">
          Orders require full payment before processing. We accept major payment methods as displayed at checkout. All orders are subject to acceptance and availability.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">5. Shipping and Delivery</h2>
        <p class="text-sm leading-relaxed">
          Shipping times vary by location and method. Delivery estimates are not guarantees. Risk of loss transfers upon delivery to the carrier.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">6. Returns and Refunds</h2>
        <p class="text-sm leading-relaxed">
          Returns are accepted within 30 days for unused items in original packaging (unless defective). Refunds process 5–10 business days after we receive the item.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">7. Warranty</h2>
        <p class="text-sm leading-relaxed">
          Warranty coverage follows product descriptions and excludes misuse, accidents, or normal wear and tear.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">8. Limitation of Liability</h2>
        <p class="text-sm leading-relaxed">
          Tanzanite is not liable for indirect or consequential damages arising from site or product use. Total liability is limited to the amount paid for the product in question.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">9. Intellectual Property</h2>
        <p class="text-sm leading-relaxed">
          All site content is owned by Tanzanite or its suppliers and protected by copyright and trademark laws.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">10. Governing Law</h2>
        <p class="text-sm leading-relaxed">
          These terms are governed by applicable laws without regard to conflict of law principles. Disputes are resolved in the appropriate courts.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">11. Changes to Terms</h2>
        <p class="text-sm leading-relaxed">
          We may modify these terms at any time. Continued use of the site after changes constitutes acceptance.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">12. Contact Us</h2>
        <p class="text-sm leading-relaxed">
          Questions about these Terms? Email
          <a href="mailto:support@tanzanite.site" class="text-cyan-400 hover:text-cyan-300 underline">
            support@tanzanite.site
          </a>.
        </p>
      </section>

      <div class="pt-6 border-t border-slate-800">
        <p class="text-xs tz-text-muted">Last updated: December 2024</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead } from '#imports'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Terms of Service',
})
</script>

<style scoped>
.company-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0.5rem;
}

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
