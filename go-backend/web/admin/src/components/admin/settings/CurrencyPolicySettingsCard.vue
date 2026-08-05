<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Currency Policy</p>
        <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">币种策略</h3>
        <p class="mt-1 max-w-2xl text-xs leading-relaxed text-muted-foreground">
          这里维护本站订单和支付请求使用的币种；Stripe、PayPal 的实际可用能力仍由网关侧校验。
        </p>
      </div>
      <div class="flex items-center gap-2">
        <span class="rounded-full border border-admin-selected-border bg-admin-selected-soft px-2.5 py-1 text-[11px] font-black text-admin-selected">
          默认收款 {{ policy.default_order_currency || '未配置' }}
        </span>
        <Button type="button" variant="outline" size="sm" :disabled="loading || saving" @click="loadPolicy">
          <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" type="button" size="sm" :disabled="loading || saving" @click="savePolicy">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存币种策略' }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="mt-5 flex h-28 items-center justify-center text-xs text-muted-foreground">
      <LoaderCircle class="mr-2 size-4 animate-spin" />
      正在读取币种策略
    </div>

    <div v-else class="mt-5 space-y-5">
      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField label="记账基准币种" description="商品原价、ERP 与报表默认使用的基准。">
          <select v-model="policy.accounting_currency" class="h-10 w-full rounded-md border bg-background px-3 text-sm font-bold text-foreground" :disabled="!canEdit">
            <option v-for="option in catalog" :key="`accounting-${option.code}`" :value="option.code">
              {{ option.code }} · {{ option.name }}
            </option>
          </select>
        </AdminFormField>
        <AdminFormField label="默认收款币种" description="后端创建订单和支付请求时使用的默认币种。">
          <select v-model="policy.default_order_currency" class="h-10 w-full rounded-md border bg-background px-3 text-sm font-bold text-foreground" :disabled="!canEdit">
            <option v-for="option in acceptedCatalog" :key="`default-${option.code}`" :value="option.code">
              {{ option.code }} · {{ option.name }}
            </option>
          </select>
        </AdminFormField>
      </div>

      <div class="grid gap-4">
        <div class="rounded-xl border bg-background/70 p-3">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black text-foreground">允许收款币种</p>
              <p class="mt-1 text-[11px] leading-relaxed text-muted-foreground">订单、支付、退款和 webhook 只允许使用这里的币种；前台不单独提供币种切换。</p>
            </div>
            <span class="font-mono text-xs font-black text-admin-selected">{{ policy.accepted_currencies.length }}</span>
          </div>
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-for="option in catalog"
              :key="`accepted-${option.code}`"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-black transition"
              :class="policy.accepted_currencies.includes(option.code) ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected' : 'bg-background text-muted-foreground hover:border-admin-selected-border'"
              :disabled="!canEdit"
              @click="toggleAcceptedCurrency(option.code)"
            >
              {{ option.code }}
            </button>
          </div>
        </div>
      </div>

      <div class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3 text-xs leading-relaxed text-amber-800 dark:text-amber-100">
        网关自身的可收款币种仍需结合 Stripe / PayPal 账户、国家和支付方式能力确认；这里只维护 Tanzanite 的业务允许列表。
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const props = defineProps({
  canEdit: { type: Boolean, default: false },
})
const emit = defineEmits(['saved'])

const canEdit = computed(() => props.canEdit)
const loading = ref(false)
const saving = ref(false)
const policy = reactive({
  accounting_currency: '',
  default_order_currency: '',
  accepted_currencies: [],
  available_currencies: [],
})

const catalog = computed(() => policy.available_currencies || [])
const acceptedCatalog = computed(() => catalog.value)

const ensureDefaultOrderCurrency = () => {
  if (!policy.default_order_currency) return
  if (!policy.accepted_currencies.includes(policy.default_order_currency)) {
    policy.accepted_currencies = [...policy.accepted_currencies, policy.default_order_currency]
  }
}

const loadPolicy = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/currency-policy')
    const next = response.data?.policy || {}
    Object.assign(policy, {
      accounting_currency: next.accounting_currency || '',
      default_order_currency: next.default_order_currency || next.default_checkout_currency || '',
      accepted_currencies: Array.isArray(next.accepted_currencies)
        ? next.accepted_currencies
        : Array.isArray(next.checkout_currencies)
          ? next.checkout_currencies
          : [],
      available_currencies: Array.isArray(next.available_currencies) ? next.available_currencies : [],
    })
    ensureDefaultOrderCurrency()
  } catch (error) {
    toast.error(error?.response?.data?.error || '币种策略读取失败')
  } finally {
    loading.value = false
  }
}

const toggleAcceptedCurrency = (code) => {
  if (policy.accepted_currencies.includes(code)) {
    if (policy.accepted_currencies.length === 1) return
    policy.accepted_currencies = policy.accepted_currencies.filter((item) => item !== code)
  } else {
    policy.accepted_currencies = [...policy.accepted_currencies, code]
  }
  ensureDefaultOrderCurrency()
}

const savePolicy = async () => {
  ensureDefaultOrderCurrency()
  saving.value = true
  try {
    const response = await axios.put('/api/admin/settings/currency-policy', {
      accounting_currency: policy.accounting_currency,
      default_order_currency: policy.default_order_currency,
      accepted_currencies: policy.accepted_currencies,
    })
    Object.assign(policy, response.data?.policy || {})
    toast.success('币种策略已保存')
    emit('saved')
  } catch (error) {
    toast.error(error?.response?.data?.error || '币种策略保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadPolicy)
</script>
