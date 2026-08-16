<template>
  <section
    v-if="selectedProjectId"
    class="grid gap-3 xl:grid-cols-[minmax(18rem,0.72fr)_minmax(0,1.28fr)]"
  >
    <Card size="sm">
      <CardHeader class="border-b border-dashed border-border/70">
        <CardTitle>{{
          workflow?.mode === "production" ? "生产发布工作流" : "Dry-run 工作流"
        }}</CardTitle>
        <CardDescription>
          {{
            workflow?.mode === "production"
              ? "生产工作流必须经过 Preflight 和人工审批；更新或健康检查失败会停在回滚处理状态，缓存清理失败会保留失败证据。"
              : "Dry-run 只执行只读步骤，不会修改 Hostinger、Cloudflare、Docker 或网关。"
          }}
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-3 pt-3">
        <div
          v-if="workflow"
          class="rounded-xl border border-dashed border-border/70 p-3"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-sm font-black">工作流 #{{ workflow.id }}</p>
              <p
                class="mt-1 truncate font-mono text-[10px] text-muted-foreground"
              >
                {{ workflow.requested_ref }}
              </p>
            </div>
            <AdminStatusBadge :tone="workflowTone(workflow.status)">
              {{ workflowStatusLabel(workflow.status) }}
            </AdminStatusBadge>
          </div>
          <dl class="mt-3 grid gap-2 text-[10px] text-muted-foreground">
            <div class="flex justify-between gap-3">
              <dt>Preflight</dt>
              <dd>{{ statusLevelLabel(workflow.preflight_status) }}</dd>
            </div>
            <div class="flex justify-between gap-3">
              <dt>创建人</dt>
              <dd class="truncate">{{ workflow.created_by || "-" }}</dd>
            </div>
            <div v-if="workflow.approved_by" class="flex justify-between gap-3">
              <dt>审批人</dt>
              <dd class="truncate">{{ workflow.approved_by }}</dd>
            </div>
            <div
              v-if="workflow.completed_at"
              class="flex justify-between gap-3"
            >
              <dt>完成时间</dt>
              <dd>{{ formatDate(workflow.completed_at) }}</dd>
            </div>
            <div
              v-if="workflow.rollback_ref || workflow.previous_ref"
              class="flex justify-between gap-3"
            >
              <dt>回滚点</dt>
              <dd class="truncate font-mono">
                {{ workflow.rollback_ref || workflow.previous_ref }}
              </dd>
            </div>
            <div
              v-if="workflow.remote_operation_id"
              class="flex justify-between gap-3"
            >
              <dt>远程操作</dt>
              <dd class="truncate font-mono">
                {{ workflow.remote_operation_id }}
              </dd>
            </div>
            <div
              v-if="workflow.health_status"
              class="flex justify-between gap-3"
            >
              <dt>发布后健康</dt>
              <dd>{{ workflow.health_status }}</dd>
            </div>
          </dl>
          <p
            v-if="workflow.last_error"
            class="mt-3 rounded-lg bg-rose-500/5 p-2 text-[10px] text-rose-700"
          >
            {{ workflow.last_error }}
          </p>
        </div>
        <div
          v-else
          class="rounded-xl border border-dashed border-border/70 p-3 text-center"
        >
          <p class="text-xs font-bold">尚未创建工作流</p>
          <p class="mt-1 text-[10px] text-muted-foreground">
            创建后可以重新验证当前台账，再提交审批。
          </p>
        </div>

        <div class="flex flex-wrap gap-2">
          <Button
            v-if="
              workflow &&
              ['draft', 'awaiting_approval', 'validated'].includes(
                workflow.status,
              )
            "
            size="sm"
            variant="outline"
            :disabled="workflowBusy"
            @click="$emit('validate')"
          >
            <RefreshCw class="size-4" />
            重新验证
          </Button>
          <Button
            v-if="workflow?.status === 'awaiting_approval'"
            size="sm"
            :disabled="workflowBusy"
            @click="$emit('approve')"
          >
            <CircleCheck class="size-4" />
            {{
              workflow.mode === "production" ? "审批生产发布" : "审批 dry-run"
            }}
          </Button>
          <Button
            v-if="workflow?.status === 'validated'"
            size="sm"
            :disabled="workflowBusy"
            @click="$emit('execute')"
          >
            <ShieldAlert v-if="workflow.mode === 'production'" class="size-4" />
            <FileSearch v-else class="size-4" />
            {{
              workflow.mode === "production" ? "执行生产发布" : "执行只读步骤"
            }}
          </Button>
          <Button
            v-if="
              workflow &&
              ['failed', 'paused', 'rollback_required'].includes(
                workflow.status,
              )
            "
            size="sm"
            variant="outline"
            :disabled="workflowBusy || !workflowHasRetryableFailure(workflow)"
            :title="
              workflowHasRetryableFailure(workflow)
                ? '从第一个可重试失败步骤继续'
                : '当前失败步骤不可重试，需要人工处理或回滚'
            "
            @click="$emit('retry')"
          >
            <RefreshCw class="size-4" />
            重试/继续
          </Button>
          <Button
            v-if="canExecuteRollback(workflow)"
            size="sm"
            variant="destructive"
            :disabled="workflowBusy"
            :title="
              rollbackStepSucceeded(workflow)
                ? '重新执行回滚后的远程健康核验'
                : '使用已记录的完整 Commit SHA 执行受限 SSH 回滚'
            "
            @click="$emit('rollback')"
          >
            <Undo2 class="size-4" />
            {{ rollbackStepSucceeded(workflow) ? "复核回滚" : "执行回滚" }}
          </Button>
          <Button
            v-if="
              workflow &&
              ['draft', 'awaiting_approval', 'validated'].includes(
                workflow.status,
              )
            "
            size="sm"
            variant="outline"
            :disabled="workflowBusy"
            @click="$emit('cancel')"
          >
            <XCircle class="size-4" />
            取消
          </Button>
        </div>

        <div v-if="workflow?.steps?.length" class="space-y-2">
          <div class="flex items-center justify-between gap-3">
            <p class="text-xs font-black">步骤证据</p>
            <span class="text-[10px] text-muted-foreground"
              >{{ workflow.steps.length }} 步</span
            >
          </div>
          <div
            v-for="step in workflow.steps"
            :key="step.id"
            class="flex items-start justify-between gap-3 rounded-lg border border-dashed border-border/70 px-3 py-2"
          >
            <div class="min-w-0">
              <p class="truncate text-xs font-bold">
                {{ step.sequence }}. {{ step.label }}
              </p>
              <p class="mt-1 line-clamp-2 text-[10px] text-muted-foreground">
                {{ step.output_summary || "等待执行" }}
              </p>
            </div>
            <AdminStatusBadge :tone="workflowStepTone(step.status)">
              {{ workflowStepStatusLabel(step.status) }}
            </AdminStatusBadge>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card size="sm">
      <CardHeader class="border-b border-dashed border-border/70">
        <CardTitle>最近工作流</CardTitle>
        <CardDescription
          >保留最近 25 次工作流记录，便于查看审批和 dry-run
          证据。</CardDescription
        >
      </CardHeader>
      <CardContent class="space-y-2 pt-3">
        <div
          v-if="!workflows.length"
          class="rounded-xl border border-dashed border-border/70 p-4 text-center text-xs text-muted-foreground"
        >
          暂无工作流记录
        </div>
        <button
          v-for="item in workflows"
          :key="item.id"
          type="button"
          class="w-full rounded-xl border border-dashed px-3 py-2 text-left transition hover:bg-muted/40"
          :class="
            item.id === workflow?.id
              ? 'border-primary/40 bg-muted/60'
              : 'border-border/70'
          "
          @click="$emit('select', item)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate text-xs font-black">
                #{{ item.id }} · {{ item.requested_ref }}
              </p>
              <p class="mt-1 text-[10px] text-muted-foreground">
                {{ formatDate(item.created_at) }} · {{ item.created_by || "-" }}
              </p>
            </div>
            <AdminStatusBadge :tone="workflowTone(item.status)">
              {{ workflowStatusLabel(item.status) }}
            </AdminStatusBadge>
          </div>
        </button>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import {
  CircleCheck,
  FileSearch,
  RefreshCw,
  ShieldAlert,
  Undo2,
  XCircle,
} from "@lucide/vue";
import AdminStatusBadge, {
  type AdminStatusTone,
} from "@/components/admin/AdminStatusBadge.vue";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { OpsDeploymentWorkflow } from "@/api/ops";
import {
  formatDeploymentPreflightDate as formatDate,
  statusLevelLabel,
} from "@/lib/deploymentPreflightPresentation";

