<template>
  <Card size="sm">
    <CardHeader class="border-b border-dashed border-border/70">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <CardTitle>Preflight 报告</CardTitle>
          <CardDescription>{{
            report
              ? `生成于 ${formatDate(report.generated_at)}`
              : "选择项目后生成只读检查报告"
          }}</CardDescription>
        </div>
        <div v-if="report" class="flex items-center gap-2">
          <Button variant="outline" size="sm" type="button" @click="$emit('copy-summary')">
            <Copy class="size-4" />
            复制摘要
          </Button>
          <AdminStatusBadge :tone="statusLevelTone(report.status_level)">
            {{ statusLevelLabel(report.status_level) }}
          </AdminStatusBadge>
        </div>
      </div>
    </CardHeader>
    <CardContent class="space-y-4 pt-3">
      <div
        v-if="!report && !generating"
        class="flex min-h-56 items-center justify-center rounded-xl border border-dashed border-border/70 text-center"
      >
        <div>
          <FileSearch class="mx-auto size-8 text-muted-foreground/50" />
          <p class="mt-3 text-sm font-bold">尚未生成报告</p>
          <p class="mt-1 text-xs text-muted-foreground">
            报告会显示阻断项、警告项和每个只读检查的证据摘要。
          </p>
        </div>
      </div>

      <div
        v-if="generating"
        class="flex min-h-56 items-center justify-center rounded-xl border border-dashed border-border/70"
      >
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <LoaderCircle class="size-4 animate-spin" />
          正在读取项目台账并生成报告
        </div>
      </div>

      <template v-if="report && !generating">
        <section class="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            size="sm"
            :variant="detailMode === 'all' ? 'default' : 'outline'"
            @click="$emit('update:detail-mode', 'all')"
          >
            全部
          </Button>
          <Button
            type="button"
            size="sm"
            :variant="detailMode === 'needs-action' ? 'default' : 'outline'"
            @click="$emit('update:detail-mode', 'needs-action')"
          >
            需处理项
          </Button>
          <Button type="button" size="sm" variant="outline" @click="$emit('copy-checklist')">
            <Copy class="size-4" />
            复制清单
          </Button>
        </section>

        <section class="grid gap-3 sm:grid-cols-4">
          <div class="rounded-xl border border-dashed border-border/80 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">阻断项</p>
 <p class="mt-2 text-2xl font-black" :class="report.blocking_count ? 'text-rose-600': 'text-emerald-600'">
              {{ report.blocking_count }}
            </p>
          </div>
          <div class="rounded-xl border border-dashed border-border/80 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">警告项</p>
 <p class="mt-2 text-2xl font-black" :class="report.warning_count ? 'text-amber-600': 'text-emerald-600'">
              {{ report.warning_count }}
            </p>
          </div>
          <div class="rounded-xl border border-dashed border-border/80 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">通过项</p>
            <p class="mt-2 text-2xl font-black text-emerald-600">{{ report.pass_count }}</p>
          </div>
          <div class="rounded-xl border border-dashed border-border/80 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">信息项</p>
            <p class="mt-2 text-2xl font-black text-blue-600">{{ report.info_count }}</p>
          </div>
        </section>

        <section class="rounded-xl border border-dashed border-border/70 bg-muted/20 p-3">
          <p class="text-sm font-bold">{{ report.summary }}</p>
          <p class="mt-1 text-[10px] text-muted-foreground">
            项目：{{ report.project }} · {{ environmentLabel(report.environment) }} ·
            {{ statusLevelLabel(report.status_level) }}
          </p>
        </section>

        <section v-if="nextActions.length" class="rounded-xl border border-dashed border-border/70 bg-background p-3">
          <div class="flex items-center justify-between gap-3">
            <p class="text-xs font-black">下一步建议</p>
            <span class="text-[10px] text-muted-foreground">{{ nextActions.length }} 条</span>
          </div>
          <ul class="mt-2 space-y-1">
            <li v-for="action in nextActions" :key="action" class="text-[10px] text-muted-foreground">
              {{ action }}
            </li>
          </ul>
        </section>

        <section v-if="categories.length" class="rounded-xl border border-dashed border-border/70 bg-background p-3">
          <div class="flex items-center justify-between gap-3">
            <p class="text-xs font-black">类别分布</p>
            <span class="text-[10px] text-muted-foreground">{{ categories.length }} 类</span>
          </div>
          <div class="mt-3 grid gap-2 md:grid-cols-2">
            <button
              v-for="group in categories"
              :key="group.category"
              type="button"
              class="rounded-lg border border-dashed px-3 py-2 text-left transition hover:bg-muted/40"
 :class="activeCategory === group.category ? 'border-primary/40 bg-muted/60': 'border-border/70'"
              @click="$emit('update:active-category', group.category)"
            >
              <div class="flex items-center justify-between gap-3">
                <p class="truncate text-xs font-black">{{ group.label }}</p>
                <span class="font-mono text-[10px] text-muted-foreground">{{ group.total_count }}</span>
              </div>
              <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
                <div class="h-full rounded-full" :class="groupBarClass(group)" :style="{ width: groupRiskWidth(group) }" />
              </div>
              <p class="mt-1 text-[10px] text-muted-foreground">
                阻断 {{ group.blocking_count }} · 警告 {{ group.warning_count }}
              </p>
            </button>
          </div>
        </section>

        <section class="flex flex-wrap gap-2">
          <Button
            v-for="item in categoryCards"
            :key="item.key"
            type="button"
            size="sm"
            :variant="activeCategory === item.key ? 'default' : 'outline'"
            @click="$emit('update:active-category', item.key)"
          >
            <component :is="categoryIcon(item.key)" class="size-4" />
            {{ item.label }}
          </Button>
        </section>

        <section class="overflow-hidden rounded-xl border border-dashed border-border/70">
          <div
            v-for="check in visibleChecks"
            :key="check.key"
            class="grid gap-3 border-b border-dashed border-border/70 p-3 last:border-b-0 md:grid-cols-[10rem_minmax(0,1fr)_6rem]"
          >
            <div class="flex items-center gap-2">
              <span class="flex size-7 shrink-0 items-center justify-center rounded-full" :class="checkIconClass(check.status)">
                <component :is="checkIcon(check.status)" class="size-4" />
              </span>
              <span class="min-w-0">
                <span class="block truncate text-sm font-black">{{ check.label }}</span>
                <span class="mt-0.5 block text-[9px] font-bold uppercase tracking-widest text-muted-foreground/60">
                  {{ categoryLabel(check.category) }}
                </span>
              </span>
            </div>
            <div class="min-w-0">
              <p class="text-xs font-bold">{{ check.message }}</p>
              <p v-if="check.detail" class="mt-1 break-words font-mono text-[10px] text-muted-foreground">
                {{ check.detail }}
              </p>
            </div>
            <div class="flex items-start md:justify-end">
              <AdminStatusBadge :tone="checkTone(check.status)">
                {{ checkStatusLabel(check.status) }}
              </AdminStatusBadge>
            </div>
          </div>
        </section>
      </template>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  Boxes,
  CircleCheck,
  Cloud,
  Copy,
  FileSearch,
  GitCommit,
  Info,
  LoaderCircle,
  Network,
  Server,
  TriangleAlert,
  XCircle,
} from "@lucide/vue";
import AdminStatusBadge, { type AdminStatusTone } from "@/components/admin/AdminStatusBadge.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { OpsDeploymentPreflight, OpsDeploymentPreflightGroup } from "@/api/ops";
import {
  categoryLabel,
  checkStatusLabel,
  environmentLabel,
  formatDeploymentPreflightDate as formatDate,
  statusLevelLabel,
} from "@/lib/deploymentPreflightPresentation";

