<template>
  <div class="wheelset-policy-flow">
    <article class="wheelset-guide-step wheelset-safety-notice">
      <div class="wheelset-guide-step__number wheelset-safety-notice__icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
      </div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title wheelset-safety-notice__title">Important Safety Notice</h3>
        <p class="wheelset-guide-step__body">
            Improper installation, use, or maintenance may cause serious damage to the wheels and pose a danger to the rider.
             To ensure a reliable riding experience, please read and follow the instructions below.
        </p>
        <div class="wheelset-guide-panel">
          <p>Even if you have years of cycling or wheelset assembly experience, these instructions still apply to these carbon fiber products.</p>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">1</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Installation & Inspection</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Professional installation</strong>
            <p>Have the wheelset installed by a qualified mechanic. Incompatibility with brakes or drivetrain can damage wheels or frame.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Tire installation</strong>
            <p>Use provided plastic tire levers only. Metal levers will permanently damage the rim.</p>
            <div class="wheelset-guide-actions">
              <button type="button" class="wheelset-guide-button" @click="goToTubelessInstallation">
                  Open Tubeless Tire Installation Guide
              </button>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Torque specs</strong>
            <p>Follow manufacturer recommended torque for all bolts and axles.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Brake pads</strong>
            <p>Use carbon-specific pads only to prevent rim damage.</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">2</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Usage & Riding</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Weight limits</strong>
            <p>Strictly follow max rider weight/load limits. Overloading causes rim cracking and injury risk.</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Road conditions</strong>
            <p>Avoid severe potholes and gravel. Impact damage may not be immediately visible.</p>
            <div class="wheelset-guide-note">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
              <span>After impact: clean the rim, remove the tire, and inspect carefully.</span>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Braking technique</strong>
            <p>Avoid continuous drag braking on long descents. Use intermittent braking to prevent carbon overheating.</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">3</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Regular Maintenance</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Pre-ride inspection</strong>
            <ul>
              <li>Check rims for cracks, bulges, or delamination.</li>
              <li>Inspect brake pads and remove metal shavings or stones.</li>
            </ul>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Spoke tension</strong>
            <p>Regularly check for even tension distribution.</p>
            <div class="wheelset-guide-actions">
              <button type="button" class="wheelset-guide-button" @click="goToTechnicalTension">
                      Open Spoke Tension Settings
              </button>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Cleaning</strong>
            <p>Use neutral detergent and soft cloth only. Do not use strong acids, alkalis, or high-pressure water.</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">4</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">Prohibited & Emergency</h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>Strictly prohibited</strong>
            <ul>
              <li>No non-original parts modification.</li>
              <li>No sanding or grinding with metal tools.</li>
              <li>No extreme-temperature storage such as ovens or freezing environments.</li>
            </ul>
          </div>
          <div class="wheelset-guide-panel">
            <strong>Emergency protocol</strong>
            <p>If cracks or abnormal noises occur: <strong>stop riding immediately</strong>.</p>
            <p class="mt-2">After any crash, perform thorough inspection even if no damage is visible.</p>
          </div>
        </div>

        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button"
            @click="goToWarranty"
          >
                    Open Advanced Warranty Policy
          </button>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
const { goToTubelessInstallation, goToTechnicalTension, goToWarranty } = defineProps<{
  goToTubelessInstallation: () => void | Promise<void>
  goToTechnicalTension: () => void | Promise<void>
  goToWarranty: () => void | Promise<void>
  }>()
</script>
