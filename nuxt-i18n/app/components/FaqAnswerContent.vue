<template>
  <div
    class="faq-answer-content"
    :class="{ 'faq-answer-content--with-image': canShowImage }"
  >
    <figure v-if="canShowImage" class="faq-answer-content__media">
      <img
        :src="imageUrl"
        :alt="imageAlt || 'FAQ illustration'"
        width="800"
        height="800"
        loading="lazy"
        decoding="async"
      >
    </figure>
    <div class="tz-rich-text faq-answer-content__body" v-html="answer" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  answer: string
  imageUrl?: string
  imageAlt?: string
  imageWidth?: number
  imageHeight?: number
}>()

const canShowImage = computed(() => {
  if (!props.imageUrl) return false
  return Number(props.imageWidth || 800) === 800 && Number(props.imageHeight || 800) === 800
})
</script>

<style scoped>
.faq-answer-content {
  display: block;
  min-width: 0;
}

.faq-answer-content--with-image {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: clamp(1rem, 2vw, 2rem);
  align-items: start;
}

.faq-answer-content__media {
  width: 100%;
  max-width: 800px;
  margin: 0;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border-radius: 1rem;
  background: rgba(15, 23, 42, 0.12);
}

.faq-answer-content__media img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.faq-answer-content__body {
  min-width: 0;
}

@media (max-width: 767px) {
  .faq-answer-content--with-image {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
