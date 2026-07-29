<template>
  <div>
    <!-- Page Header -->
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-lg font-bold text-slate-100 tracking-tight">Settings</h1>
    </div>

    <!-- Instance Selector (when no :id param) -->
    <div v-if="!instanceId" class="mb-4">
      <div v-if="instancesStore.loading" class="text-center py-12">
        <svg class="animate-spin h-7 w-7 text-accent mx-auto" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        <p class="mt-2 text-slate-500 text-xs">Loading instances...</p>
      </div>

      <div v-else-if="instancesStore.instances.length === 0" class="text-center py-14">
        <svg class="w-14 h-14 mx-auto text-slate-700 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <h2 class="text-base font-semibold text-slate-300 mb-1.5">No instances yet</h2>
        <p class="text-slate-500 text-xs mb-4">Add an instance first to configure its settings.</p>
        <router-link
          to="/instances"
          class="inline-flex items-center px-4 py-2 text-xs font-medium rounded bg-accent text-base hover:bg-accent-hover transition-base"
        >
          <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          Add Instance
        </router-link>
      </div>

      <div v-else>
        <label class="text-xs font-medium text-slate-400 block mb-1.5">Select Instance</label>
        <select
          v-model="selectedInstanceId"
          @change="onInstanceSelected"
          class="bg-elevated text-slate-100 border border-white/[0.06] rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent w-full max-w-md font-mono"
        >
          <option value="" disabled>Choose an instance...</option>
          <option
            v-for="inst in instancesStore.instances"
            :key="inst.id"
            :value="inst.id"
          >
            {{ inst.name }} ({{ inst.type }})
          </option>
        </select>
      </div>
    </div>

    <!-- Settings Content -->
    <div v-if="instanceId && currentInstance">
      <!-- Loading -->
      <div v-if="store.loading && !store.settings[instanceId]" class="text-center py-12">
        <svg class="animate-spin h-7 w-7 text-accent mx-auto" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        <p class="mt-2 text-slate-500 text-xs">Loading settings...</p>
      </div>

      <div v-else @input="formDirty = true">
        <!-- Instance Info -->
        <div class="card-glass p-3 mb-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-sm font-semibold text-slate-100">{{ currentInstance.name }}</h2>
              <p class="text-xs text-slate-500 font-mono capitalize">{{ currentInstance.type }} &middot; {{ currentInstance.host }}</p>
            </div>
            <span
              class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium capitalize"
              :class="instanceTypeBadge"
            >
              {{ currentInstance.type }}
            </span>
          </div>
        </div>

        <!-- Toast Notifications -->
        <div v-if="toast" class="mb-4" :class="toast.type === 'success' ? 'bg-accent/10 border border-accent/20 text-accent rounded p-3' : 'bg-danger/10 border border-danger/20 text-danger rounded p-3'">
          <div class="flex items-center justify-between">
            <p class="text-xs">{{ toast.message }}</p>
            <button @click="toast = null" class="text-xs opacity-70 hover:opacity-100">&times;</button>
          </div>
        </div>

        <!-- Quality Settings -->
        <div class="card-glass p-4 mb-3">
          <h3 class="text-sm font-semibold text-slate-100 mb-3">Quality Settings</h3>
          <p class="text-xs text-slate-500 mb-3">JPEG quality for image roles (1-100). Higher values = better quality, larger files.</p>

          <!-- Default Quality -->
          <div class="mb-3">
            <label class="text-xs font-medium text-slate-400 block mb-1">Default Quality</label>
            <input
              type="number"
              min="1"
              max="100"
              v-model.number="form.quality.default"
              class="bg-elevated border border-white/[0.06] rounded px-3 py-1.5 text-slate-100 focus:ring-1 focus:ring-accent focus:border-transparent w-full max-w-xs font-mono text-sm"
            />
          </div>

          <!-- Role Overrides -->
          <div v-if="form.quality.overrides.length > 0">
            <label class="text-xs font-medium text-slate-400 block mb-1.5">Role Overrides</label>
            <div class="space-y-1.5">
              <div
                v-for="(override, index) in form.quality.overrides"
                :key="index"
                class="flex items-center space-x-2"
              >
                <select
                  v-model="override.role"
                  class="bg-elevated text-slate-100 border border-white/[0.06] rounded px-2.5 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent flex-1 max-w-[200px]"
                >
                  <option v-for="role in availableRoles" :key="role" :value="role">{{ role }}</option>
                </select>
                <input
                  type="number"
                  min="1"
                  max="100"
                  v-model.number="override.value"
                  class="bg-elevated border border-white/[0.06] rounded px-2.5 py-1.5 text-slate-100 focus:ring-1 focus:ring-accent focus:border-transparent w-20 font-mono text-sm"
                />
                <button
                  @click="removeQualityOverride(index)"
                  class="text-danger hover:text-red-400 transition-base p-1"
                  title="Remove override"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <button
            @click="addQualityOverride"
            class="mt-2 inline-flex items-center text-xs text-accent hover:text-accent-hover transition-base"
          >
            <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Add role override
          </button>
        </div>

        <!-- Max Width Settings -->
        <div class="card-glass p-4 mb-3">
          <h3 class="text-sm font-semibold text-slate-100 mb-3">Max Width Settings</h3>
          <p class="text-xs text-slate-500 mb-3">Maximum pixel width for image roles (100-8000). Images wider than this will be resized.</p>

          <!-- Default Max Width -->
          <div class="mb-3">
            <label class="text-xs font-medium text-slate-400 block mb-1">Default Max Width</label>
            <input
              type="number"
              min="100"
              max="8000"
              v-model.number="form.max_width.default"
              class="bg-elevated border border-white/[0.06] rounded px-3 py-1.5 text-slate-100 focus:ring-1 focus:ring-accent focus:border-transparent w-full max-w-xs font-mono text-sm"
            />
          </div>

          <!-- Role Overrides -->
          <div v-if="form.max_width.overrides.length > 0">
            <label class="text-xs font-medium text-slate-400 block mb-1.5">Role Overrides</label>
            <div class="space-y-1.5">
              <div
                v-for="(override, index) in form.max_width.overrides"
                :key="index"
                class="flex items-center space-x-2"
              >
                <select
                  v-model="override.role"
                  class="bg-elevated text-slate-100 border border-white/[0.06] rounded px-2.5 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent flex-1 max-w-[200px]"
                >
                  <option v-for="role in availableRoles" :key="role" :value="role">{{ role }}</option>
                </select>
                <input
                  type="number"
                  min="100"
                  max="8000"
                  v-model.number="override.value"
                  class="bg-elevated border border-white/[0.06] rounded px-2.5 py-1.5 text-slate-100 focus:ring-1 focus:ring-accent focus:border-transparent w-20 font-mono text-sm"
                />
                <button
                  @click="removeMaxWidthOverride(index)"
                  class="text-danger hover:text-red-400 transition-base p-1"
                  title="Remove override"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <button
            @click="addMaxWidthOverride"
            class="mt-2 inline-flex items-center text-xs text-accent hover:text-accent-hover transition-base"
          >
            <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Add role override
          </button>
        </div>

        <!-- General Settings -->
        <div class="card-glass p-4 mb-3">
          <h3 class="text-sm font-semibold text-slate-100 mb-3">General Settings</h3>

          <!-- Backup Toggle -->
          <div class="flex items-center justify-between py-2.5 border-b border-white/[0.04]">
            <div>
              <label class="text-xs font-medium text-slate-300">Backup original files</label>
              <p class="text-[11px] text-slate-500 mt-0.5">Create a .bak copy before overwriting images</p>
            </div>
            <button
              type="button"
              role="switch"
              :aria-checked="form.backup"
              @click="form.backup = !form.backup; formDirty = true"
              class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0 focus:outline-none focus:ring-1 focus:ring-accent focus:ring-offset-1 focus:ring-offset-base"
              :class="form.backup ? 'bg-accent' : 'bg-slate-600'"
            >
              <span
                class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform"
                :class="form.backup ? 'translate-x-[18px]' : 'translate-x-[3px]'"
              />
            </button>
          </div>

          <!-- Min Saving Threshold -->
          <div class="py-2.5 border-b border-white/[0.04]">
            <label class="text-xs font-medium text-slate-300 block mb-1">Minimum saving threshold (KB)</label>
            <p class="text-[11px] text-slate-500 mb-1.5">Skip compression if estimated savings are below this amount</p>
            <input
              type="number"
              min="0"
              v-model.number="form.min_saving_kb"
              class="bg-elevated border border-white/[0.06] rounded px-3 py-1.5 text-slate-100 focus:ring-1 focus:ring-accent focus:border-transparent w-full max-w-xs font-mono text-sm"
            />
          </div>

          <!-- Auto-lock Plex Metadata (Plex only) -->
          <div v-if="currentInstance.type === 'plex'" class="flex items-center justify-between py-2.5">
            <div>
              <label class="text-xs font-medium text-slate-300">Auto-lock Plex metadata</label>
              <p class="text-[11px] text-slate-500 mt-0.5">Automatically lock Plex metadata before compression to prevent Plex from overwriting compressed images</p>
            </div>
            <button
              type="button"
              role="switch"
              :aria-checked="form.lock_plex"
              @click="form.lock_plex = !form.lock_plex; formDirty = true"
              class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0 focus:outline-none focus:ring-1 focus:ring-accent focus:ring-offset-1 focus:ring-offset-base"
              :class="form.lock_plex ? 'bg-accent' : 'bg-slate-600'"
            >
              <span
                class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform"
                :class="form.lock_plex ? 'translate-x-[18px]' : 'translate-x-[3px]'"
              />
            </button>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex items-center justify-between pt-1 pb-6">
          <button
            @click="resetToDefaults"
            class="bg-white/[0.04] hover:bg-white/[0.08] text-slate-400 font-medium px-3 py-1.5 rounded transition-base text-xs"
          >
            Reset to Defaults
          </button>
          <button
            @click="saveSettings"
            :disabled="store.loading"
            class="bg-accent hover:bg-accent-hover disabled:bg-accent/50 disabled:cursor-not-allowed text-base font-medium px-3 py-1.5 rounded transition-base text-xs inline-flex items-center"
          >
            <svg v-if="store.loading" class="animate-spin h-3.5 w-3.5 mr-1.5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Save Settings
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useInstancesStore } from '../composables/useInstances'
import { useSettingsStore } from '../composables/useSettings'
import type { InstanceSettings } from '../types'

