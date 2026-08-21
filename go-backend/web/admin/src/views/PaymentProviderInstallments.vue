<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      :title="`${providerLabel} 分期`"
      :description="providerDescription"
    >
      <template #actions>
        <div class="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="loadSettings"
          >
            <RefreshCw :class="['size-4', { 'animate-spin': loading }]" />
            刷新
          </Button>
          <Button
            type="button"
            variant="outline"
            :disabled="loading || saving || !canEdit"
            @click="clearSettings"
          >
            <Trash2 class="size-4" />
            清空
          </Button>
          <Button
            type="button"
            :disabled="saving || loading || !canEdit"
            @click="saveSettings"
          >
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? "保存中" : "保存分期设置" }}
          </Button>
        </div>
      </template>
    </AdminPageHeader>

    <div class="min-h-0 flex-1 overflow-auto">
      <div class="space-y-6">
        <section
          class="max-w-4xl space-y-4 border-b border-dashed border-border/80 pb-6"
        >
          <div class="grid gap-4 md:grid-cols-2">
            <div class="rounded-lg border bg-background p-4">
              <p
                class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60"
              >
                Provider
              </p>
              <p class="mt-2 text-sm font-black text-foreground">
                {{ providerLabel }}
              </p>
            </div>
            <div class="rounded-lg border bg-background p-4">
              <p
                class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60"
              >
                Status
              </p>
              <p
                class="mt-2 text-sm font-black"
                :class="
                  settings.enabled
                    ? 'text-emerald-600 dark:text-emerald-300'
                    : 'text-muted-foreground'
                "
              >
                {{ settings.enabled ? "已启用" : "未启用" }}
              </p>
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="启用分期能力">
              <div
                class="flex min-h-10 items-center justify-between rounded-md border border-input bg-background px-3"
              >
                <span class="text-sm text-foreground"
                  >控制该渠道是否允许分期相关配置生效</span
                >
                <Switch
                  v-model="settings.enabled"
                  :disabled="loading || saving || !canEdit"
                />
              </div>
            </AdminFormField>

            <AdminFormField
              label="适用国家"
              description="逗号分隔，例如 US, GB, DE。"
            >
              <Textarea
                v-model="settings.countries_text"
                :disabled="loading || saving || !canEdit"
                class="min-h-24 font-mono"
                placeholder="US, GB, DE"
              />
              <div class="mt-2 flex flex-wrap gap-2">
                <Button
                  v-for="country in commonCountryPresets"
                  :key="country"
                  type="button"
                  variant="outline"
                  size="xs"
                  :disabled="loading || saving || !canEdit"
                  @click="appendCsvValue('countries_text', country, true)"
                >
                  <Plus class="size-3" />
                  {{ country }}
                </Button>
              </div>
            </AdminFormField>

            <AdminFormField
              label="适用币种"
              description="逗号分隔，例如 USD, GBP, EUR。"
            >
              <Textarea
                v-model="settings.currencies_text"
                :disabled="loading || saving || !canEdit"
                class="min-h-24 font-mono"
                placeholder="USD, GBP, EUR"
              />
              <div class="mt-2 flex flex-wrap gap-2">
                <Button
                  v-for="currency in commonCurrencyPresets"
                  :key="currency"
                  type="button"
                  variant="outline"
                  size="xs"
                  :disabled="loading || saving || !canEdit"
                  @click="appendCsvValue('currencies_text', currency, true)"
                >
                  <Plus class="size-3" />
                  {{ currency }}
                </Button>
              </div>
            </AdminFormField>

            <AdminFormField
              v-if="providerKey === 'stripe'"
              label="Payment Method Types"
              description="逗号分隔，直接写入 Stripe PaymentIntent，例如 card, klarna, affirm, afterpay_clearpay。"
            >
              <Textarea
                v-model="settings.payment_method_types_text"
                :disabled="loading || saving || !canEdit"
                class="min-h-24 font-mono"
                placeholder="card, klarna"
              />
              <div class="mt-2 flex flex-wrap gap-2">
                <Button
                  v-for="method in stripePaymentMethodPresets"
                  :key="method"
                  type="button"
                  variant="outline"
                  size="xs"
                  :disabled="loading || saving || !canEdit"
                  @click="appendCsvValue('payment_method_types_text', method)"
                >
                  <Plus class="size-3" />
                  {{ method }}
                </Button>
              </div>
            </AdminFormField>

            <AdminFormField
              label="最小金额"
              description="低于此金额时自动回退普通卡支付。0 表示不限制。"
            >
              <Input
                v-model="settings.min_amount_text"
                type="number"
                min="0"
                step="0.01"
                :disabled="loading || saving || !canEdit"
                placeholder="100.00"
              />
            </AdminFormField>

            <AdminFormField
              label="最大金额"
              description="高于此金额时自动回退普通卡支付。0 表示不限制。"
            >
              <Input
                v-model="settings.max_amount_text"
                type="number"
                min="0"
                step="0.01"
                :disabled="loading || saving || !canEdit"
                placeholder="5000.00"
              />
            </AdminFormField>

            <AdminFormField class="md:col-span-2" label="备注">
              <Textarea
                v-model="settings.notes"
                :disabled="loading || saving || !canEdit"
                class="min-h-24"
                placeholder="这里可以写这个渠道分期策略的内部说明"
              />
            </AdminFormField>
          </div>
        </section>

        <section
          class="max-w-4xl border-l-2 border-primary/40 pl-4 text-sm text-muted-foreground"
        >
          <p v-if="providerKey === 'stripe'">
            Stripe 会把这里的 `payment_method_types` 直接用在 PaymentIntent
            上；如果国家或币种不匹配，就会回退到普通 card。
          </p>
          <p v-else-if="providerKey === 'paypal'">
            PayPal 这页先保留分期可见性和适用范围，后续再接 Pay Later
            的前台展示。
          </p>
          <p v-else>这个渠道目前只保留分期设置占位和说明。</p>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { LoaderCircle, Plus, RefreshCw, Save, Trash2 } from "@lucide/vue";
