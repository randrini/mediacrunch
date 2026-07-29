<template>
  <button
    :disabled="disabled || loading"
    class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md transition-base"
    :class="buttonClass"
    @click="$emit('click')"
  >
    <svg
      v-if="loading"
      class="animate-spin -ml-1 mr-2 h-4 w-4"
      fill="none" viewBox="0 0 24 24"
    >
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
    </svg>
    <svg v-else class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
    </svg>
    <slot>{{ loading ? 'Compressing...' : 'Compress' }}</slot>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  disabled?: boolean
  loading?: boolean
  variant?: 'primary' | 'accent'
}>()

defineEmits<{
  click: []
}>()

const buttonClass = computed(() => {
  if (props.disabled && !props.loading) {
    return 'bg-elevated text-slate-600 cursor-not-allowed'
  }
  const variant = props.variant || 'accent'
  if (variant === 'primary') {
    return 'bg-accent hover:bg-accent-hover text-base'
  }
  return 'bg-accent hover:bg-accent-hover text-base'
})
</script>