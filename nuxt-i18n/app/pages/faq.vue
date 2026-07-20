<script setup lang="ts">
import { useLocalePath } from '#imports'

const localePath = useLocalePath()

await navigateTo(localePath('/support/faqs'), { redirectCode: 301 })
</script>
