import { computed, ref, type Ref } from 'vue'
import {
  groupGlobalAllFaqTopics,
  type GlobalAllFaqFlatItem,
  type GlobalAllFaqSearchTopic,
} from '~/data/faq'

export function useFaqTopics(
  allItems: Readonly<Ref<GlobalAllFaqFlatItem[]>>,
) {
  const activeTopicId = ref('')
  const topicGroups = computed<GlobalAllFaqSearchTopic[]>(() => (
    groupGlobalAllFaqTopics(allItems.value)
  ))
  const featuredTopics = computed(() => topicGroups.value.slice(0, 4))
  const featuredItems = computed(() => {
    const selected = featuredTopics.value
      .map(topic => topic.items[0])
      .filter((item): item is GlobalAllFaqFlatItem => Boolean(item))

    if (selected.length >= 4) return selected.slice(0, 4)

    const selectedIds = new Set(selected.map(item => item.id))
    for (const item of allItems.value) {
      if (selectedIds.has(item.id)) continue
      selected.push(item)
      selectedIds.add(item.id)
      if (selected.length >= 4) break
    }

    return selected
  })
  const activeTopic = computed(() => (
    featuredTopics.value.find(topic => topic.id === activeTopicId.value) || null
  ))
  const topicItems = computed(() => (
    activeTopic.value?.items.slice(0, 6) || featuredItems.value
  ))

  const selectTopic = (topicId: string) => {
    activeTopicId.value = activeTopicId.value === topicId ? '' : topicId
  }

  const resetTopic = () => {
    activeTopicId.value = ''
  }

  return {
    activeTopicId,
    topicGroups,
    featuredTopics,
    featuredItems,
    activeTopic,
    topicItems,
    selectTopic,
    resetTopic,
  }
}
