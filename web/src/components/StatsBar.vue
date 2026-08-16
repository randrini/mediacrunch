<template>
  <div class="card-glass px-4 py-2.5">
    <div class="grid grid-cols-2 gap-3 md:flex md:items-center md:justify-around md:gap-3">
      <!-- Total Items -->
      <div class="flex items-center space-x-2.5">
        <div class="w-7 h-7 rounded-sm bg-highlight flex items-center justify-center">
          <svg class="w-3.5 h-3.5 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
          </svg>
        </div>
        <div>
          <p class="text-[16px] font-semibold text-text-primary font-mono tabular-nums">{{ stats.total_items }}</p>
          <p class="text-[10px] text-text-tertiary uppercase tracking-wider">Items</p>
        </div>
      </div>

      <!-- Divider -->
      <div class="hidden md:block w-px h-6 bg-border" />

      <!-- Total Size -->
      <div class="flex items-center space-x-2.5">
        <div class="w-7 h-7 rounded-sm bg-highlight flex items-center justify-center">
          <svg class="w-3.5 h-3.5 text-warning" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
        </div>
        <div>
          <p class="text-[16px] font-semibold text-warning font-mono tabular-nums">{{ formatBytes(stats.total_size) }}</p>
          <p class="text-[10px] text-text-tertiary uppercase tracking-wider">Total Size</p>
        </div>
      </div>

      <!-- Divider -->
      <div class="hidden md:block w-px h-6 bg-border" />

      <!-- Total Saved -->
      <div class="flex items-center space-x-2.5">
        <div class="w-7 h-7 rounded-sm bg-highlight flex items-center justify-center">
          <svg class="w-3.5 h-3.5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div>
          <p class="text-[16px] font-semibold text-accent font-mono tabular-nums">{{ formatBytes(stats.total_savings) }}</p>
          <p class="text-[10px] text-text-tertiary uppercase tracking-wider">Saved</p>
        </div>
      </div>

      <!-- Divider -->
      <div class="hidden md:block w-px h-6 bg-border" />

      <!-- Savings Percentage -->
      <div class="flex items-center space-x-2.5">
        <div class="w-7 h-7 rounded-sm bg-highlight flex items-center justify-center">
          <svg class="w-3.5 h-3.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </div>
        <div>
          <p class="text-[16px] font-semibold text-success font-mono tabular-nums">{{ savingsPercent }}%</p>
          <p class="text-[10px] text-text-tertiary uppercase tracking-wider">Savings</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Stats } from '../types'

const props = defineProps<{
  stats: Stats
}>()

const savingsPercent = computed(() => {
  if (props.stats.total_size === 0) return 0
  return Math.round((props.stats.total_savings / props.stats.total_size) * 100)
})

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}
</script>