<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 部署中心"
      description="先生成 Preflight，再通过工作流完成验证、审批、受控发布、发布后健康检查和缓存清理。"
    >
      <template #actions>
        <Button
          variant="outline"
          :disabled="loadingProjects || generating"
          @click="loadProjects"
          v-if="activeTab === 'overview'"
        >
          <RefreshCw
 :class="['size-4', loadingProjects ? 'animate-spin': '']"
          />
          刷新项目
        </Button>
        <Button
          variant="outline"
          :disabled="!overview || loadingProjects || generating"
          @click="copyOverview"
          v-if="activeTab === 'overview'"
        >
          <Copy class="size-4" />
          复制总览
        </Button>
        <Button
          :disabled="!selectedProjectId || generating"
          @click="generateReport"
          v-if="activeTab === 'overview'"
        >
          <LoaderCircle v-if="generating" class="size-4 animate-spin" />
          <FileSearch v-else class="size-4" />
          生成报告
        </Button>
        <Button
          :disabled="!selectedProjectId || workflowBusy"
          @click="createDryRun"
          v-if="activeTab === 'workflow'"
        >
          <LoaderCircle v-if="workflowBusy" class="size-4 animate-spin" />
          <CircleCheck v-else class="size-4" />
          创建 dry-run
        </Button>
        <Button
          variant="outline"
          :disabled="!selectedProjectId || workflowBusy || selectedProject?.environment !== 'production'"
          title="仅生产环境项目可以创建生产发布工作流"
          @click="createProduction"
          v-if="activeTab === 'workflow'"
        >
          <ShieldAlert class="size-4" />
          创建生产工作流
        </Button>
      </template>
    </AdminPageHeader>

    <nav class="flex flex-wrap gap-1 border-b border-border/70" aria-label="部署中心视图">
      <button
        v-for="tab in deploymentTabs"
        :key="tab.value"
        type="button"
        class="border-b-2 px-3 py-2 text-xs font-black transition-colors"
        :class="activeTab === tab.value
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="selectTab(tab.value)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <section v-if="activeTab === 'overview'" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="item in overviewStats"
        :key="item.key"
        class="rounded-xl border border-dashed border-border/80 bg-background p-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p
              class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60"
            >
              {{ item.label }}
            </p>
            <p class="mt-2 text-2xl font-black" :class="item.valueClass">
              {{ item.value }}
            </p>
          </div>
          <span
            class="flex size-8 shrink-0 items-center justify-center rounded-full"
            :class="item.iconClass"
          >
            <component :is="item.icon" class="size-4" />
          </span>
        </div>
        <p class="mt-2 truncate text-[10px] text-muted-foreground">
          {{ item.detail }}
        </p>
      </div>
    </section>

    <section
      v-if="activeTab === 'overview' && overviewRiskCategories.length"
      class="grid gap-3 md:grid-cols-2 xl:grid-cols-4"
    >
      <button
        v-for="group in overviewRiskCategories"
        :key="group.category"
        type="button"
        class="rounded-xl border border-dashed border-border/80 bg-background p-3 text-left transition hover:bg-muted/40"
        @click="focusCategory(group.category)"
      >
        <div class="flex items-center justify-between gap-3">
          <p class="text-xs font-black">{{ group.label }}</p>
          <AdminStatusBadge :tone="groupTone(group)">{{
            groupStatus(group)
          }}</AdminStatusBadge>
        </div>
        <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-muted">
          <div
            class="h-full rounded-full"
            :class="groupBarClass(group)"
            :style="{ width: groupRiskWidth(group) }"
          />
        </div>
        <p class="mt-2 text-[10px] text-muted-foreground">
          阻断 {{ group.blocking_count }} · 警告 {{ group.warning_count }} ·
          总检查 {{ group.total_count }}
        </p>
      </button>
    </section>

    <section v-if="activeTab === 'workflow'" class="space-y-3">
      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>工作流项目</CardTitle>
          <CardDescription>选择要验证、审批或执行的项目。</CardDescription>
        </CardHeader>
        <CardContent class="pt-3">
          <select
            v-model.number="selectedProjectId"
            class="h-10 w-full rounded-md border bg-background px-3 text-sm"
            :disabled="loadingProjects || workflowBusy"
          >
            <option :value="0">请选择项目</option>
            <option v-for="project in projectOptions" :key="project.id" :value="project.id">
              {{ project.name }} · {{ environmentLabel(project.environment) }}
            </option>
          </select>
        </CardContent>
      </Card>
      <OpsDeploymentWorkflowPanel
        :selected-project-id="selectedProjectId"
        :workflow="workflow"
        :workflows="workflows"
        :workflow-busy="workflowBusy"
        @validate="validateWorkflow"
        @approve="approveWorkflow"
        @execute="executeWorkflow"
        @retry="retryWorkflow"
        @rollback="rollbackWorkflow"
        @cancel="cancelWorkflow"
        @select="workflow = $event"
      />
    </section>

    <section
      v-if="activeTab === 'overview'"
      class="grid gap-3 xl:grid-cols-[minmax(18rem,0.64fr)_minmax(0,1.36fr)]"
    >
      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>项目选择</CardTitle>
          <CardDescription>{{
            overview
              ? `总览生成于 ${formatDate(overview.generated_at)}`
              : "报告按项目生成，覆盖 Compose、镜像、边界、VPS、连接器、健康和备份记录"
          }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 pt-3">
          <div class="grid gap-2 sm:grid-cols-3 xl:grid-cols-1 2xl:grid-cols-3">
            <select
              v-model="overviewStatusFilter"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="all">全部状态</option>
              <option value="blocked">只看阻断</option>
              <option value="review">只看 REVIEW</option>
              <option value="ready">只看 READY</option>
            </select>
            <select
              v-model="overviewEnvironmentFilter"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="">全部环境</option>
              <option
                v-for="option in opsEnvironmentOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
            <select
              v-model="overviewSort"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="risk">风险优先</option>
              <option value="name">项目名</option>
              <option value="generated">生成时间</option>
            </select>
          </div>

          <select
            v-model.number="selectedProjectId"
            class="h-10 w-full rounded-md border bg-background px-3 text-sm"
            :disabled="loadingProjects || generating"
          >
            <option :value="0">请选择项目</option>
            <option
              v-for="project in projectOptions"
              :key="project.id"
              :value="project.id"
            >
              {{ project.name }} · {{ environmentLabel(project.environment) }}
            </option>
          </select>

          <div
            v-if="selectedProject"
            class="rounded-xl border border-dashed border-border/70 p-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="truncate text-sm font-black">
                  {{ selectedProject.name }}
                </p>
                <p
                  class="mt-1 truncate font-mono text-[10px] text-muted-foreground"
                >
                  {{
                    selectedProject.compose_project_name ||
                    "未登记 Compose 项目名"
                  }}
                </p>
              </div>
              <AdminStatusBadge
                :tone="healthTone(selectedProject.health_status)"
              >
                {{ healthLabel(selectedProject.health_status) }}
              </AdminStatusBadge>
            </div>
            <dl class="mt-3 grid gap-2 text-[10px] text-muted-foreground">
              <div class="flex justify-between gap-3">
                <dt>Compose</dt>
                <dd class="truncate font-mono">
                  {{ selectedProject.compose_source || "-" }}
                </dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>镜像</dt>
                <dd class="truncate font-mono">
                  {{ selectedProject.current_image_tag || "-" }}
                </dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>Commit</dt>
                <dd class="truncate font-mono">
                  {{ selectedProject.current_commit_sha || "-" }}
                </dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>VPS</dt>
 <dd class="truncate">{{ selectedProject.vps_name ||"-" }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>最近观察</dt>
                <dd class="truncate">
                  {{
                    selectedProject.last_checked_at
                      ? formatDate(selectedProject.last_checked_at)
                      : "未同步"
                  }}
                </dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>容器</dt>
                <dd class="truncate font-mono">
                  {{ containerSummary(selectedProject) }}
                </dd>
              </div>
            </dl>
            <div class="mt-3 border-t border-dashed border-border/70 pt-3">
              <div class="flex items-center justify-between gap-3">
                <label
                  for="ops-deployment-requested-ref"
                  class="text-xs font-black"
                  >发布引用</label
                >
                <AdminStatusBadge tone="blue">Dry-run</AdminStatusBadge>
              </div>
              <input
                id="ops-deployment-requested-ref"
                v-model.trim="requestedRef"
                type="text"
                class="mt-2 h-9 w-full rounded-md border bg-background px-3 font-mono text-xs"
                placeholder="master、sha-... 或 sha256:..."
              />
              <div
                v-if="requestedRefCandidates.length"
                class="mt-2 flex flex-wrap gap-2"
              >
                <Button
                  v-for="candidate in requestedRefCandidates"
                  :key="candidate.value"
                  type="button"
                  size="sm"
                  variant="outline"
                  :disabled="candidate.value === requestedRef"
                  @click="requestedRef = candidate.value"
                >
                  {{ candidate.label }}
                </Button>
              </div>
            </div>
          </div>

          <div v-if="overviewProjects.length" class="space-y-2">
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs font-black">项目 Preflight</p>
              <span class="text-[10px] text-muted-foreground"
                >{{ filteredOverviewProjects.length }} /
                {{ overviewProjects.length }} 个项目</span
              >
            </div>
            <button
              v-for="summary in filteredOverviewProjects"
              :key="summary.project_id"
              type="button"
              class="w-full rounded-xl border border-dashed px-3 py-2 text-left transition hover:bg-muted/40 disabled:pointer-events-none disabled:opacity-60"
              :class="
                summary.project_id === selectedProjectId
                  ? 'border-primary/40 bg-muted/60'
                  : 'border-border/70'
              "
              :disabled="generating"
              @click="selectOverviewProject(summary.project_id)"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-xs font-black">
                    {{ summary.project }}
                  </p>
                  <p class="mt-0.5 truncate text-[10px] text-muted-foreground">
                    {{ environmentLabel(summary.environment) }} · 阻断
                    {{ summary.blocking_count }} · 警告
                    {{ summary.warning_count }}
                  </p>
                </div>
                <AdminStatusBadge :tone="statusLevelTone(summary.status_level)">
                  {{ statusLevelLabel(summary.status_level) }}
                </AdminStatusBadge>
              </div>
              <p class="mt-2 line-clamp-2 text-[10px] text-muted-foreground">
                {{ summaryReason(summary) }}
              </p>
            </button>
            <div
              v-if="!filteredOverviewProjects.length"
              class="rounded-xl border border-dashed border-border/70 p-3 text-center"
            >
              <p class="text-xs font-bold">没有匹配的项目</p>
              <p class="mt-1 text-[10px] text-muted-foreground">
                调整状态或环境筛选后再查看。
              </p>
            </div>
          </div>

          <div
            class="rounded-xl border border-dashed border-amber-500/30 bg-amber-500/5 p-3"
          >
            <div class="flex items-start gap-2">
              <ShieldAlert class="mt-0.5 size-4 shrink-0 text-amber-600" />
              <p class="text-xs text-muted-foreground">
                Dry-run 不会调用 Hostinger 更新、Cloudflare 写入、SSH、Docker
                restart
                或缓存清理接口；生产工作流只会在健康检查成功后按已绑定域名清理
                Cloudflare 缓存。
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <OpsDeploymentPreflightReportPanel
        :report="report"
        :generating="generating"
        :active-category="activeCategory"
        :detail-mode="detailMode"
        @copy-summary="copyReportSummary"
        @copy-checklist="copyFullChecklist"
        @update:active-category="activeCategory = $event"
        @update:detail-mode="detailMode = $event"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { toast } from "vue-sonner";
import { useRoute, useRouter } from "vue-router";
import {
  CircleCheck,
  Copy,
  FileSearch,
  LoaderCircle,
  RefreshCw,
  Server,
  ShieldAlert,
  TriangleAlert,
  XCircle,
} from "@lucide/vue";
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue";
import AdminStatusBadge, {
  type AdminStatusTone,
} from "@/components/admin/AdminStatusBadge.vue";
import OpsDeploymentWorkflowPanel from "@/components/admin/ops/OpsDeploymentWorkflowPanel.vue";
import OpsDeploymentPreflightReportPanel from "@/components/admin/ops/OpsDeploymentPreflightReportPanel.vue";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import opsApi, {
  type OpsDeploymentPreflight,
  type OpsDeploymentPreflightGroup,
  type OpsDeploymentPreflightOverview,
  type OpsDeploymentWorkflow,
  type OpsProject,
} from "@/api/ops";
import {
  useDeploymentPreflightOverview,
  type DeploymentPreflightOverviewSort,
  type DeploymentPreflightOverviewStatus,
} from "@/composables/useDeploymentPreflightOverview";
import {
  buildPreflightChecklistMarkdown,
  buildPreflightOverviewMarkdown,
  buildPreflightReportMarkdown,
  environmentLabel,
  formatDeploymentPreflightDate as formatDate,
  statusLevelLabel,
  summaryReason,
} from "@/lib/deploymentPreflightPresentation";
import {
  opsEnvironmentOptions,
  readOpsEnvironmentQuery,
} from "@/lib/opsEnvironment";

const route = useRoute();
const router = useRouter();

const queryValue = (value: unknown): string =>
  typeof value === "string" ? value : "";
const queryNumber = (value: unknown): number => {
  const parsed = Number.parseInt(queryValue(value), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
};
const queryChoice = <T extends string>(
  value: unknown,
  allowed: readonly T[],
  fallback: T,
): T => {
  const candidate = queryValue(value) as T;
  return allowed.includes(candidate) ? candidate : fallback;
};

type DeploymentTab = "overview" | "workflow";

const deploymentTabs: Array<{ value: DeploymentTab; label: string }> = [
  { value: "overview", label: "发布总览" },
  { value: "workflow", label: "工作流" },
];

const projects = ref<OpsProject[]>([]);
const overview = ref<OpsDeploymentPreflightOverview | null>(null);
const selectedProjectId = ref(queryNumber(route.query.project));
const report = ref<OpsDeploymentPreflight | null>(null);
const loadingProjects = ref(false);
const generating = ref(false);
const workflows = ref<OpsDeploymentWorkflow[]>([]);
const workflow = ref<OpsDeploymentWorkflow | null>(null);
const workflowBusy = ref(false);
const requestedRef = ref("");
const activeCategory = ref(queryValue(route.query.category) || "all");
const detailMode = ref(
  queryChoice(route.query.mode, ["all", "needs-action"] as const, "all"),
);
const overviewStatusFilter = ref<DeploymentPreflightOverviewStatus>(
  queryChoice(
    route.query.status,
    ["all", "blocked", "review", "ready"] as const,
    "all",
  ),
);
const overviewEnvironmentFilter = ref(
  readOpsEnvironmentQuery(route.query.environment, "production"),
);
const overviewSort = ref<DeploymentPreflightOverviewSort>(
  queryChoice(route.query.sort, ["risk", "name", "generated"] as const, "risk"),
);
const activeTab = ref<DeploymentTab>(
  queryChoice(route.query.tab, ["overview", "workflow"] as const, "overview"),
);
let projectLoadSequence = 0;
let workflowLoadSequence = 0;
let reportLoadSequence = 0;

const selectedProject = computed(
  () =>
    projects.value.find((project) => project.id === selectedProjectId.value) ||
    null,
);

const selectTab = (tab: DeploymentTab): void => {
  activeTab.value = tab;
};
const requestedRefCandidates = computed(() => {
  const values = [
    { label: "Commit", value: selectedProject.value?.current_commit_sha || "" },
    { label: "镜像", value: selectedProject.value?.current_image_tag || "" },
    { label: "master", value: "master" },
  ];
  const seen = new Set<string>();
  return values.reduce<Array<{ label: string; value: string }>>(
    (items, item) => {
      const value = item.value.trim();
      if (!value || seen.has(value)) return items;
      seen.add(value);
      items.push({ label: item.label, value });
      return items;
    },
    [],
  );
});
const {
  overviewProjects,
  overviewCategories,
  overviewRiskCategories,
  filteredOverviewProjects,
  projectOptions,
  overviewStats: overviewStatsBase,
  mergeReportSummary,
} = useDeploymentPreflightOverview({
  projects,
  overview,
  statusFilter: overviewStatusFilter,
  environmentFilter: overviewEnvironmentFilter,
  sort: overviewSort,
});
const overviewStats = computed(() =>
  overviewStatsBase.value.map((item) => ({
    ...item,
    icon:
      item.key === "projects"
        ? FileSearch
        : item.key === "ready"
          ? CircleCheck
          : item.key === "review"
            ? TriangleAlert
            : XCircle,
    valueClass:
      item.key === "projects"
        ? "text-foreground"
        : item.key === "ready"
          ? "text-emerald-600"
          : item.key === "review"
            ? item.value > 0
              ? "text-amber-600"
              : "text-emerald-600"
            : item.value > 0
              ? "text-rose-600"
              : "text-emerald-600",
    iconClass:
      item.key === "projects"
        ? "bg-blue-500/10 text-blue-700"
        : item.key === "ready"
          ? "bg-emerald-500/10 text-emerald-700"
          : item.key === "review"
            ? "bg-amber-500/10 text-amber-700"
            : "bg-rose-500/10 text-rose-700",
  })),
);
const loadWorkflows = async (): Promise<void> => {
  const projectID = selectedProjectId.value;
  const requestSequence = ++workflowLoadSequence;
  if (!projectID) {
    workflows.value = [];
    workflow.value = null;
    return;
  }
  try {
    const nextWorkflows = await opsApi.listWorkflows(projectID);
    if (requestSequence !== workflowLoadSequence || projectID !== selectedProjectId.value) {
      return;
    }
    workflows.value = nextWorkflows;
    if (workflow.value) {
      const refreshed = workflows.value.find(
        (item) => item.id === workflow.value?.id,
      );
      workflow.value = refreshed || workflows.value[0] || null;
    } else {
      workflow.value = workflows.value[0] || null;
    }
  } catch (error: any) {
    if (requestSequence !== workflowLoadSequence || projectID !== selectedProjectId.value) {
      return;
    }
    workflows.value = [];
    workflow.value = null;
    toast.warning(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "工作流记录加载失败",
    );
  }
};

const loadProjects = async (): Promise<void> => {
  const requestSequence = ++projectLoadSequence;
  const environment = overviewEnvironmentFilter.value || undefined;
  loadingProjects.value = true;
  reportLoadSequence++;
  report.value = null;
  generating.value = false;
  overview.value = null;
  workflows.value = [];
  workflow.value = null;
  try {
    const [projectsResult, overviewResult] = await Promise.allSettled([
      opsApi.listProjects(environment),
      opsApi.getDeploymentPreflightOverview(environment),
    ]);

    if (requestSequence !== projectLoadSequence) {
      return;
    }

    if (projectsResult.status === "fulfilled") {
      projects.value = Array.isArray(projectsResult.value?.projects)
        ? projectsResult.value.projects
        : [];
    } else {
      projects.value = [];
    }

    if (overviewResult.status === "fulfilled") {
      overview.value = overviewResult.value;
    }

    if (
      projectsResult.status === "rejected" &&
      overviewResult.status === "rejected"
    ) {
      throw overviewResult.reason || projectsResult.reason;
    }

    if (projectsResult.status === "rejected") {
      toast.warning("项目台账明细加载失败，已显示 preflight 总览");
    }
    if (overviewResult.status === "rejected") {
      toast.warning("preflight 总览加载失败，仍可手动生成单项目报告");
    }

    const currentProjectIsAvailable = projectOptions.value.some(
      (project) => project.id === selectedProjectId.value,
    );
    const nextProjectID = currentProjectIsAvailable
      ? selectedProjectId.value
      : projectOptions.value[0]?.id || 0;
    const projectChanged = nextProjectID !== selectedProjectId.value;
    selectedProjectId.value = nextProjectID;
    if (!projectChanged) {
      requestedRef.value = defaultRequestedRef(selectedProject.value);
      await loadWorkflows();
    }
  } catch (error: any) {
    if (requestSequence !== projectLoadSequence) {
      return;
    }
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "部署项目列表加载失败",
    );
  } finally {
    if (requestSequence === projectLoadSequence) {
      loadingProjects.value = false;
    }
  }
};

const handleProjectChange = (): void => {
  report.value = null;
  workflow.value = null;
  workflows.value = [];
  requestedRef.value = defaultRequestedRef(selectedProject.value);
  activeCategory.value = "all";
  detailMode.value = "all";
  void loadWorkflows();
};

const selectOverviewProject = async (projectID: number): Promise<void> => {
  if (!projectID || projectID === selectedProjectId.value) {
    if (!report.value) {
      await generateReport();
    }
    return;
  }
  selectedProjectId.value = projectID;
  await generateReport();
};

const generateReport = async (): Promise<void> => {
  if (!selectedProjectId.value) {
    toast.error("请选择项目");
    return;
  }
  const projectID = selectedProjectId.value;
  const requestSequence = ++reportLoadSequence;
  generating.value = true;
  try {
    const nextReport = await opsApi.getProjectPreflight(projectID);
    if (
      requestSequence !== reportLoadSequence ||
      projectID !== selectedProjectId.value ||
      (overviewEnvironmentFilter.value && nextReport.environment !== overviewEnvironmentFilter.value)
    ) {
      return;
    }
    report.value = nextReport;
    mergeReportSummary(nextReport);
    activeCategory.value = "all";
    detailMode.value = "all";
  } catch (error: any) {
    if (requestSequence !== reportLoadSequence) {
      return;
    }
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "preflight 报告生成失败",
    );
  } finally {
    if (requestSequence === reportLoadSequence) {
      generating.value = false;
    }
  }
};

const refreshWorkflow = async (
  operation: () => Promise<OpsDeploymentWorkflow>,
  successMessage: string,
): Promise<void> => {
  const projectID = selectedProjectId.value;
  workflowBusy.value = true;
  try {
    const result = await operation();
    if (projectID !== selectedProjectId.value) {
      return;
    }
    workflow.value = result;
    await loadWorkflows();
    toast.success(successMessage);
  } catch (error: any) {
    toast.error(
      error?.response?.data?.message ||
        error?.response?.data?.error ||
        "工作流操作失败",
    );
  } finally {
    workflowBusy.value = false;
  }
};

const createDryRun = async (): Promise<void> => {
  if (!selectedProjectId.value) return;
  await refreshWorkflow(
    () =>
      opsApi.createDryRun(selectedProjectId.value, requestedDeploymentRef()),
    "dry-run 工作流已创建",
  );
};

const createProduction = async (): Promise<void> => {
  if (!selectedProjectId.value) return;
  if (selectedProject.value?.environment !== "production") {
    toast.error("仅生产环境项目可以创建生产发布工作流");
    return;
  }
  const requested = requestedDeploymentRef();
  const refNotice =
    requested !== "master"
      ? `当前发布引用 ${requested} 仅用于 dry-run；生产执行器仍使用 master。`
      : "";
  if (
    !window.confirm(
      `${refNotice}将创建生产发布工作流。真正更新 Hostinger 仍需重新验证并人工审批，是否继续？`,
    )
  )
    return;
  await refreshWorkflow(
    () => opsApi.createProduction(selectedProjectId.value, "master"),
    "生产工作流已创建，等待 Preflight 和审批",
  );
};

const validateWorkflow = async (): Promise<void> => {
  if (!workflow.value) return;
  await refreshWorkflow(
    () => opsApi.validateWorkflow(workflow.value!.id),
    "Preflight 已重新验证",
  );
};

const approveWorkflow = async (): Promise<void> => {
  if (!workflow.value) return;
  await refreshWorkflow(
    () => opsApi.approveWorkflow(workflow.value!.id),
    workflow.value.mode === "production" ? "生产发布已审批" : "dry-run 已审批",
  );
};

const executeWorkflow = async (): Promise<void> => {
  if (!workflow.value) return;
  const isProduction = workflow.value.mode === "production";
  if (
    isProduction &&
    !window.confirm(
      "该操作会更新 Hostinger Docker 项目并触发容器重建，发布后还会按绑定域名清理 Cloudflare 缓存；缓存清理失败不会自动回滚源站，是否确认执行？",
    )
  )
    return;
  await refreshWorkflow(
    () => opsApi.executeWorkflow(workflow.value!.id),
    isProduction
      ? "生产发布执行完成或已进入人工处理状态"
      : "只读 dry-run 步骤已执行",
  );
};

const retryWorkflow = async (): Promise<void> => {
  if (!workflow.value) return;
  if (!workflowHasRetryableFailure(workflow.value)) {
    toast.error("当前失败步骤不可重试，需要人工处理或回滚");
    return;
  }
  const isProduction = workflow.value.mode === "production";
  if (
    isProduction &&
    !window.confirm(
      "将从可重试失败步骤继续生产工作流；已成功步骤不会重跑，是否确认？",
    )
  )
    return;
  await refreshWorkflow(
    () => opsApi.retryWorkflow(workflow.value!.id),
    isProduction
      ? "生产工作流已从失败步骤继续执行"
      : "dry-run 已从失败步骤继续执行",
  );
};

const rollbackWorkflow = async (): Promise<void> => {
  if (!workflow.value || workflow.value.mode !== "production") return;
  const rollbackRef = (
    workflow.value.rollback_ref ||
    workflow.value.previous_ref ||
    ""
  ).trim();
  if (!/^[0-9a-f]{40}$/i.test(rollbackRef)) {
    toast.error("当前工作流没有可执行的完整 Commit SHA 回滚点");
    return;
  }
  if (
    !window.confirm(
      `将通过受限 SSH 在绑定 VPS 上执行 DEPLOY_REF=${rollbackRef} ./deploy.sh，并重新核验健康状态，是否继续？`,
    )
  )
    return;
  await refreshWorkflow(
    () => opsApi.rollbackWorkflow(workflow.value!.id),
    "回滚执行完成或已保留人工处理证据",
  );
};

const cancelWorkflow = async (): Promise<void> => {
  if (!workflow.value) return;
  await refreshWorkflow(
    () => opsApi.cancelWorkflow(workflow.value!.id),
    "工作流已取消",
  );
};

const focusCategory = (category: string): void => {
  activeCategory.value = category || "all";
  detailMode.value = "needs-action";
};

const copyReportSummary = async (): Promise<void> => {
  if (!report.value) return;
  const markdown = buildPreflightReportMarkdown(report.value);
  try {
    await navigator.clipboard.writeText(markdown);
    toast.success("preflight 摘要已复制");
  } catch {
    toast.error("复制失败，请检查浏览器剪贴板权限");
  }
};

const copyFullChecklist = async (): Promise<void> => {
  if (!report.value) return;
  const markdown = buildPreflightChecklistMarkdown(report.value);
  try {
    await navigator.clipboard.writeText(markdown);
    toast.success("完整清单已复制");
  } catch {
    toast.error("复制失败，请检查浏览器剪贴板权限");
  }
};

const copyOverview = async (): Promise<void> => {
  if (!overview.value) return;
  try {
    await navigator.clipboard.writeText(
      buildPreflightOverviewMarkdown({
        overview: overview.value,
        projects: filteredOverviewProjects.value,
        filterLabel: overviewFilterLabel(),
      }),
    );
    toast.success("当前总览已复制");
  } catch {
    toast.error("复制失败，请检查浏览器剪贴板权限");
  }
};

const overviewFilterLabel = (): string =>
  [
    overviewStatusFilter.value === "all"
      ? ""
      : `状态=${statusLevelLabel(overviewStatusFilter.value)}`,
    !overviewEnvironmentFilter.value
      ? ""
      : `环境=${environmentLabel(overviewEnvironmentFilter.value)}`,
    `排序=${overviewSort.value === "risk" ? "风险优先" : overviewSort.value === "name" ? "项目名" : "生成时间"}`,
  ]
    .filter(Boolean)
    .join(" · ");

const defaultRequestedRef = (project: OpsProject | null): string =>
  project?.current_commit_sha?.trim() ||
  project?.current_image_tag?.trim() ||
  "master";

const requestedDeploymentRef = (): string =>
  requestedRef.value.trim() || defaultRequestedRef(selectedProject.value);

const groupTone = (group: OpsDeploymentPreflightGroup): AdminStatusTone => {
  if (group.blocking_count > 0) return "coral";
  if (group.warning_count > 0) return "amber";
  return "green";
};

const groupStatus = (group: OpsDeploymentPreflightGroup): string => {
  if (group.blocking_count > 0) return "BLOCK";
  if (group.warning_count > 0) return "WARN";
  return "PASS";
};

const groupBarClass = (group: OpsDeploymentPreflightGroup): string => {
  if (group.blocking_count > 0) return "bg-rose-500";
  if (group.warning_count > 0) return "bg-amber-500";
  return "bg-emerald-500";
};

const groupRiskWidth = (group: OpsDeploymentPreflightGroup): string => {
  const risky = group.blocking_count + group.warning_count;
  const total = Math.max(group.total_count, 1);
  return `${Math.max(Math.round((risky / total) * 100), risky > 0 ? 12 : 0)}%`;
};

const statusLevelTone = (value?: string): AdminStatusTone => {
  if (value === "ready") return "green";
  if (value === "review") return "amber";
  if (value === "blocked") return "coral";
  return "gray";
};

const workflowHasRetryableFailure = (item: OpsDeploymentWorkflow): boolean =>
  ["failed", "paused", "rollback_required"].includes(item.status) &&
  (item.steps || []).some(
    (step) =>
      (step.status === "failed" || step.status === "running") && step.retryable,
  );

const healthLabel = (value: string): string =>
  ({
    healthy: "健康",
    degraded: "降级",
    unknown: "未同步",
    offline: "离线",
  })[value] ||
  value ||
  "-";

const healthTone = (value: string): AdminStatusTone => {
  if (value === "healthy") return "green";
  if (value === "degraded" || value === "unknown") return "amber";
  if (value === "offline") return "coral";
  return "gray";
};

const containerSummary = (project: OpsProject): string => {
  if (!project.last_checked_at) return "未同步";
  return `${project.observed_container_count || 0} / ${project.observed_running_container_count || 0} / ${project.observed_healthy_container_count || 0}`;
};

watch(
  selectedProjectId,
  (projectID, previousProjectID) => {
    if (projectID === previousProjectID) return;
    handleProjectChange();
  },
);

watch(
  overviewEnvironmentFilter,
  (environment, previousEnvironment) => {
    if (environment === previousEnvironment) return;
    selectedProjectId.value = 0;
    requestedRef.value = "";
    activeCategory.value = "all";
    detailMode.value = "all";
    void loadProjects();
  },
);

watch(
  () => route.query.environment,
  (value) => {
    const nextEnvironment = readOpsEnvironmentQuery(value, "production");
    if (nextEnvironment !== overviewEnvironmentFilter.value) {
      overviewEnvironmentFilter.value = nextEnvironment;
    }
  },
);

watch(
  () => route.query.tab,
  (value) => {
    const nextTab = queryChoice(
      value,
      ["overview", "workflow"] as const,
      "overview",
    );
    if (nextTab !== activeTab.value) activeTab.value = nextTab;
  },
);

watch(
  [
    selectedProjectId,
    activeCategory,
    detailMode,
    overviewStatusFilter,
    overviewEnvironmentFilter,
    overviewSort,
    activeTab,
  ],
  async ([projectID, category, mode, status, environment, sort, tab]) => {
    const query: Record<string, string> = {};
    if (projectID) query.project = String(projectID);
    if (category && category !== "all") query.category = category;
    if (mode && mode !== "all") query.mode = mode;
    if (status && status !== "all") query.status = status;
    query.environment = environment || "all";
    if (sort && sort !== "risk") query.sort = sort;
    if (tab && tab !== "overview") query.tab = tab;
    if (JSON.stringify(route.query) !== JSON.stringify(query)) {
      await router.replace({ query });
    }
  },
);

onMounted(async () => {
  await loadProjects();
  if (!selectedProjectId.value && projectOptions.value.length) {
    selectedProjectId.value = projectOptions.value[0].id;
  }
  if (selectedProjectId.value) {
    await generateReport();
    await loadWorkflows();
  }
});
</script>
