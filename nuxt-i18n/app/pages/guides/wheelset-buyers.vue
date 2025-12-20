<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Wheelset Buyers Guide</h2>

    <div class="wheelset-page">
      <div class="wheelset-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="wheelset-tabs__item"
          :class="{ 'wheelset-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Safety instructions -->
      <section
        v-show="activeTab === 'safety-instructions'"
        id="safety-instructions"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Safety instructions</h3>
        <p class="products-page__intro">
          Improper installation, use, or maintenance may cause serious damage to the wheels and pose a danger to the rider.
          To ensure a reliable riding experience with TANZANITE products, please read and follow the instructions below.
          Even if you have years of cycling or wheelset assembly experience, these instructions and recommendations apply only
          to TANZANITE carbon fiber products. Therefore, please carefully read them before your first ride or wheel assembly.
        </p>
        <p class="products-page__intro">
          If you have any questions, feel free to contact us at
          <a class="wheelset-link" href="mailto:support@tanzanite.site">support@tanzanite.site</a>
          or
          <button
            type="button"
            class="wheelset-inline-button"
            @click="openWhatsAppChat"
          >
            leave a message
          </button>
          on any page of our website.
        </p>
        <ol class="wheelset-section__list wheelset-section__list--ordered">
          <li>
            <strong>Installation &amp; Inspection</strong>
            <p class="wheelset-safety-item">
              <strong>Professional Installation:</strong>
              Have the wheelset installed by a qualified mechanic or experienced user, ensuring compatibility with the braking and
              drivetrain systems. Incompatibility may reduce performance and damage the wheels or the frame/fork.
            </p>
            <p class="wheelset-safety-item">
              <strong>Tire Installation:</strong>
              We provide professional tire levers for installing tires. Do not use metal tire levers, as they will permanently damage
              the rim.
              <button
                type="button"
                class="wheelset-inline-button"
                @click="goToTubelessInstallation"
              >
                Tubeless tire installation instructions
              </button>
            </p>
            <p class="wheelset-safety-item">
              <strong>Torque Specifications:</strong>
              All bolts, quick releases, or thru-axles must be tightened according to the manufacturer’s recommended torque values to
              avoid being too tight or too loose.
            </p>
            <p class="wheelset-safety-item">
              <strong>Brake Pad Compatibility:</strong>
              Use only carbon-specific brake pads to prevent rim damage from metal or incompatible materials.
            </p>
          </li>
          <li>
            <strong>Usage &amp; Riding</strong>
            <p class="wheelset-safety-item">
              <strong>Weight Limit:</strong>
              Follow the manufacturer’s specified maximum rider weight and load limits. Our wheels are designed and tested for
              specific riding purposes. Improper use may shorten product lifespan or even cause rim cracking, leading to rider
              injury. Always ensure the product you purchase suits your riding needs.
            </p>
            <p class="wheelset-safety-item">
              <strong>Road Conditions:</strong>
              Avoid riding on severe potholes, gravel, or extremely rough terrain to reduce impact damage. After a strong impact or
              collision, failure may not be immediately visible. Wipe the rim clean, consider removing the tire for inspection, and
              report any potential serious damage to us.
            </p>
            <p class="wheelset-safety-item">
              <strong>Braking Technique:</strong>
              On long descents, avoid continuous braking. Use intermittent braking to prevent overheating of the carbon fiber.
            </p>
          </li>
          <li>
            <strong>Regular Maintenance</strong>
            <p class="wheelset-safety-item">
              <strong>Surface Inspection:</strong>
              Before each ride, check the rims for cracks, bulges, or delamination. Regularly inspect brake pads and remove
              contaminants such as metal shavings or stones, which can damage the braking surface. Do not use excessively worn brake
              pads.
            </p>
            <p class="wheelset-safety-item">
              <strong>Spoke Tension:</strong>
              Regularly check spoke tension to ensure even distribution and prevent structural imbalance.
              <button
                type="button"
                class="wheelset-inline-button"
                @click="goToTechnicalTension"
              >
                Check out our targeted settings for spoke tension
              </button>
            </p>
            <p class="wheelset-safety-item">
              <strong>Cleaning Method:</strong>
              Clean with a neutral detergent and soft cloth. Avoid strong acids, strong alkalis, or high-pressure water jets.
            </p>
          </li>
          <li>
            <strong>Prohibited Actions</strong>
            <p class="wheelset-safety-item">Do not use non-original parts or modify the braking system.</p>
            <p class="wheelset-safety-item">Do not sand or grind carbon rims with sandpaper or metal tools.</p>
            <p class="wheelset-safety-item">
              Do not store or use wheels in extreme temperature environments (e.g., ovens or freezing conditions).
            </p>
          </li>
          <li>
            <strong>Emergency Handling</strong>
            <p class="wheelset-safety-item">
              <strong>Damage Detection:</strong>
              If cracks or abnormal noises are found, stop using the wheel immediately and seek professional repair.
            </p>
            <p class="wheelset-safety-item">
              <strong>Post-Accident Inspection:</strong>
              After a crash or collision, conduct a thorough inspection even if no visible damage is present.
            </p>
          </li>
        </ol>

        <div class="wheelset-safety-footer">
          <button
            type="button"
            class="wheelset-inline-button"
            @click="goToAfterSales"
          >
            Check out our after-sales instructions for rims, wheelsets and more
          </button>
        </div>
      </section>

      <!-- Mullet wheelsets -->
      <section
        v-show="activeTab === 'mullet-wheelsets'"
        id="mullet-wheelsets"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Mullet wheelsets</h3>
        <p class="products-page__intro">
          Many mountain bikers have fallen in love with the combination of a 29-inch front wheel paired with a 27.5-inch rear
          wheel, or a 27.5-inch front wheel with a 26-inch rear wheel. This wheel size mix is particularly well-suited for
          off-road riding. It offers easy handling on steep descents, agile and responsive cornering, and the ability to roll
          smoothly over rough terrain. This setup is commonly known as the “mullet”—and for good reason: the front wheel
          provides stability and reliability, while the rear wheel delivers agility and performance!
        </p>
        <button
          type="button"
          class="wheelset-inline-button"
          @click="openQuickBuy"
        >
          You can use our quick customization to match
        </button>
      </section>

      <!-- Custom Front & Rear Wheels -->
      <section
        v-show="activeTab === 'custom-front-rear-wheels'"
        id="custom-front-rear-wheels"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Custom Front &amp; Rear Wheels</h3>
        <p class="products-page__intro">
          We specialize in customized services, including building individual wheels tailored to your specific needs. While we only
          showcase complete wheelsets online, single wheels can be custom-made upon request. Please contact us at
          <a class="wheelset-link" href="mailto:support@tanzanite.site">support@tanzanite.site</a>
          for special orders—or
          <button
            type="button"
            class="wheelset-inline-button"
            @click="openWhatsAppChat"
          >
            use our faster Instant Chat option
          </button>
          !
        </p>
      </section>

      <!-- Sample assembly -->
      <section
        v-show="activeTab === 'sample-assembly'"
        id="sample-assembly"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Sample assembly</h3>
        <p class="products-page__intro">
          We provide professional assembly services, supporting customers who wish to send in carbon fiber rims, wheelsets, and bicycle components. Our skilled technicians will complete the assembly to the highest standards.
        </p>
        <p class="products-page__intro">
          Please contact our sales team at
          <a class="wheelset-link" href="mailto:support@tanzanite.site">support@tanzanite.site</a>
          to begin the quotation process, or
          <button
            type="button"
            class="wheelset-inline-button"
            @click="openWhatsAppChat"
          >
            reach us directly via online chat
          </button>
          .
        </p>
        <p class="products-page__intro">
          <strong style="color: #38bdf8;">Assembly Fee: USD 30 per set</strong>
        </p>

        <ol class="wheelset-section__list wheelset-section__list--ordered">
          <li>
            <strong>Service Process</strong>
            <p class="wheelset-safety-item">
              <strong>Shipping:</strong> Customers send their components to the designated address.
            </p>
            <p class="wheelset-safety-item">
              <strong>Confirmation:</strong> Upon receipt, we will check and confirm all items.
            </p>
            <p class="wheelset-safety-item">
              <strong>Assembly:</strong> Professional technicians perform installation and adjustment.
            </p>
            <p class="wheelset-safety-item">
              <strong>Testing:</strong> Safety and performance checks are conducted to ensure compliance with riding standards.
            </p>
            <p class="wheelset-safety-item">
              <strong>Delivery:</strong> The assembled product will be shipped back or made available for pickup.
            </p>
          </li>
          <li>
            <strong>Accepted Components</strong>
            <ul class="wheelset-section__list">
              <li>Wheelsets (including hubs)</li>
              <li>Other related accessories (please confirm with us in advance)</li>
            </ul>
          </li>
          <li>
            <strong>Important Notes</strong>
            <ul class="wheelset-section__list">
              <li>Please ensure that all components sent are new or in good condition.</li>
              <li>A detailed packing list is recommended to avoid missing items.</li>
              <li>Special customization requests should be communicated in advance.</li>
              <li>Typical assembly time is 3–7 business days, depending on actual circumstances.</li>
            </ul>
          </li>
          <li>
            <strong>Value Proposition</strong>
            <p class="wheelset-safety-item">
              <strong>Expert Technicians:</strong> Years of experience in carbon fiber wheelset and complete bike assembly.
            </p>
            <p class="wheelset-safety-item">
              <strong>Safety Assurance:</strong> Rigorous testing procedures to guarantee riding safety.
            </p>
            <p class="wheelset-safety-item">
              <strong>Efficient Delivery:</strong> Fast assembly and return shipping, saving customers valuable time.
            </p>
          </li>
        </ol>

        <p class="products-page__intro" style="margin-top: 1rem;">
          For smooth service, shipping costs for sending and returning components are the responsibility of the customer. We will assist by recommending reliable logistics options to ensure your parts arrive safely.
        </p>
        <p class="products-page__intro" style="margin-top: 1rem; color: #38bdf8;">
          If you wish to assemble your own wheels and purchase parts exclusively from TANZANITE , we are happy to assist you. We will guide you through every step of the assembly process, including spoke length calculation, assembly methods, and tire installation.
          <button
            type="button"
            class="wheelset-inline-button"
            @click="goToTechnicalSpokePattern"
          >
            Click to view documentation on spoke weaving
          </button>
          <button
            type="button"
            class="wheelset-inline-button"
            @click="goToHolePatterns"
          >
            Check out the rim weaving holes we can customize
          </button>
        </p>
      </section>

      <!-- Mixed rim -->
      <section
        v-show="activeTab === 'mixed-rim'"
        id="mixed-rim"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Mixed rim</h3>
      </section>

      <!-- Appearance Logo -->
      <section
        v-show="activeTab === 'appearance-logo'"
        id="appearance-logo"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Appearance Logo</h3>
        <ul class="wheelset-section__list">
          <li>
            <strong>Customized decals</strong>
            <p>
              Our skilled graphics department uses the latest technology to design and execute any
              custom decal or rim personalization you can imagine! Our most popular options are
              laser engraving, waterslide decals, or removable laser-cut vinyl stickers. Our team
              will work closely with you to design the perfect size, location, font, and color of
              your custom graphics. Whether you're looking for a bold statement or subtle
              customization, we'll help you achieve the look you want.
            </p>
          </li>
          <li>
            <strong>Laser engraving: Eco-friendly personalization</strong>
            <p>
              Staying true to our sustainability mission by substantially reducing our plastic
              waste, we’re ecstatic about our new laser engraved graphic option. Our engraving
              machine delivers precision designs in a sleek light gray color, all while preserving
              the structural integrity of your carbon rim and reducing the need for plastic/vinyl
              graphics.
            </p>
          </li>
        </ul>
        <div class="wheelset-appearance-images">
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO.webp"
              alt="Carbon rim with customized laser engraved logo detail"
              class="wheelset-appearance-images__image"
            />
          </div>
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO1.webp"
              alt="Close-up of carbon rim laser engraving process"
              class="wheelset-appearance-images__image"
            />
          </div>
          <div class="wheelset-appearance-images__item">
            <img
              src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO2.webp"
              alt="Different examples of carbon rim laser engraved graphics"
              class="wheelset-appearance-images__image"
            />
          </div>
        </div>
      </section>

      <!-- Optional -->
      <section
        v-show="activeTab === 'optional'"
        id="optional"
        class="wheelset-section"
      >
        <h3 class="wheelset-section__title">Optional</h3>
        <p>Here is some new content.</p>
      </section>

      <!-- FAQ Section - 放在所有 tab 内容之后 -->
      <section class="wheelset-section wheelset-faq">
        <PageFaq
          page-id="guides-wheelset-buyers"
          theme="dark"
          :show-categories="true"
        />
      </section>

      <!-- Feedback / Leave a message -->
      <section class="wheelset-feedback">
        <UserFeedbackThread
          threadKey="guides-wheelset-buyers"
          title="Share your feedback or leave a message about the Wheelset Buyers Guide"
        />
      </section>
    </div>
  </div>
  <QuickBuyModal v-if="quickOpen" :config="null" @close="quickOpen = false" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useLocalePath, useRouter } from '#imports'
