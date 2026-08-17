<template>
  <teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center"
    >
      <!-- Backdrop -->
      <div
        class="absolute inset-0 bg-base/80 backdrop-blur-md"
        @click="$emit('cancel')"
      />
      <!-- Modal -->
      <div class="relative card-glass w-full max-w-lg mx-4 p-4">
        <h2 class="text-lg font-semibold text-text-primary mb-4">
          {{ isEdit ? 'Edit Instance' : 'Add Instance' }}
        </h2>

        <form @submit.prevent="handleSubmit" class="space-y-3">
          <!-- Name -->
          <div>
            <label class="block text-sm font-medium text-text-secondary mb-1">Name</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 placeholder-text-tertiary"
              placeholder="My Radarr"
            />
          </div>

          <!-- Type -->
          <div>
            <label class="block text-sm font-medium text-text-secondary mb-1">Type</label>
            <select
              v-model="form.type"
              required
              class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50"
            >
              <option value="" disabled>Select type...</option>
              <option value="radarr">Radarr</option>
              <option value="sonarr">Sonarr</option>
              <option value="plex">Plex</option>
            </select>
          </div>

          <!-- Host -->
          <div>
            <label class="block text-sm font-medium text-text-secondary mb-1">
              {{ isPlex ? 'Plex Server URL' : 'Host' }}
            </label>
            <input
              v-model="form.host"
              type="text"
              required
              class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 placeholder-text-tertiary"
              :placeholder="isPlex ? 'http://plex:32400' : 'http://radarr:7878'"
            />
            <p v-if="hostError" class="mt-1 text-xs text-danger">{{ hostError }}</p>
          </div>

          <!-- API Key / Plex Token -->
          <div v-if="!isPlex">
            <label class="block text-sm font-medium text-text-secondary mb-1">API Key</label>
            <div class="relative">
              <input
                v-model="form.api_key"
                :type="showKey ? 'text' : 'password'"
                required
                class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 placeholder-text-tertiary pr-10"
                placeholder="Your API key"
              />
              <button
                type="button"
                @click="showKey = !showKey"
                class="absolute inset-y-0 right-0 pr-3 flex items-center text-text-secondary hover:text-text-primary"
              >
                <svg v-if="!showKey" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Plex Auth Section -->
          <div v-if="isPlex" class="space-y-3">
            <label class="block text-sm font-medium text-text-secondary">Plex Token</label>

            <!-- Sign in with Plex button -->
            <button
              v-if="!plexToken && !plexManualEntry"
              type="button"
              @click="startPlexAuth"
              :disabled="plexAuthLoading"
              class="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-md bg-warning hover:bg-warning/80 text-ink font-medium text-sm transition-base disabled:opacity-50"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 15l-1.41-1.41L14.17 11H8V9h8v2l-5 6z"/>
              </svg>
              {{ plexAuthLoading ? 'Waiting for Plex...' : 'Sign in with Plex' }}
            </button>

            <!-- Plex auth status -->
            <div v-if="plexAuthLoading" class="flex items-center gap-2 text-sm text-warning">
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              Authenticating with plex.tv — check your browser window
            </div>

            <!-- Token acquired -->
            <div v-if="plexToken" class="flex items-center gap-2 p-2 rounded-md bg-accent-muted border border-accent/40">
              <svg class="w-4 h-4 text-accent shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
              <span class="text-sm text-accent">Connected as <strong>{{ plexUsername }}</strong></span>
              <button type="button" @click="clearPlexToken" class="ml-auto text-xs text-text-secondary hover:text-danger">Disconnect</button>
            </div>

            <!-- Manual token entry toggle -->
            <button
              v-if="!plexToken && !plexAuthLoading"
              type="button"
              @click="plexManualEntry = !plexManualEntry"
              class="text-xs text-text-secondary hover:text-text-primary underline"
            >
              {{ plexManualEntry ? 'Hide manual entry' : 'Or enter token manually' }}
            </button>

            <!-- Manual token input -->
            <div v-if="plexManualEntry && !plexToken">
              <input
                v-model="form.api_key"
                type="text"
                class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 placeholder-text-tertiary"
                placeholder="X-Plex-Token"
              />
              <p class="mt-1 text-xs text-text-tertiary">Find your token at plex.tv → Account → Authorized devices</p>
            </div>
          </div>

          <!-- Path Prefix -->
          <div>
            <label class="block text-sm font-medium text-text-secondary mb-1">Path Prefix</label>
            <input
              v-model="form.path_prefix"
              type="text"
              required
              class="w-full bg-elevated border border-border rounded-md px-3 py-2 text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 placeholder-text-tertiary"
              :placeholder="isPlex ? '/etc/komodo/stacks/plex' : '/etc/komodo/stacks/arr/radarr'"
            />
          </div>

          <!-- Test Connection -->
          <div>
            <button
              type="button"
              @click="handleTestConnection"
              :disabled="!canTest || testing"
              class="w-full px-3 py-1.5 text-sm font-medium rounded-md transition-base"
              :class="testStatusClass"
            >
              <span v-if="testing" class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Testing...
              </span>
              <span v-else-if="testResult !== null && testResult.success" class="inline-flex items-center">
                <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                Connection OK
              </span>
              <span v-else-if="testResult !== null && !testResult.success" class="inline-flex items-center">
                <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
                Connection Failed
              </span>
              <span v-else>Test Connection</span>
            </button>
            <p v-if="testResult && !testResult.success" class="mt-1 text-xs text-danger">{{ testResult.message }}</p>
            <p v-if="testResult && testResult.success && testResult.details" class="mt-1 text-xs text-accent">
              {{ testResult.details.name }} v{{ testResult.details.version }}
            </p>
          </div>

          <!-- Actions -->
          <div class="flex justify-end space-x-3 pt-2">
            <button
              type="button"
              @click="$emit('cancel')"
              class="px-3 py-1.5 text-sm font-medium text-text-secondary bg-elevated rounded-md hover:bg-highlight transition-base"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-3 py-1.5 text-sm font-medium text-ink bg-accent rounded-md hover:bg-accent-hover transition-base disabled:opacity-50"
              :disabled="!isValid"
            >
              {{ isEdit ? 'Save Changes' : 'Add Instance' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Instance, TestConnectionResponse } from '../types'
import { testConnection, createPlexPIN, checkPlexPIN } from '../api'

const props = defineProps<{
  visible: boolean
  instance?: Instance | null
}>()

const emit = defineEmits<{
  save: [data: {
    type: 'radarr' | 'sonarr' | 'plex'
    name: string
    host: string
    api_key: string
    path_prefix: string
  }]
  cancel: []
}>()

const showKey = ref(false)
const hostError = ref('')
const testing = ref(false)
const testResult = ref<TestConnectionResponse | null>(null)

const isPlex = computed(() => form.value.type === 'plex')
const plexManualEntry = ref(false)
const plexAuthLoading = ref(false)
const plexToken = ref('')
const plexUsername = ref('')

const form = ref<{
  name: string
  type: 'radarr' | 'sonarr' | 'plex' | ''
  host: string
  api_key: string
  path_prefix: string
}>({
  name: '',
  type: '',
  host: '',
  api_key: '',
  path_prefix: '',
})

const isEdit = computed(() => !!props.instance)

const canTest = computed(() => {
  if (isPlex.value) {
    return form.value.type && form.value.host && (plexToken.value || form.value.api_key) && !hostError.value
  }
  return form.value.type && form.value.host && form.value.api_key && !hostError.value
})

const testStatusClass = computed(() => {
  if (testing.value) return 'bg-elevated text-text-secondary cursor-wait'
  if (testResult.value === null) return 'bg-elevated text-text-secondary hover:bg-highlight'
  if (testResult.value.success) return 'bg-accent-muted text-accent'
  return 'bg-danger/20 text-danger'
})

const isValid = computed(() => {
  if (isPlex.value) {
    return (
      form.value.name &&
      form.value.type &&
      form.value.host &&
      (plexToken.value || form.value.api_key) &&
      form.value.path_prefix &&
      !hostError.value
    )
  }
  return (
    form.value.name &&
    form.value.type &&
    form.value.host &&
    form.value.api_key &&
    form.value.path_prefix &&
    !hostError.value
  )
})

watch(
  () => props.visible,
  (visible) => {
    if (visible && props.instance) {
      form.value = {
        name: props.instance.name,
        type: props.instance.type,
        host: props.instance.host,
        api_key: props.instance.api_key,
        path_prefix: props.instance.path_prefix,
      }
    } else if (visible) {
      form.value = {
        name: '',
        type: '',
        host: '',
        api_key: '',
        path_prefix: '',
      }
    }
    hostError.value = ''
    testResult.value = null
  }
)

watch(
  () => form.value.host,
  (host) => {
    if (host && !host.startsWith('http://') && !host.startsWith('https://')) {
      hostError.value = 'Host must start with http:// or https://'
    } else {
      hostError.value = ''
    }
  }
)

// When type changes to plex, clear the api_key
watch(
  () => form.value.type,
  (newType) => {
    if (newType === 'plex') {
      plexToken.value = ''
      plexUsername.value = ''
      plexManualEntry.value = false
      plexAuthLoading.value = false
    }
  }
)

async function handleTestConnection() {
  if (!canTest.value) return
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testConnection({
      type: form.value.type!,
      host: form.value.host!,
      api_key: isPlex.value ? (plexToken.value || form.value.api_key!) : form.value.api_key!,
    })
  } catch (e: any) {
    testResult.value = {
      success: false,
      message: e?.response?.data?.error || e?.message || 'Request failed',
    }
  } finally {
    testing.value = false
  }
}

