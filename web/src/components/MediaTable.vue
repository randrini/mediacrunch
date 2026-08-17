<template>
  <div>
    <!-- Selection Bar -->
    <div
      v-if="someSelected"
      class="bg-accent-muted border border-accent/40 rounded-md px-4 py-2 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2"
    >
      <span class="text-sm text-text-primary">
        <template v-if="selectAllMode">
          All <span class="font-semibold">{{ totalItems }}</span> items selected
        </template>
        <template v-else>
          <span class="font-semibold">{{ selectedCount }}</span> item{{ selectedCount !== 1 ? 's' : '' }} selected
        </template>
      </span>
      <div class="flex items-center flex-wrap gap-3">
        <button
          v-if="!selectAllMode && selectedCount < totalItems"
          @click="$emit('selectAllAcrossPages')"
          class="text-sm text-accent hover:text-accent-hover transition-base"
        >
          Select all {{ totalItems }} items
        </button>
        <button
          v-else
          @click="$emit('clearSelection')"
          class="text-sm text-text-secondary hover:text-text-primary transition-base"
        >
          Clear selection
        </button>
        <button
          @click="$emit('compress', selectedIds)"
          class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded-md bg-accent text-ink hover:bg-accent-hover transition-base"
        >
          <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
          Compress Selected
        </button>
      </div>
    </div>

    <!-- Table (desktop) -->
    <div class="card-glass overflow-hidden hidden lg:block">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-border">
          <thead class="bg-elevated">
            <tr>
              <th class="px-2 py-1.5 text-left">
                <input
                  type="checkbox"
                  :checked="allSelected"
                  :indeterminate="someSelected && !allSelected"
                  @change="onSelectAll"
                  class="rounded-sm border-border-strong bg-elevated text-accent focus:ring-accent/50"
                />
              </th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Type</th>
              <th
                class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider cursor-pointer hover:text-text-primary"
                @click="sortBy('title')"
              >
                <span class="inline-flex items-center">
                  Title
                  <svg v-if="sortField === 'title'" class="w-3 h-3 ml-1 text-accent" :class="sortOrder === 'asc' ? '' : 'rotate-180'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-text-tertiary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Year</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Images</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Fanart</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Poster</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Clear Logo</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Season Poster</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Banner</th>
              <th
                class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider cursor-pointer hover:text-text-primary"
                @click="sortBy('original_size')"
              >
                <span class="inline-flex items-center">
                  Size
                  <svg v-if="sortField === 'original_size' && sortOrder === 'asc'" class="w-3 h-3 ml-1 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else-if="sortField === 'original_size'" class="w-3 h-3 ml-1 text-accent rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-text-tertiary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th
                class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider cursor-pointer hover:text-text-primary"
                @click="sortBy('total_size')"
              >
                <span class="inline-flex items-center">
                  After
                  <svg v-if="sortField === 'total_size' && sortOrder === 'asc'" class="w-3 h-3 ml-1 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else-if="sortField === 'total_size'" class="w-3 h-3 ml-1 text-accent rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else class="w-3 h-3 ml-1 text-text-tertiary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                </span>
              </th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Saved</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Compressed</th>
              <th v-if="instanceType === 'plex'" class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Locked</th>
              <th class="px-2 py-1.5 text-left text-[11px] font-medium text-text-tertiary uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <!-- Loading skeleton -->
            <tr v-if="loading">
              <td v-for="i in (instanceType === 'plex' ? 16 : 15)" :key="i" class="px-2 py-1">
                <div class="h-4 bg-elevated rounded animate-pulse" :style="{ width: i === 3 ? '60%' : i === 5 ? '40%' : '80%' }" />
              </td>
            </tr>
            <!-- Empty state -->
            <tr v-else-if="items.length === 0">
              <td :colspan="instanceType === 'plex' ? 16 : 15" class="px-3 py-12 text-center">
                <svg class="w-12 h-12 mx-auto text-text-tertiary mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <p class="text-text-secondary text-sm">No media items found.</p>
                <p class="text-text-tertiary text-xs mt-1">Try scanning your instance.</p>
              </td>
            </tr>
            <!-- Data rows -->
            <tr
              v-for="item in items"
              :key="item.id"
              class="hover:bg-highlight/50 transition-base"
              :class="{ 'bg-accent-muted': selectedIds.includes(item.id) }"
            >
              <td class="px-2 py-1">
                <input
                  type="checkbox"
                  :checked="selectedIds.includes(item.id)"
                  @click="onRowCheck(item, $event)"
                  class="rounded-sm border-border-strong bg-elevated text-accent focus:ring-accent/50"
                />
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <span class="inline-flex items-center justify-center text-text-secondary" :title="item.media_type">
                  <svg v-if="typeIcon(item.media_type) === 'movie'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="5" width="18" height="14" rx="1" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M7 5v14M17 5v14M3 9h4M3 15h4M17 9h4M17 15h4" />
                  </svg>
                  <svg v-else-if="typeIcon(item.media_type) === 'series'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <rect x="2" y="7" width="20" height="13" rx="2" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M17 2l-5 5-5-5" />
                  </svg>
                  <svg v-else-if="typeIcon(item.media_type) === 'season'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="4" width="18" height="18" rx="2" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M16 2v4M8 2v4M3 10h18" />
                  </svg>
                  <svg v-else-if="typeIcon(item.media_type) === 'episode'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <rect x="4" y="3" width="16" height="18" rx="1" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4 9h16M4 15h16" />
                  </svg>
                  <svg v-else-if="typeIcon(item.media_type) === 'collection'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M14 2v6h6" />
                  </svg>
                </span>
              </td>
              <td class="px-2 py-1">
                <div class="text-xs font-medium text-text-primary truncate max-w-xs" :title="item.title">
                  {{ item.title }}
                </div>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                {{ item.year ?? '-' }}
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <div class="relative group">
                  <span class="text-xs text-text-secondary cursor-help font-mono">{{ item.total_images }}</span>
                  <div
                    v-if="item.images && item.images.length > 0"
                    class="absolute bottom-full left-0 mb-2 hidden group-hover:block z-10"
                  >
                    <div class="bg-elevated border border-border-strong rounded-md shadow-lg px-2 py-1 text-xs text-text-primary whitespace-nowrap">
                      <div v-for="img in item.images" :key="img.role" class="py-0.5">
                        <span class="font-medium">{{ img.role }}:</span>
                        <span class="font-mono">{{ formatBytes(img.size_bytes) }} ({{ img.width }}x{{ img.height }})</span>
                      </div>
                    </div>
                  </div>
                </div>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                <span v-if="item.fanart_size">{{ formatBytes(item.fanart_size) }}</span>
                <span v-else class="text-text-tertiary">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                <span v-if="item.poster_size">{{ formatBytes(item.poster_size) }}</span>
                <span v-else class="text-text-tertiary">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                <span v-if="item.clear_logo_size">{{ formatBytes(item.clear_logo_size) }}</span>
                <span v-else class="text-text-tertiary">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                <span v-if="item.season_poster_size">{{ formatBytes(item.season_poster_size) }}</span>
                <span v-else class="text-text-tertiary">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap text-xs text-text-secondary font-mono">
                <span v-if="item.banner_size">{{ formatBytes(item.banner_size) }}</span>
                <span v-else class="text-text-tertiary">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <span :class="sizeBadgeClass(item.compressed && item.original_size > 0 ? item.original_size : item.total_size)">
                  {{ formatBytes(item.compressed && item.original_size > 0 ? item.original_size : item.total_size) }}
                </span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <span v-if="item.compressed && item.original_size > 0" class="size-badge-small">
                  {{ formatBytes(item.total_size) }}
                </span>
                <span v-else class="text-text-tertiary text-xs">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <span v-if="item.compressed && item.original_size > 0" class="text-xs font-medium text-accent font-mono">
                  −{{ formatBytes(item.original_size - item.total_size) }}
                  <span class="text-[11px] text-text-tertiary">(<span class="font-mono">{{ Math.round((1 - item.total_size / item.original_size) * 100) }}%</span>)</span>
                </span>
                <span v-else class="text-text-tertiary text-xs">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <span
                  v-if="item.compressed"
                  class="inline-flex items-center px-1.5 py-0.5 rounded-sm text-[11px] font-medium bg-accent-muted text-accent"
                >
                  <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  Compressed
                </span>
                <span v-else class="text-text-tertiary text-xs">—</span>
              </td>
              <td v-if="instanceType === 'plex'" class="px-2 py-1 whitespace-nowrap">
                <svg
                  v-if="item.locked"
                  class="w-4 h-4 text-warning"
                  fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
                  title="Locked"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span v-else class="text-text-tertiary text-xs">—</span>
              </td>
              <td class="px-2 py-1 whitespace-nowrap">
                <button
                  @click="$emit('compress', [item.id])"
                  :disabled="item.compressed"
                  class="inline-flex items-center px-1.5 py-0.5 text-[11px] font-medium rounded-sm transition-base"
                  :class="item.compressed ? 'bg-elevated text-text-tertiary cursor-not-allowed' : 'bg-accent-muted text-accent hover:bg-accent/20'"
                >
                  <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                  Compress
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Card list (mobile/tablet) -->
    <div class="lg:hidden space-y-2">
      <!-- Loading skeleton -->
      <div v-if="loading" class="space-y-2">
        <div v-for="i in 4" :key="i" class="card-glass p-3">
          <div class="h-4 bg-elevated rounded animate-pulse w-2/3 mb-2" />
          <div class="h-3 bg-elevated rounded animate-pulse w-1/3 mb-3" />
          <div class="h-3 bg-elevated rounded animate-pulse w-full" />
        </div>
      </div>

      <!-- Empty state -->
      <div v-else-if="items.length === 0" class="card-glass p-8 text-center">
        <svg class="w-12 h-12 mx-auto text-text-tertiary mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <p class="text-text-secondary text-sm">No media items found.</p>
        <p class="text-text-tertiary text-xs mt-1">Try scanning your instance.</p>
      </div>

      <!-- Data cards -->
      <div
        v-for="item in items"
        :key="item.id"
        class="card-glass p-2.5 relative"
        :class="{ 'bg-accent-muted border-accent/40': selectedIds.includes(item.id) }"
      >
        <!-- Selection toggle (top-right) -->
        <label class="absolute top-2.5 right-2.5 z-10 cursor-pointer" @click.stop>
          <input
            type="checkbox"
            :checked="selectedIds.includes(item.id)"
            @click="onRowCheck(item, $event)"
            class="rounded-sm border-border-strong bg-elevated text-accent focus:ring-accent/50"
          />
        </label>

        <!-- Title + type + year -->
        <div class="flex items-start gap-2 pr-8">
          <span class="inline-flex items-center justify-center text-text-secondary" :title="item.media_type">
            <svg v-if="typeIcon(item.media_type) === 'movie'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <rect x="3" y="5" width="18" height="14" rx="1" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 5v14M17 5v14M3 9h4M3 15h4M17 9h4M17 15h4" />
            </svg>
            <svg v-else-if="typeIcon(item.media_type) === 'series'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <rect x="2" y="7" width="20" height="13" rx="2" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 2l-5 5-5-5" />
            </svg>
            <svg v-else-if="typeIcon(item.media_type) === 'season'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 2v4M8 2v4M3 10h18" />
            </svg>
            <svg v-else-if="typeIcon(item.media_type) === 'episode'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <rect x="4" y="3" width="16" height="18" rx="1" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 9h16M4 15h16" />
            </svg>
            <svg v-else-if="typeIcon(item.media_type) === 'collection'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
            </svg>
            <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 2v6h6" />
            </svg>
          </span>
          <div class="min-w-0">
            <p class="text-xs font-semibold text-text-primary truncate" :title="item.title">{{ item.title }}</p>
            <p class="text-[11px] text-text-tertiary font-mono mt-0.5">{{ item.year ?? '-' }}</p>
          </div>
        </div>

        <!-- Key stats -->
        <div class="mt-3 grid grid-cols-3 gap-2 text-center">
          <div class="bg-highlight rounded-sm py-1 px-1">
            <p class="text-xs font-semibold text-text-primary font-mono tabular-nums">{{ item.total_images }}</p>
            <p class="text-[9px] text-text-tertiary uppercase tracking-wider">Images</p>
          </div>
          <div class="bg-highlight rounded-sm py-1 px-1">
            <p class="text-xs font-semibold text-text-primary font-mono tabular-nums">
              {{ formatBytes(item.compressed && item.original_size > 0 ? item.original_size : item.total_size) }}
            </p>
            <p class="text-[9px] text-text-tertiary uppercase tracking-wider">Size</p>
          </div>
          <div class="bg-highlight rounded-sm py-1 px-1">
            <p v-if="item.compressed && item.original_size > 0" class="text-xs font-semibold text-accent font-mono tabular-nums">
              −{{ formatBytes(item.original_size - item.total_size) }}
            </p>
            <p v-else class="text-xs font-semibold text-text-tertiary font-mono tabular-nums">—</p>
            <p class="text-[9px] text-text-tertiary uppercase tracking-wider">Saved</p>
          </div>
        </div>

        <!-- Compressed badge + action -->
        <div class="mt-3 flex items-center justify-between gap-2">
          <span
            v-if="item.compressed"
            class="inline-flex items-center px-1.5 py-0.5 rounded-sm text-[11px] font-medium bg-accent-muted text-accent"
          >
            <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            Compressed
          </span>
          <span v-else class="text-text-tertiary text-xs">Not compressed</span>
          <button
            @click="$emit('compress', [item.id])"
            :disabled="item.compressed"
            class="inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-sm transition-base"
            :class="item.compressed ? 'bg-elevated text-text-tertiary cursor-not-allowed' : 'bg-accent-muted text-accent hover:bg-accent/20'"
          >
            <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
            Compress
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mt-4">
      <p class="text-sm text-text-secondary">
        Showing page <span class="font-mono">{{ currentPage }}</span> of <span class="font-mono">{{ totalPages }}</span> (<span class="font-mono">{{ total }}</span> total)
      </p>
      <div class="flex items-center flex-wrap gap-2">
        <button
          @click="$emit('pageChange', currentPage - 1)"
          :disabled="currentPage <= 1"
          class="px-3 py-1.5 text-sm rounded-md bg-elevated text-text-secondary hover:bg-highlight disabled:opacity-50 disabled:cursor-not-allowed transition-base"
        >
          Previous
        </button>
        <button
          v-for="p in visiblePages"
          :key="p"
          @click="$emit('pageChange', p)"
          class="px-3 py-1.5 text-sm rounded-md transition-base font-mono"
          :class="p === currentPage ? 'bg-accent text-ink' : 'bg-elevated text-text-secondary hover:bg-highlight'"
        >
          {{ p }}
        </button>
        <button
          @click="$emit('pageChange', currentPage + 1)"
          :disabled="currentPage >= totalPages"
          class="px-3 py-1.5 text-sm rounded-md bg-elevated text-text-secondary hover:bg-highlight disabled:opacity-50 disabled:cursor-not-allowed transition-base"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MediaItem } from '../types'

