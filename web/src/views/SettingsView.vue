<template>
  <div>
    <!-- Page Header -->
    <div class="flex items-end justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-text-primary tracking-tight">Settings</h1>
        <p class="text-xs text-text-tertiary mt-1">
          {{ instancesStore.instances.length }} instance{{ instancesStore.instances.length === 1 ? '' : 's' }} configured
        </p>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-16">
      <svg class="animate-spin h-8 w-8 text-accent mx-auto" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
      <p class="mt-3 text-text-tertiary text-xs">Loading instance settings...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="instancesStore.instances.length === 0" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-text-tertiary mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
        <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
      <h2 class="text-[16px] font-semibold text-text-secondary mb-1.5">No instances yet</h2>
      <p class="text-text-tertiary text-xs mb-5">Add an instance first to configure its compression settings.</p>
      <router-link
        to="/instances"
        class="inline-flex items-center px-4 py-2 text-xs font-medium rounded-md bg-accent text-ink hover:bg-accent-hover transition-base"
      >
        <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        Add Instance
      </router-link>
    </div>

    <!-- Instance Cards Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="inst in instancesStore.instances"
        :key="inst.id"
        :ref="(el) => { if (el) cardRefs[inst.id] = el as HTMLElement }"
        class="card-glass flex flex-col"
        :class="{ 'ring-1 ring-accent/40': inst.id === instanceIdFromRoute }"
      >
        <!-- Card Header -->
        <div class="flex items-start justify-between gap-3 p-4 border-b border-border">
          <div class="flex items-start gap-3 min-w-0">
            <div
              class="w-9 h-9 rounded-md flex items-center justify-center text-sm font-bold font-mono shrink-0 border"
              :class="typeIconClass(inst.type)"
            >
              {{ typeIconText(inst.type) }}
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-text-primary truncate">{{ inst.name }}</h3>
              <p class="text-xs text-text-tertiary font-mono truncate">{{ inst.host }}</p>
            </div>
          </div>
          <span
            class="inline-flex items-center px-2 py-0.5 rounded-sm text-[10px] font-medium capitalize shrink-0 border"
            :class="typeBadgeClass(inst.type)"
          >
            {{ inst.type }}
          </span>
        </div>

        <!-- Card Body -->
        <div v-if="forms[inst.id]" class="p-4 space-y-5 flex-1" @input="dirty[inst.id] = true">
          <!-- Defaults -->
          <div class="space-y-3">
            <h4 class="text-xs font-semibold text-text-secondary uppercase tracking-wider">Defaults</h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <label class="block">
                <span class="text-[10px] text-text-tertiary block mb-1">Quality</span>
                <input
                  type="number"
                  min="1"
                  max="100"
                  v-model.number="forms[inst.id].quality_default"
                  class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                />
              </label>
              <label class="block">
                <span class="text-[10px] text-text-tertiary block mb-1">Max Width</span>
                <input
                  type="number"
                  min="100"
                  max="8000"
                  v-model.number="forms[inst.id].max_width_default"
                  class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                />
              </label>
              <label class="block">
                <span class="text-[10px] text-text-tertiary block mb-1">Min Saving (KB)</span>
                <input
                  type="number"
                  min="0"
                  v-model.number="forms[inst.id].min_saving_kb"
                  class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                />
              </label>
            </div>
          </div>

          <!-- Role Overrides -->
          <div>
            <button
              type="button"
              @click="expandedRoles[inst.id] = !expandedRoles[inst.id]"
              class="w-full flex items-center justify-between text-xs font-semibold text-text-secondary uppercase tracking-wider hover:text-text-primary transition-base py-1"
            >
              <span>Role Overrides</span>
              <svg
                class="w-4 h-4 text-text-tertiary transition-transform duration-200"
                :class="{ 'rotate-90': expandedRoles[inst.id] }"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </button>

            <div v-show="expandedRoles[inst.id]" class="mt-3 space-y-2">
              <div
                v-for="role in rolesForType(inst.type)"
                :key="role"
                class="bg-highlight border border-border rounded-md p-3"
              >
                <div class="flex items-center justify-between mb-2.5">
                  <span class="text-xs font-medium text-text-secondary">{{ ROLE_LABELS[role] || role }}</span>
                  <span class="inline-flex items-center text-[10px] text-text-tertiary font-mono">
                    <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                    </svg>
                    Skip if &lt; {{ forms[inst.id].roles[role].min_size_kb }} KB
                  </span>
                </div>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <label class="block">
                    <span class="text-[10px] text-text-tertiary block mb-1">Quality</span>
                    <input
                      type="number"
                      min="1"
                      max="100"
                      v-model.number="forms[inst.id].roles[role].quality"
                      class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                    />
                  </label>
                  <label class="block">
                    <span class="text-[10px] text-text-tertiary block mb-1">Max Width</span>
                    <input
                      type="number"
                      min="100"
                      max="8000"
                      v-model.number="forms[inst.id].roles[role].max_width"
                      class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                    />
                  </label>
                  <label class="block">
                    <span class="text-[10px] text-text-tertiary block mb-1">Min Size (KB)</span>
                    <input
                      type="number"
                      min="0"
                      v-model.number="forms[inst.id].roles[role].min_size_kb"
                      class="bg-elevated border border-border rounded-sm px-2 py-1.5 text-text-primary focus:ring-1 focus:ring-accent focus:border-transparent w-full font-mono text-xs"
                    />
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Options -->
          <div class="space-y-2">
            <h4 class="text-xs font-semibold text-text-secondary uppercase tracking-wider">Options</h4>

            <!-- Backup Toggle -->
            <div class="flex items-center justify-between py-2">
              <div class="pr-4">
                <label class="text-xs font-medium text-text-secondary">Create backup</label>
                <p class="text-[10px] text-text-tertiary mt-0.5">Create a .bak copy before overwriting images</p>
              </div>
              <button
                type="button"
                role="switch"
                :aria-checked="forms[inst.id].backup"
                @click="forms[inst.id].backup = !forms[inst.id].backup; dirty[inst.id] = true"
                class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0 focus:outline-none focus:ring-1 focus:ring-accent focus:ring-offset-1 focus:ring-offset-base"
                :class="forms[inst.id].backup ? 'bg-accent' : 'bg-highlight'"
              >
                <span
                  class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform"
                  :class="forms[inst.id].backup ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                />
              </button>
            </div>

            <!-- Lock Plex Toggle (Plex only) -->
            <div v-if="inst.type === 'plex'" class="flex items-center justify-between py-2">
              <div class="pr-4">
                <label class="text-xs font-medium text-text-secondary">Lock Plex metadata</label>
                <p class="text-[10px] text-text-tertiary mt-0.5">Auto-lock Plex metadata before compression</p>
              </div>
              <button
                type="button"
                role="switch"
                :aria-checked="forms[inst.id].lock_plex"
                @click="forms[inst.id].lock_plex = !forms[inst.id].lock_plex; dirty[inst.id] = true"
                class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0 focus:outline-none focus:ring-1 focus:ring-accent focus:ring-offset-1 focus:ring-offset-base"
                :class="forms[inst.id].lock_plex ? 'bg-accent' : 'bg-highlight'"
              >
                <span
                  class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform"
                  :class="forms[inst.id].lock_plex ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                />
              </button>
            </div>
          </div>

          <!-- Inline Toast -->
          <div
            v-if="toasts[inst.id]"
            class="rounded-sm p-2.5 text-xs"
            :class="toasts[inst.id]!.type === 'success' ? 'bg-accent-muted border border-accent/20 text-accent' : 'bg-danger/10 border border-danger/30 text-danger'"
          >
            <div class="flex items-center justify-between">
              <p>{{ toasts[inst.id]!.message }}</p>
              <button @click="toasts[inst.id] = null" class="opacity-70 hover:opacity-100 ml-2">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Card Footer -->
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 p-4 border-t border-border mt-auto">
          <button
            type="button"
            @click="resetInstance(inst.id)"
            class="w-full sm:w-auto justify-center bg-highlight hover:bg-elevated text-text-secondary font-medium px-3 py-1.5 rounded-md transition-base text-xs"
          >
            Reset to Defaults
          </button>
          <button
            type="button"
            @click="saveInstance(inst.id)"
            :disabled="saving[inst.id]"
            class="w-full sm:w-auto justify-center bg-accent hover:bg-accent-hover disabled:bg-accent/50 disabled:cursor-not-allowed text-ink font-medium px-3 py-1.5 rounded-md transition-base text-xs inline-flex items-center"
          >
            <svg v-if="saving[inst.id]" class="animate-spin h-3.5 w-3.5 mr-1.5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Save
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useInstancesStore } from '../composables/useInstances'
import { useSettingsStore } from '../composables/useSettings'
import type { Instance, InstanceSettings } from '../types'
import { ROLE_DEFAULTS } from '../constants/defaults'

