import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { customsClassificationApi } from '@/api/customsClassifications'
import type { LookupCandidate } from '@/modules/customs/customsClassificationTypes'

export const useCustomsLookup = () => {
  const lookupProvider = ref('us_hts')
  const lookupQuery = ref('')
  const lookupLoading = ref(false)
  const lookupCompleted = ref(false)
  const lookupCandidates = ref<LookupCandidate[]>([])

  const runLookup = async () => {
    if (!lookupQuery.value.trim()) return
    lookupLoading.value = true
    lookupCompleted.value = false
    try {
      lookupCandidates.value = await customsClassificationApi.lookup({
        provider: lookupProvider.value,
        q: lookupQuery.value.trim(),
        limit: 10,
      })
      lookupCompleted.value = true
    } catch (error) {
      console.error('Failed to lookup customs classification:', error)
      lookupCandidates.value = []
      lookupCompleted.value = true
      toast.error('编码查询失败，请检查接口设置或稍后重试')
    } finally {
      lookupLoading.value = false
    }
  }

  return {
    lookupProvider,
    lookupQuery,
    lookupLoading,
    lookupCompleted,
    lookupCandidates,
    runLookup,
  }
}

export default useCustomsLookup

