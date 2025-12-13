<template>
  <section class="mt-10 rounded-2xl bg-slate-900/80 backdrop-blur-xl p-4 md:p-6 shadow-[0_12px_30px_-18px_rgba(0,0,0,0.98)]">
    <header class="mb-4 md:mb-6">
      <h2 class="text-lg md:text-xl font-semibold text-slate-50">
        {{ titleText }}
      </h2>
      <p v-if="subtitleText" class="mt-1 text-sm text-slate-400">
        {{ subtitleText }}
      </p>
    </header>

    <!-- Search -->
    <div v-if="showSearchComputed" class="mb-4 flex flex-col gap-2 md:flex-row md:items-center">
      <label class="text-xs font-medium uppercase tracking-wide text-slate-400">
        {{ $t('feedback.searchLabel', 'Search feedback') }}
      </label>
      <div class="flex flex-1 items-center gap-2">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('feedback.searchPlaceholder', 'Type to filter comments on this page...')"
          class="w-full rounded-lg border-none bg-slate-900/70 px-3 py-2 text-sm text-slate-50 placeholder:text-slate-500 shadow-[0_2px_6px_rgba(0,0,0,0.9)] focus:outline-none focus:ring-2 focus:ring-[#40ffaa]"
        />
      </div>
    </div>

    <!-- List -->
    <div class="space-y-3">
      <p v-if="loadingList" class="text-sm text-slate-400">
        {{ $t('feedback.loading', 'Loading feedback...') }}
      </p>
      <p v-else-if="filteredItems.length === 0" class="text-sm text-slate-400">
        {{ $t('feedback.empty', 'No feedback has been submitted for this page yet.') }}
      </p>
      <article
        v-for="item in filteredItems"
        v-else
        :key="item.id"
        class="rounded-xl border border-slate-700/60 bg-slate-900/60 px-3 py-2.5 md:px-4 md:py-3 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]"
      >
        <header class="mb-1.5 flex items-center justify-between gap-2">
          <span class="text-sm font-medium text-slate-100">
            {{ item.name || $t('feedback.anonymous', 'Member') }}
          </span>
          <span class="text-xs text-slate-500">
            {{ formatDate(item.created_at) }}
          </span>
        </header>
        <p class="text-sm leading-relaxed text-slate-200 whitespace-pre-line">
          {{ item.content }}
        </p>
      </article>
    </div>

    <!-- Divider -->
    <div class="mt-6 border-t border-white/10 pt-4 md:pt-5">
      <!-- Eligibility message -->
      <div v-if="eligibilityState && !eligibilityState.can_post" class="space-y-3">
        <p class="text-sm text-slate-300">
          {{ $t('feedback.loginRequired', 'Please sign in to leave feedback.') }}
        </p>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-full bg-white px-4 py-2 text-sm font-semibold text-black shadow-[8px_8px_22px_rgba(0,0,0,0.92)] hover:bg-white/90 hover:shadow-[10px_10px_26px_rgba(0,0,0,0.95)] transition-all"
            @click="showAuth = true"
          >
            {{ $t('feedback.loginCta', 'Sign in or create an account') }}
          </button>
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-full bg-white px-4 py-2 text-sm font-semibold text-black shadow-[8px_8px_22px_rgba(0,0,0,0.92)] hover:bg-white/90 hover:shadow-[10px_10px_26px_rgba(0,0,0,0.95)] transition-all"
            @click="openWhatsApp"
          >
            {{ $t('feedback.liveChatCta', 'To live chat') }}
          </button>
        </div>
      </div>

      <!-- Form -->
      <form v-else @submit.prevent="handleSubmit" class="space-y-3">
        <p class="text-sm text-slate-300">
          {{ $t('feedback.formIntro', 'Share your thoughts to help us improve this page.') }}
        </p>

        <div class="space-y-2">
          <label class="block text-xs font-medium text-slate-400">
            {{ $t('feedback.messageLabel', 'Your feedback') }}
          </label>
          <textarea
            v-model="message"
            rows="3"
            class="w-full rounded-lg border-none bg-slate-900/70 px-3 py-2 text-sm text-slate-50 placeholder:text-slate-500 shadow-[0_2px_6px_rgba(0,0,0,0.9)] focus:outline-none focus:ring-2 focus:ring-[#40ffaa]"
            :placeholder="$t('feedback.messagePlaceholder', 'Tell us what worked well and what could be improved...')"
          />
        </div>

        <div class="grid gap-3 md:grid-cols-2">
          <div class="space-y-1.5">
            <label class="block text-xs font-medium text-slate-400">
              {{ $t('feedback.optionalName', 'Name (optional)') }}
            </label>
            <input
              v-model="name"
              type="text"
              class="w-full rounded-lg border-none bg-slate-900/70 px-3 py-2 text-sm text-slate-50 placeholder:text-slate-500 shadow-[0_2px_6px_rgba(0,0,0,0.9)] focus:outline-none focus:ring-2 focus:ring-[#40ffaa]"
              :placeholder="$t('feedback.namePlaceholder', 'How should we address you?')"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block text-xs font-medium text-slate-400">
              {{ $t('feedback.optionalEmail', 'Email (optional, not public)') }}
            </label>
            <input
              v-model="email"
              type="email"
              class="w-full rounded-lg border-none bg-slate-900/70 px-3 py-2 text-sm text-slate-50 placeholder:text-slate-500 shadow-[0_2px_6px_rgba(0,0,0,0.9)] focus:outline-none focus:ring-2 focus:ring-[#40ffaa]"
              :placeholder="$t('feedback.emailPlaceholder', 'For follow-up only, never shared publicly.')"
            />
          </div>
        </div>

        <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <button
            type="submit"
            class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] px-5 py-2.5 text-sm font-semibold text-black shadow-md hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loadingSubmit || !message.trim()"
          >
            <span v-if="loadingSubmit">
              {{ $t('feedback.submitting', 'Submitting...') }}
            </span>
            <span v-else>
              {{ $t('feedback.submit', 'Submit feedback') }}
            </span>
          </button>

          <p v-if="submitMessage" class="text-xs text-emerald-400">
            {{ submitMessage }}
          </p>
          <p v-else-if="submitError" class="text-xs text-red-400">
            {{ submitError }}
          </p>
        </div>
      </form>
    </div>

    <!-- Auth modal -->
    <AuthModal v-model="showAuth" @success="onAuthSuccess" />
    <WhatsAppChatModal
      v-if="showWhatsApp"
      :conversation="{ showAgentList: true }"
      @close="showWhatsApp = false"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import AuthModal from '~/components/AuthModal.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'