const route = useRoute()
const instancesStore = useInstancesStore()
const settingsStore = useSettingsStore()

interface RoleForm {
  quality: number
  max_width: number
  min_size_kb: number
}

interface SettingsForm {
  quality_default: number
  max_width_default: number
  min_saving_kb: number
  backup: boolean
  lock_plex: boolean
  roles: Record<string, RoleForm>
}

interface Toast {
  type: 'success' | 'error'
  message: string
}

const ROLE_LABELS: Record<string, string> = {
  poster: 'Poster',
  fanart: 'Fanart',
  season_poster: 'Season Poster',
  banner: 'Banner',
  clearLogo: 'Clear Logo',
  episode_thumb: 'Episode Thumb',
  art: 'Art',
  squareArt: 'Square Art',
}

const PLEX_ROLES = ['poster', 'fanart', 'banner', 'clearLogo', 'season_poster', 'episode_thumb', 'art', 'squareArt']
const RADARR_ROLES = ['poster', 'fanart']
const SONARR_ROLES = ['poster', 'fanart', 'banner', 'clearLogo']

const loading = ref(false)
const forms = reactive<Record<string, SettingsForm>>({})
const dirty = reactive<Record<string, boolean>>({})
const toasts = reactive<Record<string, Toast | null>>({})
const saving = reactive<Record<string, boolean>>({})
const expandedRoles = reactive<Record<string, boolean>>({})
const cardRefs = reactive<Record<string, HTMLElement | null>>({})

