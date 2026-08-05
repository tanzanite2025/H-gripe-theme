<template>
  <div class="admin-login-shell relative min-h-screen min-h-dvh overflow-hidden">
    <div aria-hidden="true" class="admin-login-rail absolute inset-y-0 left-0 hidden w-[34vw] lg:block" />
    <div aria-hidden="true" class="admin-login-blueprint absolute right-[7vw] top-[12vh] hidden size-[22rem] lg:block" />

    <main class="relative z-10 flex min-h-screen min-h-dvh items-center justify-center px-4 py-8 sm:px-6">
      <Card class="w-full max-w-[430px] rounded-xl border-border/70 bg-card/95 py-5 shadow-2xl backdrop-blur-sm">
        <CardHeader class="space-y-3 px-5 text-center sm:px-7">
          <div v-if="brandName || brandInitial || panelLabel" class="flex items-center justify-between gap-3 text-left">
            <div v-if="brandName || brandInitial" class="flex min-w-0 items-center gap-2.5">
              <span v-if="brandInitial" class="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-black text-primary-foreground shadow-xs">
                {{ brandInitial }}
              </span>
              <strong v-if="brandName" class="truncate text-sm font-black italic uppercase">{{ brandName }}</strong>
            </div>
            <span v-if="panelLabel" class="ml-auto shrink-0 text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ panelLabel }}</span>
          </div>
          <span class="block text-[9px] font-black uppercase tracking-[0.16em] text-muted-foreground/60">CONTROL ACCESS / 账户认证</span>
          <CardTitle v-if="loginTitle" class="text-lg font-black italic uppercase text-foreground">{{ loginTitle }}</CardTitle>
          <CardDescription class="text-[10px] font-medium leading-5 text-muted-foreground">请输入管理员账号和密码进入控制面板</CardDescription>
        </CardHeader>

        <CardContent class="px-5 sm:px-7">
          <form class="space-y-5" @submit="onSubmit">
            <FormField v-slot="{ componentField }" name="email">
              <FormItem>
                <FormLabel>邮箱</FormLabel>
                <FormControl>
                  <div class="relative">
                    <Mail class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/60" />
                    <Input
                      v-bind="componentField"
                      type="email"
                      autocomplete="username"
                      placeholder="name@company.com"
                      class="h-10 pl-9"
                    />
                  </div>
                </FormControl>
                <div class="min-h-5 text-left" aria-live="polite">
                  <FormMessage class="text-[11px] leading-5" />
                </div>
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="password">
              <FormItem>
                <FormLabel>密码</FormLabel>
                <FormControl>
                  <div class="relative">
                    <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/60" />
                    <Input
                      v-bind="componentField"
                      :type="showPassword ? 'text' : 'password'"
                      autocomplete="current-password"
                      placeholder="请输入密码"
                      class="h-10 px-9"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      class="absolute right-1 top-1/2 -translate-y-1/2 rounded-full"
                      :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                      @click="showPassword = !showPassword"
                    >
                      <EyeOff v-if="showPassword" class="size-4" />
                      <Eye v-else class="size-4" />
                    </Button>
                  </div>
                </FormControl>
                <div class="min-h-5 text-left" aria-live="polite">
                  <FormMessage class="text-[11px] leading-5" />
                </div>
              </FormItem>
            </FormField>

            <Button type="submit" size="lg" class="h-11 w-full rounded-full font-black text-xs uppercase tracking-widest" :disabled="loading">
              <LoaderCircle v-if="loading" class="size-4 animate-spin" />
              <LogIn v-else class="size-4" />
              {{ loading ? '正在登录' : '登录' }}
            </Button>
          </form>

          <div v-if="googleAvailable" class="mt-5 space-y-3">
            <div class="flex items-center gap-3 text-[10px] font-bold uppercase tracking-[0.16em] text-muted-foreground/60">
              <span class="h-px flex-1 bg-border/70" />
              <span>或</span>
              <span class="h-px flex-1 bg-border/70" />
            </div>
            <Button
              type="button"
              variant="outline"
              class="h-11 w-full rounded-lg border-border/80 bg-background/75 font-bold text-xs hover:bg-muted"
              :disabled="loading || googleLoading"
              @click="onGoogleSubmit"
            >
              <LoaderCircle v-if="googleLoading" class="size-4 animate-spin" />
              <svg v-else viewBox="0 0 48 48" class="size-4" aria-hidden="true">
                <path fill="#FFC107" d="M43.611 20.083H42V20H24v8h11.303C33.565 32.664 29.177 36 24 36c-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-9 20-20c0-1.341-.138-2.651-.389-3.917z"/>
                <path fill="#FF3D00" d="M6.306 14.691l6.571 4.819C14.655 15.108 19 12 24 12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 15.322 4 8.135 9.069 6.306 14.691z"/>
                <path fill="#4CAF50" d="M24 44c5.114 0 9.725-1.961 13.261-5.174l-6.132-5.198C29.16 34.488 26.715 35.5 24 35.5c-5.139 0-9.479-3.335-11.029-8.014l-6.57 5.055C8.122 38.897 15.348 44 24 44z"/>
                <path fill="#1976D2" d="M43.611 20.083H42V20H24v8h11.303c-.685 2.316-2.172 4.285-4.134 5.628l.003-.001 6.132 5.198C39.846 35.896 44 30.5 44 24c0-1.341-.138-2.651-.389-3.917z"/>
              </svg>
              {{ googleLoading ? '正在连接' : '使用 Google 登录' }}
            </Button>
            <p v-if="googleError" class="text-center text-xs text-destructive">{{ googleError }}</p>
          </div>
        </CardContent>
      </Card>
    </main>

    <footer v-if="footerText" class="absolute inset-x-0 bottom-5 z-10 px-4 text-center text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">{{ footerText }}</footer>
  </div>
