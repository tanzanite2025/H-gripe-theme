<template>
  <section class="feedback-card">
    <div class="feedback-card__header">
      <div>
        <p class="feedback-card__eyebrow">{{ t('feedbackForm.eyebrow') }}</p>
        <h2 class="feedback-card__title">{{ t('feedbackForm.title') }}</h2>
        <p class="feedback-card__subtitle">
          {{ t('feedbackForm.subtitle') }}
        </p>
      </div>
      <div class="feedback-card__stats">
        <div>
          <p class="feedback-card__stat-label">{{ t('feedbackForm.stat.responses') }}</p>
          <p class="feedback-card__stat-value">{{ t('feedbackForm.stat.responsesValue') }}</p>
        </div>
        <div>
          <p class="feedback-card__stat-label">{{ t('feedbackForm.stat.reviewTime') }}</p>
          <p class="feedback-card__stat-value">{{ t('feedbackForm.stat.reviewTimeValue') }}</p>
        </div>
      </div>
    </div>

    <form class="feedback-form" @submit.prevent="handleSubmit">
      <div class="feedback-form__grid">
        <label class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.fullName') }}</span>
          <input
            v-model.trim="form.fullName"
            type="text"
            :placeholder="t('feedbackForm.placeholders.fullName')"
            autocomplete="name"
          />
        </label>

        <label class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.email') }}</span>
          <input
            v-model.trim="form.email"
            type="email"
            :placeholder="t('feedbackForm.placeholders.email')"
            autocomplete="email"
          />
        </label>

        <label class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.country') }}</span>
          <select v-model="form.country">
            <option value="">{{ t('feedbackForm.placeholders.country') }}</option>
            <option
              v-for="country in countryOptions"
              :key="country.value"
              :value="country.value"
            >
              {{ country.label }}
            </option>
          </select>
        </label>

        <label class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.orderNumber') }}</span>
          <input
            v-model.trim="form.orderNumber"
            type="text"
            :placeholder="t('feedbackForm.placeholders.orderNumber')"
            autocomplete="off"
          />
        </label>
      </div>

      <div class="feedback-form__grid">
        <label class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.productCategory') }}</span>
          <select v-model="form.productCategory">
            <option value="">{{ t('feedbackForm.placeholders.productCategory') }}</option>
            <option
              v-for="category in productCategories"
              :key="category.value"
              :value="category.value"
            >
              {{ category.label }}
            </option>
          </select>
        </label>

        <div class="feedback-form__field">
          <span>{{ t('feedbackForm.fields.requestType') }}</span>
          <div class="feedback-form__pills">
            <button
              v-for="type in requestTypes"
              :key="type.value"
              type="button"
              class="feedback-pill"
              :class="{
                'feedback-pill--active': form.requestType === type.value,
              }"
              @click="form.requestType = type.value"
            >
              {{ type.label }}
            </button>
          </div>
        </div>
      </div>

      <label class="feedback-form__field">
        <div class="feedback-form__label-row">
          <span>{{ t('feedbackForm.fields.message') }}</span>
          <span class="feedback-form__char-counter">
            {{ t('feedbackForm.charactersLeft', { count: messageCharactersLeft }) }}
          </span>
        </div>
        <textarea
          v-model.trim="form.message"
          :placeholder="t('feedbackForm.placeholders.message')"
          :maxlength="messageMaxLength"
          rows="6"
        />
      </label>

      <div v-if="props.showAttachments" class="feedback-form__field">
        <div class="feedback-form__label-row">
          <span>{{ t('feedbackForm.fields.attachments') }}</span>
          <span class="feedback-form__tag">{{ t('feedbackForm.fields.membersOnly') }}</span>
        </div>
        <div
          class="feedback-upload"
          role="button"
          tabindex="0"
          :class="{ 'feedback-upload--disabled': !allowAttachments }"
          @keydown.enter.prevent="allowAttachments && triggerFileInput()"
          @click="allowAttachments && triggerFileInput()"
        >
          <div>
            <p>{{ t('feedbackForm.upload.prompt') }}</p>
            <p>{{ attachmentHint }}</p>
          </div>
          <input
            ref="fileInputRef"
            class="feedback-upload__input"
            type="file"
            accept="image/*"
            multiple
            @change="handleFileSelect"
          />
        </div>
        <ul v-if="form.attachments.length" class="feedback-upload__list">
          <li v-for="(file, index) in form.attachments" :key="file.name">
            <span>{{ file.name }}</span>
            <button type="button" @click="removeAttachment(index)">
              {{ t('feedbackForm.actions.removeAttachment') }}
            </button>
          </li>
        </ul>
      </div>

      <label class="feedback-form__consent">
        <input v-model="form.consent" type="checkbox" />
        <span>{{ t('feedbackForm.fields.consent') }}</span>
      </label>

      <div class="feedback-form__actions">
        <button
          type="submit"
          class="feedback-form__submit"
          :disabled="isSubmitting"
        >
          <span v-if="isSubmitting" class="feedback-form__spinner" />
          <span v-else>{{ t('feedbackForm.actions.submit') }}</span>
        </button>
        <button type="button" class="feedback-form__secondary" :disabled="isSubmitting" @click="resetForm">
          {{ t('feedbackForm.actions.reset') }}
        </button>
      </div>

      <p v-if="infoMessage" class="feedback-form__info">{{ infoMessage }}</p>
      <p v-else-if="errorMessage" class="feedback-form__info feedback-form__info--error">{{ errorMessage }}</p>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, watch } from 'vue'
