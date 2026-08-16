import axios from 'axios'
import type {
  Instance,
  InstanceSettings,
  MediaItem,
  CompressionJob,
  CompressionResult,
  Stats,
  PaginatedResponse,
  TestConnectionResponse,
  LogEntry,
  CleanupResult,
  CleanupAllResult,
} from '../types'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// --- Instances ---
export function getInstances(): Promise<Instance[]> {
  return api.get('/instances').then((r) => r.data)
}

export function getInstance(id: string): Promise<Instance> {
  return api.get(`/instances/${id}`).then((r) => r.data)
}

export function createInstance(data: {
  type: 'radarr' | 'sonarr' | 'plex'
  name: string
  host: string
  api_key: string
  path_prefix: string
}): Promise<Instance> {
  return api.post('/instances', data).then((r) => r.data)
}

export function updateInstance(id: string, data: Partial<Instance>): Promise<Instance> {
  return api.put(`/instances/${id}`, data).then((r) => r.data)
}

export function deleteInstance(id: string): Promise<void> {
  return api.delete(`/instances/${id}`).then((r) => r.data)
}

export function scanInstance(id: string): Promise<void> {
  return api.post(`/instances/${id}/scan`).then((r) => r.data)
}

export function lockInstance(id: string): Promise<void> {
  return api.post(`/instances/${id}/lock`).then((r) => r.data)
}

export function testConnection(data: {
  type: string
  host: string
  api_key: string
}): Promise<TestConnectionResponse> {
  return api.post('/instances/test', data).then((r) => r.data)
}

// --- Plex Auth ---
export interface PlexPINResponse {
  pin_id: number
  code: string
  auth_url: string
  client_id: string
}

export interface PlexPINStatusResponse {
  token?: string
  username?: string
  claimed: boolean
}

export function createPlexPIN(): Promise<PlexPINResponse> {
  return api.post('/plex/pin').then(r => r.data)
}

export function checkPlexPIN(pinId: number): Promise<PlexPINStatusResponse> {
  return api.get(`/plex/pin/${pinId}`).then(r => r.data)
}

// --- Media ---
export interface MediaQueryParams {
  type?: string
  search?: string
  sort?: string
  order?: string
  page?: number
  per_page?: number
  compressed?: '0' | '1'
  locked?: '0' | '1'
}

export function getMediaItems(
  instanceId: string,
  params?: MediaQueryParams
): Promise<PaginatedResponse<MediaItem>> {
  return api.get(`/instances/${instanceId}/media`, { params }).then((r) => r.data)
}

// --- Compression ---
export interface CompressRequest {
  instance_id: string
  media_item_ids?: string[] | null
  quality?: Record<string, number>
  max_width?: Record<string, number>
  min_saving_kb?: number
  min_size_kb?: Record<string, number>
  backup?: boolean
  lock_plex?: boolean
  recompress?: boolean
}

export function compressItems(data: CompressRequest): Promise<CompressionJob> {
  return api.post('/compress', data).then((r) => r.data)
}

export function getCompressJob(id: string): Promise<CompressionJob> {
  return api.get(`/compress/${id}`).then((r) => r.data)
}

export function cancelCompressJob(id: string): Promise<void> {
  return api.post(`/compress/${id}/cancel`).then((r) => r.data)
}

export function getCompressResults(id: string): Promise<CompressionResult[]> {
  return api.get(`/compress/${id}/results`).then((r) => r.data)
}

export function getRecentJobs(limit = 20, instanceId?: string): Promise<CompressionJob[]> {
  const params: Record<string, any> = { limit }
  if (instanceId) params.instance_id = instanceId
  return api.get('/compress', { params }).then((r) => r.data)
}

// --- Settings ---
export function getInstanceSettings(id: string): Promise<InstanceSettings> {
  return api.get(`/instances/${id}/settings`).then(r => r.data)
}

export function updateInstanceSettings(id: string, data: Partial<InstanceSettings>): Promise<InstanceSettings> {
  return api.put(`/instances/${id}/settings`, data).then(r => r.data)
}

// --- Stats ---
export function getStats(): Promise<Stats> {
  return api.get('/stats').then((r) => r.data)
}

export function getInstanceStats(id: string): Promise<Stats> {
  return api.get(`/instances/${id}/stats`).then((r) => r.data)
}

// --- Logs ---
export interface LogQueryParams {
  level?: string;
  source?: string;
  limit?: number;
  offset?: number;
  search?: string;
  instance_id?: string;
}

export interface LogsResponse {
  logs: LogEntry[];
  total: number;
  limit: number;
  offset: number;
}

export function getLogs(params?: LogQueryParams): Promise<LogsResponse> {
  return api.get('/logs', { params }).then(r => r.data)
}

export function clearLogs(): Promise<{ message: string; deleted: number }> {
  return api.delete('/logs').then(r => r.data)
}

// --- Backup Cleanup ---
export function cleanupBackups(instanceId: string, dryRun: boolean = false): Promise<CleanupResult> {
  return api.post(`/instances/${instanceId}/cleanup-backups`, { dry_run: dryRun }).then(r => r.data)
}

export function cleanupAllBackups(dryRun: boolean = false): Promise<CleanupAllResult> {
  return api.post('/cleanup-backups', { dry_run: dryRun }).then(r => r.data)
}
