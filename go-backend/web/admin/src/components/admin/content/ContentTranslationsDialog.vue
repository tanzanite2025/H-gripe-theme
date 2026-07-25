<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[85dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>翻译管理</DialogTitle>
        <DialogDescription v-if="currentPost">
          {{ currentPost.title }} · {{ localeName(currentPost.locale) }}
        </DialogDescription>
      </DialogHeader>

      <div class="relative min-h-32 overflow-x-auto rounded-lg border">
        <div v-if="translationsLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载翻译版本" />
        </div>
        <Table class="min-w-[520px]">
          <TableHeader>
            <TableRow>
              <TableHead class="w-24">语言</TableHead>
              <TableHead>标题</TableHead>
              <TableHead class="w-24">状态</TableHead>
              <TableHead class="w-16 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="!translationsLoading && translations.length === 0" :colspan="4">暂无翻译版本</TableEmpty>
            <TableRow v-for="translation in translations" :key="translation.id">
              <TableCell>{{ localeName(translation.locale) }}</TableCell>
              <TableCell class="font-medium">{{ translation.title }}</TableCell>
              <TableCell>
                <AdminStatusBadge :tone="statusTone(translation.status)">
                  {{ getStatusName(translation.status) }}
                </AdminStatusBadge>
              </TableCell>
              <TableCell class="text-right">
                <Button variant="ghost" size="icon" :aria-label="`编辑翻译 ${translation.title}`" @click="emit('edit', translation)">
                  <Pencil class="size-4" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle, Pencil } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  open: { type: Boolean, default: false },
  currentPost: { type: Object, default: null },
  translationsLoading: { type: Boolean, default: false },
  translations: { type: Array, default: () => [] },
  localeName: { type: Function, required: true },
  statusTone: { type: Function, required: true },
  getStatusName: { type: Function, required: true }
})

const emit = defineEmits(['update:open', 'edit'])
</script>
