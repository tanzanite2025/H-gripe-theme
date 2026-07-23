<template>
  <div>
    <h2 class="support-page__title">Shipping instructions</h2>
    <p class="support-page__intro">
      This page provides an overview of Tanzanite shipping, including handling time, delivery
      options, and what to expect for international orders.
    </p>

    <section class="support-section">
      <h3 class="support-section__title">Overview</h3>
      <p class="support-section__body">
        Here we will collect the main shipping guidelines for Tanzanite wheelsets and
        components. In the future this page will include detailed examples for different
        regions, typical delivery times, and links to more specific guides.
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Processing &amp; handling</h3>
      <p class="support-section__body">
        Order processing time can vary depending on whether the wheelset is in stock or built
        to order. We will describe how long it usually takes to prepare an order before
        handing it over to the carrier, and how this interacts with custom wheel building
        services.
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">International shipping &amp; customs</h3>
      <p class="support-section__body">
        For international deliveries, customs clearance and local import duties can affect the
        final delivery time. This section will later list common destinations, examples of how
        duties are handled, and any documents we provide with the shipment.
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Coming soon</h3>
      <p class="support-section__body">
        If you need specific shipping information right now, please contact our support team
        with your country and the products you are interested in so we can share the latest
        details while this page is being completed.
      </p>
    </section>

  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Shipping instructions',
})
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.support-section {
  margin-top: 1.75rem;
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--tz-text-secondary);
}
</style>
