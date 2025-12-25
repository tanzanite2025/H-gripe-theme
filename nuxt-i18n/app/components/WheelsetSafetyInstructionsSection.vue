<template>
  <div>
    <p class="sizecharts-section__intro">
      Improper installation, use, or maintenance may cause serious damage to the wheels and pose a danger to the rider.
      To ensure a reliable riding experience with TANZANITE products, please read and follow the instructions below.
      Even if you have years of cycling or wheelset assembly experience, these instructions and recommendations apply only
      to TANZANITE carbon fiber products. Therefore, please carefully read them before your first ride or wheel assembly.
    </p>
    <p class="sizecharts-section__intro">
      If you have any questions, feel free to contact us at
      <a
        class="wheelset-link text-teal-300 underline hover:text-sky-400"
        href="mailto:support@tanzanite.site"
      >
        support@tanzanite.site
      </a>
      or
      <button
        type="button"
        class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
        @click="openWhatsAppChat"
      >
        leave a message
      </button>
      on any page of our website.
    </p>
    <ol class="sizecharts-section__list sizecharts-section__list--numbered">
      <li>
        <strong style="color: #38bdf8;">Installation &amp; Inspection</strong>
        <p class="wheelset-safety-item mt-1">
          <strong>Professional Installation:</strong>
          Have the wheelset installed by a qualified mechanic or experienced user, ensuring compatibility with the braking and
          drivetrain systems. Incompatibility may reduce performance and damage the wheels or the frame/fork.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Tire Installation:</strong>
          We provide professional tire levers for installing tires. Do not use metal tire levers, as they will permanently damage
          the rim.
          <button
            type="button"
            class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
            @click="goToTubelessInstallation"
          >
            Tubeless tire installation instructions
          </button>
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Torque Specifications:</strong>
          All bolts, quick releases, or thru-axles must be tightened according to the manufacturer’s recommended torque values to
          avoid being too tight or too loose.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Brake Pad Compatibility:</strong>
          Use only carbon-specific brake pads to prevent rim damage from metal or incompatible materials.
        </p>
      </li>
      <li>
        <strong style="color: #38bdf8;">Usage &amp; Riding</strong>
        <p class="wheelset-safety-item mt-1">
          <strong>Weight Limit:</strong>
          Follow the manufacturer’s specified maximum rider weight and load limits. Our wheels are designed and tested for
          specific riding purposes. Improper use may shorten product lifespan or even cause rim cracking, leading to rider
          injury. Always ensure the product you purchase suits your riding needs.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Road Conditions:</strong>
          Avoid riding on severe potholes, gravel, or extremely rough terrain to reduce impact damage. After a strong impact or
          collision, failure may not be immediately visible. Wipe the rim clean, consider removing the tire for inspection, and
          report any potential serious damage to us.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Braking Technique:</strong>
          On long descents, avoid continuous braking. Use intermittent braking to prevent overheating of the carbon fiber.
        </p>
      </li>
      <li>
        <strong style="color: #38bdf8;">Regular Maintenance</strong>
        <p class="wheelset-safety-item mt-1">
          <strong>Surface Inspection:</strong>
          Before each ride, check the rims for cracks, bulges, or delamination. Regularly inspect brake pads and remove
          contaminants such as metal shavings or stones, which can damage the braking surface. Do not use excessively worn brake
          pads.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Spoke Tension:</strong>
          Regularly check spoke tension to ensure even distribution and prevent structural imbalance.
          <button
            type="button"
            class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
            @click="goToTechnicalTension"
          >
            Check out our targeted settings for spoke tension
          </button>
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Cleaning Method:</strong>
          Clean with a neutral detergent and soft cloth. Avoid strong acids, strong alkalis, or high-pressure water jets.
        </p>
      </li>
      <li>
        <strong style="color: #38bdf8;">Prohibited Actions</strong>
        <p class="wheelset-safety-item mt-1">Do not use non-original parts or modify the braking system.</p>
        <p class="wheelset-safety-item mt-1">Do not sand or grind carbon rims with sandpaper or metal tools.</p>
        <p class="wheelset-safety-item mt-1">
          Do not store or use wheels in extreme temperature environments (e.g., ovens or freezing conditions).
        </p>
      </li>
      <li>
        <strong style="color: #38bdf8;">Emergency Handling</strong>
        <p class="wheelset-safety-item mt-1">
          <strong>Damage Detection:</strong>
          If cracks or abnormal noises are found, stop using the wheel immediately and seek professional repair.
        </p>
        <p class="wheelset-safety-item mt-1">
          <strong>Post-Accident Inspection:</strong>
          After a crash or collision, conduct a thorough inspection even if no visible damage is present.
        </p>
      </li>
    </ol>

    <div class="wheelset-safety-footer mt-3">
      <p class="guide-section__cta-wrapper">
        <button
          type="button"
          class="wheelset-inline-button inline-flex items-center justify-center px-2.5 py-1 mx-1 rounded-full border-0 bg-slate-800/80 text-white text-xs font-semibold shadow-lg hover:bg-slate-700/80 active:scale-95 transition duration-150"
          @click="goToAfterSales"
        >
          Check out our after-sales instructions for rims, wheelsets and more
        </button>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const { openWhatsAppChat, goToTubelessInstallation, goToTechnicalTension, goToAfterSales } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  goToTubelessInstallation: () => void | Promise<void>
  goToTechnicalTension: () => void | Promise<void>
  goToAfterSales: () => void | Promise<void>
}>()
</script>
