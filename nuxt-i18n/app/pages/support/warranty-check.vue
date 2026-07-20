<template>
  <div class="warranty-check">
    <WarrantyCheckPanel
      :is-logged-in="isLoggedIn"
      @login-request="handleLogin"
    />

    <!-- FAQ Section -->
    <section class="warranty-check__faq">
      <PageFaq 
        page-id="support-warranty-check"
        theme="dark"
        :show-categories="true"
      />
    </section>

    <!-- 登录弹窗：复用全局 AuthModal，嵌入模式 -->
    <AuthModal
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleAuthSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'
import PageFaq from '~/components/PageFaq.vue'
import WarrantyCheckPanel from '~/components/WarrantyCheckPanel.vue'
import AuthModal from '~/components/AuthModal.vue'

definePageMeta({
  layout: 'support',
})

const { t, locale } = useI18n()

useHead({
  title: t('warranty.title'),
})

// 登录状态：来源于全局 auth
const auth = useAuth()
const isLoggedIn = computed(() => !!auth.user.value)

// 登录弹窗状态
const showAuthModal = ref(false)
const authMode = ref<'login' | 'register'>('login')

// 处理登录：打开 AuthModal
const handleLogin = () => {
  authMode.value = 'login'
  showAuthModal.value = true
}

const handleAuthSuccess = () => {
  showAuthModal.value = false
}
</script>

<style>
.warranty-check {
  min-height: 60vh;
  /* Removed padding to align with global support layout (padding handled by layout) */
}

.warranty-check__faq {
  width: 100%;
  max-width: none;
  margin: 3rem auto 0;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