async function startPlexAuth() {
  plexAuthLoading.value = true
  plexToken.value = ''
  plexUsername.value = ''

  try {
    // 1. Create a PIN
    const pinData = await createPlexPIN()

    // 2. Open the auth URL in a new tab
    window.open(pinData.auth_url, '_blank', 'width=800,height=600')

    // 3. Poll for PIN status
    const pollInterval = setInterval(async () => {
      try {
        const status = await checkPlexPIN(pinData.pin_id)
        if (status.claimed && status.token) {
          clearInterval(pollInterval)
          plexToken.value = status.token
          plexUsername.value = status.username || 'Plex User'
          form.value.api_key = status.token
          plexAuthLoading.value = false
        }
      } catch {
        // Keep polling
      }
    }, 2000)

    // Timeout after 5 minutes
    setTimeout(() => {
      clearInterval(pollInterval)
      plexAuthLoading.value = false
    }, 300000)

  } catch (e: any) {
    plexAuthLoading.value = false
    // Fall back to manual entry
    plexManualEntry.value = true
  }
}

function clearPlexToken() {
  plexToken.value = ''
  plexUsername.value = ''
  form.value.api_key = ''
}

function handleSubmit() {
  if (!isValid.value) return
  emit('save', {
    type: form.value.type as 'radarr' | 'sonarr' | 'plex',
    name: form.value.name,
    host: form.value.host,
    api_key: form.value.api_key,
    path_prefix: form.value.path_prefix,
  })
}
</script>