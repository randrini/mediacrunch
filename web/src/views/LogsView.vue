<template>
  <div class="space-y-4">
    <!-- Page Header -->
    <div>
      <h1 class="text-lg font-bold text-slate-100 tracking-tight">Activity Logs</h1>
      <p class="mt-0.5 text-xs text-slate-500 font-mono">
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
          class="px-2 py-0.5 rounded text-[11px] font-mono font-medium transition-base"
          :class="activeLevel === level.value ? 'bg-accent text-base' : 'bg-white/[0.04] text-slate-400 hover:bg-white/[0.08]'"
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
          class="w-full bg-elevated border border-white/[0.06] rounded px-2.5 py-1 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent font-mono"
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
          class="bg-danger/10 border border-danger/20 text-danger hover:bg-danger/20 px-2.5 py-1 rounded text-[11px] font-medium transition-base"
        >
          Clear Logs
        </button>
        <div v-else class="flex items-center space-x-1.5">
          <span class="text-[11px] text-danger">Clear all logs?</span>
          <button
            @click="handleClearLogs"
            class="bg-danger text-white px-2 py-0.5 rounded text-[11px] font-medium hover:bg-red-600 transition-base"
          >
            Yes
          </button>
          <button
            @click="showClearConfirm = false"
            class="bg-white/[0.04] text-slate-400 px-2 py-0.5 rounded text-[11px] font-medium hover:bg-white/[0.08] transition-base"
          >
            No
          </button>
        </div>
      </div>
    </div>

    <!-- Error banner -->
    <div
      v-if="error"
      class="bg-danger/10 border border-danger/20 rounded px-3 py-2 text-xs text-danger"
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
      class="flex flex-col items-center justify-center py-14 text-slate-600"
    >
      <svg class="w-10 h-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <p class="text-xs font-medium text-slate-500">No logs found</p>
      <p class="text-[11px] mt-0.5 text-slate-600">Try adjusting your filters or wait for new events</p>
    </div>

    <!-- Log list -->
    <div v-else class="space-y-0.5">
      <div
        v-for="entry in logs"
        :key="entry.id"
        class="bg-white/[0.015] rounded p-2.5 hover:bg-white/[0.03] border border-white/[0.03] hover:border-white/[0.06] transition-base"
      >
        <div class="flex flex-col sm:flex-row sm:items-start gap-2.5">
          <!-- Timestamp -->
          <div class="font-mono text-[11px] text-slate-600 whitespace-nowrap pt-0.5 sm:min-w-[4rem]">
            {{ formatTime(entry.created_at) }}
          </div>

          <!-- Badges -->
          <div class="flex items-center flex-wrap gap-1 flex-shrink-0 pt-0.5">
            <!-- Level badge -->
            <span
              class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium"
              :class="levelClass(entry.level)"
            >
              {{ entry.level.toUpperCase() }}
            </span>
            <!-- Source badge -->
            <span
              class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium"
              :class="sourceClass(entry.source)"
            >
              {{ entry.source }}
            </span>
            <!-- Instance badge -->
            <span
              v-if="entry.instance_id && instanceMap.get(entry.instance_id)"
              class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/20"
            >
              {{ instanceMap.get(entry.instance_id) }}
            </span>
          </div>

          <!-- Message -->
          <div class="flex-1 min-w-0">
            <p class="text-xs text-slate-200 break-words">{{ entry.message }}</p>

            <!-- Collapsible details -->
            <div v-if="entry.details" class="mt-1">
              <button
                @click="toggleExpand(entry.id)"
                class="text-[11px] text-slate-600 hover:text-slate-400 transition-base font-mono"
              >
                {{ expandedIds.has(entry.id) ? '▲ Hide details' : '▼ Show details' }}
              </button>
              <div
                v-if="expandedIds.has(entry.id)"
                class="mt-1 p-2 rounded bg-base border border-white/[0.04]"
              >
                <pre class="font-mono text-[11px] text-slate-500 whitespace-pre-wrap break-words">{{ entry.details }}</pre>
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
      <p class="text-[11px] text-slate-600 font-mono">
        Showing <span class="text-slate-400">{{ showingFrom }}</span>
        <span v-if="showingFrom !== showingTo">–<span class="text-slate-400">{{ showingTo }}</span></span>
        of <span class="text-slate-400">{{ total }}</span>
      </p>
      <div class="flex items-center flex-wrap gap-1.5">
        <button
          @click="prevPage"
          :disabled="offset === 0"
          class="px-2.5 py-1 rounded text-[11px] font-medium transition-base"
          :class="offset === 0 ? 'bg-white/[0.02] text-slate-700 cursor-not-allowed' : 'bg-white/[0.04] text-slate-400 hover:bg-white/[0.08]'"
        >
          Previous
        </button>
        <button
          @click="nextPage"
          :disabled="offset + limit >= total"
          class="px-2.5 py-1 rounded text-[11px] font-medium transition-base"
          :class="offset + limit >= total ? 'bg-white/[0.02] text-slate-700 cursor-not-allowed' : 'bg-white/[0.04] text-slate-400 hover:bg-white/[0.08]'"
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
    case 'error': return 'bg-danger/10 text-danger border border-danger/20'
    case 'warn': return 'bg-warning/10 text-warning border border-warning/20'
    case 'info': return 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
    case 'debug': return 'bg-white/[0.03] text-slate-500 border border-white/[0.06]'
    default: return 'bg-white/[0.03] text-slate-500 border border-white/[0.06]'
  }
}

function sourceClass(source: string): string {
  switch (source) {
    case 'scanner': return 'bg-accent/10 text-accent border border-accent/20'
    case 'compressor': return 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
    case 'api': return 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
    case 'system': return 'bg-white/[0.03] text-slate-500 border border-white/[0.06]'
    default: return 'bg-white/[0.03] text-slate-500 border border-white/[0.06]'
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