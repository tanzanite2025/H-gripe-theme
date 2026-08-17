<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Markets</p>
        <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">市场与本地化语种</h3>
        <p class="mt-1 max-w-2xl text-xs leading-relaxed text-muted-foreground">
          国家/IP 只用于匹配市场；默认语种、支持语种、展示币种、物流和税务策略从市场上下文继续向后传递。支付按钮由支付设置里启用的收款方式决定。
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Button type="button" variant="outline" size="sm" :disabled="loading" @click="loadMarkets">
 <RefreshCw :class="['size-3.5', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button v-if="canEdit" type="button" size="sm" :disabled="loading" @click="openCreateDialog">
          <Plus class="size-3.5" />
          新增市场
        </Button>
      </div>
    </div>

    <div v-if="loading" class="mt-5 flex h-28 items-center justify-center text-xs text-muted-foreground">
      <LoaderCircle class="mr-2 size-4 animate-spin" />
      正在读取市场配置
    </div>

    <div v-else class="mt-5 overflow-hidden rounded-xl border bg-background/70">
      <div v-if="markets.length === 0" class="flex h-28 items-center justify-center text-xs text-muted-foreground">
        暂无市场配置
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[860px] text-left text-xs">
          <thead class="border-b bg-muted/60 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            <tr>
              <th class="px-3 py-2">市场</th>
              <th class="px-3 py-2">国家</th>
              <th class="px-3 py-2">本地化语种</th>
              <th class="px-3 py-2">展示币种</th>
              <th class="px-3 py-2">状态</th>
              <th class="px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="market in markets" :key="market.id || market.code" class="border-b last:border-b-0">
              <td class="px-3 py-3 align-top">
                <div class="font-black text-foreground">{{ market.code }}</div>
                <div class="mt-0.5 text-muted-foreground">{{ market.name || market.code }}</div>
              </td>
              <td class="px-3 py-3 align-top">
                <div class="flex max-w-72 flex-wrap gap-1.5">
                  <Badge v-for="country in market.countries || []" :key="country" variant="outline">{{ country }}</Badge>
                  <span v-if="!market.countries?.length" class="text-muted-foreground">GLOBAL</span>
                </div>
              </td>
              <td class="px-3 py-3 align-top">
                <div class="font-mono font-bold">{{ market.default_locale }}</div>
                <div class="mt-1 flex max-w-72 flex-wrap gap-1.5">
                  <Badge v-for="languageLocale in market.supported_locales || []" :key="languageLocale" variant="secondary">{{ languageLocale }}</Badge>
                </div>
              </td>
              <td class="px-3 py-3 align-top">
                <div class="font-mono font-bold">{{ market.default_currency }}</div>
                <div class="mt-1 flex max-w-72 flex-wrap gap-1.5">
                  <Badge v-for="currency in market.display_currencies || []" :key="currency" variant="outline">{{ currency }}</Badge>
                </div>
              </td>
              <td class="px-3 py-3 align-top">
                <Badge :variant="market.enabled ? 'default' : 'outline'">{{ market.enabled ? '启用' : '停用' }}</Badge>
                <div class="mt-1 text-[10px] text-muted-foreground">{{ market.priority || 100 }}</div>
              </td>
              <td class="px-3 py-3 align-top">
                <div class="flex justify-end gap-1.5">
                  <Button type="button" variant="ghost" size="icon" title="编辑市场" :disabled="!canEdit" @click="openEditDialog(market)">
                    <Pencil class="size-4" />
                  </Button>
                  <Button type="button" variant="ghost" size="icon" title="删除市场" :disabled="!canEdit" @click="deleteMarket(market)">
                    <Trash2 class="size-4" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Dialog :open="dialogOpen" @update:open="dialogOpen = $event">
      <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
        <form class="space-y-6" @submit.prevent="saveMarket">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? '新增市场' : '编辑市场' }}</DialogTitle>
            <DialogDescription>市场配置会影响前台国家匹配、默认语种、支持语种、展示币种、物流策略和税务策略解析。</DialogDescription>
          </DialogHeader>

          <div class="grid gap-4 lg:grid-cols-2">
            <AdminFormField label="市场代码" required :error="errors.code">
              <Input v-model="form.code" class="font-mono uppercase" placeholder="US" @input="clearError('code')" />
            </AdminFormField>
            <AdminFormField label="市场名称" :error="errors.name">
              <Input v-model="form.name" placeholder="United States" @input="clearError('name')" />
            </AdminFormField>

            <AdminFormField label="国家代码" class="lg:col-span-2" :error="errors.countries" description="用逗号分隔，例如 US, CA, MX；同一个国家只能归属一个市场。">
              <Textarea v-model="countryInput" class="min-h-20 font-mono text-xs uppercase" placeholder="US, CA" @input="clearError('countries')" />
            </AdminFormField>

            <AdminFormField label="默认语种" required :error="errors.default_locale">
              <Select v-model="form.default_locale">
                <SelectTrigger class="w-full"><SelectValue placeholder="选择默认语种" /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="languageLocale in availableLanguageLocales" :key="languageLocale.code" :value="languageLocale.code">
                    {{ languageLocale.code }} · {{ languageLocale.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>

            <AdminFormField label="默认展示币种" required :error="errors.default_currency">
              <Select v-model="form.default_currency">
                <SelectTrigger class="w-full"><SelectValue placeholder="选择默认币种" /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="currency in availableCurrencies" :key="currency.code" :value="currency.code">
                    {{ currency.code }} · {{ currency.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>

            <AdminFormField label="支持语种" class="lg:col-span-2" :error="errors.supported_locales">
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="languageLocale in availableLanguageLocales"
                  :key="languageLocale.code"
                  type="button"
                  class="rounded-full border px-3 py-1.5 text-xs font-black transition hover:border-admin-selected-border hover:bg-admin-selected-soft"
 :class="form.supported_locales.includes(languageLocale.code) ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected shadow-[var(--admin-control-selected-surface-shadow)]': 'bg-background/70 text-foreground'"
                  @click="toggleListValue(form.supported_locales, languageLocale.code)"
                >
                  {{ languageLocale.code }}
                </button>
              </div>
            </AdminFormField>

            <AdminFormField label="展示币种" class="lg:col-span-2" :error="errors.display_currencies">
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="currency in availableCurrencies"
                  :key="currency.code"
                  type="button"
                  class="rounded-full border px-3 py-1.5 text-xs font-black transition hover:border-admin-selected-border hover:bg-admin-selected-soft"
 :class="form.display_currencies.includes(currency.code) ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected shadow-[var(--admin-control-selected-surface-shadow)]': 'bg-background/70 text-foreground'"
                  @click="toggleListValue(form.display_currencies, currency.code)"
                >
                  {{ currency.code }}
                </button>
              </div>
            </AdminFormField>

            <AdminFormField label="物流策略键">
              <Input v-model="form.logistics_policy" class="font-mono" placeholder="standard" />
            </AdminFormField>
            <AdminFormField label="税务策略键">
              <Input v-model="form.tax_policy" class="font-mono" placeholder="vat_eu" />
            </AdminFormField>
            <AdminFormField label="优先级">
              <Input v-model.number="form.priority" type="number" min="1" />
            </AdminFormField>

            <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5 lg:col-span-2">
              <div>
                <span class="text-xs font-bold uppercase tracking-wider">启用市场 / ENABLED</span>
                <p class="mt-0.5 text-xs text-muted-foreground">停用后公开上下文解析不会命中该市场。</p>
              </div>
              <Switch v-model="form.enabled" aria-label="启用市场" />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              {{ saving ? '保存中' : '保存市场' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { LoaderCircle, Pencil, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import axios from '@/utils/axios'
import type {
  MarketCurrencyOption,
  MarketLanguageOption,
  StorefrontMarket,
  StorefrontMarketForm,
} from './settingsTypes'

const props = withDefaults(defineProps<{
  canEdit?: boolean
}>(), {
  canEdit: false,
})

const loading = ref(false)
const saving = ref(false)
const markets = ref<StorefrontMarket[]>([])
const options = reactive<{
  available_locales: MarketLanguageOption[]
  available_currencies: MarketCurrencyOption[]
}>({
  available_locales: [],
  available_currencies: [],
})
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingID = ref(0)
const countryInput = ref('')
const errors = reactive<Record<string, string>>({})

const form = reactive<StorefrontMarketForm>({
  code: '',
  name: '',
  supported_locales: ['en'],
  default_locale: 'en',
  display_currencies: ['USD'],
  default_currency: 'USD',
  payment_method_policy: '',
  logistics_policy: '',
  tax_policy: '',
  enabled: true,
  priority: 100,
})

const canEdit = computed(() => props.canEdit)
const availableLanguageLocales = computed(() => options.available_locales || [])
const availableCurrencies = computed(() => options.available_currencies || [])

const responsePayload = (response: { data?: any }): Record<string, any> => response.data?.data?.data || response.data?.data || response.data || {}

const normalizeCodes = (value: unknown, length: number): string[] => {
  const seen = new Set<string>()
  return String(value || '')
    .split(/[\s,;，；]+/)
    .map((code) => code.trim().toUpperCase())
    .filter((code) => code.length === length && /^[A-Z0-9_-]+$/.test(code))
    .filter((code) => {
      if (seen.has(code)) return false
      seen.add(code)
      return true
    })
}

const clearErrors = () => {
  Object.keys(errors).forEach((key) => delete errors[key])
}

const clearError = (key: string): void => {
  delete errors[key]
}

const applyMarketToForm = (market: Partial<StorefrontMarket> = {}): void => {
  form.code = market.code || ''
  form.name = market.name || ''
  form.supported_locales = Array.isArray(market.supported_locales) && market.supported_locales.length ? [...market.supported_locales] : ['en']
  form.default_locale = market.default_locale || form.supported_locales[0] || 'en'
  form.display_currencies = Array.isArray(market.display_currencies) && market.display_currencies.length ? [...market.display_currencies] : ['USD']
  form.default_currency = market.default_currency || form.display_currencies[0] || 'USD'
  form.payment_method_policy = market.payment_method_policy || ''
  form.logistics_policy = market.logistics_policy || ''
  form.tax_policy = market.tax_policy || ''
  form.enabled = market.enabled !== false
  form.priority = Number(market.priority || 100)
  countryInput.value = Array.isArray(market.countries) ? market.countries.join(', ') : ''
}

const loadOptions = async () => {
  const response = await axios.get('/api/admin/storefront/markets/options')
  const payload = responsePayload(response)
  options.available_locales = Array.isArray(payload.available_locales) ? payload.available_locales : []
  options.available_currencies = Array.isArray(payload.available_currencies) ? payload.available_currencies : []
}

const loadMarkets = async () => {
  loading.value = true
  try {
    await loadOptions()
    const response = await axios.get('/api/admin/storefront/markets')
    const payload = responsePayload(response)
    markets.value = Array.isArray(payload) ? payload : []
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '市场配置读取失败')
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  clearErrors()
  dialogMode.value = 'create'
  editingID.value = 0
  applyMarketToForm({})
  dialogOpen.value = true
}

const openEditDialog = (market: StorefrontMarket): void => {
  clearErrors()
  dialogMode.value = 'edit'
  editingID.value = Number(market?.id || 0)
  applyMarketToForm(market)
  dialogOpen.value = true
}

const toggleListValue = (list: string[], value: string): void => {
  const index = list.indexOf(value)
  if (index >= 0) list.splice(index, 1)
  else list.push(value)
}

const validateForm = (): boolean => {
  clearErrors()
  if (!form.code.trim()) errors.code = '请输入市场代码'
  if (!form.default_locale) errors.default_locale = '请选择默认语种'
  if (!form.default_currency) errors.default_currency = '请选择默认展示币种'
  if (!form.supported_locales.length) errors.supported_locales = '至少选择一种支持语种'
  if (!form.display_currencies.length) errors.display_currencies = '至少选择一种展示币种'
  if (!form.supported_locales.includes(form.default_locale)) form.supported_locales.unshift(form.default_locale)
  if (!form.display_currencies.includes(form.default_currency)) form.display_currencies.unshift(form.default_currency)
  return Object.keys(errors).length === 0
}

const marketPayload = (): StorefrontMarket & { countries: string[] } => ({
  code: form.code.trim().toUpperCase(),
  name: form.name.trim(),
  countries: normalizeCodes(countryInput.value, 2),
  default_locale: form.default_locale,
  supported_locales: [...form.supported_locales],
  default_currency: form.default_currency,
  display_currencies: [...form.display_currencies],
  payment_method_policy: form.payment_method_policy.trim(),
  logistics_policy: form.logistics_policy.trim(),
  tax_policy: form.tax_policy.trim(),
  enabled: form.enabled,
  priority: Number(form.priority || 100),
})

const saveMarket = async () => {
  if (!validateForm()) return
  saving.value = true
  try {
    const payload = marketPayload()
    if (dialogMode.value === 'edit' && editingID.value) {
      await axios.put(`/api/admin/storefront/markets/${editingID.value}`, payload)
      toast.success('市场已保存')
    } else {
      await axios.post('/api/admin/storefront/markets', payload)
      toast.success('市场已创建')
    }
    dialogOpen.value = false
    await loadMarkets()
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '市场保存失败')
  } finally {
    saving.value = false
  }
}

const deleteMarket = async (market: StorefrontMarket): Promise<void> => {
  if (!market?.id) return
  const expected = String(market.code || '').trim().toUpperCase()
  const typed = window.prompt(`删除市场 ${expected} 会让相关国家回退到 GLOBAL。请输入 ${expected} 确认删除。`)
  if (typed !== expected) return
  try {
    await axios.delete(`/api/admin/storefront/markets/${market.id}`)
    toast.success('市场已删除')
    await loadMarkets()
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '市场删除失败')
  }
}

onMounted(loadMarkets)
</script>
