import { reactive, ref } from 'vue'
import shippingApi from '@/api/shipping'
import {
  shippingServiceCollectionReferenceApi,
  type FpxPublishedCollectionReference,
  type YanwenPublishedCollectionReference,
} from '@/api/shippingServiceCollectionReferenceApi'

export const useShippingResources = () => {
  const templates = ref<any[]>([])
  const carriers = ref<any[]>([])
  const carrierServices = ref<any[]>([])
  const fpxChannels = ref<FpxPublishedCollectionReference[]>([])
  const yanwenPublishedChannels = ref<YanwenPublishedCollectionReference[]>([])
  const trackingProviders = ref<any[]>([])
  const trackingCarrierMappings = ref<any[]>([])
  const trackingShipmentsCount = ref(0)
  const packagingRules = ref<any[]>([])
  const refreshing = ref(false)
  const loading = reactive({
    templates: false,
    carriers: false,
    services: false,
    tracking: false,
    trackingMappings: false,
    trackingShipments: false,
    packaging: false,
  })

  const handleTrackingShipmentsCountChange = (count: number) => {
    trackingShipmentsCount.value = Number(count || 0)
  }

  const fetchTemplates = async () => {
    loading.templates = true
    try {
      templates.value = await shippingApi.listTemplates()
    } catch (error) {
      console.error('Failed to fetch shipping templates:', error)
    } finally {
      loading.templates = false
    }
  }

  const fetchCarriers = async () => {
    loading.carriers = true
    try {
      carriers.value = await shippingApi.listCarriers()
    } catch (error) {
      console.error('Failed to fetch carriers:', error)
    } finally {
      loading.carriers = false
    }
  }

  const fetchCarrierServices = async () => {
    loading.services = true
    try {
      carrierServices.value = await shippingApi.listCarrierServices()
    } catch (error) {
      console.error('Failed to fetch carrier services:', error)
    } finally {
      loading.services = false
    }
  }

  const fetchFpxPublishedChannels = async () => {
    try {
      fpxChannels.value = await shippingServiceCollectionReferenceApi.listFpxPublishedCollection()
    } catch (error) {
      fpxChannels.value = []
      console.error('Failed to fetch enabled 4PX service references:', error)
    }
  }

  const fetchYanwenPublishedChannels = async () => {
    try {
      yanwenPublishedChannels.value = await shippingServiceCollectionReferenceApi.listYanwenPublishedCollection()
    } catch (error) {
      yanwenPublishedChannels.value = []
      console.error('Failed to fetch published Yanwen service references:', error)
    }
  }

  const fetchTrackingProviders = async () => {
    loading.tracking = true
    try {
      trackingProviders.value = await shippingApi.listTrackingProviders()
    } catch (error) {
      console.error('Failed to fetch tracking providers:', error)
    } finally {
      loading.tracking = false
    }
  }

  const fetchTrackingCarrierMappings = async () => {
    loading.trackingMappings = true
    try {
      trackingCarrierMappings.value = await shippingApi.listTrackingCarrierMappings()
    } catch (error) {
      console.error('Failed to fetch tracking carrier mappings:', error)
    } finally {
      loading.trackingMappings = false
    }
  }

  const fetchPackagingRules = async () => {
    loading.packaging = true
    try {
      packagingRules.value = await shippingApi.listPackagingRules()
    } catch (error) {
      console.error('Failed to fetch packaging rules:', error)
    } finally {
      loading.packaging = false
    }
  }

  const refreshCurrentTab = async (activeTab: string, trackingShipmentsPanelRef: any) => {
    refreshing.value = true
    try {
      if (activeTab === 'templates') {
        await fetchTemplates()
      } else if (activeTab === 'carriers') {
        await fetchCarriers()
      } else if (activeTab === 'services') {
        await Promise.all([fetchCarrierServices(), fetchCarriers(), fetchTemplates(), fetchFpxPublishedChannels(), fetchYanwenPublishedChannels()])
      } else if (activeTab === 'tracking') {
        await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings(), fetchCarriers(), fetchCarrierServices()])
      } else if (activeTab === 'trackingShipments') {
        await Promise.all([
          trackingShipmentsPanelRef.value?.refresh?.(),
          fetchTrackingProviders(),
          fetchCarriers(),
          fetchCarrierServices(),
        ])
      } else if (activeTab === 'packaging') {
        await fetchPackagingRules()
      } else {
        await Promise.all([
          fetchTemplates(),
          fetchCarriers(),
          fetchCarrierServices(),
          fetchFpxPublishedChannels(),
          fetchYanwenPublishedChannels(),
          fetchTrackingProviders(),
          fetchTrackingCarrierMappings(),
          fetchPackagingRules(),
        ])
      }
    } finally {
      refreshing.value = false
    }
  }

  const fetchAllShippingResources = () => Promise.all([
    fetchTemplates(),
    fetchCarriers(),
    fetchCarrierServices(),
    fetchFpxPublishedChannels(),
    fetchYanwenPublishedChannels(),
    fetchTrackingProviders(),
    fetchTrackingCarrierMappings(),
    fetchPackagingRules(),
  ])

  return {
    templates,
    carriers,
    carrierServices,
    fpxChannels,
    yanwenPublishedChannels,
    trackingProviders,
    trackingCarrierMappings,
    trackingShipmentsCount,
    packagingRules,
    refreshing,
    loading,
    handleTrackingShipmentsCountChange,
    fetchTemplates,
    fetchCarriers,
    fetchCarrierServices,
    fetchFpxPublishedChannels,
    fetchYanwenPublishedChannels,
    fetchTrackingProviders,
    fetchTrackingCarrierMappings,
    fetchPackagingRules,
    refreshCurrentTab,
    fetchAllShippingResources,
  }
}

export default useShippingResources
