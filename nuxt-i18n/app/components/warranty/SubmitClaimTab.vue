<template>
  <section id="submit-warranty" class="support-section">
    <h2 class="support-section__title text-center mb-6 tz-text-primary">Submit Warranty Claim</h2>
    
    <div class="w-full max-w-none">
      <div v-if="submitMessage" :class="['p-4 rounded mb-6 text-center', submitStatus === 'success' ? 'bg-green-500/20 text-green-200 border border-green-500/30' : 'bg-red-500/20 text-red-200 border border-red-500/30']">
        {{ submitMessage }}
      </div>

      <form @submit.prevent="submitClaim" class="space-y-4">
        <TurnstileChallenge ref="turnstileChallenge" action="warranty" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Order Number -->
          <div>
            <label class="block tz-text-primary text-sm font-bold mb-1">Order Number</label>
            <input 
              v-model="form.order_number" 
              type="text" 
              required
              :disabled="!isFormLocked"
              class="w-full tz-surface-panel border tz-border-subtle rounded px-3 py-2 tz-text-primary focus:outline-none focus:border-emerald-500 transition-colors disabled:opacity-50"
              placeholder="e.g. TANZ-12345"
            />
          </div>

          <!-- Email -->
          <div>
            <label class="block tz-text-primary text-sm font-bold mb-1">Email Address</label>
            <input 
              v-model="form.email" 
              type="email" 
              required
              :disabled="!isFormLocked"
              class="w-full tz-surface-panel border tz-border-subtle rounded px-3 py-2 tz-text-primary focus:outline-none focus:border-emerald-500 transition-colors disabled:opacity-50"
              placeholder="Your email address"
            />
          </div>
        </div>

        <!-- Verify Button -->
        <div v-if="isFormLocked" class="flex justify-end">
          <button 
            type="button"
            @click="verifyOrder"
            :disabled="isVerifying || !form.order_number || !form.email"
            class="bg-[var(--tz-action-primary)] hover:bg-[var(--tz-action-primary-hover)] text-white font-bold py-2 px-6 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="isVerifying" class="flex items-center gap-2">
              <svg class="animate-spin h-4 w-4 text-white" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
              Verifying...
            </span>
            <span v-else>Verify Order</span>
          </button>
        </div>

        <div v-if="!isFormLocked" class="bg-green-500/10 border border-green-500/20 text-green-400 px-4 py-2 rounded text-sm flex items-center gap-2">
           <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
           Order Verified. Please complete the details below.
        </div>

        <!-- Locked Section: Details & Uploads -->
        <div :class="['space-y-4 transition-all duration-500', isFormLocked ? 'opacity-40 grayscale pointer-events-none select-none filter blur-[1px]' : 'opacity-100']">
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Tire Pressure -->
          <div>
            <label class="block tz-text-primary text-sm font-bold mb-1">Tire Pressure (PSI)</label>
            <input 
              v-model="form.tire_pressure" 
              type="text" 
              class="w-full tz-surface-panel border tz-border-subtle rounded px-3 py-2 tz-text-primary focus:outline-none focus:border-emerald-500 transition-colors"
              placeholder="e.g. 80"
            />
          </div>

          <!-- Tubeless -->
          <div>
            <label class="block tz-text-primary text-sm font-bold mb-1">Tubeless Setup?</label>
            <div class="flex items-center space-x-6 mt-2">
               <label class="flex items-center cursor-pointer">
                 <input type="radio" v-model="form.is_tubeless" value="yes" class="form-radio text-emerald-600 tz-surface-panel tz-border-subtle focus:ring-emerald-600/30">
                 <span class="ml-2 tz-text-secondary">Yes</span>
               </label>
               <label class="flex items-center cursor-pointer">
                 <input type="radio" v-model="form.is_tubeless" value="no" class="form-radio text-emerald-600 tz-surface-panel tz-border-subtle focus:ring-emerald-600/30">
                 <span class="ml-2 tz-text-secondary">No</span>
               </label>
            </div>
          </div>
        </div>

        <!-- Description -->
        <div>
          <label class="block tz-text-primary text-sm font-bold mb-1">Issue Description</label>
          <textarea 
            v-model="form.issue_description" 
            rows="4"
            required
            class="w-full tz-surface-panel border tz-border-subtle rounded px-3 py-2 tz-text-primary focus:outline-none focus:border-emerald-500 transition-colors"
            placeholder="Please describe the issue in detail..."
          ></textarea>
        </div>

        <!-- Images -->
        <div>
          <label class="block tz-text-primary text-sm font-bold mb-1">Upload Images</label>
          <div class="flex items-center gap-3">
            <button 
              type="button" 
              @click="triggerImageUpload"
              class="tz-surface-panel hover:tz-surface-subtle text-emerald-600 font-semibold py-1.5 px-4 rounded text-sm transition-colors"
            >
              Choose Files
            </button>
            <span class="tz-text-muted text-sm">
              {{ imageFiles.length > 0 ? `${imageFiles.length} file(s) selected` : 'No file chosen' }}
            </span>
          </div>
          <input 
            ref="imageInput"
            type="file" 
            multiple 
            :accept="uploadSpecAccept('warranty_evidence')"
            @change="handleImages"
            class="hidden"
          />
          <p class="mt-1 text-xs tz-text-muted">{{ uploadSpecHint('warranty_evidence') }}</p>
        </div>

        <!-- Video -->
        <div>
          <label class="block tz-text-primary text-sm font-bold mb-1">Upload Video (Optional)</label>
          <div class="flex items-center gap-3">
            <button 
              type="button" 
              @click="triggerVideoUpload"
              class="tz-surface-panel hover:tz-surface-subtle text-emerald-600 font-semibold py-1.5 px-4 rounded text-sm transition-colors"
            >
              Choose File
            </button>
            <span class="tz-text-muted text-sm">
              {{ videoFile ? videoFile.name : 'No file chosen' }}
            </span>
          </div>
          <input 
            ref="videoInput"
            type="file" 
            accept=".mp4,.mov,.webm,video/mp4,video/quicktime,video/webm" 
            @change="handleVideo"
            class="hidden"
          />
          <p class="mt-1 text-xs tz-text-muted">Optional. MP4, MOV or WebM. Max 50MB. For larger videos, provide a link in the description.</p>
        </div>

        <!-- Submit Button -->
        <div class="pt-2">
          <button 
            type="submit" 
            :disabled="isSubmitting || isFormLocked"
            class="w-full bg-[var(--tz-action-primary)] hover:bg-[var(--tz-action-primary-hover)] text-white font-bold py-2.5 px-4 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-slate-900/20"
          >
            <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
              <svg class="animate-spin h-5 w-5 text-white" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l-3-2.647z"></path></svg>
              Processing...
            </span>
            <span v-else>Submit Claim</span>
          </button>
        </div>
        </div> <!-- End Locked Section -->
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import { useAuth } from '~/composables/useAuth'
import {
  uploadSpecAccept,
  uploadSpecHint,
  validateStorefrontUploadFiles,
} from '~/utils/uploadSpecs'

