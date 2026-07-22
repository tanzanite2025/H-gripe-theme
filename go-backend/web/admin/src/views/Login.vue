<template>
  <div class="grid min-h-screen min-h-dvh grid-rows-[auto_1fr_auto] bg-muted/45">
    <header class="flex h-16 items-center justify-between border-b border-dashed border-border/70 bg-card px-4 sm:px-8">
      <div class="flex items-center gap-2.5">
        <span class="flex size-8 items-center justify-center rounded-full bg-primary text-sm font-black text-primary-foreground shadow-xs">
          T
        </span>
        <strong class="text-sm font-black italic tracking-tighter uppercase">Tanzanite</strong>
      </div>
      <span class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">CONTROL PANEL</span>
    </header>

    <main class="flex items-start justify-center px-3 py-9 sm:items-center sm:px-5 sm:py-12">
      <Card class="w-full max-w-[420px] rounded-[32px] border-dashed border-border/80 shadow-xl">
        <CardHeader class="space-y-1 text-center">
          <span class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 block">AUTHENTICATION / 账户认证</span>
          <CardTitle class="text-lg font-black tracking-tighter italic uppercase text-foreground">登录 Tanzanite</CardTitle>
          <CardDescription class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">请输入管理员账号和密码进入控制面板</CardDescription>
        </CardHeader>

        <CardContent>
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
                <FormMessage />
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
                <FormMessage />
              </FormItem>
            </FormField>

            <Button type="submit" size="lg" class="h-11 w-full rounded-full font-black text-xs uppercase tracking-widest" :disabled="loading">
              <LoaderCircle v-if="loading" class="size-4 animate-spin" />
              <LogIn v-else class="size-4" />
              {{ loading ? '正在登录' : '登录' }}
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>

    <footer class="pb-5 text-center text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">Tanzanite Operations System</footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
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
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const showPassword = ref(false)

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
</script>