import { useFeedback } from '~/composables/useFeedback'

const props = defineProps<{
  threadKey: string
  title?: string
  subtitle?: string
  showSearch?: boolean
}>()

const { t: $t } = useI18n()

const {
  items,
  loadingList,
  loadingSubmit,
  error,
  search,
  eligibility,
  fetchList,
  submitFeedback,
  loadEligibility,
} = useFeedback(props.threadKey)

const searchQuery = ref('')

const filteredItems = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((item) => item.content.toLowerCase().includes(q))
})

const message = ref('')
const name = ref('')
const email = ref('')
const submitMessage = ref('')
const submitError = ref('')
const showAuth = ref(false)
const showWhatsApp = ref(false)

const titleText = computed(
  () => props.title || $t('feedback.defaultTitle', 'Share your feedback')
)

const subtitleText = computed(
  () =>
    props.subtitle ||
    $t(
      'feedback.defaultSubtitle',
      'Help us improve this page and the tools you use every day.'
    )
)

const showSearchComputed = computed(() => props.showSearch !== false)

const eligibilityState = computed(() => eligibility.value)

const formatDate = (iso: string) => {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

onMounted(async () => {
  await Promise.all([fetchList(1), loadEligibility()])
})

const handleSubmit = async () => {
  submitError.value = ''
  submitMessage.value = ''

  const content = message.value.trim()
  if (!content) {
    submitError.value = $t('feedback.required', 'Please enter your feedback before submitting.')
    return
  }

  const result = await submitFeedback({
    content,
    name: name.value || undefined,
    email: email.value || undefined,
  })

  if (!result.success) {
    submitError.value =
      (result.error && (result.error as any).message) ||
      $t('feedback.submitFailed', 'Failed to submit feedback. Please try again.')
    return
  }

  message.value = ''
  name.value = ''
  email.value = ''
  submitMessage.value =
    result.message ||
    $t('feedback.pendingMessage', 'Submitted successfully. Your feedback will appear after it is reviewed.')
}

const onAuthSuccess = async () => {
  showAuth.value = false
  await loadEligibility()
}

const openWhatsApp = () => {
  showWhatsApp.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}
</script>
