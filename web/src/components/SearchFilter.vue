<template>
  <div class="flex flex-wrap items-center gap-2">
    <!-- Search -->
    <div class="relative flex-1 min-w-[180px] max-w-xs">
      <svg
        class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-tertiary"
        fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        v-model="searchText"
        type="text"
        placeholder="Search by title..."
        class="w-full bg-elevated border border-border rounded-sm pl-8 pr-3 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent placeholder-text-tertiary"
      />
    </div>

    <!-- Type filter -->
    <select
      v-model="typeFilter"
      aria-label="Filter by type"
      class="bg-elevated border border-border rounded-sm px-2.5 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
    >
      <option value="">All Types</option>
      <option value="movie">Movie</option>
      <option value="series">Series</option>
      <option value="season">Season</option>
      <option value="episode">Episode</option>
      <option value="collection">Collection</option>
    </select>

    <!-- Compressed filter -->
    <select
      v-model="compressedFilter"
      aria-label="Filter by compression status"
      class="bg-elevated border border-border rounded-sm px-2.5 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
    >
      <option value="">All Status</option>
      <option :value="'1'">Compressed</option>
      <option :value="'0'">Not Compressed</option>
    </select>

    <!-- Locked filter (plex only) -->
    <select
      v-if="showLockFilter"
      v-model="lockedFilter"
      aria-label="Filter by lock status"
      class="bg-elevated border border-border rounded-sm px-2.5 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
    >
      <option value="">All Lock</option>
      <option :value="'1'">Locked</option>
      <option :value="'0'">Unlocked</option>
    </select>

    <!-- Clear filters -->
    <button
      @click="clearFilters"
      class="inline-flex items-center px-2 py-1 text-xs rounded-sm bg-elevated text-text-tertiary hover:text-text-primary hover:bg-highlight transition-base"
    >
      <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
      </svg>
      Clear
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'

defineProps<{
  showLockFilter?: boolean
}>()

const emit = defineEmits<{
  filterChange: [filters: {
    search?: string
    type?: string
    compressed?: '0' | '1' | undefined
    locked?: '0' | '1' | undefined
  }]
}>()

const searchText = ref('')
const typeFilter = ref('')
const compressedFilter = ref<'0' | '1' | undefined>(undefined)
const lockedFilter = ref<'0' | '1' | undefined>(undefined)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function emitFilters() {
  emit('filterChange', {
    search: searchText.value || undefined,
    type: typeFilter.value || undefined,
    compressed: compressedFilter.value,
    locked: lockedFilter.value,
  })
}

function clearFilters() {
  searchText.value = ''
  typeFilter.value = ''
  compressedFilter.value = undefined
  lockedFilter.value = undefined
}

watch([searchText, typeFilter, compressedFilter, lockedFilter], () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(emitFilters, 300)
})

onMounted(() => {
  emitFilters()
})

onUnmounted(() => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
})
</script>
