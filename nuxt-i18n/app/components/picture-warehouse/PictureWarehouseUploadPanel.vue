<template>
  <section class="mt-10 border-t tz-border-subtle pt-8">
    <div class="picture-upload-card mx-auto max-w-3xl rounded-2xl px-4 py-4 sm:px-6 sm:py-5">
      <div class="mb-4 text-center">
        <h2 class="text-sm font-semibold tz-text-primary">
          {{ t('resourcesPictureWarehouseShared.upload.title') }}
        </h2>
        <p class="tz-caption tz-text-secondary">
          {{ t('resourcesPictureWarehouseShared.upload.intro') }}
        </p>
      </div>

      <form class="space-y-3" @submit.prevent="emit('submit')">
        <div class="flex flex-col gap-1">
          <label class="tz-caption tz-text-secondary">
            {{ t('resourcesPictureWarehouseShared.upload.order') }}
            <span class="text-red-400">*</span>
          </label>
          <select
            :value="props.uploadOrderId"
            class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
            :disabled="props.uploadOrdersLoading || props.uploading || !props.uploadOrders.length"
            required
            @change="emit('update:uploadOrderId', ($event.target as HTMLSelectElement).value)"
          >
            <option value="" disabled>
              {{ props.uploadOrdersLoading
                ? t('resourcesPictureWarehouseShared.upload.loadingOrders')
                : t('resourcesPictureWarehouseShared.upload.selectOrder') }}
            </option>
            <option
              v-for="order in props.uploadOrders"
              :key="order.id"
              :value="String(order.id)"
              :disabled="!order.eligible"
            >
              {{ order.order_number }} · {{ order.status }}
            </option>
          </select>
          <p v-if="props.uploadOrdersError" class="tz-caption text-red-400">
            {{ props.uploadOrdersError }}
          </p>
          <p
            v-else-if="!props.uploadOrdersLoading && !props.hasEligibleUploadOrder"
            class="tz-caption tz-text-muted"
          >
            {{ t('resourcesPictureWarehouseShared.upload.completedOrderRequired') }}
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">
              {{ t('resourcesPictureWarehouseShared.upload.region') }}
              <span class="text-red-400">*</span>
            </label>
            <input
              :value="props.uploadRegion"
              type="text"
              class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
              :placeholder="t('resourcesPictureWarehouseShared.upload.regionPlaceholder')"
              required
              @input="emit('update:uploadRegion', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">
              {{ t('resourcesPictureWarehouseShared.upload.location') }}
            </label>
            <input
              :value="props.uploadLocation"
              type="text"
              class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
              :placeholder="t('resourcesPictureWarehouseShared.upload.locationPlaceholder')"
              @input="emit('update:uploadLocation', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>

        <div class="flex flex-col gap-1">
          <label class="tz-caption tz-text-secondary">
            {{ t('resourcesPictureWarehouseShared.upload.nickname') }}
          </label>
          <input
            :value="props.uploadNickname"
            type="text"
            class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
            :placeholder="t('resourcesPictureWarehouseShared.upload.nicknamePlaceholder')"
            @input="emit('update:uploadNickname', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="flex flex-col gap-1">
          <label class="tz-caption tz-text-secondary">
            {{ t('resourcesPictureWarehouseShared.upload.notes') }}
          </label>
          <textarea
            :value="props.uploadNotes"
            rows="2"
            class="picture-upload-control rounded-lg px-2.5 py-1.5 text-xs"
            :placeholder="t('resourcesPictureWarehouseShared.upload.notesPlaceholder')"
            @input="emit('update:uploadNotes', ($event.target as HTMLTextAreaElement).value)"
          ></textarea>
        </div>

        <div class="flex flex-col gap-1">
          <label class="tz-caption tz-text-secondary">
            {{ t('resourcesPictureWarehouseShared.upload.photos') }}
          </label>
          <div class="picture-upload-picker">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <button
                type="button"
                class="picture-upload-file-trigger inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="props.uploading || props.uploadFiles.length >= 10 || !props.selectedUploadOrderEligible"
                @click="openUploadFilePicker"
              >
                <Icon name="lucide:image-plus" class="h-3.5 w-3.5" aria-hidden="true" />
                <span>
                  {{ props.uploadFiles.length >= 10
                    ? t('resourcesPictureWarehouseShared.upload.maximumReached')
                    : t('resourcesPictureWarehouseShared.upload.chooseFiles') }}
                </span>
              </button>
              <span class="tz-caption tz-text-muted">
                {{ props.uploadFiles.length
                  ? t('resourcesPictureWarehouseShared.upload.selectedCount', {
                    count: props.uploadFiles.length,
                  })
                  : t('resourcesPictureWarehouseShared.upload.noPhotosSelected') }}
              </span>
            </div>

            <input
              ref="uploadFileInput"
              type="file"
              :accept="props.uploadAccept"
              multiple
              :disabled="props.uploading || !props.selectedUploadOrderEligible"
              class="hidden"
              @change="emit('file-change', $event)"
            />
            <p class="mt-2 tz-caption tz-text-muted">
              {{ props.uploadHint }}
            </p>

            <div
              v-if="props.uploadPreviews.length"
              class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4 md:grid-cols-5"
            >
              <div
                v-for="(preview, index) in props.uploadPreviews"
                :key="preview.key"
                class="picture-upload-preview group"
              >
                <div class="relative aspect-square overflow-hidden rounded-lg tz-surface-panel">
                  <StorefrontImage
                    :src="preview.url"
                    :alt="preview.file.name"
                    class="size-full object-cover"
                    preset="thumbnail"
                  />
                  <button
                    type="button"
                    class="picture-upload-preview-remove"
                    :aria-label="t('resourcesPictureWarehouseShared.upload.removePhoto', {
                      name: preview.file.name,
                    })"
                    :title="t('resourcesPictureWarehouseShared.upload.removePhotoTitle')"
                    @click="emit('remove-file', index)"
                  >
                    <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
                  </button>
                </div>
                <p class="mt-1 truncate text-[10px] tz-text-muted" :title="preview.file.name">
                  {{ preview.file.name }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-2 flex items-center justify-between gap-2 border-t tz-border-subtle pt-2">
          <div class="flex-1">
            <p v-if="props.uploadSuccess" class="tz-caption text-emerald-600">
              {{ props.uploadSuccess }}
            </p>
            <p v-else-if="props.uploadError" class="tz-caption text-red-400">
              {{ props.uploadError }}
            </p>
          </div>
          <button
            type="submit"
            class="picture-upload-submit inline-flex h-9 items-center justify-center rounded-full px-5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="props.uploading || !props.selectedUploadOrderEligible"
          >
            {{ props.uploading
              ? t('resourcesPictureWarehouseShared.upload.uploading')
              : t('resourcesPictureWarehouseShared.upload.submit') }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import type { UploadOrderOption, UploadPreview } from '~/types/pictureWarehouse'

const props = defineProps<{
  uploadOrderId: string
  uploadOrders: UploadOrderOption[]
  uploadOrdersLoading: boolean
  uploadOrdersError: string | null
  hasEligibleUploadOrder: boolean
  selectedUploadOrderEligible: boolean
  uploadRegion: string
  uploadLocation: string
  uploadNickname: string
  uploadNotes: string
  uploadFiles: File[]
  uploadPreviews: UploadPreview[]
  uploadAccept: string
  uploadHint: string
  uploading: boolean
  uploadError: string | null
  uploadSuccess: string | null
}>()

const emit = defineEmits<{
  'update:uploadOrderId': [value: string]
  'update:uploadRegion': [value: string]
  'update:uploadLocation': [value: string]
  'update:uploadNickname': [value: string]
  'update:uploadNotes': [value: string]
  'file-change': [event: Event]
  'remove-file': [index: number]
  submit: []
}>()

const { t } = useI18n()
const uploadFileInput = ref<HTMLInputElement | null>(null)

const openUploadFilePicker = () => {
  if (!props.selectedUploadOrderEligible) return
  uploadFileInput.value?.click()
}

watch(
  () => props.uploadFiles.length,
  (count) => {
    if (count === 0 && uploadFileInput.value) {
      uploadFileInput.value.value = ''
    }
  },
)
</script>

<style scoped>
.picture-upload-card {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-card-surface);
  box-shadow: 0 3px 9px rgb(15 23 42 / 0.08);
}

.picture-upload-control {
  box-sizing: border-box;
  border: 1px solid var(--tz-form-control-border);
  color: var(--tz-text-primary);
  background: var(--tz-form-control-surface);
  box-shadow: none;
}

.picture-upload-control::placeholder {
  color: var(--tz-text-muted);
}

.picture-upload-control:focus {
  outline: none;
  border-color: var(--tz-form-control-focus-border);
  box-shadow: 0 0 0 1px var(--tz-form-control-focus-ring);
}

.picture-upload-picker {
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.75rem;
  background: var(--tz-form-control-surface);
  padding: 0.625rem;
}

.picture-upload-file-trigger {
  border: 1px solid var(--tz-border-strong);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
  transition: background-color 160ms ease, border-color 160ms ease;
}

.picture-upload-file-trigger:hover:not(:disabled) {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
}

.picture-upload-preview {
  min-width: 0;
}

.picture-upload-preview-remove {
  position: absolute;
  top: 0.35rem;
  right: 0.35rem;
  display: inline-flex;
  height: 1.5rem;
  width: 1.5rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 9999px;
  background: rgba(0, 0, 0, 0.68);
  color: white;
  opacity: 0;
  transition: opacity 160ms ease, background-color 160ms ease;
}

.picture-upload-preview:hover .picture-upload-preview-remove,
.picture-upload-preview-remove:focus-visible {
  opacity: 1;
}

.picture-upload-preview-remove:hover {
  background: rgba(220, 38, 38, 0.9);
}

.picture-upload-submit {
  color: #ffffff;
  background: var(--tz-action-primary);
  border-color: var(--tz-action-primary);
  box-shadow: 0 4px 12px -4px rgb(15 23 42 / 0.14);
  transition: background-color 160ms ease, opacity 160ms ease;
}

.picture-upload-submit:hover:not(:disabled) {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
}
</style>
