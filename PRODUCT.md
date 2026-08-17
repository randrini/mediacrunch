# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Go 1.26+ backend (Gin HTTP framework, SQLite via go-sqlite3, disintegration/imaging for JPEG processing). Vue 3 + Composition API, Pinia state, Tailwind CSS, Heroicons, Axios frontend. Single-container Docker deployment serving the Vue SPA from the Go binary.

## Users

Self-hosted media server administrators running Plex, Radarr, or Sonarr on home-lab or NAS hardware. They manage large media libraries (hundreds to thousands of titles) and want to reduce disk usage by compressing poster, fanart, and other media images without visible quality loss. They are technical enough to deploy Docker containers and configure API keys but want a tool that handles the details automatically. A hosted/cloud offering may serve the same audience later, but the primary user today is the self-hoster.

## Product Purpose

MediaCrunch reduces storage consumption by compressing media images (posters, fanart, clear logos, season posters, banners) across Radarr, Sonarr, and Plex instances. Success means significant disk savings with no perceptible quality degradation, handled automatically after a one-time setup.

## Positioning

MediaCrunch is the only tool that understands media image roles (poster, fanart, clearLogo, season_poster, banner) and applies role-specific compression settings automatically, while managing batch compression across multiple instances from a single dashboard. Manual tools like ImageMagick or Squoosh compress individual files; MediaCrunch scans entire libraries, detects the right images per role, and compresses them in bulk with per-role quality and size thresholds.

## Operating Context

- Users connect Radarr, Sonarr, and/or Plex instances via API keys
- Scans discover media items and their associated images by role
- Compression jobs run asynchronously with progress tracking
- Results show per-image savings (original size, compressed size, percentage saved)
- Backup files (.bak) can be created before compression and cleaned up later
- Re-compression prevention: already-compressed items are skipped by default
- Deployed via Docker Compose pulling from GHCR (`ghcr.io/randrini/mediacrunch:latest`)
- Arr and Plex volumes mounted read-write for in-place compression
- Health check via `/api/health`

## Capabilities and Constraints

- Supports Radarr, Sonarr, and Plex instance types
- Per-role compression settings (quality, max width, min size threshold, min saving threshold)
- Batch compression with select-all across pages
- Lock Plex metadata fields to prevent overwrites
- Backup creation and cleanup for .bak files
- Scan/re-scan instances to detect new or changed media
- Async job system with progress tracking and cancellation
- Role-specific defaults: poster/fanart quality 82, banner 85, clearLogo 90
- Sync/lock toggle for role-specific settings vs global defaults
- Logs view: shipped feature for reviewing compression job history and output
- Settings view: shipped feature for configuring compression defaults and application behavior
- Docker deployment via `ghcr.io/randrini/mediacrunch:latest`, port 8970→8080
- SQLite database, Go backend, Vue 3 + Tailwind frontend
- Environment configuration: `MC_DATA_DIR`, `MC_PORT`, `MC_QUALITY_DEFAULT`, `MC_MAX_WIDTH_DEFAULT`, `MC_MIN_SAVING_KB`

## Brand Commitments

- Name: MediaCrunch
- Tone: professional, data-dense, media-tool aesthetic — not SaaS dashboard
- Design direction (user-stated): avoid AI slop; use modern design. Specific visual world (palette, typography, surfaces) is established in DESIGN.md, not here.

## Evidence on Hand

- Production deployment serving 4 real instances (2 Radarr, 1 Sonarr, 1 Plex)
- Active CI/CD pipeline with GitHub Actions, GHCR, and Docker Compose
- Real compression data showing savings across thousands of media items
- `compress_mediacover.py` at repo root is the legacy CLI predecessor (Radarr/Sonarr MediaCover compressor, Pillow-based); MediaCrunch is the Go web rewrite that replaced it. Kept in repo as history.

## Product Principles

1. **Automate the details** — role-aware compression means users set quality targets, not pixel dimensions per image type
2. **Transparent operations** — every compression job shows what was compressed, what was skipped, and why
3. **Safe by default** — already-compressed items are skipped; re-compression requires explicit opt-in; backups available before any destructive operation
4. **Multi-instance from day one** — one dashboard for all Radarr, Sonarr, and Plex instances
5. **Data-dense, not data-overwhelming** — compact, scannable interfaces that respect screen real estate

## Accessibility & Inclusion

- Dark theme with sufficient contrast ratios
- Keyboard-navigable modals with focus trapping
- Semantic HTML with ARIA labels on interactive elements