const route = useRoute()
const instancesStore = useInstancesStore()
const store = useSettingsStore()

const selectedInstanceId = ref('')
const toast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const formDirty = ref(false)

interface OverrideEntry {
  role: string
  value: number
}

interface SettingsForm {
  quality: { default: number; overrides: OverrideEntry[] }
  max_width: { default: number; overrides: OverrideEntry[] }
  backup: boolean
  min_saving_kb: number
  lock_plex: boolean
}

const form = ref<SettingsForm>({
  quality: { default: 80, overrides: [] },
  max_width: { default: 1920, overrides: [] },
  backup: false,
  min_saving_kb: 50,
  lock_plex: false,
})

// All possible roles
const allRoles = [
  'poster',
  'art',
  'clearLogo',
  'banner',
  'squareArt',
  'season_poster',
  'episode_thumb',
  'fanart',
] as const

// Roles relevant per instance type
const plexRoles = ['poster', 'art', 'clearLogo', 'banner', 'squareArt', 'season_poster', 'episode_thumb', 'fanart']
const radarrRoles = ['poster', 'fanart']
const sonarrRoles = ['poster', 'fanart', 'banner', 'clearLogo']

const availableRoles = computed(() => {
  if (!currentInstance.value) return allRoles as unknown as string[]
  switch (currentInstance.value.type) {
    case 'plex': return plexRoles
    case 'radarr': return radarrRoles
    case 'sonarr': return sonarrRoles
    default: return allRoles as unknown as string[]
  }
})

