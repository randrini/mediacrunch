<template>
  <div class="space-y-4">
    <!-- Page Header -->
    <div>
      <h1 class="text-2xl font-bold text-text-primary tracking-tight">Activity Logs</h1>
      <p class="mt-0.5 text-xs text-text-tertiary font-mono">
        Real-time stream of compression jobs, scans, and system events
      </p>
    </div>

    <!-- Filter Bar -->
    <div class="flex flex-wrap items-center gap-2">
      <!-- Level filter chips -->
      <div class="flex items-center space-x-1.5">
        <button
          v-for="level in levels"
          :key="level.value"
          @click="setLevel(level.value)"
          class="px-2 py-0.5 rounded-sm text-[11px] font-mono font-medium transition-base"
          :class="activeLevel === level.value ? 'bg-accent text-ink' : 'bg-highlight text-text-secondary hover:bg-elevated'"
        >
          {{ level.label }}
        </button>
      </div>

      <!-- Search input -->
      <div class="flex-1 min-w-[160px] max-w-xs">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search logs..."
          @input="onSearch"
          class="w-full bg-elevated border border-border rounded-sm px-2.5 py-1 text-xs text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent font-mono"
        />
      </div>

      <!-- Live indicator -->
      <div v-if="isLive" class="flex items-center space-x-1.5 text-[11px] text-accent">
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-accent"></span>
        </span>
        <span class="font-mono">LIVE</span>
      </div>

      <!-- Clear button -->
      <div class="relative">
        <button
          v-if="!showClearConfirm"
          @click="showClearConfirm = true"
          class="bg-danger/10 border border-danger/30 text-danger hover:bg-danger/20 px-2.5 py-1 rounded-sm text-[11px] font-medium transition-base"
        >
          Clear Logs
        </button>
        <div v-else class="flex items-center space-x-1.5">
          <span class="text-[11px] text-danger">Clear all logs?</span>
          <button
            @click="handleClearLogs"
            class="bg-danger text-ink px-2 py-0.5 rounded-sm text-[11px] font-medium hover:bg-danger/80 transition-base"
          >
            Yes
          </button>
          <button
            @click="showClearConfirm = false"
            class="bg-highlight text-text-secondary px-2 py-0.5 rounded-sm text-[11px] font-medium hover:bg-elevated transition-base"
          >
            No
          </button>
        </div>
      </div>
    </div>

    <!-- Error banner -->
    <div
      v-if="error"
      class="bg-danger/10 border border-danger/30 rounded-sm px-3 py-2 text-xs text-danger"
    >
      {{ error }}
    </div>

    <!-- Loading state -->
    <div v-if="loading && logs.length === 0" class="flex justify-center py-12">
      <svg class="animate-spin h-7 w-7 text-accent" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!loading && logs.length === 0"
      class="flex flex-col items-center justify-center py-14 text-text-tertiary"
    >
      <svg class="w-10 h-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <p class="text-xs font-medium text-text-tertiary">No logs found</p>
      <p class="text-[11px] mt-0.5 text-text-tertiary">Try adjusting your filters or wait for new events</p>
    </div>

    <!-- Log list -->
    <div v-else class="space-y-0.5">
      <div
        v-for="entry in logs"
        :key="entry.id"
        class="bg-surface rounded-sm p-2.5 hover:bg-elevated border border-border hover:border-border-strong transition-base"
      >
        <div class="flex flex-col sm:flex-row sm:items-start gap-2.5">
          <!-- Timestamp -->
          <div class="font-mono text-[11px] text-text-tertiary whitespace-nowrap pt-0.5 sm:min-w-[4rem]">
            {{ formatTime(entry.created_at) }}
          </div>

          <!-- Badges -->
          <div class="flex items-center flex-wrap gap-1 flex-shrink-0 pt-0.5">
            <!-- Level badge -->
            <span
              class="inline-flex items-center px-1.5 py-0.5 rounded-sm text-[10px] font-mono font-medium"
              :class="levelClass(entry.level)"
            >
              {{ entry.level.toUpperCase() }}
            </span>
            <!-- Source badge -->
            <span
              class="inline-flex items-center px-1.5 py-0.5 rounded-sm text-[10px] font-mono font-medium"
              :class="sourceClass(entry.source)"
            >
              {{ entry.source }}
            </span>
            <!-- Instance badge -->
            <span
              v-if="entry.instance_id && instanceMap.get(entry.instance_id)"
              class="inline-flex items-center px-1.5 py-0.5 rounded-sm text-[10px] font-mono font-medium bg-accent-muted text-accent border border-accent/40"
            >
              {{ instanceMap.get(entry.instance_id) }}
            </span>
          </div>

          <!-- Message -->
          <div class="flex-1 min-w-0">
            <p class="text-xs text-text-primary break-words">{{ entry.message }}</p>

            <!-- Collapsible details -->
            <div v-if="entry.details" class="mt-1">
              <button
                @click="toggleExpand(entry.id)"
                class="text-[11px] text-text-tertiary hover:text-text-secondary transition-base font-mono"
              >
                <span class="inline-flex items-center">
                  <svg
                    class="w-3 h-3 mr-1 transition-transform duration-150"
                    :class="expandedIds.has(entry.id) ? 'rotate-180' : ''"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                  {{ expandedIds.has(entry.id) ? 'Hide details' : 'Show details' }}
                </span>
              </button>
              <div
                v-if="expandedIds.has(entry.id)"
                class="mt-1 p-2 rounded-sm bg-base border border-border"
              >
                <pre class="font-mono text-[11px] text-text-tertiary whitespace-pre-wrap break-words">{{ entry.details }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div
      v-if="total > 0"
      class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 pt-2"
    >
      <p class="text-[11px] text-text-tertiary font-mono">
        Showing <span class="text-text-secondary">{{ showingFrom }}</span>
        <span v-if="showingFrom !== showingTo">–<span class="text-text-secondary">{{ showingTo }}</span></span>
        of <span class="text-text-secondary">{{ total }}</span>
      </p>
      <div class="flex items-center flex-wrap gap-1.5">
        <button
          @click="prevPage"
          :disabled="offset === 0"
          class="px-2.5 py-1 rounded-sm text-[11px] font-medium transition-base"
          :class="offset === 0 ? 'bg-surface text-text-tertiary cursor-not-allowed' : 'bg-highlight text-text-secondary hover:bg-elevated'"
        >
          Previous
        </button>
        <button
          @click="nextPage"
          :disabled="offset + limit >= total"
          class="px-2.5 py-1 rounded-sm text-[11px] font-medium transition-base"
          :class="offset + limit >= total ? 'bg-surface text-text-tertiary cursor-not-allowed' : 'bg-highlight text-text-secondary hover:bg-elevated'"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { getLogs, clearLogs } from '../api'
import { useInstancesStore } from '../composables/useInstances'
import type { LogEntry } from '../types'
import type { LogQueryParams } from '../api'

const logs = ref<LogEntry[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)
const activeLevel = ref<string>('all')
const searchQuery = ref('')
const offset = ref(0)
const limit = ref(100)
const showClearConfirm = ref(false)
const expandedIds = ref<Set<number>>(new Set())
const instanceStore = useInstancesStore()
let pollTimer: ReturnType<typeof setInterval> | null = null

const levels = [
  { label: 'All', value: 'all' },
  { label: 'Info', value: 'info' },
  { label: 'Warn', value: 'warn' },
  { label: 'Error', value: 'error' },
  { label: 'Debug', value: 'debug' },
]

const instanceMap = computed(() => {
  const map = new Map<string, string>()
  for (const inst of instanceStore.instances) {
    map.set(inst.id, inst.name)
  }
  return map
})

async function fetchLogs() {
  loading.value = true
  error.value = null
  try {
    const params: LogQueryParams = {
      limit: limit.value,
      offset: offset.value,
    }
    if (activeLevel.value !== 'all') params.level = activeLevel.value
    if (searchQuery.value) params.search = searchQuery.value
    const res = await getLogs(params)
    logs.value = res.logs
    total.value = res.total
  } catch (e: any) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to fetch logs'
  } finally {
    loading.value = false
  }
}

