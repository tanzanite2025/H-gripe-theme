<script lang="ts">
import { computed, defineComponent, h } from 'vue'
import { useRuntimeConfig } from '#imports'
import {
  renderRichText,
  safeRichTextMediaOriginsFromRuntimeConfig,
} from '~/utils/security/safeRichText'

export default defineComponent({
  name: 'SafeRichText',
  inheritAttrs: false,
  props: {
    as: {
      type: String,
      default: 'div',
    },
    html: {
      type: String,
      default: '',
    },
  },
  setup(props, { attrs }) {
    const runtimeConfig = useRuntimeConfig()
    const mediaOrigins = computed(() => safeRichTextMediaOriginsFromRuntimeConfig(
      runtimeConfig,
      import.meta.client ? window.location.origin : '',
    ))

    return () => h(props.as, attrs, renderRichText(props.html, {
      mediaOrigins: mediaOrigins.value,
    }))
  },
})
</script>
