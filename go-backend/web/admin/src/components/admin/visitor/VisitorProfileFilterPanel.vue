<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_120px_120px_120px_120px_120px_120px_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block md:col-span-2 xl:col-span-1">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="邮箱、用户 ID、visitor hash、cart session、地区" />
        </div>
      </label>

      <AdminFilterSelect v-model="filters.identity" label="IDENTITY / 身份" :options="identityOptions" />
      <AdminFilterSelect v-model="filters.email" label="EMAIL / 邮箱" :options="emailOptions" />
      <AdminFilterSelect v-model="filters.cartSession" label="CART / 购物车" :options="cartOptions" />
      <AdminFilterSelect v-model="filters.customerServiceVisitor" label="CHAT / 聊天" :options="chatOptions" />
      <AdminFilterSelect v-model="filters.lastSeen" label="LAST SEEN / 活跃" :options="lastSeenOptions" />

      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">COUNTRY / 国家</span>
        <Input v-model="filters.countryCode" class="h-9 font-mono uppercase" placeholder="US" maxlength="8" />
      </label>

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION</span>
        <div class="flex items-center gap-2">
          <Button type="submit" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" :disabled="loading">
            <Search class="size-3.5" />
            查询
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup>
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

defineProps({
  filters: { type: Object, required: true },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['apply', 'reset'])

const identityOptions = [
  { label: '全部', value: 'all' },
  { label: '会员', value: 'account' },
  { label: '匿名', value: 'anonymous' },
]
const emailOptions = [
  { label: '全部', value: 'all' },
  { label: '已采集', value: 'yes' },
  { label: '未采集', value: 'no' },
]
const cartOptions = [
  { label: '全部', value: 'all' },
  { label: '已绑定', value: 'yes' },
  { label: '未绑定', value: 'no' },
]
const chatOptions = cartOptions
const lastSeenOptions = [
  { label: '全部', value: 'all' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
]
</script>
