<template>
  <div>
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4">
      <h1 class="text-xl font-bold text-slate-100">Instances</h1>
      <button
        @click="openCreateModal"
        class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md bg-accent text-white hover:bg-accent-hover transition-base"
      >
        <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        Add Instance
      </button>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-12">
      <svg class="animate-spin h-8 w-8 text-accent mx-auto" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
      <p class="mt-3 text-slate-400 text-sm">Loading instances...</p>
    </div>

    <!-- Error -->
    <div v-else-if="store.error" class="bg-danger/10 border border-danger/30 rounded-lg p-4 mb-4">
      <p class="text-danger text-sm">{{ store.error }}</p>
      <button @click="store.fetchInstances()" class="mt-2 text-sm text-danger hover:text-danger underline">Retry</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="store.instances.length === 0" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-slate-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
        <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
      <h2 class="text-xl font-semibold text-slate-300 mb-2">No instances yet</h2>
      <p class="text-slate-500 text-sm mb-4">Add your first Radarr, Sonarr, or Plex instance to get started.</p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-accent text-white hover:bg-accent-hover transition-base"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        Add Instance
      </button>
    </div>

    <!-- Instance Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
      <InstanceCard
        v-for="inst in store.instances"
        :key="inst.id"
        :instance="inst"
        :stats="instanceStatsMap[inst.id] || null"
        @scan="handleScan"
        @view-media="goToMedia"
        @lock="handleLock"
        @edit="openEditModal"
        @delete="confirmDelete"
      />
    </div>

    <!-- Create/Edit Modal -->
    <InstanceForm
      :visible="showForm"
      :instance="editingInstance"
      @save="handleSave"
      @cancel="closeForm"
    />

    <!-- Delete Confirmation -->
    <teleport to="body">
      <div
        v-if="deletingInstance"
        class="fixed inset-0 z-50 flex items-center justify-center"
      >
        <div class="absolute inset-0 bg-base/80 backdrop-blur-md" @click="deletingInstance = null" />
        <div class="relative card-glass w-full max-w-sm mx-4 p-5">
          <h3 class="text-lg font-semibold text-slate-100 mb-2">Delete Instance</h3>
          <p class="text-sm text-slate-400 mb-4">
            Are you sure you want to delete <span class="font-medium text-slate-200">{{ deletingInstance.name }}</span>?
            This will also remove all associated media items and compression results.
          </p>
          <div class="flex justify-end space-x-3">
            <button
              @click="deletingInstance = null"
              class="px-3 py-1.5 text-sm font-medium text-slate-300 bg-elevated rounded-md hover:bg-slate-700 transition-base"
            >
              Cancel
            </button>
            <button
              @click="handleDelete"
              class="px-3 py-1.5 text-sm font-medium text-white bg-danger rounded-md hover:bg-red-600 transition-base"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useInstancesStore } from '../composables/useInstances'
import { getInstanceStats } from '../api'
import type { Instance, Stats } from '../types'
import InstanceCard from '../components/InstanceCard.vue'
import InstanceForm from '../components/InstanceForm.vue'

const router = useRouter()
const store = useInstancesStore()

const showForm = ref(false)
const editingInstance = ref<Instance | null>(null)
const deletingInstance = ref<Instance | null>(null)
const instanceStatsMap = ref<Record<string, Stats>>({})

function openCreateModal() {
  editingInstance.value = null
  showForm.value = true
}

function openEditModal(instance: Instance) {
  editingInstance.value = instance
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingInstance.value = null
}

async function handleSave(data: {
  type: 'radarr' | 'sonarr' | 'plex'
  name: string
  host: string
  api_key: string
  path_prefix: string
}) {
  try {
    if (editingInstance.value) {
      await store.updateInstance(editingInstance.value.id, data)
    } else {
      await store.createInstance(data)
    }
    closeForm()
    await loadInstanceStats()
  } catch (e: any) {
    // error handled by store
  }
}

function confirmDelete(instance: Instance) {
  deletingInstance.value = instance
}

async function handleDelete() {
  if (!deletingInstance.value) return
  try {
    await store.deleteInstance(deletingInstance.value.id)
    deletingInstance.value = null
  } catch {
    // error handled by store
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

onMounted(async () => {
  await store.fetchInstances()
  await loadInstanceStats()
})
</script>