import PageFaq from '~/components/PageFaq.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import QuickBuyModal from '@/components/QuickBuy.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Wheelset Buyers Guide',
})

type WheelsetTabId =
  | 'safety-instructions'
  | 'mullet-wheelsets'
  | 'custom-front-rear-wheels'
  | 'sample-assembly'
  | 'mixed-rim'
  | 'appearance-logo'
  | 'optional'

const tabs: { id: WheelsetTabId; label: string }[] = [
  { id: 'safety-instructions', label: 'Safety instructions' },
  { id: 'mullet-wheelsets', label: 'Mullet wheelsets' },
  { id: 'custom-front-rear-wheels', label: 'Custom Front & Rear Wheels' },
  { id: 'sample-assembly', label: 'Sample assembly' },
  { id: 'mixed-rim', label: 'Mixed rim' },
  { id: 'appearance-logo', label: 'Appearance Logo' },
  { id: 'optional', label: 'Optional' },
]

const activeTab = ref<WheelsetTabId>('safety-instructions')
const quickOpen = ref(false)

const setActiveTab = (id: WheelsetTabId) => {
  activeTab.value = id
}

const { openChat } = useChatWidget()
const router = useRouter()
const localePath = useLocalePath()

const openWhatsAppChat = () => {
  openChat({ showAgentList: true })
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

const goToTubelessInstallation = async () => {
  await router.push(`${localePath('/guides/sizecharts')}#installation`)
}

const goToAfterSales = async () => {
  await router.push(localePath('/support/after-sales'))
}

const goToTechnicalTension = async () => {
  await router.push(`${localePath('/guides/technical')}#tension`)
}

const goToTechnicalSpokePattern = async () => {
  await router.push(`${localePath('/guides/technical')}#spoke-pattern`)
}

const goToHolePatterns = async () => {
  await router.push(`${localePath('/company/about')}#hole-patterns`)
}

const openQuickBuy = () => {
  quickOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'wheelset-quick' } }))
  }
}
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.wheelset-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.wheelset-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.wheelset-tabs::-webkit-scrollbar {
  display: none;
}

