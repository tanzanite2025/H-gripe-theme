<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto">
      <form @submit="emit('submit', $event)">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '添加用户' : '编辑用户' }}</DialogTitle>
          <DialogDescription>
            {{ mode === 'create' ? '创建新的后台用户并分配角色。' : '更新账号资料、角色和状态。' }}
          </DialogDescription>
        </DialogHeader>

        <div class="grid grid-cols-1 gap-4 py-5 sm:grid-cols-2">
          <FormField v-slot="{ componentField }" name="email">
            <FormItem>
              <FormLabel>邮箱</FormLabel>
              <FormControl><Input v-bind="componentField" type="email" autocomplete="email" /></FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="username">
            <FormItem>
              <FormLabel>用户名</FormLabel>
              <FormControl><Input v-bind="componentField" autocomplete="username" /></FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="password">
            <FormItem class="sm:col-span-2">
              <FormLabel>密码</FormLabel>
              <FormControl>
                <div class="relative">
                  <Input
                    v-bind="componentField"
                    :type="showPassword ? 'text' : 'password'"
                    :placeholder="mode === 'create' ? '至少 6 位' : '留空则不修改'"
                    autocomplete="new-password"
                    class="pr-9"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    class="absolute right-1 top-1/2 -translate-y-1/2"
                    :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                    @click="emit('update:showPassword', !showPassword)"
                  >
                    <EyeOff v-if="showPassword" class="size-4" />
                    <Eye v-else class="size-4" />
                  </Button>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="first_name">
            <FormItem>
              <FormLabel>名字</FormLabel>
              <FormControl><Input v-bind="componentField" /></FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="last_name">
            <FormItem>
              <FormLabel>姓氏</FormLabel>
              <FormControl><Input v-bind="componentField" /></FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="role">
            <FormItem>
              <FormLabel>角色</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger class="w-full"><SelectValue placeholder="请选择角色" /></SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="admin">超级管理员</SelectItem>
                  <SelectItem value="manager">经理</SelectItem>
                  <SelectItem value="editor">编辑</SelectItem>
                  <SelectItem value="support">客服</SelectItem>
                  <SelectItem value="viewer">查看者</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="locale">
            <FormItem>
              <FormLabel>语言</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger class="w-full"><SelectValue placeholder="请选择语言" /></SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem v-for="language in languageOptions" :key="language.value" :value="language.value">
                    {{ language.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="status">
            <FormItem class="sm:col-span-2">
              <FormLabel>状态</FormLabel>
              <FormControl>
                <RadioGroup v-bind="componentField" class="grid grid-cols-1 gap-2 sm:grid-cols-3">
                  <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                    <RadioGroupItem value="active" />活跃
                  </label>
                  <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                    <RadioGroupItem value="inactive" />未激活
                  </label>
                  <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                    <RadioGroupItem value="suspended" />已停用
                  </label>
                </RadioGroup>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '正在保存' : '保存' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Eye, EyeOff, LoaderCircle } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { UserDialogMode } from '@/modules/user/userTypes'

withDefaults(defineProps<{
  open?: boolean
  mode?: UserDialogMode
  submitting?: boolean
  showPassword?: boolean
  languageOptions?: LanguageOption[]
}>(), {
  open: false,
  mode: 'create',
  submitting: false,
  showPassword: false,
  languageOptions: () => []
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:showPassword', value: boolean): void
  (event: 'submit', value: Event): void
}>()
</script>