defineProps<{
  selectedProjectId: number;
  workflow: OpsDeploymentWorkflow | null;
  workflows: OpsDeploymentWorkflow[];
  workflowBusy: boolean;
}>();

defineEmits<{
  validate: [];
  approve: [];
  execute: [];
  retry: [];
  rollback: [];
  cancel: [];
  select: [workflow: OpsDeploymentWorkflow];
}>();

const workflowStatusLabel = (value: string): string =>
  ({
    draft: "草稿",
    awaiting_approval: "待审批",
    validated: "已审批",
    running: "执行中",
    succeeded: "已完成",
    failed: "失败",
    cancelled: "已取消",
    paused: "已暂停",
    rollback_required: "需回滚",
    rolled_back: "已回滚",
  })[value] ||
  value ||
  "-";

const workflowTone = (value: string): AdminStatusTone => {
  if (value === "succeeded" || value === "validated" || value === "rolled_back")
    return "green";
  if (value === "awaiting_approval" || value === "running") return "amber";
  if (value === "failed" || value === "rollback_required") return "coral";
  if (value === "cancelled" || value === "draft") return "gray";
  return "blue";
};

const workflowStepStatusLabel = (value: string): string =>
  ({
    pending: "待执行",
    running: "执行中",
    succeeded: "完成",
    failed: "失败",
    skipped: "跳过",
  })[value] ||
  value ||
  "-";

const workflowStepTone = (value: string): AdminStatusTone => {
  if (value === "succeeded") return "green";
  if (value === "running") return "amber";
  if (value === "failed") return "coral";
  if (value === "skipped") return "gray";
  return "blue";
};

const workflowHasRetryableFailure = (item: OpsDeploymentWorkflow): boolean =>
  ["failed", "paused", "rollback_required"].includes(item.status) &&
  (item.steps || []).some(
    (step) =>
      (step.status === "failed" || step.status === "running") && step.retryable,
  );

const rollbackStepSucceeded = (item: OpsDeploymentWorkflow | null): boolean =>
  Boolean(
    item?.steps?.some(
      (step) => step.key === "execute_rollback" && step.status === "succeeded",
    ),
  );

const canExecuteRollback = (item: OpsDeploymentWorkflow | null): boolean =>
  Boolean(
    item &&
    item.mode === "production" &&
    item.status === "rollback_required" &&
    /^[0-9a-f]{40}$/i.test((item.rollback_ref || "").trim()),
  );
</script>
