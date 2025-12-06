<template>
  <div>
    <h1 class="sr-only">Spoke Calculator</h1>
    <p class="support-page__intro">
      Use the Tanzanite spoke length calculator to design wheelsets and estimate spoke lengths
      before you order.
    </p>
    <p
      class="support-page__intro support-page__intro--secondary support-page__technical-hint"
    >
      Not sure about all the parameters? See the Technical guide for rim, hub and ERD definitions
      before you calculate.
      <NuxtLink
        :to="localePath('/guides/technical')"
        class="support-page__technical-hint-button"
      >
        Open Technical guide
      </NuxtLink>
    </p>
    <div class="support-page__technical-image">
      <img
        src="/public/technical/lengthcalculationparameters.webp"
        alt="Overview of spoke length calculation parameters used in the Technical guide"
      />
    </div>

    <div class="support-page__calculator-wrapper">
      <SpokeCalculatorCore />

      <div class="mt-8">
        <SpokeHistorySearch />
      </div>
    </div>

    <div class="mt-10">
      <UserFeedbackThread
        threadKey="products-spoke-calculator"
        title="Share your feedback about the Spoke Calculator"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import SpokeHistorySearch from '~/components/SpokeHistorySearch.vue'
import { useLocalePath } from '#imports'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Spoke Calculator',
})

const localePath = useLocalePath()
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-page__intro--secondary {
  margin-top: 0.5rem;
}

.support-page__link {
  color: #38bdf8;
  text-decoration: underline;
}

.support-page__technical-hint {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.support-page__technical-hint-button {
  padding: 0.15rem 0.7rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  font-size: 0.8rem;
  font-weight: 500;
  color: #e0f2fe;
  text-decoration: none;
  background: rgba(15, 23, 42, 0.9);
  transition: background-color 0.15s ease, border-color 0.15s ease,
    color 0.15s ease;
}

.support-page__technical-hint-button:hover {
  background: rgba(56, 189, 248, 0.2);
  border-color: rgba(56, 189, 248, 1);
  color: #ffffff;
}

.support-page__technical-image {
  margin-top: 0.5rem;
}

.support-page__technical-image img {
  display: block;
  max-width: 480px;
  width: 100%;
  height: auto;
  border-radius: 0.5rem;
}

.support-page__calculator-wrapper {
  margin-top: 1.5rem;
}
</style>
