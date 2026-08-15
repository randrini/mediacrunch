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
  const lastSelectedId = ref<string | null>(null)
  const selectAllMode = ref(false)

  let fetchId = 0

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

  async function fetchItems(instanceId: string, preserveSelection = false) {
    const currentFetchId = ++fetchId
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
      if (currentFetchId !== fetchId) return // stale response, discard
      items.value = result.items
      total.value = result.total
      page.value = result.page
      perPage.value = result.per_page
      totalPages.value = result.total_pages
      if (!preserveSelection) {
        selectedIds.value = new Set()
        selectAllMode.value = false
      }
    } catch (e: any) {
      if (currentFetchId !== fetchId) return // stale error, discard
      error.value = e?.response?.data?.error || e?.message || 'Failed to fetch media items'
    } finally {
      loading.value = false
    }
  }

  function setPage(p: number) {
    page.value = p
  }

  function setFilter<K extends keyof MediaQueryParams>(key: K, value: MediaQueryParams[K]) {
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
      lastSelectedId.value = null
    } else {
      selectedIds.value.add(id)
      lastSelectedId.value = id
    }
  }

  function selectRange(fromId: string, toId: string) {
    const ids = items.value.map((i) => i.id)
    const fromIdx = ids.indexOf(fromId)
    const toIdx = ids.indexOf(toId)
    if (fromIdx === -1 || toIdx === -1) return
    const start = Math.min(fromIdx, toIdx)
    const end = Math.max(fromIdx, toIdx)
    for (let i = start; i <= end; i++) {
      selectedIds.value.add(ids[i])
    }
  }

  function selectAll() {
    items.value.forEach((i) => selectedIds.value.add(i.id))
  }

  function selectAllAcrossPages() {
    selectAllMode.value = true
    selectedIds.value = new Set()
  }

  function clearSelection() {
    selectAllMode.value = false
    selectedIds.value = new Set()
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
    lastSelectedId,
    selectAllMode,
    filters,
    selectedItems,
    fetchItems,
    setPage,
    setFilter,
    resetFilters,
    toggleSelect,
    selectRange,
    selectAll,
    selectAllAcrossPages,
    clearSelection,
    deselectAll,
  }
})
