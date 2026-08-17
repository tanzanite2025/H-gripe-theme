<template>
  <section class="bg-transparent py-8 text-white sm:py-12 lg:py-20" aria-hidden="true">
    <div class="page-content-shell px-0 md:px-6">
      <div class="grid items-start gap-5 sm:gap-10 lg:grid-cols-12 lg:gap-16">
        <div class="self-start space-y-3 sm:space-y-6 lg:sticky lg:top-32 lg:col-span-3">
          <div class="inline-flex h-6 w-28 animate-pulse rounded-full bg-white/10"></div>
          <div class="space-y-3">
            <div class="h-8 w-5/6 animate-pulse rounded bg-white/10"></div>
            <div class="h-4 w-full animate-pulse rounded bg-white/5"></div>
            <div class="h-4 w-4/5 animate-pulse rounded bg-white/5"></div>
          </div>
          <div class="h-12 w-36 animate-pulse rounded-full bg-white/10"></div>
        </div>

        <div class="lg:col-span-9">
          <div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
            <div
              v-for="index in 4"
              :key="index"
              class="aspect-square animate-pulse rounded-2xl border border-white/10 bg-white/5"
            ></div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
