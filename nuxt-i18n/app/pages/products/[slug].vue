<script setup lang="ts">
const route = useRoute()
const slug = route.params.slug

await navigateTo({ path: `/shop/${slug}` }, { redirectCode: 301 })
</script>
