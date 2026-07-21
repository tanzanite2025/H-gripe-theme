<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增配送区域' : '编辑配送区域' }}</DialogTitle>
          <DialogDescription>
            配送区域是运费模板匹配的基础维度。当前先兼容 JSON 数组或逗号分隔文本。
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 lg:grid-cols-2">
          <AdminFormField label="区域名称" required :error="errors.name">
            <Input v-model.trim="form.name" placeholder="例如 北美、欧盟、东南亚" @input="emit('clear-error', 'name')" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-medium">启用区域</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不参与运费区域匹配。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用配送区域" />
          </div>

          <AdminFormField label="国家/地区代码" required :error="errors.countries" class="lg:col-span-2" description='建议 JSON 数组，例如 ["US", "CA"]；也兼容 US, CA。'>
            <Textarea
              v-model="form.countries"
              class="min-h-24 font-mono text-xs"
              placeholder='["US", "CA", "MX"]'
              @input="emit('clear-error', 'countries')"
            />
          </AdminFormField>

          <AdminFormField label="州/省代码" description="可选；JSON 数组或逗号分隔。">
            <Textarea v-model="form.states" class="min-h-20 font-mono text-xs" placeholder='["CA", "NY"]' />
          </AdminFormField>

          <AdminFormField label="邮编范围" description="可选；后续会接更细的范围校验。">
            <Textarea v-model="form.postal_codes" class="min-h-20 font-mono text-xs" placeholder='["90000-96162", "10001"]' />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存区域' }}
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