.wheelset-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow:
    0 3px 9px -6px rgba(0, 0, 0, 0.9),
    0 0 9px rgba(0, 0, 0, 0.85);
}

.wheelset-tabs__item:active {
  transform: scale(0.96);
}

.wheelset-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.wheelset-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .wheelset-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}


.wheelset-section {
  margin-top: 0.75rem;
}

.wheelset-section__title {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.wheelset-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.88rem;
  color: rgba(148, 163, 184, 0.9);
}

.wheelset-section__list--ordered {
  list-style-type: decimal;
}

.wheelset-safety-item {
  margin: 0.25rem 0 0;
}

.wheelset-section__list ul {
  margin-top: 0.15rem;
  padding-left: 1.1rem;
  list-style-type: disc;
}

.wheelset-section__list li + li {
  margin-top: 0.25rem;
}

.wheelset-appearance-images {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.wheelset-appearance-images__item {
  flex: 1 1 260px;
  max-width: 100%;
  border-radius: 0.75rem;
  overflow: hidden;
  background: rgba(15, 23, 42, 0.9);
}

.wheelset-appearance-images__image {
  width: 100%;
  height: auto;
  display: block;
  aspect-ratio: 2 / 1;
  object-fit: cover;
}

.wheelset-link {
  color: #2dd4bf;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: #60a5fa;
}

.wheelset-inline-note {
  margin-left: 0.25rem;
  color: rgba(148, 163, 184, 0.9);
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

.wheelset-safety-footer {
  margin-top: 0.85rem;
}

@media (min-width: 768px) {
  .wheelset-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .wheelset-tabs {
    justify-content: flex-start;
  }
}
</style>
