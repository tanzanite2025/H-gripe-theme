<template>
  <section class="bg-transparent py-8 tz-text-primary sm:py-12 lg:py-20" aria-hidden="true">
    <div class="page-content-shell px-0 md:px-6">
      <div class="grid items-start gap-5 sm:gap-10 lg:grid-cols-12 lg:gap-16">
        <div class="self-start space-y-3 sm:space-y-6 lg:sticky lg:top-32 lg:col-span-3">
          <div class="h-6 w-28 animate-pulse rounded-full tz-surface-subtle"></div>
          <div class="space-y-3">
            <div class="h-8 w-5/6 animate-pulse rounded tz-surface-subtle"></div>
            <div class="h-4 w-full animate-pulse rounded tz-surface-subtle"></div>
            <div class="h-4 w-4/5 animate-pulse rounded tz-surface-subtle"></div>
          </div>
          <div class="h-12 w-40 animate-pulse rounded-full tz-surface-subtle"></div>
        </div>

        <div class="min-w-0 lg:col-span-9">
          <div class="grid grid-cols-2 gap-2 sm:gap-4 xl:grid-cols-4">
            <article
              v-for="index in 4"
              :key="index"
              class="overflow-hidden rounded-lg border tz-border-subtle tz-surface-subtle"
            >
              <div class="aspect-square animate-pulse tz-surface-subtle"></div>
              <div class="space-y-3 p-3.5">
                <div class="h-4 w-5/6 animate-pulse rounded tz-surface-subtle"></div>
                <div class="h-3 w-3/5 animate-pulse rounded tz-surface-subtle"></div>
                <div class="h-4 w-2/5 animate-pulse rounded tz-surface-subtle"></div>
              </div>
            </article>
          </div>

          <section class="mt-6 pt-1 lg:mt-8">
            <div class="flex items-end justify-between gap-4 border-b border-[#059669]/20 pb-3">
              <div class="space-y-2">
                <div class="h-3 w-28 animate-pulse rounded bg-[#059669]/15"></div>
                <div class="h-5 w-60 animate-pulse rounded tz-surface-subtle"></div>
              </div>
              <div class="flex gap-2">
                <div class="h-8 w-8 animate-pulse rounded-full tz-surface-subtle"></div>
                <div class="h-8 w-8 animate-pulse rounded-full tz-surface-subtle"></div>
              </div>
            </div>

            <div class="mt-4 flex gap-3 overflow-hidden">
              <article
                v-for="index in 3"
                :key="index"
                class="aspect-square w-[min(82vw,20rem)] shrink-0 overflow-hidden rounded-lg border tz-border-subtle tz-surface-subtle sm:w-[min(29vw,19.5rem)]"
              >
                <div class="h-[56%] animate-pulse border-b tz-border-subtle tz-surface-subtle"></div>
                <div class="space-y-2 p-3.5">
                  <div class="h-3 w-16 animate-pulse rounded bg-[#059669]/15"></div>
                  <div class="h-4 w-4/5 animate-pulse rounded tz-surface-subtle"></div>
                  <div class="h-3 w-full animate-pulse rounded tz-surface-subtle"></div>
                  <div class="h-3 w-3/4 animate-pulse rounded tz-surface-subtle"></div>
                </div>
              </article>
            </div>

            <div class="mt-2 flex justify-center gap-2">
              <div v-for="index in 5" :key="index" class="h-12 w-12 animate-pulse rounded-full bg-white/[0.035]"></div>
            </div>
          </section>
        </div>
      </div>

      <section class="mt-8 pt-1 lg:mt-12">
        <div class="flex items-end justify-between gap-4 border-b border-[#059669]/20 pb-3">
          <div class="space-y-2">
            <div class="h-3 w-24 animate-pulse rounded bg-[#059669]/15"></div>
            <div class="h-5 w-40 animate-pulse rounded tz-surface-subtle"></div>
          </div>
          <div class="flex gap-2">
            <div class="h-8 w-8 animate-pulse rounded-full tz-surface-subtle"></div>
            <div class="h-8 w-8 animate-pulse rounded-full tz-surface-subtle"></div>
          </div>
        </div>

        <div class="mt-4 flex gap-3 overflow-hidden">
          <div
            v-for="index in 3"
            :key="index"
            class="aspect-[4/3] w-[min(82vw,23rem)] shrink-0 animate-pulse rounded-lg border tz-border-subtle tz-surface-subtle sm:w-[min(62vw,23rem)] lg:w-[min(29vw,25rem)]"
          ></div>
        </div>

        <div class="mt-2 flex justify-center gap-2">
          <div v-for="index in 5" :key="index" class="h-12 w-12 animate-pulse rounded-full bg-white/[0.035]"></div>
        </div>
      </section>
    </div>
  </section>
</template>
