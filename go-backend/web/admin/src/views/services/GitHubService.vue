<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="服务中心 / GitHub"
      description="独立管理 GitHub 仓库授权、GHCR 镜像读取和项目发布引用状态。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 border bg-background px-3 text-sm"
          aria-label="筛选 GitHub 环境"
          :disabled="loading || Boolean(oauthPending)"
          @change="changeEnvironment"
        >
          <option value="">全部环境</option>
          <option
            v-for="option in opsConnectorEnvironmentOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
        <Button
          size="icon"
          variant="outline"
          title="刷新 GitHub / GHCR 状态"
          :disabled="loading"
          @click="loadGitHub"
        >
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
        <Button
          v-if="canManage"
          :disabled="Boolean(oauthPending)"
          :title="connections.length ? '重新授权 GitHub' : '连接 GitHub'"
          @click="startOAuth()"
        >
          <LoaderCircle v-if="oauthPending" class="size-4 animate-spin" />
          <KeyRound v-else class="size-4" />
          {{ connections.length ? "重新授权" : "连接 GitHub" }}
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <div
        v-for="item in summaryItems"
        :key="item.label"
        class="border bg-card p-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">
              {{ item.label }}
            </p>
            <p class="mt-2 text-2xl font-black" :class="item.tone">
              {{ item.value }}
            </p>
          </div>
          <component :is="item.icon" class="size-4 text-muted-foreground" />
        </div>
        <p class="mt-2 truncate text-[10px] text-muted-foreground">
          {{ item.detail }}
        </p>
      </div>
    </section>

    <section
      class="flex flex-col gap-3 border border-dashed border-border/80 bg-muted/20 p-4 sm:flex-row sm:items-center sm:justify-between"
    >
      <div class="flex items-start gap-3">
        <GitBranch class="mt-0.5 size-5 shrink-0 text-foreground" />
        <div>
          <p class="text-sm font-black">GitHub OAuth</p>
          <p class="mt-1 text-xs text-muted-foreground">
            {{ oauthSummary }}
          </p>
        </div>
      </div>
      <AdminStatusBadge :tone="attentionCount ? 'amber' : activeConnectionCount ? 'green' : 'gray'">
        {{ statusSummary }}
      </AdminStatusBadge>
    </section>

    <OpsGitHubIntegrationPanel
      :connectors="connections"
      :projects="projects"
      :github-loading="loading"
      :github-o-auth-loading="oauthPending"
      :testing-connector-id="testingID"
      :can-manage="canManage"
      :show-actions="false"
      @connect="startOAuth"
      @refresh="loadGitHub"
      @configure="openConnectorConfig"
      @test="testGitHubConnector"
    />

    <section class="border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-dashed px-4 py-3">
        <div class="flex items-start gap-3">
          <Package class="mt-0.5 size-5 shrink-0 text-primary" />
          <div>
            <h2 class="text-base font-black">已读取仓库</h2>
            <p class="mt-1 text-xs text-muted-foreground">
              {{ repositorySummary }}
            </p>
          </div>
        </div>
        <AdminStatusBadge :tone="repositories.length ? 'blue' : 'gray'">
          {{ generatedLabel }}
        </AdminStatusBadge>
      </div>

      <div
        v-if="repositoryReadErrors.length"
        class="border-b border-dashed border-amber-500/30 bg-amber-500/5 px-4 py-3"
      >
        <div class="flex items-start gap-2">
          <TriangleAlert class="mt-0.5 size-4 shrink-0 text-amber-600" />
          <div class="min-w-0 text-xs text-amber-800">
            <p class="font-black">仓库读取未全部完成</p>
            <p
              v-for="item in repositoryReadErrors"
              :key="item"
              class="mt-1 truncate"
              :title="item"
            >
              {{ item }}
            </p>
          </div>
        </div>
      </div>

      <div
        v-if="loading"
        class="p-8 text-center text-sm text-muted-foreground"
      >
        正在读取 GitHub 仓库
      </div>
      <div
        v-else-if="repositories.length === 0"
        class="p-8 text-center text-sm text-muted-foreground"
      >
        尚未读取到 GitHub 仓库
      </div>
      <div v-else class="divide-y">
        <a
          v-for="repository in repositories"
          :key="`${repository.connector_id}:${repository.id}`"
          :href="repository.html_url || '#'"
          target="_blank"
          rel="noreferrer"
          class="grid gap-3 px-4 py-3 text-xs transition hover:bg-muted/30 sm:grid-cols-[minmax(0,1fr)_auto_auto_auto] sm:items-center"
          :class="repository.html_url ? '' : 'pointer-events-none'"
        >
          <span class="min-w-0">
            <span class="block truncate font-black">
              {{ repository.full_name || repository.name }}
            </span>
            <span class="mt-1 block truncate text-[10px] text-muted-foreground">
              {{ repository.connector_name }} · {{ repository.default_branch || "-" }}
            </span>
          </span>
          <span>{{ visibilityLabel(repository) }}</span>
          <span>{{ formatDate(repository.pushed_at || repository.updated_at) }}</span>
          <ExternalLink class="size-4 text-muted-foreground" />
        </a>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { toast } from "vue-sonner";
