<template>
  <div>
    <p class="text-center text-slate-50">
      We provide professional assembly services, supporting customers who wish to send in carbon fiber rims, wheelsets, and bicycle components.
      Our skilled technicians will complete the assembly to the highest standards.
    </p>
    <p class="mt-2 text-center text-slate-50">
      Please contact our sales team at
      <a
        class="wheelset-link text-teal-300 underline hover:text-sky-400"
        href="mailto:support@tanzanite.site"
      >
        support@tanzanite.site
      </a>
      to begin the quotation process, or
      <button
        type="button"
        class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
        @click="openWhatsAppChat"
      >
        reach us directly via online chat
      </button>
      .
    </p>
    <p class="mt-2 text-center text-slate-50">
      <strong style="color: #38bdf8;">Assembly Fee: USD 30 per set</strong>
    </p>

    <ol class="sizecharts-section__list sizecharts-section__list--numbered mt-3">
      <li>
        <strong style="color: #38bdf8;">Service Process</strong>
        <ul class="sizecharts-section__list">
          <li>
            <strong>Shipping:</strong>
            Customers send their components to the designated address.
          </li>
          <li>
            <strong>Confirmation:</strong>
            Upon receipt, we will check and confirm all items.
          </li>
          <li>
            <strong>Assembly:</strong>
            Professional technicians perform installation and adjustment.
          </li>
          <li>
            <strong>Testing:</strong>
            Safety and performance checks are conducted to ensure compliance with riding standards.
          </li>
          <li>
            <strong>Delivery:</strong>
            The assembled product will be shipped back or made available for pickup.
          </li>
        </ul>
      </li>
      <li>
        <strong style="color: #38bdf8;">Accepted Components</strong>
        <ul class="sizecharts-section__list">
          <li>Wheelsets (including hubs)</li>
          <li>Other related accessories (please confirm with us in advance)</li>
        </ul>
      </li>
      <li>
        <strong style="color: #38bdf8;">Important Notes</strong>
        <ul class="sizecharts-section__list">
          <li>Please ensure that all components sent are new or in good condition.</li>
          <li>A detailed packing list is recommended to avoid missing items.</li>
          <li>Special customization requests should be communicated in advance.</li>
          <li>Typical assembly time is 3–7 business days, depending on actual circumstances.</li>
        </ul>
      </li>
      <li>
        <strong style="color: #38bdf8;">Value Proposition</strong>
        <ul class="sizecharts-section__list">
          <li>
            <strong>Expert Technicians:</strong>
            Years of experience in carbon fiber wheelset and complete bike assembly.
          </li>
          <li>
            <strong>Safety Assurance:</strong>
            Rigorous testing procedures to guarantee riding safety.
          </li>
          <li>
            <strong>Efficient Delivery:</strong>
            Fast assembly and return shipping, saving customers valuable time.
          </li>
        </ul>
      </li>
    </ol>

    <p class="mt-4 text-center text-slate-50">
      For smooth service, shipping costs for sending and returning components are the responsibility of the customer.
      We will assist by recommending reliable logistics options to ensure your parts arrive safely.
    </p>
    <p class="mt-4 text-center text-sky-400">
      If you wish to assemble your own wheels and purchase parts exclusively from TANZANITE , we are happy to assist you.
      We will guide you through every step of the assembly process, including spoke length calculation, assembly methods, and tire installation.
      <button
        type="button"
        class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
        @click="goToTechnicalSpokePattern"
      >
        Click to view documentation on spoke weaving
      </button>
      <button
        type="button"
        class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
        @click="goToHolePatterns"
      >
        Check out the rim weaving holes we can customize
      </button>
    </p>
  </div>
</template>

<script setup lang="ts">
const { openWhatsAppChat, goToTechnicalSpokePattern, goToHolePatterns } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  goToTechnicalSpokePattern: () => void | Promise<void>
  goToHolePatterns: () => void | Promise<void>
}>()
</script>
