<template>
  <div class="h-full overflow-y-auto px-1 pt-1 pb-3 md:p-6">
    <div class="w-full md:max-w-md md:mx-auto rounded-2xl p-3 md:p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md space-y-3 md:space-y-4">
      <!-- 顶部：当前等级 / 提示 -->
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 md:w-10 md:h-10 rounded-full bg-white/10 flex items-center justify-center text-[11px] md:text-xs font-semibold text-white/80 border border-white/20">
          {{ isMemberLogged ? (levelName || '—') : 'Guest' }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-[11px] text-white/50 truncate">
            {{ isMemberLogged ? 'Your membership' : 'Membership program' }}
          </div>
          <div class="text-sm font-semibold text-white truncate">
            <span v-if="isMemberLogged">Level {{ levelName }}</span>
            <span v-else>Log in to unlock member prices</span>
          </div>
        </div>
      </div>

      <!-- 核心指标网格 -->
      <div class="grid grid-cols-2 gap-2 md:gap-3 text-[11px]">
        <div class="rounded-xl px-2.5 md:px-3 py-2 bg-white/5">
          <div class="text-white/50">Points</div>
          <div class="text-sm font-semibold text-white">
            {{ isMemberLogged ? points : '—' }}
          </div>
        </div>
        <div class="rounded-xl px-2.5 md:px-3 py-2 bg-white/5">
          <div class="text-white/50">Product discount</div>
          <div class="text-sm font-semibold text-white">
            {{ isMemberLogged ? (levelDiscounts.product + '%') : '—' }}
          </div>
        </div>
        <div class="rounded-xl px-2.5 md:px-3 py-2 bg-white/5">
          <div class="text-white/50">Points discount</div>
          <div class="text-sm font-semibold text-white">
            {{ isMemberLogged ? (levelDiscounts.points + '%') : '—' }}
          </div>
        </div>
        <div class="rounded-xl px-2.5 md:px-3 py-2 bg-white/5">
          <div class="text-white/50">Coupons / Cards</div>
          <div class="text-sm font-semibold text-white">
            {{ isMemberLogged ? `× ${userCoupons} / × ${userPointCards}` : '—' }}
          </div>
        </div>
      </div>

      <!-- 等级进度条 -->
      <div v-if="isMemberLogged" class="space-y-1.5">
        <div class="h-1.5 rounded-full bg-white/10 overflow-hidden">
          <div
            class="h-full bg-[linear-gradient(90deg,#40ffaa,#6b73ff)]"
            :style="{ width: tierInfo.pct + '%' }"
          ></div>
        </div>
        <div class="flex items-center justify-between text-[10px] md:text-[11px] text-white/60">
          <span>{{ tierInfo.current ? tierInfo.current.min : 0 }}</span>
          <span class="font-semibold text-white/80">{{ tierInfo.pct }}%</span>
          <span>
            {{
              tierInfo.next
                ? tierInfo.next.min
                : (tierInfo.current && tierInfo.current.max !== -1 ? tierInfo.current.max : 'MAX')
            }}
          </span>
        </div>
      </div>

      <div v-else class="text-[11px] text-white/60 space-y-2 md:space-y-3">
        <p>Log in or sign up to see your member prices, points and progress.</p>
        <div class="flex gap-1.5 md:gap-2">
          <button
            type="button"
            class="flex-1 h-8 md:h-9 rounded-full bg-[linear-gradient(135deg,#40ffaa,#6b73ff)] text-slate-950 text-[11px] font-semibold hover:brightness-110 transition-all shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95)]"
            @click="$emit('openAuth', 'register')"
          >
            Sign up
          </button>
          <button
            type="button"
            class="flex-1 h-8 md:h-9 rounded-full bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-[11px] font-semibold shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
            @click="$emit('openAuth', 'login')"
          >
            Log in
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  isMemberLogged: boolean
  levelName: string | number
  points: number | string
  tierInfo: any
  levelDiscounts: any
  userCoupons: number
  userPointCards: number
}>()

defineEmits<{
  'openAuth': [mode: 'login' | 'register']
}>()
</script>
