<template>
  <div>
    <!-- Instance Header -->
    <div v-if="instance" class="mb-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <button
            @click="router.push('/instances')"
            class="text-slate-500 hover:text-slate-200 transition-base"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-xl font-bold text-slate-100">{{ instance.name }}</h1>
            <div class="flex items-center space-x-2 mt-1">
              <span
                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                :class="typeBadgeClass"
              >
                {{ instance.type }}
              </span>
              <span class="text-sm text-slate-500">{{ instance.host }}</span>
            </div>
          </div>
        </div>
        <div class="flex items-center space-x-2">
          <button
            @click="handleScan"
            class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md bg-accent text-white hover:bg-accent-hover transition-base"
          >
            <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Scan
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Bar -->
    <div v-if="stats" class="mb-4">
      <StatsBar :stats="stats" />
    </div>

    <!-- Search & Filters -->
    <div class="mb-4">
      <SearchFilter
        :show-lock-filter="instance?.type === 'plex'"
        @filter-change="onFilterChange"
      />
    </div>

    <!-- Active compression job indicator -->
    <div
      v-if="currentJob && (currentJob.status === 'pending' || currentJob.status === 'running')"
      class="bg-sky-900/20 border border-sky-700/30 rounded-lg px-3 py-2 mb-4"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <svg class="animate-spin h-5 w-5 text-sky-400" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <div>
            <p class="text-sm text-sky-200 font-medium">Compression in progress...</p>
            <p class="text-xs text-sky-300/70">
              {{ currentJob.processed_items }}/{{ currentJob.total_items }} items
              <span v-if="currentJob.total_images > 0">
                &middot; {{ currentJob.processed_images }}/{{ currentJob.total_images }} images
              </span>
              &middot; {{ formatBytes(currentJob.saved_bytes) }} saved
            </p>
          </div>
        </div>
        <button
          @click="cancelCurrentJob"
          class="text-sm text-danger hover:text-red-400 underline"
        >
          Cancel
        </button>
      </div>
      <!-- Progress bar -->
      <div class="mt-2 w-full bg-elevated rounded-full h-1.5">
        <div
          class="bg-sky-500 h-1.5 rounded-full transition-all duration-500"
          :style="{ width: progressPercent + '%' }"
        />
      </div>
    </div>

    <!-- Media Table -->
    <MediaTable
      :items="mediaStore.items"
      :loading="mediaStore.loading"
      :instance-id="instanceId"
      :instance-type="instance?.type || ''"
      :selected-ids="[...mediaStore.selectedIds]"
      :current-page="mediaStore.page"
      :total-pages="mediaStore.totalPages"
      :total="mediaStore.total"
      :sort-field="mediaStore.filters.sort || 'total_size'"
      :sort-order="mediaStore.filters.order || 'desc'"
      @compress="handleCompress"
      @select-all="mediaStore.selectAll()"
      @deselect-all="mediaStore.deselectAll()"
      @toggle-select="mediaStore.toggleSelect"
      @page-change="handlePageChange"
      @sort="handleSort"
    />

    <!-- Compress Config Modal -->
    <teleport to="body">
      <div
        v-if="showCompressModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
      >
        <div class="absolute inset-0 bg-base/80 backdrop-blur-md" @click="showCompressModal = false" />
        <div class="relative card-glass w-full max-w-md mx-4 p-5">
          <h3 class="text-lg font-semibold text-slate-100 mb-4">Compress Settings</h3>

          <div class="space-y-3">
            <!-- Quality -->
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1">Quality (default)</label>
              <input
                v-model.number="compressConfig.quality_default"
                type="number"
                min="1"
                max="100"
                class="w-full bg-elevated border border-white/[0.08] rounded-md px-3 py-2 text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50"
              />
            </div>

            <!-- Max Width -->
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1">Max Width (default)</label>
              <input
                v-model.number="compressConfig.max_width_default"
                type="number"
                min="100"
                step="100"
                class="w-full bg-elevated border border-white/[0.08] rounded-md px-3 py-2 text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50"
              />
            </div>

            <!-- Min Saving -->
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1">Min Saving (KB)</label>
              <input
                v-model.number="compressConfig.min_saving_kb"
                type="number"
                min="0"
                class="w-full bg-elevated border border-white/[0.08] rounded-md px-3 py-2 text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50"
              />
            </div>

            <!-- Backup toggle -->
            <label class="flex items-center space-x-3">
              <input
                v-model="compressConfig.backup"
                type="checkbox"
                class="rounded border-white/[0.08] bg-elevated text-accent focus:ring-accent/50"
              />
              <span class="text-sm text-slate-300">Create backup before compressing</span>
            </label>

            <!-- Lock Plex toggle -->
            <label v-if="instance?.type === 'plex'" class="flex items-center space-x-3">
              <input
                v-model="compressConfig.lock_plex"
                type="checkbox"
                class="rounded border-white/[0.08] bg-elevated text-accent focus:ring-accent/50"
              />
              <span class="text-sm text-slate-300">Lock Plex metadata fields</span>
            </label>
          </div>

          <div class="flex justify-end space-x-3 mt-4">
            <button
              @click="showCompressModal = false"
              class="px-3 py-1.5 text-sm font-medium text-slate-300 bg-elevated rounded-md hover:bg-slate-700 transition-base"
            >
              Cancel
            </button>
            <button
              @click="startCompression"
              :disabled="compressLoading"
              class="px-3 py-1.5 text-sm font-medium text-white bg-accent rounded-md hover:bg-accent-hover transition-base disabled:opacity-50"
            >
              {{ compressLoading ? 'Starting...' : 'Start Compression' }}
            </button>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useInstancesStore } from '../composables/useInstances'
