<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '创建图库' : '编辑图库' }}</DialogTitle>
          <DialogDescription>图库的封面由其中首张排序图片自动提供。</DialogDescription>
        </DialogHeader>

        <AdminFormField label="标题" required :error="errors.title">
          <Input v-model="form.title" placeholder="请输入图库标题" @input="emit('clear-error', 'title')" />
        </AdminFormField>
        <AdminFormField label="Slug" required :error="errors.slug">
          <Input v-model="form.slug" placeholder="例如 customer-stories" @input="emit('clear-error', 'slug')" />
        </AdminFormField>
        <AdminFormField label="描述">
          <Textarea v-model="form.description" class="min-h-24" placeholder="请输入图库描述" />
        </AdminFormField>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存图库' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, default: () => ({}) },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