const auth = useAuth()

const form = ref({
  order_number: '',
  email: '',
  verification_token: '',
  tire_pressure: '',
  is_tubeless: 'no',
  issue_description: '',
})

const imageFiles = ref<File[]>([])
const videoFile = ref<File | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const isSubmitting = ref(false)
const isVerifying = ref(false)
const isFormLocked = ref(true)
const submitMessage = ref('')
const submitStatus = ref<'success' | 'error' | ''>('')
const route = useRoute()
const turnstileChallenge = ref<{ execute: () => Promise<string> } | null>(null)

const MAX_IMAGE_FILES = 10
const MAX_VIDEO_SIZE = 50 * 1024 * 1024
const ALLOWED_VIDEO_EXTENSIONS = ['.mp4', '.mov', '.webm']

const formatUploadSize = (bytes: number) => `${Math.round(bytes / 1024 / 1024)}MB`

const fileExtension = (file: File) => {
  const name = file.name || ''
  const dotIndex = name.lastIndexOf('.')
  return dotIndex >= 0 ? name.slice(dotIndex).toLowerCase() : ''
}

const hasAllowedExtension = (file: File, extensions: string[]) => extensions.includes(fileExtension(file))

const setUploadError = (message: string) => {
  submitStatus.value = 'error'
  submitMessage.value = message
}

const clearUploadError = () => {
  if (submitStatus.value === 'error') {
    submitStatus.value = ''
    submitMessage.value = ''
  }
}

const validateImages = async (files: File[]) => {
  const result = await validateStorefrontUploadFiles(files, 'warranty_evidence')
  return result.ok ? '' : result.error || `You can upload up to ${MAX_IMAGE_FILES} images.`
}

const validateVideo = (file: File | null) => {
  if (!file) return ''
  if (!hasAllowedExtension(file, ALLOWED_VIDEO_EXTENSIONS)) {
    return `${file.name} is not supported. Use MP4, MOV or WebM.`
  }
  if (file.size > MAX_VIDEO_SIZE) {
    return `${file.name} exceeds the ${formatUploadSize(MAX_VIDEO_SIZE)} video limit.`
  }
  return ''
}

