<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent size="full" class="max-w-[1400px]">
      <DialogHeader>
        <DialogTitle>PayPal 商业发票样式预览</DialogTitle>
        <DialogDescription>填写样例数据后生成 PDF，只用于查看排版，不会上传或提交 PayPal。</DialogDescription>
      </DialogHeader>

      <div class="grid min-h-0 gap-5 lg:grid-cols-[minmax(360px,0.8fr)_minmax(0,1.2fr)]">
        <form class="min-h-0 space-y-4 overflow-auto pr-1" @submit.prevent="renderPreview">
          <section class="space-y-3">
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Document</h3>
            <div class="grid gap-3 sm:grid-cols-2">
              <Input v-model="form.document_number" placeholder="Document number" />
              <Input v-model="form.document_date" type="date" />
              <Input v-model="form.currency" placeholder="Currency" />
              <Input v-model="form.payment_reference" placeholder="PayPal payment reference" />
              <Input v-model="form.payment_method" placeholder="Payment method" />
              <Input v-model="form.payment_status" placeholder="Payment status" />
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Seller</h3>
            <Input v-model="form.seller.name" placeholder="Seller name" />
            <Textarea v-model="form.seller.address" rows="3" placeholder="Seller address" />
            <div class="grid gap-3 sm:grid-cols-2">
              <Input v-model="form.seller.email" placeholder="Seller email" />
              <Input v-model="form.seller.phone" placeholder="Seller phone" />
              <Input v-model="form.seller.website" placeholder="Seller website" />
              <Input v-model="form.seller.tax_id" placeholder="Seller tax ID / VAT ID / GST ID" />
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Bill To</h3>
            <div class="grid gap-3 sm:grid-cols-2">
              <Input v-model="form.bill_to.name" placeholder="Customer name" />
              <Input v-model="form.bill_to.email" placeholder="Customer email" />
              <Input v-model="form.bill_to.line1" placeholder="Address line 1" />
              <Input v-model="form.bill_to.line2" placeholder="Address line 2" />
              <Input v-model="form.bill_to.city" placeholder="City" />
              <Input v-model="form.bill_to.state" placeholder="State / Province" />
              <Input v-model="form.bill_to.postal_code" placeholder="Postal code" />
              <Input v-model="form.bill_to.country" placeholder="Country" />
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Ship To</h3>
            <div class="grid gap-3 sm:grid-cols-2">
              <Input v-model="form.ship_to.name" placeholder="Recipient name" />
              <Input v-model="form.ship_to.email" placeholder="Recipient email" />
              <Input v-model="form.ship_to.line1" placeholder="Address line 1" />
              <Input v-model="form.ship_to.line2" placeholder="Address line 2" />
              <Input v-model="form.ship_to.city" placeholder="City" />
              <Input v-model="form.ship_to.state" placeholder="State / Province" />
              <Input v-model="form.ship_to.postal_code" placeholder="Postal code" />
              <Input v-model="form.ship_to.country" placeholder="Country" />
            </div>
          </section>

          <section class="space-y-3">
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Items</h3>
              <Button type="button" variant="outline" size="sm" class="rounded-full" @click="addItem">
                <Plus class="size-3.5" />
                添加商品
              </Button>
            </div>
            <div v-for="(item, index) in form.items" :key="item.key" class="space-y-2 rounded-xl border border-dashed border-border/80 p-3">
              <div class="flex items-center justify-between gap-2">
                <span class="text-xs font-bold text-muted-foreground">商品 {{ index + 1 }}</span>
                <Button v-if="form.items.length > 1" type="button" variant="ghost" size="icon-sm" @click="removeItem(index)">
                  <Trash2 class="size-3.5" />
                </Button>
              </div>
              <Input v-model="item.description" placeholder="Description" />
              <div class="grid gap-2 sm:grid-cols-2">
                <Input v-model="item.sku" placeholder="SKU" />
                <Input v-model.number="item.quantity" type="number" min="1" placeholder="Qty" />
                <Input v-model.number="item.unit_price" type="number" min="0" step="0.01" placeholder="Unit price" />
                <Input v-model.number="item.total" type="number" min="0" step="0.01" placeholder="Line total" />
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Totals</h3>
            <div class="grid gap-3 sm:grid-cols-2">
              <Input v-model.number="form.subtotal" type="number" min="0" step="0.01" placeholder="Subtotal" />
              <Input v-model.number="form.shipping" type="number" min="0" step="0.01" placeholder="Shipping" />
              <Input v-model.number="form.tax" type="number" min="0" step="0.01" placeholder="Tax" />
              <Input v-model.number="form.discount" type="number" min="0" step="0.01" placeholder="Discount" />
              <Input v-model.number="form.total" type="number" min="0" step="0.01" placeholder="Total" />
            </div>
          </section>

          <Button type="submit" class="w-full rounded-full font-black uppercase tracking-wider" :disabled="loading">
            <FileText class="size-4" />
            {{ loading ? '正在生成...' : '生成 PDF 预览' }}
          </Button>
        </form>

        <section class="min-h-[560px] overflow-hidden rounded-xl border border-dashed border-border/80 bg-muted/20">
          <iframe v-if="pdfUrl" :src="pdfUrl" title="PayPal commercial invoice preview" class="h-full min-h-[560px] w-full bg-white" />
          <div v-else class="flex h-full min-h-[560px] items-center justify-center p-8 text-center text-sm text-muted-foreground">
            生成后将在这里显示 PDF。
          </div>
        </section>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { FileText, Plus, Trash2 } from '@lucide/vue'
