# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Self-hosted media server administrators running Plex, Radarr, or Sonarr on home-lab or NAS hardware. They manage large media libraries (hundreds to thousands of titles) and want to reduce disk usage by compressing poster, fanart, and other media images without visible quality loss. They are technical enough to deploy Docker containers and configure API keys but want a tool that handles the details automatically.

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
- Deployed via Docker Compose pulling from GHCR

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
- Docker deployment via `ghcr.io/randrini/mediacrunch:latest`
- SQLite database, Go backend, Vue 3 + Tailwind frontend

## Brand Commitments

- Name: MediaCrunch
- Visual identity: warm charcoal base (#141210), amber/gold accent (#e8a33d), intentional surface hierarchy
- Typography: DM Sans (body), JetBrains Mono (data/numbers)
- Favicon: SVG with image compression iconography
- Tone: professional, data-dense, media-tool aesthetic — not SaaS dashboard

## Evidence on Hand

- Production deployment serving 4 real instances (2 Radarr, 1 Sonarr, 1 Plex)
- Active CI/CD pipeline with GitHub Actions, GHCR, and Docker Compose
- Real compression data showing savings across thousands of media items

## Product Principles

1. **Automate the details** — role-aware compression means users set quality targets, not pixel dimensions per image type
2. **Transparent operations** — every compression job shows what was compressed, what was skipped, and why
3. **Safe by default** — already-compressed items are skipped; re-compression requires explicit opt-in; backups available before any destructive operation
4. **Multi-instance from day one** — one dashboard for all Radarr, Sonarr, and Plex instances
5. **Data-dense, not data-overwhelming** — compact, scannable interfaces that respect screen real estate

## Accessibility & Inclusion

- Dark theme with warm charcoal backgrounds and sufficient contrast ratios
- Keyboard-navigable modals with focus trapping
- Semantic HTML with ARIA labels on interactive elements