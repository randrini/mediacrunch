<template>
  <div class="min-h-screen bg-base flex flex-col">
    <!-- Skip to main content link -->
    <a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:bg-accent focus:text-white focus:px-4 focus:py-2 focus:rounded">
      Skip to main content
    </a>
    <!-- Top Nav -->
    <nav class="sticky top-0 z-40 bg-surface/80 backdrop-blur-md border-b border-white/[0.06]">
      <div class="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-12">
          <!-- Left: Brand + Nav Links -->
          <div class="flex items-center space-x-6">
            <router-link to="/" class="flex items-center space-x-2 group">
              <svg class="w-6 h-6 text-accent transition-base group-hover:scale-105" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <span class="text-xs font-bold text-slate-100 tracking-tight font-mono">MediaCrunch</span>
            </router-link>
            <div class="hidden md:flex space-x-1">
              <router-link
                to="/"
                class="px-2 py-0.5 rounded text-[11px] font-medium transition-base"
                :class="isActive('/') ? 'bg-accent/10 text-accent' : 'text-slate-400 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Dashboard
              </router-link>
              <router-link
                to="/instances"
                class="px-2 py-0.5 rounded text-[11px] font-medium transition-base"
                :class="isActive('/instances') ? 'bg-accent/10 text-accent' : 'text-slate-400 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Instances
              </router-link>
              <router-link
                to="/settings"
                class="px-2 py-0.5 rounded text-[11px] font-medium transition-base"
                :class="isActive('/settings') ? 'bg-accent/10 text-accent' : 'text-slate-400 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Settings
              </router-link>
              <router-link
                to="/logs"
                class="px-2 py-0.5 rounded text-[11px] font-medium transition-base"
                :class="isActive('/logs') ? 'bg-accent/10 text-accent' : 'text-slate-400 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Logs
              </router-link>
            </div>
          </div>

          <!-- Right: Instance Selector + Hamburger -->
          <div class="flex items-center space-x-2">
            <div class="hidden md:block">
              <select
                v-if="instances.length > 0"
                v-model="selectedInstanceId"
                @change="onInstanceChange"
                aria-label="Select instance"
                class="bg-elevated/80 text-slate-100 border border-white/[0.06] rounded px-2.5 py-1 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
              >
                <option value="" disabled>Select instance...</option>
                <option
                  v-for="inst in instances"
                  :key="inst.id"
                  :value="inst.id"
                >
                  {{ inst.name }} ({{ inst.type }})
                </option>
              </select>
            </div>
            <button
              @click="mobileMenuOpen = !mobileMenuOpen"
              class="md:hidden inline-flex items-center justify-center w-8 h-8 rounded text-slate-300 hover:text-slate-100 hover:bg-white/[0.06] transition-base"
              aria-label="Toggle navigation menu"
              :aria-expanded="mobileMenuOpen"
            >
              <svg v-if="!mobileMenuOpen" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Mobile Menu Panel -->
        <transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 -translate-y-2"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-2"
        >
          <div v-show="mobileMenuOpen" class="md:hidden pb-3 pt-1 border-t border-white/[0.06]">
            <nav class="flex flex-col space-y-1">
              <router-link
                to="/"
                @click="mobileMenuOpen = false"
                class="px-3 py-2 rounded text-sm font-medium transition-base"
                :class="isActive('/') ? 'bg-accent/10 text-accent' : 'text-slate-300 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Dashboard
              </router-link>
              <router-link
                to="/instances"
                @click="mobileMenuOpen = false"
                class="px-3 py-2 rounded text-sm font-medium transition-base"
                :class="isActive('/instances') ? 'bg-accent/10 text-accent' : 'text-slate-300 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Instances
              </router-link>
              <router-link
                to="/settings"
                @click="mobileMenuOpen = false"
                class="px-3 py-2 rounded text-sm font-medium transition-base"
                :class="isActive('/settings') ? 'bg-accent/10 text-accent' : 'text-slate-300 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Settings
              </router-link>
              <router-link
                to="/logs"
                @click="mobileMenuOpen = false"
                class="px-3 py-2 rounded text-sm font-medium transition-base"
                :class="isActive('/logs') ? 'bg-accent/10 text-accent' : 'text-slate-300 hover:text-slate-100 hover:bg-white/[0.04]'"
              >
                Logs
              </router-link>
            </nav>
            <div v-if="instances.length > 0" class="mt-2 px-3">
              <select
                v-model="selectedInstanceId"
                @change="onInstanceChange"
                aria-label="Select instance"
                class="w-full bg-elevated/80 text-slate-100 border border-white/[0.06] rounded px-2.5 py-1.5 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-accent focus:border-transparent"
              >
                <option value="" disabled>Select instance...</option>
                <option
                  v-for="inst in instances"
                  :key="inst.id"
                  :value="inst.id"
                >
                  {{ inst.name }} ({{ inst.type }})
                </option>
              </select>
            </div>
          </div>
        </transition>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="flex-1">
      <div class="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <slot />
      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-white/[0.06] py-2">
      <div class="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8">
        <p class="text-center text-[11px] text-slate-500 font-mono">
          MediaCrunch {{ version || '...' }} &mdash; Media image compression for Radarr, Sonarr &amp; Plex
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useInstancesStore } from '../composables/useInstances'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const store = useInstancesStore()

const { instances } = storeToRefs(store)
const selectedInstanceId = ref('')
const version = ref('')
const mobileMenuOpen = ref(false)

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function onInstanceChange() {
  if (selectedInstanceId.value) {
    router.push(`/instances/${selectedInstanceId.value}/media`)
  }
}

onMounted(async () => {
  store.fetchInstances()
  try {
    const res = await axios.get('/api/health')
    version.value = res.data?.version || ''
  } catch {
    version.value = ''
  }
})

watch(
  () => store.instances,
  (newInstances) => {
    if (newInstances.length > 0 && route.params.id) {
      selectedInstanceId.value = route.params.id as string
    }
  },
  { immediate: true }
)

watch(
  () => route.params.id,
  (id) => {
    selectedInstanceId.value = id ? (id as string) : ''
  }
)

watch(
  () => route.fullPath,
  () => {
    mobileMenuOpen.value = false
  }
)
</script>