// Resolve instance ID from route param or selection
const instanceId = computed(() => {
  return (route.params.id as string) || selectedInstanceId.value || null
})

const currentInstance = computed(() => {
  if (!instanceId.value) return null
  return instancesStore.instances.find(i => i.id === instanceId.value) || null
})

const instanceTypeBadge = computed(() => {
  if (!currentInstance.value) return ''
  switch (currentInstance.value.type) {
    case 'plex': return 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
    case 'radarr': return 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
    case 'sonarr': return 'bg-accent/10 text-accent border border-accent/20'
    default: return 'bg-white/[0.04] text-slate-400 border border-white/[0.06]'
  }
})

function settingsToForm(s: InstanceSettings | undefined) {
  const qualityOverrides: OverrideEntry[] = []
  const maxWidthOverrides: OverrideEntry[] = []

  if (s?.quality) {
    for (const [role, value] of Object.entries(s.quality)) {
      if (role === 'default') continue
      qualityOverrides.push({ role, value })
    }
  }
  if (s?.max_width) {
    for (const [role, value] of Object.entries(s.max_width)) {
      if (role === 'default') continue
      maxWidthOverrides.push({ role, value })
    }
  }

  form.value = {
    quality: {
      default: s?.quality?.default ?? 80,
      overrides: qualityOverrides,
    },
    max_width: {
      default: s?.max_width?.default ?? 1920,
      overrides: maxWidthOverrides,
    },
    backup: s?.backup ?? false,
    min_saving_kb: s?.min_saving_kb ?? 50,
    lock_plex: s?.lock_plex ?? false,
  }
}

