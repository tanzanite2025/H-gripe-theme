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
import { LoaderCircle, RefreshCw, Save, Trash2 } from "@lucide/vue";
import { toast } from "vue-sonner";
import AdminFormField from "@/components/admin/AdminFormField.vue";
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue";
import { Button } from "@/components/ui/button";
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
  notes: "",
});

const splitCSV = (value: string): string[] =>
  String(value || "")
    .split(/[\s,;，；]+/)
    .map((item) => item.trim())
    .filter(Boolean);

const joinCSV = (values?: string[] | null): string =>
  Array.isArray(values) ? values.join(", ") : "";

const assignSettings = (
  payload: PaymentProviderInstallmentsSettings | null | undefined,
) => {
  settings.enabled = payload?.enabled === true;
  settings.payment_method_types_text = joinCSV(payload?.payment_method_types);
  settings.countries_text = joinCSV(payload?.countries);
  settings.currencies_text = joinCSV(payload?.currencies);
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
