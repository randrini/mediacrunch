<template>
  <div>
    <!-- Selection Bar -->
    <div
      v-if="someSelected"
      class="bg-accent/10 border border-accent/30 rounded-lg px-4 py-2 mb-4 flex items-center justify-between"
    >
      <span class="text-sm text-slate-200">
        <template v-if="selectAllMode">
          All <span class="font-semibold">{{ totalItems }}</span> items selected
        </template>
        <template v-else>
          <span class="font-semibold">{{ selectedCount }}</span> item{{ selectedCount !== 1 ? 's' : '' }} selected
        </template>
      </span>
      <div class="flex items-center space-x-3">
        <button
          v-if="!selectAllMode && selectedCount < totalItems"
          @click="$emit('selectAllAcrossPages')"
          class="text-sm text-accent hover:text-accent-hover transition-base"
        >
          Select all {{ totalItems }} items
        </button>
        <button
          v-else
          @click="$emit('clearSelection')"
          class="text-sm text-slate-400 hover:text-slate-200 transition-base"
        >
          Clear selection
        </button>
        <button
          @click="$emit('compress', selectedIds)"
          class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md bg-accent text-base hover:bg-accent-hover transition-base"
        >
          <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
          Compress Selected
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="card-glass overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-white/[0.06]">
          <thead class="bg-elevated">
            <tr>
              <th class="px-3 py-2 text-left">
                <input
                  type="checkbox"
                  :checked="allSelected"
                  :indeterminate="someSelected && !allSelected"
                  @change="onSelectAll"
                  class="rounded border-white/[0.08] bg-elevated text-accent focus:ring-accent/50"
                />
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Type</th>
              <th
                class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:text-slate-200"
                @click="sortBy('title')"
              >
                <span class="inline-flex items-center">
                  Title
                  <svg v-if="sortField === 'title'" class="w-3 h-3 ml-1 text-accent" :class="sortOrder === 'asc' ? '' : 'rotate-180'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Year</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Images</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Fanart</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Poster</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Clear Logo</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Season Poster</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Banner</th>
              <th
                class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:text-slate-200"
                @click="sortBy('original_size')"
              >
                <span class="inline-flex items-center">
                  Size
                  <svg v-if="sortField === 'original_size' && sortOrder === 'asc'" class="w-3 h-3 ml-1 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else-if="sortField === 'original_size'" class="w-3 h-3 ml-1 text-accent rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th
                class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:text-slate-200"
                @click="sortBy('total_size')"
              >
                <span class="inline-flex items-center">
                  After
                  <svg v-if="sortField === 'total_size' && sortOrder === 'asc'" class="w-3 h-3 ml-1 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else-if="sortField === 'total_size'" class="w-3 h-3 ml-1 text-accent rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Saved</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Compressed</th>
              <th v-if="instanceType === 'plex'" class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Locked</th>
              <th class="px-3 py-2 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/[0.06]">
            <!-- Loading skeleton -->
            <tr v-if="loading">
              <td v-for="i in (instanceType === 'plex' ? 16 : 15)" :key="i" class="px-3 py-2">
                <div class="h-4 bg-elevated rounded animate-pulse" :style="{ width: i === 3 ? '60%' : i === 5 ? '40%' : '80%' }" />
              </td>
            </tr>
            <!-- Empty state -->
            <tr v-else-if="items.length === 0">
              <td :colspan="instanceType === 'plex' ? 16 : 15" class="px-3 py-12 text-center">
                <svg class="w-12 h-12 mx-auto text-slate-600 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <p class="text-slate-400 text-sm">No media items found.</p>
                <p class="text-slate-500 text-xs mt-1">Try scanning your instance.</p>
              </td>
            </tr>
            <!-- Data rows -->
            <tr
              v-for="item in items"
              :key="item.id"
              class="hover:bg-elevated/30 transition-base"
              :class="{ 'bg-accent/5': selectedIds.includes(item.id) }"
            >
              <td class="px-3 py-2">
                <input
                  type="checkbox"
                  :checked="selectedIds.includes(item.id)"
                  @click="onRowCheck(item, $event)"
                  class="rounded border-white/[0.08] bg-elevated text-accent focus:ring-accent/50"
                />
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span class="text-lg" :title="item.media_type">{{ typeIcon(item.media_type) }}</span>
              </td>
              <td class="px-3 py-2">
                <div class="text-sm font-medium text-slate-100 truncate max-w-xs" :title="item.title">
                  {{ item.title }}
                </div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-400 font-mono">
                {{ item.year ?? '-' }}
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <div class="relative group">
                  <span class="text-sm text-slate-300 cursor-help font-mono">{{ item.total_images }}</span>
                  <div
                    v-if="item.images && item.images.length > 0"
                    class="absolute bottom-full left-0 mb-2 hidden group-hover:block z-10"
                  >
                    <div class="bg-slate-700 border border-white/[0.08] rounded-md shadow-lg px-3 py-2 text-xs text-slate-200 whitespace-nowrap">
                      <div v-for="img in item.images" :key="img.role" class="py-0.5">
                        <span class="font-medium">{{ img.role }}:</span>
                        {{ formatBytes(img.size_bytes) }} ({{ img.width }}x{{ img.height }})
                      </div>
                    </div>
                  </div>
                </div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-300 font-mono">
                <span v-if="item.fanart_size">{{ formatBytes(item.fanart_size) }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-300 font-mono">
                <span v-if="item.poster_size">{{ formatBytes(item.poster_size) }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-300 font-mono">
                <span v-if="item.clear_logo_size">{{ formatBytes(item.clear_logo_size) }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-300 font-mono">
                <span v-if="item.season_poster_size">{{ formatBytes(item.season_poster_size) }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-sm text-slate-300 font-mono">
                <span v-if="item.banner_size">{{ formatBytes(item.banner_size) }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span :class="sizeBadgeClass(item.compressed && item.original_size > 0 ? item.original_size : item.total_size)" class="font-mono">
                  {{ formatBytes(item.compressed && item.original_size > 0 ? item.original_size : item.total_size) }}
                </span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span v-if="item.compressed && item.original_size > 0" class="size-badge-small font-mono">
                  {{ formatBytes(item.total_size) }}
                </span>
                <span v-else class="text-slate-500 text-sm">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span v-if="item.compressed && item.original_size > 0" class="text-sm font-medium text-accent font-mono">
                  −{{ formatBytes(item.original_size - item.total_size) }}
                  <span class="text-xs text-slate-500">({{ Math.round((1 - item.total_size / item.original_size) * 100) }}%)</span>
                </span>
                <span v-else class="text-slate-500 text-sm">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span
                  v-if="item.compressed"
                  class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-accent/10 text-accent"
                >
                  <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  Compressed
                </span>
                <span v-else class="text-slate-500 text-sm">—</span>
              </td>
              <td v-if="instanceType === 'plex'" class="px-3 py-2 whitespace-nowrap">
                <svg
                  v-if="item.locked"
                  class="w-4 h-4 text-warning"
                  fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
                  title="Locked"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span v-else class="text-slate-500 text-sm">—</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <button
                  @click="$emit('compress', [item.id])"
                  :disabled="item.compressed"
                  class="inline-flex items-center px-2 py-1 text-xs font-medium rounded transition-base"
                  :class="item.compressed ? 'bg-elevated text-slate-500 cursor-not-allowed' : 'bg-accent-dim text-accent hover:bg-accent/20'"
                >
                  <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                  Compress
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between mt-4">
      <p class="text-sm text-slate-400">
        Showing page <span class="font-mono">{{ currentPage }}</span> of <span class="font-mono">{{ totalPages }}</span> (<span class="font-mono">{{ total }}</span> total)
      </p>
      <div class="flex items-center space-x-2">
        <button
          @click="$emit('pageChange', currentPage - 1)"
          :disabled="currentPage <= 1"
          class="px-3 py-1.5 text-sm rounded-md bg-elevated text-slate-300 hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed transition-base"
        >
          Previous
        </button>
        <button
          v-for="p in visiblePages"
          :key="p"
          @click="$emit('pageChange', p)"
          class="px-3 py-1.5 text-sm rounded-md transition-base font-mono"
          :class="p === currentPage ? 'bg-accent text-base' : 'bg-elevated text-slate-300 hover:bg-slate-700'"
        >
          {{ p }}
        </button>
        <button
          @click="$emit('pageChange', currentPage + 1)"
          :disabled="currentPage >= totalPages"
          class="px-3 py-1.5 text-sm rounded-md bg-elevated text-slate-300 hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed transition-base"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MediaItem } from '../types'

const props = defineProps<{
  items: MediaItem[]
  loading: boolean
  instanceId: string
  instanceType: string
  selectedIds: string[]
  lastSelectedId: string | null
  selectAllMode: boolean
  currentPage: number
  totalPages: number
  total: number
  totalItems: number
  sortField: string
  sortOrder: string
}>()

const emit = defineEmits<{
  compress: [ids: string[]]
  selectAll: []
  selectAllAcrossPages: []
  clearSelection: []
  deselectAll: []
  toggleSelect: [id: string]
  selectRange: [fromId: string, toId: string]
  pageChange: [page: number]
  sort: [field: string, order: 'asc' | 'desc']
}>()

const selectedCount = computed(() => props.selectedIds.length)
const allSelected = computed(() =>
  props.selectAllMode || (props.items.length > 0 && props.selectedIds.length === props.items.length)
)
const someSelected = computed(() => props.selectAllMode || props.selectedIds.length > 0)

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = props.totalPages
  const current = props.currentPage
  let start = Math.max(1, current - 2)
  let end = Math.min(total, current + 2)
  if (end - start < 4) {
    if (start === 1) end = Math.min(total, start + 4)
    else start = Math.max(1, end - 4)
  }
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

function onSelectAll() {
  if (props.selectAllMode) {
    emit('clearSelection')
  } else if (allSelected.value) {
    emit('selectAllAcrossPages')
  } else {
    emit('selectAll')
  }
}

function onRowCheck(item: MediaItem, event: MouseEvent) {
  if (event.shiftKey && props.lastSelectedId) {
    event.preventDefault()
    emit('selectRange', props.lastSelectedId, item.id)
  } else {
    emit('toggleSelect', item.id)
  }
}

function sortBy(field: string) {
  // If clicking the same column, toggle direction; otherwise default to desc
  const order: 'asc' | 'desc' = props.sortField === field
    ? (props.sortOrder === 'asc' ? 'desc' : 'asc')
    : 'desc'
  emit('sort', field, order)
}

function typeIcon(type: string): string {
  switch (type) {
    case 'movie': return '🎬'
    case 'series': return '📺'
    case 'season': return '📋'
    case 'episode': return '🎞️'
    case 'collection': return '📂'
    default: return '📄'
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}

function sizeBadgeClass(bytes: number): string {
  if (bytes > 10 * 1024 * 1024) return 'size-badge-large' // > 10MB
  if (bytes > 1024 * 1024) return 'size-badge-medium' // > 1MB
  return 'size-badge-small'
}
</script>