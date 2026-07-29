import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { CompressionJob, CompressConfig } from '../types'
import {
  compressItems as apiCompressItems,
  getCompressJob,
  getRecentJobs,
  cancelCompressJob as apiCancelCompressJob,
} from '../api'

export const useCompressStore = defineStore('compress', () => {
  const currentJob = ref<CompressionJob | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  async function pollJobStatus(jobId: string) {
    stopPolling()
    // Immediate first fetch so UI updates instantly
    try {
      const job = await getCompressJob(jobId)
      currentJob.value = job
      if (job.status === 'completed' || job.status === 'failed' || job.status === 'cancelled') {
        stopPolling()
        return
      }
    } catch {
      stopPolling()
      return
    }
    // Then poll every 2 seconds
    pollTimer = setInterval(async () => {
      try {
        const job = await getCompressJob(jobId)
        currentJob.value = job
        if (
          job.status === 'completed' ||
          job.status === 'failed' ||
          job.status === 'cancelled'
        ) {
          stopPolling()
        }
      } catch {
        stopPolling()
      }
    }, 2000)
  }

  async function startCompression(
    instanceId: string,
    itemIds: string[] | null,
    config: Partial<CompressConfig>
  ) {
    loading.value = true
    error.value = null
    try {
      const job = await apiCompressItems({
        instance_id: instanceId,
        media_item_ids: itemIds,
        ...config,
      })
      currentJob.value = job
      if (job.status === 'pending' || job.status === 'running') {
        pollJobStatus(job.id)
      }
      return job
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to start compression'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function cancelJob(jobId: string) {
    try {
      await apiCancelCompressJob(jobId)
      stopPolling()
      if (currentJob.value && currentJob.value.id === jobId) {
        currentJob.value.status = 'cancelled'
      }
    } catch (e: any) {
      error.value = e?.response?.data?.error || e?.message || 'Failed to cancel job'
    }
  }

  // Check for running jobs for a given instance and resume polling if found
  async function resumeRunningJob(instanceId: string) {
    // Already tracking a running job — don't override
    if (currentJob.value && (currentJob.value.status === 'pending' || currentJob.value.status === 'running')) {
      // Make sure polling is active
      if (!pollTimer) {
        pollJobStatus(currentJob.value.id)
      }
      return
    }
    try {
      const jobs = await getRecentJobs(10, instanceId)
      const running = jobs.find(
        (j) => j.status === 'pending' || j.status === 'running'
      )
      if (running) {
        currentJob.value = running
        pollJobStatus(running.id)
      }
    } catch {
      // ignore — no running jobs found
    }
  }

  function clearCurrentJob() {
    stopPolling()
    currentJob.value = null
  }

  return {
    currentJob,
    loading,
    error,
    startCompression,
    pollJobStatus,
    cancelJob,
    stopPolling,
    resumeRunningJob,
    clearCurrentJob,
  }
})