<template>
  <div class="flex min-h-14 items-center gap-3">
    <Avatar class="size-14 shrink-0">
      <AvatarImage v-if="avatar" :src="avatar" alt="客服头像" />
      <AvatarFallback class="text-base"><UserRound class="size-5" /></AvatarFallback>
    </Avatar>

    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-1.5">
        <Button
          size="icon-sm"
          variant="outline"
          :disabled="!canManage || uploading"
          :title="avatar ? '更换客服头像' : '上传客服头像'"
          :aria-label="avatar ? '更换客服头像' : '上传客服头像'"
          @click="openFilePicker"
        >
          <LoaderCircle v-if="uploading" class="size-3.5 animate-spin" />
          <Upload v-else class="size-3.5" />
        </Button>
        <Button
          v-if="avatar"
          size="icon-sm"
          variant="outline"
          :disabled="!canManage || uploading || removing"
          title="移除客服头像"
          aria-label="移除客服头像"
          @click="removeAvatar"
        >
          <LoaderCircle v-if="removing" class="size-3.5 animate-spin" />
          <Trash2 v-else class="size-3.5" />
        </Button>
      </div>
      <p class="mt-1 text-[10px] font-bold text-muted-foreground">
        {{ profileReady ? (avatar ? '头像已保存' : '未设置头像') : '先保存 Profile 后上传头像' }}
      </p>
      <UploadSpecHint code="customer_service_avatar" />
    </div>

    <input
      ref="fileInput"
      class="sr-only"
      type="file"
      :accept="uploadSpecAccept('customer_service_avatar')"
      tabindex="-1"
      @change="uploadSelectedFile"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Trash2, Upload, UserRound } from '@lucide/vue'
import customerServiceAvatarApi from '@/api/customerServiceAvatar'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { uploadSpecAccept, validateUploadFile } from '@/lib/uploadSpecs'

const props = withDefaults(defineProps<{
  userId: string | number | null
  avatar?: string
  profileReady?: boolean
  disabled?: boolean
}>(), {
  avatar: '',
  profileReady: false,
  disabled: false,
})

const emit = defineEmits<{
  (event: 'uploaded', avatar: string): void
  (event: 'removed'): void
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const removing = ref(false)
const canManage = computed(() => Boolean(props.profileReady && props.userId && !props.disabled))

const openFilePicker = () => {
  if (!canManage.value || uploading.value) return
  fileInput.value?.click()
}

const uploadSelectedFile = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !canManage.value || !props.userId) return

  const validation = await validateUploadFile(file, 'customer_service_avatar')
  if (!validation.ok) {
    toast.error(validation.error || '头像不符合上传规范')
    return
  }
  if (validation.warning) toast.warning(validation.warning)

  uploading.value = true
  try {
    const avatar = await customerServiceAvatarApi.upload(props.userId, file)
    emit('uploaded', avatar)
    toast.success('客服头像已保存')
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '客服头像上传失败')
  } finally {
    uploading.value = false
  }
}

const removeAvatar = async () => {
  if (!canManage.value || !props.userId || !props.avatar || removing.value) return

  removing.value = true
  try {
    await customerServiceAvatarApi.remove(props.userId)
    emit('removed')
    toast.success('客服头像已移除')
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '客服头像移除失败')
  } finally {
    removing.value = false
  }
}
</script>