import { paymentRiskApi } from '@/api/paymentRisk'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (event: 'update:open', value: boolean): void }>()
const dialogOpen = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value),
})

const createItem = () => ({
  key: `${Date.now()}-${Math.random()}`,
  description: 'Carbon wheelset',
  sku: 'C50-DT240',
  quantity: 1,
  unit_price: 249.9,
  total: 249.9,
})

const form = reactive({
  document_number: 'CI-PREVIEW-001',
  document_date: new Date().toISOString().slice(0, 10),
  currency: 'USD',
  payment_reference: 'PAYPAL-SAMPLE-CAPTURE',
  payment_method: 'PayPal',
  payment_status: 'paid',
  seller: {
    name: 'H-GRIPE',
    address: '100 Factory Road\nAustin, TX 78701\nUS',
    email: 'support@example.test',
    phone: '+1 555 0100',
    website: 'https://example.test',
    tax_id: 'US-SAMPLE-TAX-ID',
  },
  bill_to: {
    name: 'Ada Lovelace',
    email: 'ada@example.test',
    line1: '1 Carbon Road',
    line2: '',
    city: 'Los Angeles',
    state: 'CA',
    postal_code: '90001',
    country: 'US',
  },
  ship_to: {
    name: 'Ada Lovelace',
    email: 'ada@example.test',
    line1: '1 Carbon Road',
    line2: '',
    city: 'Los Angeles',
    state: 'CA',
    postal_code: '90001',
    country: 'US',
  },
  items: [createItem()],
  subtotal: 249.9,
  shipping: 20,
  tax: 0,
  discount: 0,
  total: 269.9,
})

const loading = ref(false)
const pdfUrl = ref('')
const sellerProfileLoaded = ref(false)

const addItem = () => {
  form.items.push(createItem())
}

const removeItem = (index: number) => {
  form.items.splice(index, 1)
}

const releasePDF = () => {
  if (!pdfUrl.value) return
  URL.revokeObjectURL(pdfUrl.value)
  pdfUrl.value = ''
}

const renderPreview = async () => {
  loading.value = true
  try {
    const response = await paymentRiskApi.previewPayPalInvoicePDF({
      ...form,
      items: form.items.map((item) => ({
        ...item,
        subtotal: Number(item.unit_price || 0) * Number(item.quantity || 1),
        tax: 0,
        discount: 0,
        total: Number(item.total || 0),
      })),
      payment_date: form.document_date,
    })
    releasePDF()
    const blob = response.data instanceof Blob
      ? response.data
      : new Blob([response.data], { type: 'application/pdf' })
    pdfUrl.value = URL.createObjectURL(blob)
  } catch {
    toast.error('PDF 预览生成失败')
  } finally {
    loading.value = false
  }
}

const loadSellerProfile = async () => {
  if (sellerProfileLoaded.value) return
  try {
    const profile = await paymentRiskApi.getPayPalInvoiceSellerProfile()
    if (profile?.name) form.seller.name = profile.name
    if (profile?.address) form.seller.address = profile.address
    if (profile?.email) form.seller.email = profile.email
    if (profile?.phone) form.seller.phone = profile.phone
    if (profile?.website) form.seller.website = profile.website
    if (profile?.tax_id) form.seller.tax_id = profile.tax_id
  } catch {
    // The preview remains usable with its sample seller when settings are unavailable.
  } finally {
    sellerProfileLoaded.value = true
  }
}

watch(() => props.open, (open) => {
  if (open) void loadSellerProfile()
}, { immediate: true })

onBeforeUnmount(releasePDF)
</script>
