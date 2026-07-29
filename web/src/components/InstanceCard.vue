<template>
  <div
    class="card-glass p-3.5 hover:border-white/[0.12] hover:scale-[1.01] transition-all duration-150 cursor-pointer"
    :class="accentBorder"
    @click="$emit('viewMedia', instance.id)"
  >
    <div class="flex items-start justify-between">
      <div class="flex items-center space-x-2.5">
        <!-- Type Icon -->
        <div
          class="w-8 h-8 rounded flex items-center justify-center"
          :class="accentBg"
        >
          <!-- Radarr: film -->
          <svg v-if="instance.type === 'radarr'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" />
          </svg>
          <!-- Sonarr: tv -->
          <svg v-else-if="instance.type === 'sonarr'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          <!-- Plex: play icon -->
          <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
          <div>
            <h3 class="text-sm font-semibold text-slate-100 hover:text-accent transition-base">{{ instance.name }}</h3>
          <span
            class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium mt-0.5"
            :class="typeBadgeClass"
          >
            {{ instance.type }}
          </span>
        </div>
      </div>
    </div>

    <!-- Details -->
    <div class="mt-3 space-y-1 text-xs text-slate-500">
      <div class="flex items-center space-x-1.5">
        <svg class="w-3.5 h-3.5 flex-shrink-0 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
        <span class="truncate font-mono">{{ instance.host }}</span>
      </div>
      <div class="flex items-center space-x-1.5">
        <svg class="w-3.5 h-3.5 flex-shrink-0 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span class="truncate font-mono">{{ instance.path_prefix }}</span>
      </div>
    </div>

    <!-- Stats -->
    <div class="mt-3 grid grid-cols-3 gap-2 text-center">
      <div class="bg-white/[0.03] rounded py-1.5 px-1">
        <p class="text-sm font-semibold text-slate-100 font-mono tabular-nums">{{ stats?.total_items ?? '-' }}</p>
        <p class="text-[10px] text-slate-500 uppercase tracking-wider">Items</p>
      </div>
      <div class="bg-white/[0.03] rounded py-1.5 px-1">
        <p class="text-sm font-semibold text-slate-100 font-mono tabular-nums">{{ stats ? formatBytes(stats.total_size) : '-' }}</p>
        <p class="text-[10px] text-slate-500 uppercase tracking-wider">Size</p>
      </div>
      <div class="bg-white/[0.03] rounded py-1.5 px-1">
        <p class="text-sm font-semibold text-accent font-mono tabular-nums">{{ stats ? formatBytes(stats.total_savings) : '-' }}</p>
        <p class="text-[10px] text-slate-500 uppercase tracking-wider">Saved</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="mt-3 flex flex-wrap gap-1.5">
      <button
        @click.stop="$emit('scan', instance.id)"
        class="inline-flex items-center px-2.5 py-1 text-[11px] font-medium rounded bg-accent/10 text-accent hover:bg-accent/20 transition-base"
      >
        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Scan
      </button>
      <button
        @click.stop="$emit('viewMedia', instance.id)"
        class="inline-flex items-center px-2.5 py-1 text-[11px] font-medium rounded bg-accent text-base hover:bg-accent-hover transition-base"
      >
        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        Media
      </button>
      <button
        v-if="instance.type === 'plex'"
        @click.stop="$emit('lock', instance.id)"
        class="inline-flex items-center px-2.5 py-1 text-[11px] font-medium rounded bg-warning/10 text-warning hover:bg-warning/20 transition-base"
      >
        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        Lock
      </button>
      <button
        @click.stop="$emit('edit', instance)"
        class="inline-flex items-center px-2.5 py-1 text-[11px] font-medium rounded bg-white/[0.04] text-slate-400 hover:bg-white/[0.08] hover:text-slate-200 transition-base"
      >
        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
        </svg>
        Edit
      </button>
      <button
        @click.stop="$emit('delete', instance)"
        class="inline-flex items-center px-2.5 py-1 text-[11px] font-medium rounded bg-danger/10 text-danger hover:bg-danger/20 transition-base"
      >
        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Instance, Stats } from '../types'

const props = defineProps<{
  instance: Instance
  stats?: Stats | null
}>()

defineEmits<{
  scan: [id: string]
  viewMedia: [id: string]
  lock: [id: string]
  edit: [instance: Instance]
  delete: [instance: Instance]
}>()

const accentBorder = computed(() => {
  switch (props.instance.type) {
    case 'radarr': return 'border-l-2 border-l-blue-500'
    case 'sonarr': return 'border-l-2 border-l-purple-500'
    case 'plex': return 'border-l-2 border-l-orange-500'
    default: return ''
  }
})

const accentBg = computed(() => {
  switch (props.instance.type) {
    case 'radarr': return 'bg-blue-500/15 text-blue-400'
    case 'sonarr': return 'bg-purple-500/15 text-purple-400'
    case 'plex': return 'bg-orange-500/15 text-orange-400'
    default: return 'bg-slate-500/15 text-slate-400'
  }
})

const typeBadgeClass = computed(() => {
  switch (props.instance.type) {
    case 'radarr': return 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
    case 'sonarr': return 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
    case 'plex': return 'bg-orange-500/10 text-orange-400 border border-orange-500/20'
    default: return 'bg-white/[0.04] text-slate-400 border border-white/[0.06]'
  }
})

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}
</script>