const validateUploads = async () => (await validateImages(imageFiles.value)) || validateVideo(videoFile.value)

const verifyEmailToken = async (token: string) => {
  const normalizedToken = token.trim()
  if (!normalizedToken) return

  isVerifying.value = true
  submitMessage.value = ''
  submitStatus.value = ''

  try {
    await auth.request<{ verified: boolean }>(
      `/warranty/verify/${encodeURIComponent(normalizedToken)}`,
      {
        method: 'GET',
        headers: {
          accept: 'application/json',
        },
      },
      'Warranty email verification failed'
    )

    form.value.verification_token = normalizedToken
    isFormLocked.value = false
    submitStatus.value = 'success'
    submitMessage.value = 'Email verified. Please complete and submit your warranty claim.'
  } catch (err: unknown) {
    console.error(err)
    submitStatus.value = 'error'
    submitMessage.value = err instanceof Error ? err.message : 'The verification link is invalid or expired.'
  } finally {
    isVerifying.value = false
  }
}

const verifyOrder = async () => {
  isVerifying.value = true
  submitMessage.value = ''
  submitStatus.value = ''

  try {
    const captchaToken = await turnstileChallenge.value?.execute()
    await auth.request<{ success: boolean; message?: string }>(
      '/warranty/verify-order',
      {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          order_number: form.value.order_number,
          email: form.value.email,
          captcha_token: captchaToken || '',
        }),
      },
      'Order verification failed'
    )

    submitStatus.value = 'success'
    submitMessage.value = 'Please check your email and open the verification link before completing the claim.'

  } catch (err: unknown) {
    console.error(err)
    submitStatus.value = 'error'
    submitMessage.value = err instanceof Error ? err.message : 'Order verification failed. Please check your details.'
  } finally {
    isVerifying.value = false
  }
}

const triggerImageUpload = () => {
  imageInput.value?.click()
}

const triggerVideoUpload = () => {
  videoInput.value?.click()
}

const handleImages = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files) {
    const files = Array.from(target.files)
    const error = await validateImages(files)
    if (error) {
      imageFiles.value = []
      target.value = ''
      setUploadError(error)
      return
    }

    imageFiles.value = files
    clearUploadError()
  }
}

const readVerificationToken = () => {
  const rawToken = route.query.verification_token
  const token = Array.isArray(rawToken) ? rawToken[0] : rawToken
  if (typeof token === 'string' && token.trim()) {
    void verifyEmailToken(token)
  }
}

const handleVideo = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files.item(0)
    if (!file) return
    const error = validateVideo(file)
    if (error) {
      videoFile.value = null
      target.value = ''
      setUploadError(error)
      return
    }

    videoFile.value = file
    clearUploadError()
  }
}

const submitClaim = async () => {
  isSubmitting.value = true
  submitMessage.value = ''
  submitStatus.value = ''

  try {
    const captchaToken = await turnstileChallenge.value?.execute()
    const uploadError = await validateUploads()
    if (uploadError) {
      setUploadError(uploadError)
      return
    }

    const formData = new FormData()
    formData.append('order_number', form.value.order_number)
    formData.append('email', form.value.email)
    formData.append('verification_token', form.value.verification_token)
    formData.append('captcha_token', captchaToken || '')
    formData.append('tire_pressure', form.value.tire_pressure)
    formData.append('is_tubeless', form.value.is_tubeless === 'yes' ? 'yes' : 'no')
    formData.append('issue_description', form.value.issue_description)

    imageFiles.value.forEach((file) => {
      formData.append('images[]', file)
    })

    if (videoFile.value) {
      formData.append('video', videoFile.value)
    }

    const response = await auth.request<{ success: boolean; message?: string; id?: number }>(
      '/warranty/claim',
      {
        method: 'POST',
        body: formData,
      },
      'Submission failed'
    )

    submitStatus.value = 'success'
    submitMessage.value = response.message || 'Your claim has been submitted successfully. We will contact you shortly.'
    
    form.value = {
      order_number: '',
      email: '',
      verification_token: '',
      tire_pressure: '',
      is_tubeless: 'no',
      issue_description: '',
    }
    imageFiles.value = []
    videoFile.value = null
    isFormLocked.value = true
    if (imageInput.value) imageInput.value.value = ''
    if (videoInput.value) videoInput.value.value = ''

  } catch (err: unknown) {
    console.error(err)
    submitStatus.value = 'error'
    submitMessage.value = err instanceof Error ? err.message : 'An error occurred. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}

onMounted(readVerificationToken)

watch(
  () => route.query.verification_token,
  () => readVerificationToken()
)
</script>
