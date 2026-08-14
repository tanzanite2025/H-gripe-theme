<template>
  <div class="wheelset-policy-flow">
    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">1</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Professional Wheel Building Service</h3>
        <p class="wheelset-guide-step__body">
           We provide professional assembly services, supporting customers who wish to send in hubs or rims.
           Our skilled technicians will complete the assembly to the highest standards.
        </p>

        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Assembly fee</strong>
            <p>USD 30 / set for standard wheel building service.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Best for</strong>
            <p>Customers sending compatible hubs, rims, or full component kits for professional assembly.</p>
          </div>
        </div>

        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button wheelset-guide-button--solid"
            @click="openWhatsAppChat"
          >
               Get Wheel Building Quote via Chat
          </button>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">2</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Service Process</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Shipping</strong>
            <p>Send components to our designated address.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Confirmation</strong>
            <p>We inspect and confirm receipt of all listed items.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Assembly & testing</strong>
            <p>Professional build plus tension and safety analysis.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Delivery</strong>
            <p>Finished wheelset shipped back or prepared for pickup.</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">3</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Scope & Notes</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Accepted items</strong>
            <p>Wheelsets, rims, hubs, spokes, or related components. Please confirm compatibility with us first.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Important</strong>
            <ul>
              <li>Components must be new or in good condition.</li>
              <li>Include a detailed packing list.</li>
              <li>Typical turnaround is 3-7 business days.</li>
              <li>Shipping costs are the customer's responsibility.</li>
            </ul>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">4</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Building Your Own Wheels?</h3>
        <p class="wheelset-guide-step__body">
           If you prefer to assemble yourself using these parts, we support you. Access our technical guides and customization options below.
        </p>
        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button"
            @click="goToTechnicalSpokePattern"
          >
            Open Wheel Component Spoke Patterns
          </button>
          <button
            type="button"
            class="wheelset-guide-button"
            @click="goToHolePatterns"
          >
            Open Rim Hole Pattern Options
          </button>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
const { openWhatsAppChat, goToTechnicalSpokePattern, goToHolePatterns } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  goToTechnicalSpokePattern: () => void | Promise<void>
  goToHolePatterns: () => void | Promise<void>
}>()
</script>
