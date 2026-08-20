<template>
  <div
    class="faq-answer-content"
    :class="{ 'faq-answer-content--with-image': canShowImage }"
  >
    <figure v-if="canShowImage" class="faq-answer-content__media">
      <StorefrontImage
        :src="imageUrl || ''"
        :alt="imageAlt || t('faq.ui.illustrationAlt')"
        width="800"
        height="800"
        preset="content"
      />
    </figure>
    <SafeRichText
      class="tz-rich-text faq-answer-content__body"
      :html="answer"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const { t } = useI18n()

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
  background: rgba(255, 255, 255, 0.02);
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
