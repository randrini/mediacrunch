export interface Instance {
  id: string;
  type: 'radarr' | 'sonarr' | 'plex';
  name: string;
  host: string;
  api_key: string;
  path_prefix: string;
  created_at: string;
  settings?: InstanceSettings;
}

export interface InstanceSettings {
  quality?: Record<string, number>;     // role → JPEG quality (1-100)
  max_width?: Record<string, number>;   // role → max pixel width (100-8000)
  min_saving_kb?: number;              // skip if savings < this (>= 0)
  // Note: backend uses int64, but JS number (float64) is safe for values < 2^53
  min_size_kb?: Record<string, number>; // role → minimum size threshold (KB)
  backup?: boolean;                     // create .bak before overwriting
  lock_plex?: boolean;                 // auto-lock Plex metadata before compress
}

export interface ImageInfo {
  role: string;
  path: string;
  size_bytes: number;
  width: number;
  height: number;
  format: string;
}

export interface MediaItem {
  id: string;
  instance_id: string;
  media_type: string;
  title: string;
  year: number | null;
  remote_id: string;
  path: string;
  images: ImageInfo[];
  total_size: number;
  original_size: number;
  total_images: number;
  compressed: boolean;
  locked: boolean | null;
  scanned_at: string;
  poster_size?: number;
  fanart_size?: number;
  clear_logo_size?: number;
  season_poster_size?: number;
  banner_size?: number;
}

export interface CompressionJob {
  id: string;
  instance_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  config: CompressConfig;
  total_items: number;
  processed_items: number;
  total_images: number;
  processed_images: number;
  saved_bytes: number;
  error_count: number;
  skip_count: number;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
}

export interface CompressConfig {
  quality: Record<string, number>;
  max_width: Record<string, number>;
  min_saving_kb: number;
  min_size_kb: Record<string, number>;
  backup: boolean;
  lock_plex: boolean;
}

export interface CompressionResult {
  id: string;
  job_id: string;
  media_item_id: string;
  image_path: string;
  role: string;
  original_bytes: number;
  new_bytes: number;
  saved_bytes: number;
  status: 'compressed' | 'skipped' | 'error';
  skip_reason: string;
  error: string;
  created_at: string;
}

export interface Stats {
  total_instances: number;
  total_items: number;
  total_size: number;
  total_savings: number;
  total_images?: number;
  compressed_items?: number;
}

export interface TestConnectionResponse {
  success: boolean;
  message: string;
  details?: {
    version: string;
    name: string;
  };
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface LogEntry {
  id: number;
  level: 'debug' | 'info' | 'warn' | 'error';
  source: string;
  instance_id: string;
  message: string;
  details: string;
  created_at: string;
}