</template>

<script setup>
import { onMounted, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toTypedSchema } from '@vee-validate/zod'
import { useForm } from 'vee-validate'
import { z } from 'zod'
import { toast } from 'vue-sonner'
import { Eye, EyeOff, LoaderCircle, LockKeyhole, LogIn, Mail } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useAdminBranding } from '@/composables/useAdminBranding'
import { useAdminGoogleAuth } from '@/composables/useAdminGoogleAuth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const showPassword = ref(false)
const googleAvailable = ref(false)
const googleLoading = ref(false)
const googleError = ref('')
const googleAuth = useAdminGoogleAuth()
const {
  brandName,
  brandInitial,
  panelLabel,
  loginTitle,
  footerText,
  loadAdminBranding,
  setAdminDocumentTitle
} = useAdminBranding()

onMounted(async () => {
  await loadAdminBranding()
  try {
    googleAvailable.value = Boolean(await googleAuth.loadConfig())
  } catch {
    googleAvailable.value = false
  }
})

watchEffect(() => {
  setAdminDocumentTitle(loginTitle.value)
})

const formSchema = toTypedSchema(
  z.object({
    email: z.string().min(1, '请输入邮箱').email('请输入正确的邮箱格式'),
    password: z.string().min(6, '密码长度至少 6 位')
  })
)

const { handleSubmit } = useForm({
  validationSchema: formSchema,
  initialValues: {
    email: '',
    password: ''
  }
})

const onSubmit = handleSubmit(async (values) => {
  if (loading.value) return

  loading.value = true
  try {
    await authStore.login(values.email, values.password)
    toast.success('登录成功')
    await router.push(route.query.redirect || '/')
  } catch (error) {
    toast.error(error.response?.data?.error || '登录失败', { id: 'admin-login-error' })
  } finally {
    loading.value = false
  }
})

const onGoogleCredential = async (response) => {
  if (!response?.credential) {
    googleError.value = 'Google 登录未返回有效凭据'
    googleLoading.value = false
    return
  }

  try {
    await authStore.loginWithGoogle(response.credential)
    toast.success('登录成功')
    await router.push(route.query.redirect || '/')
  } catch (error) {
    googleError.value = error.response?.data?.message || error.response?.data?.error || 'Google 登录失败'
  } finally {
    googleLoading.value = false
  }
}

const onGoogleSubmit = async () => {
  if (googleLoading.value || loading.value) return

  googleLoading.value = true
  googleError.value = ''
  const initialized = await googleAuth.initialize(onGoogleCredential)
  if (!initialized) {
    googleError.value = googleAuth.error.value || 'Google 登录初始化失败'
    googleLoading.value = false
    return
  }

  googleAuth.prompt()
  window.setTimeout(() => {
    if (googleLoading.value) googleLoading.value = false
  }, 10000)
}
</script>

<style scoped>
.admin-login-shell {
  background:
    linear-gradient(135deg, color-mix(in oklch, var(--primary) 8%, var(--background)) 0%, var(--background) 46%, color-mix(in oklch, var(--accent) 42%, var(--background)) 100%);
}

.admin-login-shell::before {
  position: absolute;
  inset: 0;
  content: '';
  pointer-events: none;
  background-image:
    linear-gradient(to right, color-mix(in oklch, var(--border) 72%, transparent) 1px, transparent 1px),
    linear-gradient(to bottom, color-mix(in oklch, var(--border) 72%, transparent) 1px, transparent 1px);
  background-position: center;
  background-size: 34px 34px;
  mask-image: linear-gradient(to bottom, black 0%, rgb(0 0 0 / 0.72) 42%, transparent 86%);
  opacity: 0.62;
}

.admin-login-shell::after {
  position: absolute;
  inset: 0;
  content: '';
  pointer-events: none;
  background:
    linear-gradient(112deg, transparent 0 52%, color-mix(in oklch, var(--primary) 12%, transparent) 52.1% 52.3%, transparent 52.4%),
    linear-gradient(112deg, transparent 0 64%, color-mix(in oklch, var(--chart-2) 10%, transparent) 64.1% 64.25%, transparent 64.35%);
  opacity: 0.75;
}

.admin-login-rail {
  border-right: 1px solid color-mix(in oklch, var(--primary) 15%, var(--border));
  background:
    repeating-linear-gradient(135deg, transparent 0 22px, color-mix(in oklch, var(--primary) 7%, transparent) 22px 23px, transparent 23px 48px),
    linear-gradient(180deg, color-mix(in oklch, var(--primary) 8%, transparent), transparent 72%);
  clip-path: polygon(0 0, 76% 0, 52% 100%, 0 100%);
  opacity: 0.8;
}

.admin-login-blueprint {
  border: 1px solid color-mix(in oklch, var(--primary) 20%, var(--border));
  background:
    linear-gradient(to right, transparent 49.7%, color-mix(in oklch, var(--primary) 24%, transparent) 50%, transparent 50.3%),
    linear-gradient(to bottom, transparent 49.7%, color-mix(in oklch, var(--chart-2) 22%, transparent) 50%, transparent 50.3%),
    repeating-linear-gradient(0deg, transparent 0 23px, color-mix(in oklch, var(--border) 70%, transparent) 23px 24px),
    repeating-linear-gradient(90deg, transparent 0 23px, color-mix(in oklch, var(--border) 70%, transparent) 23px 24px);
  opacity: 0.55;
  transform: rotate(8deg);
}
</style>
