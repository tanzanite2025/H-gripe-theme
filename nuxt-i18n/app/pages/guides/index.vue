<script setup lang="ts">
const localePath = useLocalePath()

await navigateTo(localePath('/guides/tireguides'), {
  redirectCode: 301,
  replace: true,
})
</script>

<template>
  <div />
</template>