import {
  ExternalLink,
  GitBranch,
  KeyRound,
  LoaderCircle,
  Package,
  RefreshCw,
  TriangleAlert,
} from "@lucide/vue";
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue";
import AdminStatusBadge from "@/components/admin/AdminStatusBadge.vue";
import OpsGitHubIntegrationPanel from "@/components/admin/ops/OpsGitHubIntegrationPanel.vue";
import { opsConnectorEnvironmentOptions } from "@/modules/ops/opsConnectorBindingForm";
import { Button } from "@/components/ui/button";
import servicesApi, {
  type ServiceCenterGitHub,
  type ServiceCenterGitHubRepository,
} from "@/api/services";
import type { OpsConnector, OpsEnvironment } from "@/api/ops";
import {
  readOpsEnvironmentQuery,
  withOpsEnvironmentQuery,
} from "@/lib/opsEnvironment";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const canManage = computed(() => authStore.hasPermission("services:manage"));
const loading = ref(false);
const oauthPending = ref(false);
const testingID = ref(0);
const github = ref<ServiceCenterGitHub | null>(null);
const environmentFilter = ref<OpsEnvironment | "">(
  readOpsEnvironmentQuery(route.query.environment),
);

const connections = computed<OpsConnector[]>(() => github.value?.connections || []);
const projects = computed(() => github.value?.projects || []);
const repositories = computed<ServiceCenterGitHubRepository[]>(
  () => github.value?.repositories || [],
);
const repositoryReadErrors = computed(
  () => github.value?.repository_read_errors || [],
);
const gitHubConnections = computed(() =>
  connections.value.filter((connection) => connection.provider === "github"),
);
const ghcrConnections = computed(() =>
  connections.value.filter((connection) => connection.provider === "ghcr"),
);
const activeConnectionCount = computed(
  () => github.value?.active_connection_count || 0,
);
const attentionCount = computed(() => github.value?.attention_count || 0);
const generatedLabel = computed(() =>
  github.value?.generated_at ? formatDate(github.value.generated_at) : "尚未读取",
);
const summaryItems = computed(() => [
  {
    label: "GitHub",
    value: gitHubConnections.value.length,
    detail: `${activeProviderConnections("github")} 个可用仓库连接`,
    tone: "",
    icon: GitBranch,
  },
  {
    label: "GHCR",
    value: ghcrConnections.value.length,
    detail: `${activeProviderConnections("ghcr")} 个可用镜像连接`,
    tone: "",
    icon: Package,
  },
  {
    label: "授权",
    value: github.value?.credential_configured_count || 0,
    detail: "OAuth token / 凭据配置",
    tone: "text-primary",
    icon: KeyRound,
  },
  {
    label: "关联项目",
    value: github.value?.project_count || 0,
    detail: "项目台账引用 GitHub / GHCR",
    tone: "",
    icon: GitBranch,
  },
  {
    label: "已读仓库",
    value: github.value?.repository_count || 0,
    detail: repositoryReadErrors.value.length
      ? `${repositoryReadErrors.value.length} 个读取异常`
      : "按最近更新排序",
    tone: repositoryReadErrors.value.length ? "text-amber-600" : "text-emerald-600",
    icon: Package,
  },
]);
const oauthSummary = computed(() => {
  if (!connections.value.length) {
    return "当前环境还没有 GitHub / GHCR 连接。";
  }
  if (!github.value?.credential_configured_count) {
    return "连接器已登记，但还没有可用 OAuth token。";
  }
  return `已配置 ${github.value.credential_configured_count} 个凭据；仓库和镜像能力由服务中心统一读取。`;
});
const statusSummary = computed(() => {
  if (attentionCount.value) return `${attentionCount.value} 项待处理`;
  if (activeConnectionCount.value) return `${activeConnectionCount.value} 个连接可用`;
  return "未连接";
});
const repositorySummary = computed(() => {
  if (repositoryReadErrors.value.length) {
    return `${repositories.value.length} 个仓库已读取，${repositoryReadErrors.value.length} 个连接读取失败。`;
  }
  if (repositories.value.length) {
    return `${repositories.value.length} 个仓库来自 GitHub OAuth 授权。`;
  }
  return "完成 GitHub OAuth 后会显示可读取的最近仓库。";
});

