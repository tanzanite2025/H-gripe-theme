<template>
  <section class="account-tab-panel">
    <div class="tab-head">
      <div>
        <p>{{ t('accountSidebar.addresses.eyebrow', 'Shipping details') }}</p>
        <h3>{{ t('accountSidebar.addresses.title', 'Address') }}</h3>
      </div>
    </div>

    <div class="address-boundary">
      <Icon name="lucide:info" />
      <p>
        {{ t('accountSidebar.addresses.boundary', 'This edits the current checkout shipping address. A persistent account address book should connect to one backend source later.') }}
      </p>
    </div>

    <form class="address-form" @submit.prevent="saveAddress">
      <label>
        <span>{{ t('checkout.fields.name', 'Name') }}</span>
        <input v-model.trim="form.name" autocomplete="name" />
      </label>
      <label>
        <span>{{ t('checkout.fields.phone', 'Phone') }}</span>
        <input v-model.trim="form.phone" autocomplete="tel" />
      </label>
      <label class="address-form__wide">
        <span>{{ t('checkout.fields.address', 'Address') }}</span>
        <input v-model.trim="form.address" autocomplete="street-address" />
      </label>
      <label>
        <span>{{ t('checkout.fields.city', 'City') }}</span>
        <input v-model.trim="form.city" autocomplete="address-level2" />
      </label>
      <label>
        <span>{{ t('checkout.fields.state', 'State') }}</span>
        <input v-model.trim="form.state" autocomplete="address-level1" />
      </label>
      <label>
        <span>{{ t('checkout.fields.zip', 'ZIP') }}</span>
        <input v-model.trim="form.zip" autocomplete="postal-code" />
      </label>
      <label>
        <span>{{ t('checkout.fields.country', 'Country') }}</span>
        <input v-model.trim="form.country" autocomplete="country-name" />
      </label>

      <div class="address-form__actions">
        <button type="submit">
          {{ t('accountSidebar.addresses.saveToCheckout', 'Use for checkout') }}
        </button>
        <button type="button" class="address-form__ghost" @click="checkoutNow">
          {{ t('cartDrawer.actions.checkout', 'Checkout') }} →
        </button>
      </div>
    </form>

    <p v-if="savedMessage" class="account-toast">
      {{ savedMessage }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useCart, type ShippingAddress } from '~/composables/useCart'

const emit = defineEmits<{
  (event: 'close'): void
}>()

const { t } = useI18n()
const auth = useAuth()
const { shippingAddress, setShippingAddress, openCheckout } = useCart()

const emptyAddress = (): ShippingAddress => ({
  name: '',
  phone: '',
  address: '',
  city: '',
  state: '',
  zip: '',
  country: '',
})

const form = reactive<ShippingAddress>(emptyAddress())
const savedMessage = ref('')

const fillForm = (source: Partial<ShippingAddress> | null | undefined) => {
  const profile = auth.user.value?.profile || {}
  form.name = source?.name || profile.fullName || ''
  form.phone = source?.phone || profile.phone || ''
  form.address = source?.address || ''
  form.city = source?.city || ''
  form.state = source?.state || ''
  form.zip = source?.zip || ''
  form.country = source?.country || profile.country || ''
}

const flashSaved = () => {
  savedMessage.value = t('accountSidebar.addresses.saved', 'Checkout shipping address updated')
  if (typeof window === 'undefined') return
  window.setTimeout(() => {
    savedMessage.value = ''
  }, 2200)
}

const saveAddress = () => {
  setShippingAddress({ ...form })
  flashSaved()
}

const checkoutNow = () => {
  saveAddress()
  openCheckout()
  emit('close')
}

watch(
  () => shippingAddress.value,
  (address) => fillForm(address),
  { immediate: true },
)
</script>

<style scoped>
.account-tab-panel {
  display: flex;
  min-height: 0;
  flex-direction: column;
  gap: 0.85rem;
}

.tab-head p,
.tab-head h3 {
  margin: 0;
}

.tab-head p {
  color: rgba(103, 232, 249, 0.9);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}

.tab-head h3 {
  margin-top: 0.18rem;
  color: #ffffff;
  font-size: 1.05rem;
  font-weight: 850;
}

.address-boundary {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.6rem;
  border-radius: 1rem;
  background: rgba(96, 165, 250, 0.11);
  padding: 0.75rem;
}

.address-boundary :deep(svg) {
  width: 1rem;
  height: 1rem;
  color: #67e8f9;
  margin-top: 0.1rem;
}

.address-boundary p {
  margin: 0;
  color: rgba(226, 232, 240, 0.8);
  font-size: 0.74rem;
  line-height: 1.45;
}

.address-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.72rem;
}

.address-form label {
  min-width: 0;
}

.address-form span {
  display: block;
  margin-bottom: 0.32rem;
  color: rgba(226, 232, 240, 0.78);
  font-size: 0.72rem;
  font-weight: 750;
}

.address-form input {
  width: 100%;
  min-height: 2.35rem;
  border: 1px solid rgba(255, 255, 255, 0.13);
  border-radius: 0.8rem;
  background: rgba(255, 255, 255, 0.065);
  color: #ffffff;
  outline: none;
  padding: 0 0.72rem;
  font-size: 0.82rem;
}

.address-form input:focus {
  border-color: rgba(64, 255, 170, 0.7);
  box-shadow: 0 0 0 3px rgba(64, 255, 170, 0.12);
}

.address-form__wide,
.address-form__actions {
  grid-column: 1 / -1;
}

.address-form__actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
  margin-top: 0.2rem;
}

.address-form__actions button {
  min-height: 2.55rem;
  border-radius: 999px;
  background: linear-gradient(135deg, #4efce7, #60a5fa);
  color: #020617;
  font-size: 0.8rem;
  font-weight: 900;
}

.address-form__actions .address-form__ghost {
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.07);
  color: #ffffff;
}

.account-toast {
  margin: 0;
  border-radius: 999px;
  background: rgba(64, 255, 170, 0.12);
  color: #bfffe3;
  padding: 0.6rem 0.8rem;
  text-align: center;
  font-size: 0.75rem;
  font-weight: 800;
}

@media (max-width: 520px) {
  .address-form {
    grid-template-columns: 1fr;
  }

  .address-form__actions {
    grid-template-columns: 1fr;
  }
}
</style>