import { useI18n } from '#imports'
import { useSuggestionFeedback } from '~/composables/useSuggestionFeedback'

const props = withDefaults(
  defineProps<{
    threadKey?: string
    showAttachments?: boolean
  }>(),
  {
    threadKey: 'product_service_suggestion',
    showAttachments: true,
  }
)

const { t } = useI18n()
const threadKeyRef = computed(() => props.threadKey || 'product_service_suggestion')
const {
  eligibility,
  isSubmitting,
  errorMessage,
  successMessage,
  loadEligibility,
  submitSuggestion,
} = useSuggestionFeedback(threadKeyRef)

onMounted(() => {
  loadEligibility()
})

watch(threadKeyRef, () => {
  loadEligibility()
})

const messageMaxLength = 1000
const infoMessage = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const form = reactive({
  fullName: '',
  email: '',
  country: '',
  orderNumber: '',
  productCategory: '',
  requestType: 'product',
  message: '',
  consent: false,
  attachments: [] as File[],
})

const countryOptions = [
  { value: 'us', label: 'United States' },
  { value: 'cn', label: '中国' },
  { value: 'ca', label: 'Canada' },
  { value: 'de', label: 'Deutschland' },
]

const productCategories = [
  { value: 'mtb', label: t('feedbackForm.productCategories.mtb') },
  { value: 'road', label: t('feedbackForm.productCategories.road') },
  { value: 'gravel', label: t('feedbackForm.productCategories.gravel') },
  { value: 'accessories', label: t('feedbackForm.productCategories.accessories') },
  { value: 'other', label: t('feedbackForm.productCategories.other') },
]

const requestTypes = [
  { value: 'product', label: t('feedbackForm.requestTypes.product') },
  { value: 'service', label: t('feedbackForm.requestTypes.service') },
  { value: 'logistics', label: t('feedbackForm.requestTypes.logistics') },
  { value: 'warranty', label: t('feedbackForm.requestTypes.warranty') },
  { value: 'other', label: t('feedbackForm.requestTypes.other') },
]

const allowAttachments = computed(() => {
  if (!props.showAttachments) return false
  const e = eligibility.value
  return !!(e && e.loggedIn && e.canAttach)
})

const attachmentHint = computed(() => {
  const e = eligibility.value
  if (!e) return t('feedbackForm.upload.limit')
  if (!e.loggedIn) {
    return e.reason || t('feedbackForm.messages.loginRequired')
  }
  if (!e.canAttach) {
    return e.reason || t('feedbackForm.messages.membersOnly')
  }
  return t('feedbackForm.upload.limit')
})

const messageCharactersLeft = computed(() => {
  return Math.max(messageMaxLength - form.message.length, 0)
})

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (!target.files?.length) return

  form.attachments = Array.from(target.files)
}

const removeAttachment = (index: number) => {
  form.attachments.splice(index, 1)
}

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const mapAttachmentsForPayload = () => {
  if (!allowAttachments.value) {
    return []
  }
  return form.attachments.map(file => ({
    name: file.name,
    url: '',
    size: file.size,
  }))
}

const handleSubmit = async () => {
  if (isSubmitting.value) return
  infoMessage.value = ''

  try {
    await submitSuggestion({
      fullName: form.fullName,
      email: form.email,
      country: form.country,
      orderNumber: form.orderNumber,
      productCategory: form.productCategory,
      requestType: form.requestType,
      message: form.message,
      attachments: mapAttachmentsForPayload(),
      threadKey: threadKeyRef.value,
    })

    infoMessage.value = successMessage.value || t('feedbackForm.messages.submitted')
    resetForm()
    loadEligibility()
  } catch (error: any) {
    infoMessage.value = error?.message || errorMessage.value || t('feedbackForm.messages.submitError')
  }
}

const resetForm = () => {
  form.fullName = ''
  form.email = ''
  form.country = ''
  form.orderNumber = ''
  form.productCategory = ''
  form.requestType = 'product'
  form.message = ''
  form.consent = false
  form.attachments = []
  infoMessage.value = successMessage.value || ''
  fileInputRef.value && (fileInputRef.value.value = '')
}
</script>

<style scoped>
.feedback-card {
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(2, 6, 23, 0.98));
  border-radius: 24px;
  padding: 2rem;
  box-shadow: 0 18px 35px -22px rgba(0, 0, 0, 0.95);
  position: relative;
  overflow: hidden;
}

