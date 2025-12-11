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
  padding: 2rem 1rem;
}

.warranty-check__container {
  max-width: 600px;
  margin: 0 auto;
}

.warranty-check__header {
  text-align: center;
  margin-bottom: 2rem;
}

.warranty-check__title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
  margin: 0 0 0.5rem;
}

.warranty-check__subtitle {
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.9);
  margin: 0;
}

/* 未登录状态 */
.warranty-check__login-required {
  text-align: center;
  padding: 3rem 2rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 1rem;
  border: none;
  box-shadow:
    0 6px 20px rgba(0, 0, 0, 0.45),
    0 14px 40px rgba(0, 0, 0, 0.35);
}

.warranty-check__lock-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.warranty-check__login-required p {
  color: rgba(148, 163, 184, 0.9);
  margin-bottom: 1.5rem;
}

.warranty-check__login-btn {
  padding: 0.75rem 2rem;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  border: none;
  border-radius: 9999px;
  color: #020617;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.warranty-check__login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(45, 212, 191, 0.4);
}

/* 查询表单 */
.warranty-check__form {
  margin-bottom: 2rem;
}

.warranty-check__label {
  display: block;
  font-size: 0.9rem;
  color: #e5e7eb;
  margin-bottom: 0.5rem;
}

.warranty-check__input-group {
  display: flex;
  gap: 0.5rem;
}

.warranty-check__input {
  flex: 1;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(15, 23, 42, 0.96));
  border: none;
  border-radius: 9999px;
  color: #e5e7eb;
  font-size: 1rem;
  box-shadow:
    0 2px 6px -3px rgba(0, 0, 0, 0.9),
    0 0 6px rgba(15, 23, 42, 0.7);
}

.warranty-check__input:focus {
  outline: none;
  box-shadow:
    0 0 0 1px rgba(45, 212, 191, 0.75),
    0 0 14px rgba(45, 212, 191, 0.35);
}

.warranty-check__input::placeholder {
  color: rgba(148, 163, 184, 0.6);
}

.warranty-check__submit-btn {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  border: none;
  border-radius: 9999px;
  color: #020617;
  font-weight: 600;
  cursor: pointer;
  min-width: 100px;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
  transition: opacity 0.2s, transform 0.18s ease, box-shadow 0.18s ease;
}

.warranty-check__submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.warranty-check__submit-btn:not(:disabled):hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(45, 212, 191, 0.4);
}

.warranty-check__spinner {
  display: inline-block;
  width: 1rem;
  height: 1rem;
  border: 2px solid #000;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.warranty-check__help {
  font-size: 0.8rem;
  color: rgba(148, 163, 184, 0.7);
  margin-top: 0.5rem;
}

/* 错误状态 */
.warranty-check__error {
  text-align: center;
  padding: 2rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 1rem;
}

.warranty-check__error-icon {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.warranty-check__error h3 {
  color: #fca5a5;
  margin: 0 0 0.5rem;
}

.warranty-check__error p {
  color: rgba(148, 163, 184, 0.9);
  margin: 0 0 1rem;
}

.warranty-check__tips {
  text-align: left;
  padding-left: 1.5rem;
  color: rgba(148, 163, 184, 0.8);
  font-size: 0.85rem;
  margin-bottom: 1rem;
}

.warranty-check__tips li {
  margin: 0.25rem 0;
}

.warranty-check__error-contact {
  font-size: 0.85rem;
  margin-bottom: 1rem;
}

.warranty-check__contact-btn {
  display: inline-block;
  padding: 0.5rem 1.5rem;
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
  border: none;
  border-radius: 9999px;
  color: #e5e7eb;
  text-decoration: none;
  font-size: 0.85rem;
  box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
  transition: all 0.18s ease;
}

.warranty-check__contact-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 1);
}

/* 查询结果 */
.warranty-check__result {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  overflow: hidden;
}

.warranty-check__status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  font-size: 1.1rem;
  font-weight: 600;
}

.warranty-check__status--valid {
  background: rgba(34, 197, 94, 0.15);
  color: #86efac;
}

.warranty-check__status--expired {
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

.warranty-check__status-icon {
  font-size: 1.25rem;
}

.warranty-check__info {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.warranty-check__info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.warranty-check__info-row:last-child {
  border-bottom: none;
}

.warranty-check__info-label {
  color: rgba(148, 163, 184, 0.9);
  font-size: 0.85rem;
}

.warranty-check__info-value {
  color: #e5e7eb;
  font-weight: 500;
}

.warranty-check__remaining {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  font-weight: 500;
}

.warranty-check__remaining--valid {
  background: rgba(34, 197, 94, 0.1);
  color: #86efac;
}

.warranty-check__remaining--expired {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.warranty-check__remaining-icon {
  font-size: 1.1rem;
}

/* 服务记录 */
.warranty-check__records {
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.warranty-check__records h4 {
  font-size: 0.9rem;
  color: #e5e7eb;
  margin: 0 0 0.75rem;
}

.warranty-check__records-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.warranty-check__record-item {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  font-size: 0.85rem;
}

.warranty-check__record-item:last-child {
  border-bottom: none;
}

.warranty-check__record-type {
  padding: 0.15rem 0.5rem;
  background: rgba(107, 115, 255, 0.2);
  color: #a5b4fc;
  border-radius: 9999px;
  font-size: 0.75rem;
}

.warranty-check__record-date {
  color: rgba(148, 163, 184, 0.9);
}

.warranty-check__record-desc {
  width: 100%;
  color: rgba(148, 163, 184, 0.8);
}

.warranty-check__no-records {
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.warranty-check__no-records p {
  color: rgba(148, 163, 184, 0.7);
  font-size: 0.85rem;
  margin: 0;
}

/* 操作按钮 */
.warranty-check__actions {
  display: flex;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.warranty-check__action-btn {
  flex: 1;
  padding: 0.6rem 1rem;
  border-radius: 9999px;
  font-size: 0.85rem;
  font-weight: 500;
  text-align: center;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.18s ease;
}

.warranty-check__action-btn:not(.warranty-check__action-btn--secondary) {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  border: none;
  color: #020617;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

.warranty-check__action-btn--secondary {
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
  border: none;
  color: #e5e7eb;
  box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
}

.warranty-check__action-btn--secondary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 1);
}

/* FAQ 区域 */
.warranty-check__faq {
  max-width: 800px;
  margin: 3rem auto 0;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* 响应式 */
@media (max-width: 480px) {
  .warranty-check__input-group {
    flex-direction: column;
  }

  .warranty-check__submit-btn {
    width: 100%;
  }

  .warranty-check__actions {
    flex-direction: column;
  }
}
</style>
