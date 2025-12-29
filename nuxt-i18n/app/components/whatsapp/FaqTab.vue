<template>
  <div class="h-full overflow-y-auto">
    <HomeFaqPreview :max-items-per-category="5" />
  </div>
</template>

<script setup lang="ts">
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
</script>
