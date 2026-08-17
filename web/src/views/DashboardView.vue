<template>
  <div>
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4">
      <h1 class="text-2xl font-bold text-text-primary tracking-tight">Dashboard</h1>
      <button
        v-if="store.instances.length > 0"
        @click="router.push('/instances')"
        class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md bg-accent text-ink hover:bg-accent-hover transition-base"
      >
        <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        Manage Instances
      </button>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-12">
      <svg class="animate-spin h-8 w-8 text-accent mx-auto" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
      <p class="mt-3 text-text-secondary text-sm">Loading instances...</p>
    </div>

    <!-- Error -->
    <div v-else-if="store.error" class="bg-danger/10 border border-danger/30 rounded-md p-4 mb-4">
      <p class="text-danger text-sm">{{ store.error }}</p>
      <button @click="store.fetchInstances()" class="mt-2 text-sm text-danger hover:text-danger underline">Retry</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="store.instances.length === 0" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-text-tertiary mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="9" />
        <circle cx="12" cy="12" r="4.5" />
        <path d="M12 3v3M12 18v3M3 12h3M18 12h3" />
        <path d="M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M18.4 5.6l-2.1 2.1M7.7 16.3l-2.1 2.1" />
      </svg>
      <h2 class="text-xl font-semibold text-text-secondary mb-2">Welcome to MediaCrunch</h2>
      <p class="text-text-tertiary text-sm mb-4 max-w-md mx-auto">
        Get started by adding your first Radarr, Sonarr, or Plex instance to begin compressing media images.
      </p>
      <button
        @click="router.push('/instances')"
        class="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-accent text-ink hover:bg-accent-hover transition-base"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        Add Instance
      </button>
    </div>

    <!-- Dashboard Content -->
    <div v-else>
      <!-- Overall Stats -->
      <div v-if="overallStats" class="mb-4">
        <StatsBar :stats="overallStats" />
      </div>

      <!-- Instance Grid -->
      <h2 class="text-lg font-semibold text-text-primary mb-4">Instances</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 mb-8">
        <InstanceCard
          v-for="inst in store.instances"
          :key="inst.id"
          :instance="inst"
          :stats="instanceStatsMap[inst.id] || null"
          @scan="handleScan"
          @view-media="goToMedia"
          @lock="handleLock"
          @edit="handleEdit"
          @delete="handleDelete"
        />
      </div>

      <!-- Recent Jobs -->
      <div v-if="recentJobs.length > 0">
        <h2 class="text-lg font-semibold text-text-primary mb-4">Recent Compression Jobs</h2>
        <div class="card-glass overflow-hidden">
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-border">
              <thead class="bg-elevated">
                <tr>
                  <th class="px-3 py-2 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Instance</th>
                  <th class="px-3 py-2 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Status</th>
                  <th class="px-3 py-2 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Progress</th>
                  <th class="px-3 py-2 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Saved</th>
                  <th class="px-3 py-2 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Date</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border">
                <tr v-for="job in recentJobs" :key="job.id" class="hover:bg-highlight/50 transition-base">
                  <td class="px-3 py-2 text-sm text-text-secondary">{{ getInstanceName(job.instance_id) }}</td>
                  <td class="px-3 py-2">
                    <span :class="statusBadge(job.status)">{{ job.status }}</span>
                  </td>
                  <td class="px-3 py-2 text-sm text-text-secondary font-mono tabular-nums">
                    {{ job.processed_items }}/{{ job.total_items }}
                  </td>
                  <td class="px-3 py-2 text-sm text-accent font-mono tabular-nums">
                    {{ formatSize(job.saved_bytes) }}
                  </td>
                  <td class="px-3 py-2 text-sm text-text-secondary">
                    {{ formatDate(job.created_at) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useInstancesStore } from '../composables/useInstances'
import { getStats, getInstanceStats, getRecentJobs } from '../api'
import type { Stats, CompressionJob } from '../types'
import InstanceCard from '../components/InstanceCard.vue'
import StatsBar from '../components/StatsBar.vue'

const router = useRouter()
const store = useInstancesStore()

const overallStats = ref<Stats | null>(null)
const instanceStatsMap = ref<Record<string, Stats>>({})
const recentJobs = ref<CompressionJob[]>([])

async function loadStats() {
  try {
    overallStats.value = await getStats()
  } catch {
    // stats may not be available
  }
}

async function loadInstanceStats() {
  for (const inst of store.instances) {
    try {
      const s = await getInstanceStats(inst.id)
      instanceStatsMap.value[inst.id] = s
    } catch {
      // skip
    }
  }
}

function handleScan(id: string) {
  store.scanInstance(id).then(() => store.fetchInstances()).catch((e: any) => {
    console.error('Failed to scan:', e?.message || e)
  })
}

function goToMedia(id: string) {
  router.push(`/instances/${id}/media`)
}

function handleLock(id: string) {
  store.lockInstance(id).catch((e: any) => {
    console.error('Failed to lock:', e?.message || e)
  })
}

function handleEdit(instance: any) {
  router.push(`/instances/${instance.id}/media`)
}

function handleDelete(instance: any) {
  store.deleteInstance(instance.id).catch((e: any) => {
    console.error('Failed to delete:', e?.message || e)
  })
}

function statusBadge(status: string): string {
  const base = 'inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium'
  switch (status) {
    case 'completed': return `${base} bg-accent-muted text-accent`
    case 'running': return `${base} bg-accent-muted text-accent`
    case 'failed': return `${base} bg-danger/10 text-danger`
    case 'cancelled': return `${base} bg-elevated text-text-tertiary`
    default: return `${base} bg-warning/10 text-warning`
  }
}

function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return '-'
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? '-' : d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatSize(bytes: number): string {
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(1) + ' GB'
  if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(1) + ' MB'
  if (bytes >= 1 << 10) return (bytes / (1 << 10)).toFixed(1) + ' KB'
  return bytes + ' B'
}

function getInstanceName(instanceId: string): string {
  const inst = store.instances.find(i => i.id === instanceId)
  return inst ? inst.name : instanceId.slice(0, 8) + '...'
}

onMounted(async () => {
  await store.fetchInstances()
  await loadStats()
  await loadInstanceStats()
  try {
    recentJobs.value = await getRecentJobs(10)
  } catch {
    // recent jobs may not be available
  }
})
</script>