const props = defineProps<{
  items: MediaItem[]
  loading: boolean
  instanceId: string
  instanceType: string
  selectedIds: string[]
  lastSelectedId: string | null
  selectAllMode: boolean
  currentPage: number
  totalPages: number
  total: number
  totalItems: number
  sortField: string
  sortOrder: string
}>()

const emit = defineEmits<{
  compress: [ids: string[]]
  selectAll: []
  selectAllAcrossPages: []
  clearSelection: []
  deselectAll: []
  toggleSelect: [id: string]
  selectRange: [fromId: string, toId: string]
  pageChange: [page: number]
  sort: [field: string, order: 'asc' | 'desc']
}>()

const selectedCount = computed(() => props.selectedIds.length)
const allSelected = computed(() =>
  props.selectAllMode || (props.items.length > 0 && props.selectedIds.length === props.items.length)
)
const someSelected = computed(() => props.selectAllMode || props.selectedIds.length > 0)

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = props.totalPages
  const current = props.currentPage
  let start = Math.max(1, current - 2)
  let end = Math.min(total, current + 2)
  if (end - start < 4) {
    if (start === 1) end = Math.min(total, start + 4)
    else start = Math.max(1, end - 4)
  }
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

function onSelectAll() {
  if (props.selectAllMode) {
    emit('clearSelection')
  } else if (allSelected.value) {
    emit('selectAllAcrossPages')
  } else {
    emit('selectAll')
  }
}

function onRowCheck(item: MediaItem, event: MouseEvent) {
  if (event.shiftKey && props.lastSelectedId) {
    event.preventDefault()
    emit('selectRange', props.lastSelectedId, item.id)
  } else {
    emit('toggleSelect', item.id)
  }
}

function sortBy(field: string) {
  // If clicking the same column, toggle direction; otherwise default to desc
  const order: 'asc' | 'desc' = props.sortField === field
    ? (props.sortOrder === 'asc' ? 'desc' : 'asc')
    : 'desc'
  emit('sort', field, order)
}

function typeIcon(type: string): string {
  switch (type) {
    case 'movie': return 'movie'
    case 'series': return 'series'
    case 'season': return 'season'
    case 'episode': return 'episode'
    case 'collection': return 'collection'
    default: return 'file'
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}

function sizeBadgeClass(bytes: number): string {
  if (bytes > 10 * 1024 * 1024) return 'size-badge-large' // > 10MB
  if (bytes > 1024 * 1024) return 'size-badge-medium' // > 1MB
  return 'size-badge-small'
}
</script>
