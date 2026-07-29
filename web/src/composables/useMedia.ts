import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { MediaItem, PaginatedResponse } from '../types'
import { getMediaItems, type MediaQueryParams } from '../api'

export const useMediaStore = defineStore('media', () => {
  const items = ref<MediaItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const perPage = ref(50)
  const totalPages = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const selectedIds = ref<Set<string>>(new Set())

  const filters = ref<MediaQueryParams>({
    type: undefined,
    search: undefined,
    sort: 'total_size',
    order: 'desc',
    compressed: undefined,
    locked: undefined,
  })

  const selectedItems = computed(() =>
    items.value.filter((i) => selectedIds.value.has(i.id))
  )

  async function fetchItems(instanceId: string) {
    loading.value = true
    error.value = null
    try {
      const params: MediaQueryParams = {
        ...filters.value,
        page: page.value,
        per_page: perPage.value,
      }
      // Remove undefined keys
      Object.keys(params).forEach((key) => {
        if (params[key as keyof MediaQueryParams] === undefined) {
          delete params[key as keyof MediaQueryParams]
        }
      })
      const result: PaginatedResponse<MediaItem> = await getMediaItems(instanceId, params)
      items.value = result.items
      total.value = result.total
      page.value = result.page
      perPage.value = result.per_page
      totalPages.value = result.total_pages
      selectedIds.value = new Set()
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to fetch media items'
    } finally {
      loading.value = false
    }
  }

  function setPage(p: number) {
    page.value = p
  }

  function setFilter(key: keyof MediaQueryParams, value: any) {
    filters.value[key] = value
    page.value = 1
  }

  function resetFilters() {
    filters.value = {
      type: undefined,
      search: undefined,
      sort: 'total_size',
      order: 'desc',
      compressed: undefined,
      locked: undefined,
    }
    page.value = 1
  }

  function toggleSelect(id: string) {
    if (selectedIds.value.has(id)) {
      selectedIds.value.delete(id)
    } else {
      selectedIds.value.add(id)
    }
  }

  function selectAll() {
    items.value.forEach((i) => selectedIds.value.add(i.id))
  }

  function deselectAll() {
    selectedIds.value = new Set()
  }

  return {
    items,
    total,
    page,
    perPage,
    totalPages,
    loading,
    error,
    selectedIds,
    filters,
    selectedItems,
    fetchItems,
    setPage,
    setFilter,
    resetFilters,
    toggleSelect,
    selectAll,
    deselectAll,
  }
})
