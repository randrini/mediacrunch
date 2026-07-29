import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { InstanceSettings } from '../types'
import { getInstanceSettings, updateInstanceSettings } from '../api'

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Record<string, InstanceSettings>>({})
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchSettings(instanceId: string) {
    loading.value = true
    error.value = null
    try {
      settings.value[instanceId] = await getInstanceSettings(instanceId)
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to fetch settings'
    } finally {
      loading.value = false
    }
  }

  async function saveSettings(instanceId: string, data: Partial<InstanceSettings>) {
    loading.value = true
    error.value = null
    try {
      settings.value[instanceId] = await updateInstanceSettings(instanceId, data)
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to save settings'
      throw e
    } finally {
      loading.value = false
    }
  }

  return { settings, loading, error, fetchSettings, saveSettings }
})
