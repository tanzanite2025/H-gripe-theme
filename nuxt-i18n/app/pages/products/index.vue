<script setup lang="ts">
const localePath = useLocalePath()

await navigateTo(localePath('/shop'), {
  redirectCode: 301,
  replace: true,
})
</script>

<template>
  <div />
</template>
