# MediaCrunch — Architecture Specification

## Overview
MediaCrunch is a web tool for compressing metadata images in Radarr, Sonarr, and Plex installations. It provides a searchable, sortable UI to browse media items, view their image sizes, and compress them individually or in bulk.

## Stack
- **Backend**: Go 1.22+ with Gin HTTP framework, SQLite via go-sqlite3, disintegration/imaging for JPEG processing
- **Frontend**: Vue 3 + Composition API, Pinia state, Tailwind CSS, Heroicons, Axios
- **Deployment**: Docker Compose (single container: Go binary + Vue SPA served from /embed or static)

## Project Structure
```
/opt/scripts/resize/
├── ARCHITECTURE.md
├── go.mod
├── go.sum
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── db.go
│   │   └── migrations.go
│   ├── models/
│   │   └── models.go
│   ├── api/
│   │   ├── router.go
│   │   ├── instances.go
│   │   ├── media.go
│   │   ├── compress.go
│   │   └── stats.go
│   ├── scanner/
│   │   ├── scanner.go
│   │   ├── arr.go
│   │   └── plex.go
│   ├── compressor/
│   │   └── compressor.go
│   └── clients/
│       ├── radarr.go
│       ├── sonarr.go
│       └── plex.go
├── web/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── api/
│       │   └── index.ts
│       ├── types/
│       │   └── index.ts
│       ├── composables/
│       │   ├── useInstances.ts
│       │   ├── useMedia.ts
│       │   └── useCompress.ts
│       ├── components/
│       │   ├── Layout.vue
│       │   ├── InstanceForm.vue
│       │   ├── InstanceCard.vue
│       │   ├── MediaTable.vue
│       │   ├── MediaRow.vue
│       │   ├── CompressButton.vue
│       │   ├── SearchFilter.vue
│       │   └── StatsBar.vue
│       ├── views/
│       │   ├── DashboardView.vue
│       │   ├── InstancesView.vue
│       │   └── MediaView.vue
│       ├── router/
│       │   └── index.ts
│       └── assets/
│           └── main.css
├── Dockerfile
└── compose.yaml
```

## Database Schema (SQLite)

```sql
CREATE TABLE instances (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK(type IN ('radarr','sonarr','plex')),
  name TEXT NOT NULL,
  host TEXT NOT NULL,           -- e.g. http://radarr:7878 or http://plex:32400
  api_key TEXT NOT NULL,
  path_prefix TEXT NOT NULL,     -- filesystem path to config dir (e.g. /etc/komodo/stacks/arr/radarr)
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE media_items (
  id TEXT PRIMARY KEY,
  instance_id TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
  media_type TEXT NOT NULL,      -- movie, series, season, episode, collection
  title TEXT NOT NULL,
  year INTEGER,
  remote_id TEXT,                -- arr API id or plex rating key / guid
  path TEXT NOT NULL,            -- filesystem path to the metadata directory
  images TEXT NOT NULL DEFAULT '[]', -- JSON array of ImageInfo
  total_size INTEGER NOT NULL DEFAULT 0,
  total_images INTEGER NOT NULL DEFAULT 0,
  compressed INTEGER NOT NULL DEFAULT 0,
  locked INTEGER DEFAULT NULL,  -- plex only: whether metadata fields are locked
  scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(instance_id, media_type, remote_id)
);

-- ImageInfo JSON structure (stored in images column):
-- [
--   {
--     "role": "poster",        -- poster, art, clearLogo, banner, season_poster, episode_thumb, fanart
--     "path": "/path/to/file",
--     "size_bytes": 2048000,
--     "width": 2000,
--     "height": 3000,
--     "format": "jpeg"         -- jpeg, png
--   }
-- ]

CREATE TABLE compression_jobs (
  id TEXT PRIMARY KEY,
  instance_id TEXT NOT NULL REFERENCES instances(id),
  status TEXT NOT NULL DEFAULT 'pending', -- pending, running, completed, failed, cancelled
  config TEXT NOT NULL DEFAULT '{}',       -- JSON: {quality, max_width_by_role, backup, min_saving_kb}
  total_items INTEGER NOT NULL DEFAULT 0,
  processed_items INTEGER NOT NULL DEFAULT 0,
  saved_bytes INTEGER NOT NULL DEFAULT 0,
  error_count INTEGER NOT NULL DEFAULT 0,
  skip_count INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  started_at DATETIME,
  completed_at DATETIME
);

CREATE TABLE compression_results (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL REFERENCES compression_jobs(id) ON DELETE CASCADE,
  media_item_id TEXT NOT NULL REFERENCES media_items(id),
  image_path TEXT NOT NULL,
  role TEXT NOT NULL,
  original_bytes INTEGER NOT NULL,
  new_bytes INTEGER NOT NULL,
  saved_bytes INTEGER NOT NULL,
  status TEXT NOT NULL,         -- compressed, skipped, error
  skip_reason TEXT,
  error TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

### Instances
```
POST   /api/instances              Create instance (arr or plex)
GET    /api/instances              List all instances
GET    /api/instances/:id          Get instance detail
PUT    /api/instances/:id          Update instance
DELETE /api/instances/:id          Delete instance (cascades media + results)
POST   /api/instances/:id/scan     Trigger scan for instance
POST   /api/instances/:id/lock    Lock plex metadata fields (plex only)
```

### Media
```
GET    /api/instances/:id/media                List media items
  ?type=movie|series|season|episode|collection  Filter by type
  ?search=...                                    Search by title
  ?sort=total_size|title|year                   Sort field
  ?order=asc|desc                               Sort order
  ?page=1&per_page=50                           Pagination
  ?compressed=0|1                               Filter by compression status
  ?locked=0|1                                   Filter by lock status (plex)