const instanceIdFromRoute = computed(() => route.params.id as string | undefined)

function rolesForType(type: string): string[] {
  switch (type) {
    case 'radarr': return RADARR_ROLES
    case 'sonarr': return SONARR_ROLES
    case 'plex': return PLEX_ROLES
    default: return []
  }
}

function typeIconText(type: string): string {
  switch (type) {
    case 'radarr': return 'R'
    case 'sonarr': return 'S'
    case 'plex': return 'P'
    default: return type.charAt(0).toUpperCase()
  }
}

function typeIconClass(type: string): string {
  switch (type) {
    case 'radarr': return 'bg-accent-muted text-accent border-accent/40'
    case 'sonarr': return 'bg-accent-muted text-accent border-accent/40'
    case 'plex': return 'bg-accent-muted text-accent border-accent/40'
    default: return 'bg-highlight text-text-secondary border-border'
  }
}

function typeBadgeClass(type: string): string {
  switch (type) {
    case 'radarr': return 'bg-accent-muted text-accent border-accent/40'
    case 'sonarr': return 'bg-accent-muted text-accent border-accent/40'
    case 'plex': return 'bg-accent-muted text-accent border-accent/40'
    default: return 'bg-highlight text-text-secondary border-border'
  }
}

function getDefaultQuality(role: string): number {
  return ROLE_DEFAULTS.quality[role] ?? ROLE_DEFAULTS.quality.default
}

function getDefaultMaxWidth(role: string): number {
  return ROLE_DEFAULTS.max_width[role] ?? ROLE_DEFAULTS.max_width.default
}