function formToSettings(): Partial<InstanceSettings> {
  const quality: Record<string, number> = { default: form.value.quality.default }
  for (const o of form.value.quality.overrides) {
    if (o.role) quality[o.role] = o.value
  }

  const max_width: Record<string, number> = { default: form.value.max_width.default }
  for (const o of form.value.max_width.overrides) {
    if (o.role) max_width[o.role] = o.value
  }

  return {
    quality,
    max_width,
    backup: form.value.backup,
    min_saving_kb: form.value.min_saving_kb,
    lock_plex: form.value.lock_plex,
  }
}

function addQualityOverride() {
  const usedRoles = form.value.quality.overrides.map(o => o.role)
  const available = availableRoles.value.filter(r => !usedRoles.includes(r))
  if (available.length === 0) return
  form.value.quality.overrides.push({ role: available[0], value: form.value.quality.default })
}

function removeQualityOverride(index: number) {
  form.value.quality.overrides.splice(index, 1)
}

function addMaxWidthOverride() {
  const usedRoles = form.value.max_width.overrides.map(o => o.role)
  const available = availableRoles.value.filter(r => !usedRoles.includes(r))
  if (available.length === 0) return
  form.value.max_width.overrides.push({ role: available[0], value: form.value.max_width.default })
}

function removeMaxWidthOverride(index: number) {
  form.value.max_width.overrides.splice(index, 1)
}

function resetToDefaults() {
  form.value = {
    quality: { default: 80, overrides: [] },
    max_width: { default: 1920, overrides: [] },
    backup: false,
    min_saving_kb: 50,
    lock_plex: false,
  }
  toast.value = { type: 'success', message: 'Form reset to defaults (not yet saved).' }
}

async function saveSettings() {
  if (!instanceId.value) return
  try {
    await store.saveSettings(instanceId.value, formToSettings())
    formDirty.value = false
    toast.value = { type: 'success', message: 'Settings saved successfully.' }
  } catch {
    toast.value = { type: 'error', message: store.error || 'Failed to save settings.' }
  }
}

function onInstanceSelected() {
  if (selectedInstanceId.value) {
    // Settings will be loaded by the watcher on instanceId
  }
}

// Load settings when instance ID changes
watch(instanceId, (newId) => {
  if (newId) {
    formDirty.value = false
    if (store.settings[newId]) {
      settingsToForm(store.settings[newId])
    }
    store.fetchSettings(newId)
  }
})

// Update form when settings are fetched
watch(() => store.settings, (newSettings) => {
  if (instanceId.value && newSettings[instanceId.value] && !formDirty.value) {
    settingsToForm(newSettings[instanceId.value])
  }
}, { deep: true })

onMounted(async () => {
  await instancesStore.fetchInstances()

  // If route has :id param, auto-select
  if (route.params.id) {
    selectedInstanceId.value = route.params.id as string
  }
})
</script>