import { useMediaStore } from '../composables/useMedia'
import { useCompressStore } from '../composables/useCompress'
import { getInstance, getInstanceStats, getInstanceSettings } from '../api'
import type { Instance, Stats, InstanceSettings } from '../types'
import MediaTable from '../components/MediaTable.vue'
import SearchFilter from '../components/SearchFilter.vue'
import StatsBar from '../components/StatsBar.vue'

const route = useRoute()
const router = useRouter()
const instanceStore = useInstancesStore()
const mediaStore = useMediaStore()
const compressState = useCompressStore()
const currentJob = computed(() => compressState.currentJob)
const compressLoading = computed(() => compressState.loading)

const instanceId = computed(() => route.params.id as string)
const instance = ref<Instance | null>(null)
const stats = ref<Stats | null>(null)
const showCompressModal = ref(false)
const compressTargetIds = ref<string[] | null>(null)
const instanceSettings = ref<InstanceSettings | null>(null)

const compressConfig = ref({
  quality_default: 80,
  max_width_default: 1920,
  min_saving_kb: 50,
  backup: false,
  lock_plex: false,
})

const progressPercent = computed(() => {
  const job = compressState.currentJob
  if (!job) return 0
  // Prefer image-level progress when available, fall back to item-level
  if (job.total_images > 0) {
    return Math.round((job.processed_images / job.total_images) * 100)
  }
  if (job.total_items === 0) return 0
  return Math.round((job.processed_items / job.total_items) * 100)
})

const typeBadgeClass = computed(() => {
  switch (instance.value?.type) {
    case 'radarr': return 'bg-sky-900/30 text-sky-300'
    case 'sonarr': return 'bg-violet-900/30 text-violet-300'
    case 'plex': return 'bg-orange-900/30 text-orange-400'
    default: return 'bg-elevated text-slate-300'
  }
})

async function loadInstance() {
  try {
    instance.value = await getInstance(instanceId.value)
  } catch {
    router.push('/instances')
  }
}

async function loadSettings() {
  try {
    instanceSettings.value = await getInstanceSettings(instanceId.value)
    // Pre-fill compress config from instance settings
    const s = instanceSettings.value
    if (s) {
      if (s.quality?.default !== undefined) compressConfig.value.quality_default = s.quality.default
      if (s.max_width?.default !== undefined) compressConfig.value.max_width_default = s.max_width.default
      if (s.min_saving_kb !== undefined) compressConfig.value.min_saving_kb = s.min_saving_kb
      if (s.backup !== undefined) compressConfig.value.backup = s.backup
      if (s.lock_plex !== undefined) compressConfig.value.lock_plex = s.lock_plex
    }
  } catch {
    // settings may not be available yet
  }
}

async function loadStats() {
  try {
    stats.value = await getInstanceStats(instanceId.value)
  } catch {
    // stats may not be available
  }
}

function onFilterChange(filters: any) {
  if (filters.search !== undefined) mediaStore.setFilter('search', filters.search)
  if (filters.type !== undefined) mediaStore.setFilter('type', filters.type)
  if (filters.compressed !== undefined) mediaStore.setFilter('compressed', filters.compressed)
  if (filters.locked !== undefined) mediaStore.setFilter('locked', filters.locked)
  mediaStore.fetchItems(instanceId.value)
}

function handlePageChange(page: number) {
  mediaStore.setPage(page)
  mediaStore.fetchItems(instanceId.value)
}

function handleSort(field: string, order: 'asc' | 'desc') {
  mediaStore.setFilter('sort', field)
  mediaStore.setFilter('order', order)
  mediaStore.fetchItems(instanceId.value)
}

async function handleCompress(ids: string[]) {
  compressTargetIds.value = ids
  // Refresh settings before showing modal so defaults are current
  await loadSettings()
  showCompressModal.value = true
}

async function startCompression() {
  const config = {
    quality: { default: compressConfig.value.quality_default },
    max_width: { default: compressConfig.value.max_width_default },
    min_saving_kb: compressConfig.value.min_saving_kb,
    backup: compressConfig.value.backup,
    lock_plex: compressConfig.value.lock_plex,
  }
  try {
    await compressState.startCompression(instanceId.value, compressTargetIds.value, config)
    showCompressModal.value = false
    // Refresh data when job completes
    const checkComplete = setInterval(() => {
      const job = compressState.currentJob
      if (job && (job.status === 'completed' || job.status === 'failed' || job.status === 'cancelled')) {
        clearInterval(checkComplete)
        mediaStore.fetchItems(instanceId.value)
        loadStats()
      }
    }, 2000)
    // Safety: stop checking after 10 minutes
    setTimeout(() => clearInterval(checkComplete), 600000)
  } catch {
    // error handled by composable
  }
}

function cancelCurrentJob() {
  const job = compressState.currentJob
  if (job) {
    compressState.cancelJob(job.id)
  }
}

function handleScan() {
  instanceStore.scanInstance(instanceId.value).then(() => {
    mediaStore.fetchItems(instanceId.value)
    loadStats()
  }).catch((e: any) => {
    console.error('Failed to scan:', e?.message || e)
  })
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}

onMounted(async () => {
  await loadInstance()
  await loadSettings()
  await mediaStore.fetchItems(instanceId.value)
  await loadStats()
  // Resume polling for any running compression job for this instance
  compressState.resumeRunningJob(instanceId.value)
})

watch(instanceId, async () => {
  compressState.clearCurrentJob()
  await loadInstance()
  await loadSettings()
  await mediaStore.fetchItems(instanceId.value)
  await loadStats()
  // Resume polling for any running compression job for this instance
  compressState.resumeRunningJob(instanceId.value)
})
</script>