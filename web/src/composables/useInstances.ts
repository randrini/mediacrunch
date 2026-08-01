import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Instance } from '../types'
import {
  getInstances,
  createInstance as apiCreateInstance,
  updateInstance as apiUpdateInstance,
  deleteInstance as apiDeleteInstance,
  scanInstance as apiScanInstance,
  lockInstance as apiLockInstance,
} from '../api'

export const useInstancesStore = defineStore('instances', () => {
  const instances = ref<Instance[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const radarrInstances = computed(() =>
    instances.value.filter((i) => i.type === 'radarr')
  )
  const sonarrInstances = computed(() =>
    instances.value.filter((i) => i.type === 'sonarr')
  )
  const plexInstances = computed(() =>
    instances.value.filter((i) => i.type === 'plex')
  )

  async function fetchInstances() {
    loading.value = true
    error.value = null
    try {
      instances.value = await getInstances()
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to fetch instances'
    } finally {
      loading.value = false
    }
  }

  async function createInstance(data: {
    type: 'radarr' | 'sonarr' | 'plex'
    name: string
    host: string
    api_key: string
    path_prefix: string
  }) {
    const instance = await apiCreateInstance(data)
    instances.value.push(instance)
    return instance
  }

  async function updateInstance(id: string, data: Partial<Instance>) {
    const updated = await apiUpdateInstance(id, data)
    const idx = instances.value.findIndex((i) => i.id === id)
    if (idx !== -1) instances.value[idx] = updated
    return updated
  }

  async function deleteInstance(id: string) {
    await apiDeleteInstance(id)
    instances.value = instances.value.filter((i) => i.id !== id)
  }

  async function scanInstance(id: string) {
    await apiScanInstance(id)
  }

  async function lockInstance(id: string) {
    await apiLockInstance(id)
  }

  return {
    instances,
    loading,
    error,
    radarrInstances,
    sonarrInstances,
    plexInstances,
    fetchInstances,
    createInstance,
    updateInstance,
    deleteInstance,
    scanInstance,
    lockInstance,
  }
})
