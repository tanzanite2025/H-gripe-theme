<template>
  <div>
    <p class="sizecharts-section__intro">
      Example front and rear rim combinations for different riding scenarios when using mixed rim profiles.
    </p>

    <h3 class="sizecharts-section__subheading text-sky-300 font-semibold mt-3">
      1. Different rim models/specifications for front and rear
    </h3>
    <ul class="sizecharts-section__list">
      <li>
        Refers to using different rims for the front and rear wheels (for example, a lightweight front rim and a
        more durable or higher-profile rear rim).
      </li>
      <li>
        This approach is common in custom wheelsets or testing and is used to balance weight, aerodynamics, and
        strength.
      </li>
    </ul>

    <h3 class="sizecharts-section__subheading text-sky-300 font-semibold mt-4">
      2. Different rim heights or widths within the same wheelset
    </h3>
    <ul class="sizecharts-section__list">
      <li>
        For example, a 50 mm high-profile rim on the front and a 60 mm high-profile rim on the rear; this setup is
        sometimes referred to as a mixed rim setup.
      </li>
      <li>
        Purpose: the front wheel offers easier handling, while the rear wheel provides greater stiffness and
        aerodynamic advantage.
      </li>
    </ul>

    <p class="guide-section__cta-wrapper mt-3">
      You can customize your purchase through
      <button
        type="button"
        class="wheelset-inline-button"
        @click="openQuickBuy"
      >
        Quick buy
      </button>
      , or send us an email at
      <a
        class="wheelset-link"
        href="mailto:support@tanzanit.site"
      >
        support@tanzanit.site
      </a>
      . You can also use
      <button
        type="button"
        class="wheelset-inline-button"
        @click="openWhatsAppChat"
      >
        Live chat
      </button>
      to contact us.
    </p>

    <div class="mt-4 overflow-x-auto">
      <table
        class="min-w-full border-separate border-spacing-0 rounded-xl overflow-hidden bg-slate-900/70 text-slate-100 text-xs sm:text-sm"
      >
        <thead>
          <tr class="bg-slate-800/90 text-sm font-semibold text-slate-50">
            <th class="px-4 py-3 text-left border-b border-slate-700">Scenario</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">Front rim choice</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">Rear rim choice</th>
            <th class="px-4 py-3 text-left border-b border-slate-700">Purpose</th>
          </tr>
        </thead>
        <tbody class="text-sm">
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3 font-semibold align-top">Lightweight climbing</td>
            <td class="px-4 py-3 align-top">
              Low-profile lightweight rim (30–40 mm, lower height, reduced weight)
            </td>
            <td class="px-4 py-3 align-top">
              Mid-profile or higher rim (40–50 mm, balances stiffness)
            </td>
            <td class="px-4 py-3 align-top">
              Front wheel reduces weight and improves handling; rear wheel maintains stiffness and efficiency,
              ideal for climbing routes.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3 font-semibold align-top">Aerodynamics</td>
            <td class="px-4 py-3 align-top">
              Mid-to-high profile rim (50–60 mm, aerodynamic gain)
            </td>
            <td class="px-4 py-3 align-top">
              High-profile rim (60–80 mm, maximum aero advantage)
            </td>
            <td class="px-4 py-3 align-top">
              Front wheel balances handling and aerodynamics; rear wheel maximizes aero benefit, ideal for flat
              high-speed riding.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3 font-semibold align-top">Durability / training</td>
            <td class="px-4 py-3 align-top">
              Mid-profile alloy or carbon rim (stable and durable)
            </td>
            <td class="px-4 py-3 align-top">
              High-profile or wider rim (enhanced strength and stability)
            </td>
            <td class="px-4 py-3 align-top">
              Front wheel ensures comfort and durability; rear wheel bears more load, suitable for daily training
              or endurance rides.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-4 py-3 font-semibold align-top">Mixed purpose</td>
            <td class="px-4 py-3 align-top">
              Mid-profile rim (40–50 mm, balance of weight and aero)
            </td>
            <td class="px-4 py-3 align-top">
              Higher-profile rim (55–65 mm, added stiffness and aero)
            </td>
            <td class="px-4 py-3 align-top">
              Balanced setup for versatile riding (racing + training).
            </td>
          </tr>
          <tr>
            <td class="px-4 py-3 font-semibold align-top">Special customization</td>
            <td class="px-4 py-3 align-top">Lightweight carbon rim</td>
            <td class="px-4 py-3 align-top">Durable alloy rim</td>
            <td class="px-4 py-3 align-top">
              Front wheel focuses on lightweight performance; rear wheel on durability, suitable for
              budget-conscious or special needs.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
const { openQuickBuy, openWhatsAppChat } = defineProps<{
  openQuickBuy: () => void | Promise<void>
  openWhatsAppChat: () => void | Promise<void>
}>()
</script>

<style scoped>
.guide-section__cta-wrapper {
  color: #e5e7eb;
}

.wheelset-link {
  color: #2dd4bf;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: #60a5fa;
}

.wheelset-inline-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.65rem;
  margin: 0 0.25rem;
  border-radius: 9999px;
  border: none;
  background: rgba(31, 41, 55, 0.65);
  color: #ffffff;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.85);
  transition: background-color 0.15s ease, transform 0.08s ease;
}

.wheelset-inline-button:hover {
  background: rgba(51, 65, 85, 0.75);
}

.wheelset-inline-button:active {
  transform: scale(0.98);
}
</style>
