<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="PayPal 商业发票卖方资料"
      description="这些资料会自动打印到 PayPal 争议商业发票 PDF；仅用于支付证据，不会公开给店铺前台。"
    >
      <template #actions>
        <Button :disabled="loading || saving || !canEdit" @click="save">
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存设置' }}
        </Button>
      </template>
    </AdminPageHeader>

    <section class="max-w-4xl space-y-4 border-b border-dashed border-border/80 pb-5">
      <div class="grid gap-3 md:grid-cols-2">
        <AdminFormField label="卖方名称" description="商业发票必填。">
          <Input v-model="form.name" :disabled="loading || saving || !canEdit" placeholder="公司或商业主体名称" />
        </AdminFormField>
        <AdminFormField label="Tax ID / VAT ID / GST ID" description="可选，填写实际适用的税务识别号。">
          <Input v-model="form.tax_id" :disabled="loading || saving || !canEdit" placeholder="例如 VAT / GST / EIN" />
        </AdminFormField>
        <AdminFormField label="卖方商业地址" description="商业发票必填，支持多行地址。">
          <Textarea v-model="form.address" :disabled="loading || saving || !canEdit" rows="4" placeholder="街道、城市、州/省、邮编、国家" />
        </AdminFormField>
        <div class="grid gap-3">
          <AdminFormField label="卖方邮箱">
            <Input v-model="form.email" :disabled="loading || saving || !canEdit" type="email" placeholder="invoice@example.com" />
          </AdminFormField>
          <AdminFormField label="卖方电话">
            <Input v-model="form.phone" :disabled="loading || saving || !canEdit" type="tel" placeholder="+1 ..." />
          </AdminFormField>
        </div>
        <AdminFormField label="卖方网站" class="md:col-span-2">
          <Input v-model="form.website" :disabled="loading || saving || !canEdit" type="url" placeholder="https://www.example.com" />
        </AdminFormField>
      </div>
    </section>

    <section class="max-w-4xl border-l-2 border-primary/40 pl-4 text-sm text-muted-foreground">
      <p>当前优先级：后台专用设置 → 环境变量 fallback。</p>
      <p class="mt-1">如果后台设置未填写，系统仍会读取 `PAYPAL_DISPUTE_INVOICE_SELLER_*` 环境变量。</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { paymentRiskApi } from '@/api/paymentRisk'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  name: '',
  address: '',
  email: '',
  phone: '',
  website: '',
  tax_id: '',
})

const assign = (profile: Record<string, any> = {}) => {
  form.name = String(profile.name || '')
  form.address = String(profile.address || '')
  form.email = String(profile.email || '')
  form.phone = String(profile.phone || '')
  form.website = String(profile.website || '')
  form.tax_id = String(profile.tax_id || '')
}

const load = async () => {
  loading.value = true
  try {
    assign(await paymentRiskApi.getPayPalInvoiceSellerProfile())
  } catch (error) {
    console.error('Failed to load PayPal invoice seller profile:', error)
    toast.error('PayPal 商业发票卖方资料加载失败')
  } finally {
    loading.value = false
  }
}

const save = async () => {
  if (!canEdit.value || saving.value) return
  saving.value = true
  try {
    assign(await paymentRiskApi.updatePayPalInvoiceSellerProfile({ ...form }))
    toast.success('PayPal 商业发票卖方资料已保存')
  } catch (error: any) {
    console.error('Failed to save PayPal invoice seller profile:', error)
    toast.error(error?.response?.data?.error || '保存失败，请填写卖方名称和商业地址')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
