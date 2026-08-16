<template>
  <Card size="sm">
    <CardHeader class="border-b border-dashed border-border/70">
      <CardTitle>首次发布流程</CardTitle>
      <CardDescription>DEV 和生产共用同一套受审计的账号维护逻辑。</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4 pt-5 text-xs leading-5">
      <div class="space-y-1">
        <p class="font-black">1. DEV</p>
        <p class="text-muted-foreground">先用 `adminctl ensure-admin` 创建 DEV 首个账号，再登录本页维护其他账号。</p>
      </div>
      <div class="space-y-1">
        <p class="font-black">2. 发布</p>
        <p class="text-muted-foreground">镜像只提供 `/app/adminctl`，服务启动不会重置任何账号密码。</p>
      </div>
      <div class="space-y-1">
        <p class="font-black">3. 生产</p>
        <p class="text-muted-foreground">生产数据库没有后台账号时，在目标容器中完成一次 bootstrap；之后直接从本页轮换密码。</p>
      </div>
      <div class="rounded-lg border border-border/70 bg-muted/30 p-3 font-mono text-[11px] leading-5">
        ADMIN_EMAIL=admin@example.com<br>
        ADMIN_PASSWORD_FILE=/run/secrets/admin-password<br>
        ADMINCTL_CONFIRM=reset-production-admin
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
</script>
