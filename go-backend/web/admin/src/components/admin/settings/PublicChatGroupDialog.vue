<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <form class="space-y-4" @submit.prevent="emit('save')">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑客服组' : '添加客服组' }}</DialogTitle>
          <DialogDescription>
            客服组只作为前台客服列表的展示标签，帮助客户识别要选哪位客服。
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="组 Code" required description="稳定的机器标识，只允许小写字母、数字和下划线。">
            <Input v-model="form.code" :disabled="saving || Boolean(form.id)" maxlength="50" placeholder="technical_support" />
          </AdminFormField>

          <AdminFormField label="组名称" required>
            <Input v-model="form.name" :disabled="saving" maxlength="100" placeholder="Technical Support" />
          </AdminFormField>

          <AdminFormField label="排序" description="数字越小越靠前。">
            <Input v-model.number="form.sort_order" :disabled="saving" type="number" min="0" />
          </AdminFormField>

          <AdminFormField label="状态" required>
            <Select v-model="form.status" :disabled="saving">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择状态" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="active">active · 启用</SelectItem>
                <SelectItem value="inactive">inactive · 停用</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="说明" class="md:col-span-2">
            <Textarea v-model="form.description" :disabled="saving" class="min-h-24" maxlength="500" placeholder="Compatibility, specs, setup and troubleshooting." />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !form.name?.trim()">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            保存客服组
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

interface PublicChatGroupForm {
  id?: string | number | null
  code: string
  name: string
  sort_order: number | string
  status: string
  description: string
}

const props = withDefaults(defineProps<{
  open?: boolean
  form: PublicChatGroupForm
  saving?: boolean
}>(), {
  open: false,
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()
</script>