.feedback-card__header {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

@media (min-width: 768px) {
  .feedback-card__header {
    flex-direction: row;
    justify-content: space-between;
    align-items: flex-start;
  }
}

.feedback-card__eyebrow {
  font-size: 0.85rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(129, 140, 248, 0.9);
  margin-bottom: 0.5rem;
}

.feedback-card__title {
  font-size: clamp(1.5rem, 2vw, 2rem);
  font-weight: 700;
  color: #f8fafc;
  margin: 0 0 0.25rem;
}

.feedback-card__subtitle {
  color: rgba(148, 163, 184, 0.9);
  margin: 0;
}

.feedback-card__stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(120px, 1fr));
  gap: 1rem;
  background: rgba(2, 6, 23, 0.85);
  border-radius: 16px;
  padding: 1rem;
  box-shadow: inset 0 0 18px rgba(0, 0, 0, 0.65);
}

.feedback-card__stat-label {
  font-size: 0.85rem;
  color: rgba(148, 163, 184, 0.8);
  margin: 0;
}

.feedback-card__stat-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: #e0e7ff;
  margin: 0;
}

.feedback-form {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.feedback-form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.feedback-form__field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  font-size: 0.95rem;
  color: #e2e8f0;
}

.feedback-form__field input,
.feedback-form__field select,
.feedback-form__field textarea {
  background: radial-gradient(circle at top left, rgba(14, 20, 40, 0.92), rgba(2, 6, 23, 0.95));
  border: none;
  border-radius: 14px;
  padding: 0.85rem 1rem;
  color: #f1f5f9;
  font-size: 0.95rem;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05), 0 8px 14px -12px rgba(0, 0, 0, 0.95);
  transition: box-shadow 0.2s ease;
}

.feedback-form__field input:focus,
.feedback-form__field select:focus,
.feedback-form__field textarea:focus {
  outline: none;
  box-shadow:
    inset 0 0 0 1px rgba(99, 102, 241, 0.65),
    0 0 18px rgba(99, 102, 241, 0.35);
}

.feedback-form__field textarea {
  resize: vertical;
  min-height: 160px;
}

.feedback-form__label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.feedback-form__char-counter {
  font-size: 0.8rem;
  color: rgba(148, 163, 184, 0.8);
}

.feedback-form__pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.feedback-pill {
  padding: 0.45rem 1rem;
  border-radius: 9999px;
  border: none;
  background: radial-gradient(circle at top left, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.92));
  color: rgba(226, 232, 240, 0.85);
  font-size: 0.85rem;
  box-shadow: 0 4px 8px -6px rgba(0, 0, 0, 0.95);
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease;
}

.feedback-pill--active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.38), rgba(14, 165, 233, 0.35));
  color: #f0f9ff;
  transform: translateY(-1px);
}

.feedback-pill:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.55);
}

.feedback-form__tag {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(248, 250, 252, 0.9);
  background: rgba(77, 124, 15, 0.35);
  border-radius: 9999px;
  padding: 0.15rem 0.75rem;
}

.feedback-upload {
  border: none;
  border-radius: 20px;
  padding: 1.5rem;
  background: radial-gradient(circle at top left, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.92));
  color: rgba(148, 163, 184, 0.9);
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.03), 0 14px 24px -18px rgba(0, 0, 0, 0.95);
}

.feedback-upload--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.feedback-upload__list button {
  background: none;
  border: none;
}

.feedback-upload__list {
  list-style: none;
  padding: 0;
  margin: 0.75rem 0 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  color: rgba(226, 232, 240, 0.85);
}

.feedback-upload__list button {
  background: none;
  border: none;
  color: rgba(248, 113, 113, 0.9);
  cursor: pointer;
}

.feedback-form__consent {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: rgba(203, 213, 225, 0.85);
}

.feedback-form__consent input {
  width: 1.1rem;
  height: 1.1rem;
}

.feedback-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.feedback-form__submit {
  background: linear-gradient(135deg, #22d3ee, #818cf8);
  border: none;
  border-radius: 9999px;
  padding: 0.85rem 2.25rem;
  color: #020617;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 180px;
}

.feedback-form__submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.feedback-form__secondary {
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
  border: none;
  border-radius: 9999px;
  padding: 0.85rem 1.75rem;
  color: rgba(226, 232, 240, 0.9);
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.95);
}

.feedback-form__info {
  font-size: 0.85rem;
  color: rgba(148, 163, 184, 0.9);
}

.feedback-form__info--error {
  color: rgba(248, 113, 113, 0.95);
}

.feedback-form__spinner {
  width: 1rem;
  height: 1rem;
  border-radius: 9999px;
  border: 2px solid rgba(2, 6, 23, 0.3);
  border-top-color: rgba(2, 6, 23, 0.8);
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
