<template>
  <div class="company-page">
    <h1 class="sr-only">Refund & Return Policy</h1>

    <div class="policies-content space-y-8 text-slate-300">
      <p class="text-sm text-slate-400">
        How we handle returns, refunds, and exchanges to keep your experience predictable and fair.
      </p>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Eligibility</h2>
        <p class="text-sm leading-relaxed">
          We accept returns within 30 days of delivery for unused items in original packaging. Custom or personalized items are non-refundable unless defective.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Condition of Items</h2>
        <p class="text-sm leading-relaxed">
          Returned items must be unused, undamaged, and include all accessories/manuals. We reserve the right to refuse returns that do not meet these conditions.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Process</h2>
        <ul class="list-disc list-inside text-sm space-y-2 text-slate-400">
          <li>Contact our support team with your order number and issue details.</li>
          <li>We will provide return instructions and, if applicable, a return authorization.</li>
          <li>Ship the item using a trackable method; retain proof of shipment.</li>
        </ul>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Refund Method & Timing</h2>
        <p class="text-sm leading-relaxed">
          Refunds are processed to the original payment method within 5–10 business days after we receive and inspect the returned item.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Shipping Costs</h2>
        <p class="text-sm leading-relaxed">
          Return shipping is the customer's responsibility unless the item is defective or incorrect. Original shipping fees are non-refundable unless required by law.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Exchanges</h2>
        <p class="text-sm leading-relaxed">
          For exchanges, please initiate a return first, then place a new order once the return is approved. This ensures availability and faster processing.
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-white mb-3">Contact</h2>
        <p class="text-sm leading-relaxed">
          For refund or return questions, email
          <a href="mailto:support@tanzanite.com" class="text-cyan-400 hover:text-cyan-300 underline">
            support@tanzanite.com
          </a>.
        </p>
      </section>

      <div class="pt-6 border-t border-slate-800">
        <p class="text-xs text-slate-500">Last updated: December 2024</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useHead, definePageMeta } from '#imports'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Refund & Return Policy',
})
</script>

<style scoped>
.company-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0.5rem;
}
</style>