const emptyGitHubState = (): ServiceCenterGitHub => ({
  environment: environmentFilter.value || "",
  generated_at: new Date().toISOString(),
  connection_count: 0,
  active_connection_count: 0,
  credential_configured_count: 0,
  project_count: 0,
  repository_count: 0,
  attention_count: 0,
  connections: [],
  projects: [],
  repositories: [],
  repository_read_errors: [],
});

const isNotFoundError = (error: any): boolean => error?.response?.status === 404;

const activeProviderConnections = (provider: string): number =>
  connections.value.filter(
    (connection) => connection.provider === provider && connection.enabled && connection.status === "active",
  ).length;

const loadGitHub = async (): Promise<void> => {
  loading.value = true;
  try {
    github.value = await servicesApi.getGitHub(environmentFilter.value || undefined);
  } catch (error: any) {
    if (isNotFoundError(error)) {
      github.value = emptyGitHubState();
      return;
    }
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "GitHub 服务加载失败",
    );
  } finally {
    loading.value = false;
  }
};

const changeEnvironment = (): void => {
  void router.replace({
    query: withOpsEnvironmentQuery(route.query, environmentFilter.value),
  });
  void loadGitHub();
};

const oauthReturnPath = (): string => {
  const query = new URLSearchParams();
  if (environmentFilter.value) query.set("environment", environmentFilter.value);
  return `/services/github${query.size ? `?${query.toString()}` : ""}`;
};

const startOAuth = async (): Promise<void> => {
  oauthPending.value = true;
  try {
    const result = await servicesApi.startGitHubOAuth(
      undefined,
      oauthReturnPath(),
      environmentFilter.value || "production",
    );
    window.location.assign(result.authorization_url);
  } catch (error: any) {
    oauthPending.value = false;
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "GitHub OAuth 启动失败，请检查后台 OAuth 配置",
    );
  }
};

const testGitHubConnector = async (connector: OpsConnector): Promise<void> => {
  testingID.value = connector.id;
  try {
    const result = await servicesApi.testGitHubConnection(connector.id);
    toast[result.success ? "success" : "error"](
      result.message || (result.success ? "连接测试成功" : "连接测试失败"),
    );
    await loadGitHub();
  } catch (error: any) {
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "GitHub / GHCR 连接测试失败",
    );
  } finally {
    testingID.value = 0;
  }
};

const openConnectorConfig = (): void => {
  void router.push({
    path: "/services/connectors",
    query: environmentFilter.value ? { environment: environmentFilter.value } : undefined,
  });
};

const handleOAuthReturn = (): void => {
  const query = new URLSearchParams(window.location.search);
  const status = query.get("ops_oauth_status");
  if (!status) return;
  const provider = query.get("ops_oauth_provider");
  if (provider && provider !== "github") return;
  const connected = status === "connected" || status === "connected_with_warnings";
  const message =
    query.get("ops_oauth_message") ||
    (connected ? "GitHub OAuth 绑定完成" : "GitHub OAuth 绑定失败");
  if (status === "connected_with_warnings") {
    toast.warning(message);
  } else {
    toast[connected ? "success" : "error"](message);
  }
  [
    "ops_oauth_status",
    "ops_oauth_provider",
    "ops_oauth_message",
    "ops_oauth_connector_id",
    "ops_oauth_connector_name",
  ].forEach((key) => query.delete(key));
  const nextSearch = query.toString();
  window.history.replaceState(
    {},
    document.title,
    `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}${window.location.hash}`,
  );
};

const formatDate = (value?: string): string =>
  value ? new Date(value).toLocaleString("zh-CN") : "-";

const visibilityLabel = (repository: ServiceCenterGitHubRepository): string => {
  if (repository.visibility) return repository.visibility;
  return repository.private ? "private" : "public";
};

watch(
  () => route.query.environment,
  (value) => {
    const nextEnvironment = readOpsEnvironmentQuery(value);
    if (nextEnvironment === environmentFilter.value) return;
    environmentFilter.value = nextEnvironment;
    void loadGitHub();
  },
);

onMounted(() => {
  handleOAuthReturn();
  void loadGitHub();
});
</script>