import { toast } from "vue-sonner";
import AdminFormField from "@/components/admin/AdminFormField.vue";
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { useAuthStore } from "@/stores/auth";
import {
  getPaymentChannelLabel,
  normalizePaymentChannelKey,
} from "@/lib/paymentChannels";
import {
  paymentInstallmentsApi,
  type PaymentProviderInstallmentsSettings,
} from "@/api/paymentInstallments";

const props = withDefaults(
  defineProps<{
    provider?: string;
  }>(),
  {
    provider: "stripe",
  },
);

const authStore = useAuthStore();
const canEdit = computed(() => authStore.hasPermission("settings:edit"));
const providerKey = computed(() =>
  normalizePaymentChannelKey(String(props.provider || "")),
);
const providerLabel = computed(() =>
  getPaymentChannelLabel(providerKey.value, "收款渠道"),
);
const providerDescription = computed(() => {
  if (providerKey.value === "stripe") {
    return "配置 Stripe 的分期/BNPL 方法、适用国家和币种。";
  }
  if (providerKey.value === "paypal") {
    return "配置 PayPal 的分期可见性和适用边界。";
  }
  return `${providerLabel.value} 分期能力的独立路由占位，后续接入该渠道自己的分期策略。`;
});

const loading = ref(false);
const saving = ref(false);
const settings = reactive({
  enabled: false,
  payment_method_types_text: "",
  countries_text: "",
  currencies_text: "",
  min_amount_text: "",
  max_amount_text: "",
  notes: "",
});

const stripePaymentMethodPresets = [
  "card",
  "klarna",
  "affirm",
  "afterpay_clearpay",
];
const commonCountryPresets = ["US", "CA", "GB", "AU", "NZ", "DE", "FR", "NL"];
const commonCurrencyPresets = ["USD", "CAD", "GBP", "EUR", "AUD", "NZD"];

