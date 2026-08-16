import { ref } from 'vue'
import { productBrandApi, productInformationTemplateApi } from '@/api/products'
import shippingApi from '@/api/shipping'

export const useCustomsEditorResources = () => {
  const brands = ref<any[]>([])
  const shippingTemplates = ref<any[]>([])
  const informationTemplates = ref<any[]>([])

  const fetchEditorResources = async () => {
    try {
      const [brandItems, shippingItems, informationItems] = await Promise.all([
        productBrandApi.list({ include_disabled: true }),
        shippingApi.listTemplates(),
        productInformationTemplateApi.list({ include_disabled: true }),
      ])
      brands.value = brandItems
      shippingTemplates.value = shippingItems
      informationTemplates.value = informationItems
    } catch (error) {
      console.error('Failed to fetch product editor resources:', error)
    }
  }

  return {
    brands,
    shippingTemplates,
    informationTemplates,
    fetchEditorResources,
  }
}

export default useCustomsEditorResources
