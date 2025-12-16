<template>
  <div class="flex flex-col gap-5">
    <h1 class="sr-only">{{ t('company.nav.certificates') }}</h1>

    <section class="rounded-2xl border border-slate-700/40 bg-slate-950/60 p-6 shadow-lg">
      <h2 class="mb-3 text-lg font-semibold text-slate-100">
        {{ t('company.nav.certificates') }}
      </h2>
      <p class="text-slate-300">
        {{ t('company.pages.placeholder') }}
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.certificates'),
})
</script>