type DetailMode = "all" | "needs-action";

const props = defineProps<{
  report: OpsDeploymentPreflight | null;
  generating: boolean;
  activeCategory: string;
  detailMode: DetailMode;
}>();

defineEmits<{
  "copy-summary": [];
  "copy-checklist": [];
  "update:active-category": [value: string];
  "update:detail-mode": [value: DetailMode];
}>();

const nextActions = computed(() => props.report?.next_actions || []);
const categories = computed(() => props.report?.categories || []);
const filteredChecks = computed(() => {
  if (!props.report) return [];
  if (props.activeCategory === "all") return props.report.checks;
  return props.report.checks.filter((check) => (check.category || "other") === props.activeCategory);
});
const visibleChecks = computed(() => (
  props.detailMode === "needs-action"
    ? filteredChecks.value.filter((check) => check.status === "block" || check.status === "warning")
    : filteredChecks.value
));
const categoryCards = computed(() => {
  const checks = props.report?.checks || [];
  return [
    {
      key: "all",
      label: "全部",
      total: checks.length,
      block: checks.filter((check) => check.status === "block").length,
      warning: checks.filter((check) => check.status === "warning").length,
    },
    ...categories.value.map((group) => ({
      key: group.category,
      label: group.label,
      total: group.total_count,
      block: group.blocking_count,
      warning: group.warning_count,
    })),
  ];
});

const statusLevelTone = (value?: string): AdminStatusTone => {
  if (value === "ready") return "green";
  if (value === "review") return "amber";
  if (value === "blocked") return "coral";
  return "gray";
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
const checkTone = (status: string): AdminStatusTone => {
  if (status === "pass") return "green";
  if (status === "warning") return "amber";
  if (status === "block") return "coral";
  if (status === "info") return "blue";
  return "gray";
};
const categoryIcon = (value: string) => (
  {
    all: FileSearch,
    compose: Boxes,
    version: GitCommit,
    boundary: Network,
    infra: Server,
    connector: Cloud,
    domain: Cloud,
    runtime: CircleCheck,
    evidence: Info,
  }[value] || Info
);
const checkIcon = (status: string) => (
  {
    pass: CircleCheck,
    warning: TriangleAlert,
    block: XCircle,
    info: Info,
  }[status] || Info
);
const checkIconClass = (status: string): string => (
  {
    pass: "bg-emerald-500/10 text-emerald-700",
    warning: "bg-amber-500/10 text-amber-700",
    block: "bg-rose-500/10 text-rose-700",
    info: "bg-blue-500/10 text-blue-700",
  }[status] || "bg-muted text-muted-foreground"
);
</script>