```

### Compression
```
POST   /api/compress                 Start compression job
  Body: {
    "instance_id": "...",
    "media_item_ids": ["id1","id2",...] | null,   // null = all items
    "quality": {"poster": 72, "art": 75, "default": 80},
    "max_width": {"poster": 1000, "art": 1920, "default": 1920},
    "backup": true,
    "min_saving_kb": 50,
    "lock_plex": true            // auto-lock plex fields before compressing
  }
GET    /api/compress/:id             Get job status
POST   /api/compress/:id/cancel     Cancel running job
GET    /api/compress/:id/results    Get job results
```

### Stats
```
GET    /api/stats                    Overall stats
GET    /api/instances/:id/stats      Per-instance stats
```

### System
```
GET    /api/health                   Health check
```

## Scanner Logic

### Arr Scanner (Radarr/Sonarr)
1. Call arr API `/api/v3/movie` (Radarr) or `/api/v3/series` (Sonarr) to get list of items with titles
2. For each item, resolve `path_prefix + /MediaCover/` + item ID or title slug
3. Walk the MediaCover directory for fanart.jpg, poster.jpg, and any subfolder images (season posters for Sonarr)
4. Build media_items with image info

### Plex Scanner
1. Call Plex API `/library/sections` to get library list
2. For each library, call `/library/sections/{id}/all` to get items with metadata
3. For each item, compute the bundle path: `path_prefix + /Metadata/{Movies|TV Shows}/{guid_sha1[0]}/{guid_sha1}.bundle/Contents/_combined/`
4. Walk each bundle's _combined directory for all image types: posters/, art/, clearLogos/, banners/
5. Also scan `path_prefix + /Cache/PhotoTranscoder/` for cached derivatives
6. Build media_items with image info

### Plex GUID to Bundle Path
- Compute SHA1 of the item's `guid` field (e.g. `plex://movie/5d776b2f177e4e5c88e290f9f9e1c36d`)
- Bundle path = `{path_prefix}/Metadata/{MediaType}/{sha1[0]}/{sha1}.bundle/Contents/_combined/`

## Compressor Logic

### Image Compression
- Open image with disintegration/imaging
- Convert to RGB if needed (RGBA, palette, etc.)
- Resize if width > max_width for the role
- Save as JPEG with quality setting, optimize=true
- If saving < min_saving_kb, skip
- If backup=true, copy original to .bak before overwriting
- Update media_item images JSON with new sizes
- Mark media_item compressed=1

### Plex-Specific: Auto-Lock
Before compressing plex media items:
1. For each item, call Plex API `PUT /library/metadata/{ratingKey}` with `thumb.locked=1&art.locked=1`
2. After compression completes, mark the item as locked in the database

## Frontend Design

### Pages
1. **Dashboard** (`/`) — Overview: total instances, total items, total potential savings, recent jobs
2. **Instances** (`/instances`) — Add/edit/remove arr and plex instances, trigger scans
3. **Media List** (`/instances/:id/media`) — Searchable sortable filterable table of media items

### Media Table Columns
- Checkbox (for bulk select)
- Type icon (movie/series/season/episode)
- Title
- Year
- Images count
- Total size (with size breakdown by role on hover)
- Compressed status (badge)
- Locked status (plex, badge)
- Estimated savings
- Compress action (individual button)

### Features
- Search by title
- Filter by: type, compressed status, locked status, size range
- Sort by: title, year, total_size, image_count
- Select all / select visible
- Bulk compress button (appears when items selected)
- Compression progress bar (per-job, updated via polling or SSE)
- Instance selector dropdown (switch between instances)

## Docker Compose

```yaml
services:
  mediacrunch:
    build: .
    container_name: mediacrunch
    ports:
      - "8970:8080"
    volumes:
      - mediacrunch_data:/app/data
      - /etc/komodo/stacks/arr:/data/arr:ro  # Read-only access to arr configs
      - /etc/komodo/stacks/plex:/data/plex:ro  # Read-only initially; rw when compressing
    environment:
      - MC_DATA_DIR=/app/data
      - MC_PORT=8080
    restart: unless-stopped

volumes:
  mediacrunch_data:
```

Note: When compressing, the plex volume needs rw access. The tool should warn if the volume is read-only and compression is attempted.

## Configuration (Environment Variables)
- `MC_DATA_DIR` — Directory for SQLite database (default: `./data`)
- `MC_PORT` — HTTP port (default: `8080`)
- `MC_QUALITY_DEFAULT` — Default JPEG quality (default: `80`)
- `MC_MAX_WIDTH_DEFAULT` — Default max width (default: `1920`)
- `MC_MIN_SAVING_KB` — Minimum saving to compress (default: `50`)