const splitCSV = (value: string): string[] =>
  String(value || "")
    .split(/[\s,;，；]+/)
    .map((item) => item.trim())
    .filter(Boolean);

const joinCSV = (values?: string[] | null): string =>
  Array.isArray(values) ? values.join(", ") : "";

const formatAmount = (value?: number | null): string =>
  typeof value === "number" && Number.isFinite(value) && value > 0
    ? value.toFixed(2)
    : "";

const parseAmount = (value: string): number => {
  const amount = Number(String(value || "").trim());
  return Number.isFinite(amount) && amount > 0 ? amount : 0;
};

type CsvField =
  | "payment_method_types_text"
  | "countries_text"
  | "currencies_text";

const appendCsvValue = (field: CsvField, value: string, upper = false) => {
  const normalized = upper
    ? value.trim().toUpperCase()
    : value.trim().toLowerCase();
  if (!normalized) return;
  const items = splitCSV(settings[field]).map((item) =>
    upper ? item.toUpperCase() : item.toLowerCase(),
  );
  if (items.includes(normalized)) return;
  if (field === "payment_method_types_text" && normalized === "card") {
    settings[field] = joinCSV([
      normalized,
      ...items.filter((item) => item !== normalized),
    ]);
    return;
  }
  settings[field] = joinCSV([...items, normalized]);
};

const assignSettings = (
  payload: PaymentProviderInstallmentsSettings | null | undefined,
) => {
  settings.enabled = payload?.enabled === true;
  settings.payment_method_types_text = joinCSV(payload?.payment_method_types);
  settings.countries_text = joinCSV(payload?.countries);
  settings.currencies_text = joinCSV(payload?.currencies);
  settings.min_amount_text = formatAmount(payload?.min_amount);
  settings.max_amount_text = formatAmount(payload?.max_amount);
  settings.notes = String(payload?.notes || "");
};

const loadSettings = async () => {
  loading.value = true;
  try {
    const payload = await paymentInstallmentsApi.get(
      providerKey.value || String(props.provider || ""),
    );
    assignSettings(payload);
  } catch (error) {
    console.error("Failed to load payment installments settings:", error);
    toast.error("分期设置加载失败");
    assignSettings({
      provider: providerKey.value || String(props.provider || ""),
      enabled: false,
    });
  } finally {
    loading.value = false;
  }
};

const buildPayload = () => ({
  enabled: settings.enabled,
  payment_method_types:
    providerKey.value === "stripe"
      ? splitCSV(settings.payment_method_types_text)
      : [],
  countries: splitCSV(settings.countries_text).map((item) =>
    item.toUpperCase(),
  ),
  currencies: splitCSV(settings.currencies_text).map((item) =>
    item.toUpperCase(),
  ),
  min_amount: parseAmount(settings.min_amount_text),
  max_amount: parseAmount(settings.max_amount_text),
  notes: String(settings.notes || "").trim(),
});

const saveSettings = async () => {
  if (!canEdit.value || saving.value || !providerKey.value) return;
  saving.value = true;
  try {
    const payload = buildPayload();
    const result = await paymentInstallmentsApi.update(
      providerKey.value,
      payload,
    );
    assignSettings(result);
    toast.success("分期设置已保存");
  } catch (error) {
    console.error("Failed to save payment installments settings:", error);
    toast.error("分期设置保存失败");
  } finally {
    saving.value = false;
  }
};

const clearSettings = async () => {
  if (!canEdit.value || saving.value || !providerKey.value) return;
  if (!window.confirm(`确认清空 ${providerLabel.value} 的分期设置？`)) return;
  saving.value = true;
  try {
    await paymentInstallmentsApi.remove(providerKey.value);
    assignSettings({ provider: providerKey.value, enabled: false });
    toast.success("分期设置已清空");
  } catch (error) {
    console.error("Failed to clear payment installments settings:", error);
    toast.error("分期设置清空失败");
  } finally {
    saving.value = false;
  }
};

watch(providerKey, () => {
  void loadSettings();
});

onMounted(() => {
  void loadSettings();
});
</script>