function getDefaultMinSize(role: string): number {
  return ROLE_DEFAULTS.min_size_kb[role] ?? ROLE_DEFAULTS.min_size_kb.default
}

function mergeWithDefaults(settings: InstanceSettings | undefined, type: string): SettingsForm {
  const roles: Record<string, RoleForm> = {}
  const rolesList = rolesForType(type)

  for (const role of rolesList) {
    roles[role] = {
      quality: settings?.quality?.[role] ?? getDefaultQuality(role),
      max_width: settings?.max_width?.[role] ?? getDefaultMaxWidth(role),
      min_size_kb: settings?.min_size_kb?.[role] ?? getDefaultMinSize(role),
    }
  }

  return {
    quality_default: settings?.quality?.default ?? ROLE_DEFAULTS.quality.default,
    max_width_default: settings?.max_width?.default ?? ROLE_DEFAULTS.max_width.default,
    min_saving_kb: settings?.min_saving_kb ?? 50,
    backup: settings?.backup ?? false,
    lock_plex: settings?.lock_plex ?? false,
    roles,
  }
}

function formToSettings(form: SettingsForm): Partial<InstanceSettings> {
  const quality: Record<string, number> = { default: form.quality_default }
  const max_width: Record<string, number> = { default: form.max_width_default }
  const min_size_kb: Record<string, number> = { default: ROLE_DEFAULTS.min_size_kb.default }

  for (const [role, values] of Object.entries(form.roles)) {
    quality[role] = values.quality
    max_width[role] = values.max_width
    min_size_kb[role] = values.min_size_kb
  }

  return {
    quality,
    max_width,
    min_size_kb,
    min_saving_kb: form.min_saving_kb,
    backup: form.backup,
    lock_plex: form.lock_plex,
  }
}

function populateForm(instance: Instance) {
  const settings = settingsStore.settings[instance.id]
  forms[instance.id] = mergeWithDefaults(settings, instance.type)
}

async function loadAllSettings() {
  await Promise.all(
    instancesStore.instances.map(async (inst) => {
      try {
        await settingsStore.fetchSettings(inst.id)
      } catch {
        toasts[inst.id] = { type: 'error', message: 'Failed to load settings.' }
      } finally {
        populateForm(inst)
      }
    })
  )
}

async function saveInstance(id: string) {
  saving[id] = true
  toasts[id] = null

  try {
    const form = forms[id]
    if (!form) return

    await settingsStore.saveSettings(id, formToSettings(form))
    dirty[id] = false
    populateForm({ ...instancesStore.instances.find(i => i.id === id)!, settings: settingsStore.settings[id] })
    toasts[id] = { type: 'success', message: 'Settings saved successfully.' }
  } catch (e: any) {
    const message = e?.response?.data?.error || e?.message || settingsStore.error || 'Failed to save settings.'
    toasts[id] = { type: 'error', message }
  } finally {
    saving[id] = false
  }
}

function resetInstance(id: string) {
  const inst = instancesStore.instances.find(i => i.id === id)
  if (!inst) return
  forms[id] = mergeWithDefaults(undefined, inst.type)
  dirty[id] = true
  toasts[id] = { type: 'success', message: 'Reset to defaults (not yet saved).' }
}

// Update forms if settings are fetched/refreshed externally, but only for cards that haven't been edited
// Only update form when user hasn't made changes (dirty guard prevents infinite loop)
watch(() => settingsStore.settings, (newSettings) => {
  for (const [id, settings] of Object.entries(newSettings)) {
    if (dirty[id]) continue
    const inst = instancesStore.instances.find(i => i.id === id)
    if (inst) {
      forms[id] = mergeWithDefaults(settings, inst.type)
    }
  }
}, { deep: true })

onMounted(async () => {
  loading.value = true
  await instancesStore.fetchInstances()
  await loadAllSettings()
  loading.value = false

  if (instanceIdFromRoute.value && cardRefs[instanceIdFromRoute.value]) {
    await nextTick()
    cardRefs[instanceIdFromRoute.value]?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
})
</script>
