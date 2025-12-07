<template>
  <div class="min-h-screen bg-slate-950 text-slate-50 pt-24 pb-16 md:pt-28">
    <div class="max-w-4xl mx-auto px-4">
      <!-- H1 隐藏，SEO 保留 -->
      <h1 class="sr-only">Privacy Policy</h1>
      <p class="text-sm text-slate-400 mb-6">
        This page describes how Tanzanite collects, uses, and protects your personal data
        when you browse our site or use our services.
      </p>

      <div class="mt-4">
        <PrivacyStatementContent />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PrivacyStatementContent from '~/components/PrivacyStatementContent.vue'

useHead({
  title: 'Privacy Policy',
})
</script>