function toggleExpand(id: number) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}

function setLevel(level: string) {
  activeLevel.value = level
  offset.value = 0
  fetchLogs()
}

function onSearch() {
  offset.value = 0
  fetchLogs()
}

function nextPage() {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    fetchLogs()
  }
}

function prevPage() {
  if (offset.value > 0) {
    offset.value = Math.max(0, offset.value - limit.value)
    fetchLogs()
  }
}

async function handleClearLogs() {
  try {
    await clearLogs()
    showClearConfirm.value = false
    fetchLogs()
  } catch (e: any) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to clear logs'
  }
}

function formatTime(ts: string): string {
  const d = new Date(ts)
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function levelClass(level: string): string {
  switch (level) {
    case 'error': return 'bg-danger/10 text-danger border border-danger/30'
    case 'warn': return 'bg-warning/10 text-warning border border-warning/20'
    case 'info': return 'bg-success/10 text-success border border-success/20'
    case 'debug': return 'bg-highlight text-text-tertiary border border-border'
    default: return 'bg-highlight text-text-tertiary border border-border'
  }
}

function sourceClass(source: string): string {
  switch (source) {
    case 'scanner': return 'bg-accent-muted text-accent border border-accent/40'
    case 'compressor': return 'bg-success/10 text-success border border-success/20'
    case 'api': return 'bg-highlight text-text-secondary border border-border'
    case 'system': return 'bg-highlight text-text-secondary border border-border'
    default: return 'bg-highlight text-text-tertiary border border-border'
  }
}

const showingFrom = computed(() => total.value === 0 ? 0 : offset.value + 1)
const showingTo = computed(() => Math.min(offset.value + limit.value, total.value))
const isLive = computed(() => offset.value === 0 && activeLevel.value === 'all' && !searchQuery.value)

// Auto-refresh on first page
watch(isLive, (live) => {
  if (live) {
    pollTimer = setInterval(fetchLogs, 5000)
  } else if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}, { immediate: true })

onMounted(() => {
  instanceStore.fetchInstances()
  fetchLogs()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>