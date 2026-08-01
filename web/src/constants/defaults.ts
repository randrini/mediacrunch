export const ROLE_DEFAULTS: {
  quality: Record<string, number>
  max_width: Record<string, number>
  min_size_kb: Record<string, number>
} = {
  quality: { default: 82, poster: 82, fanart: 82, season_poster: 82, banner: 85, clearLogo: 90 },
  max_width: { default: 1920, poster: 1000, season_poster: 1000 },
  min_size_kb: { default: 30, poster: 50, fanart: 75, season_poster: 50, banner: 15, clearLogo: 10 },
}

export const MIN_SIZES: Record<string, number> = {
  default: 30,
  poster: 50,
  fanart: 75,
  season_poster: 50,
  banner: 15,
  clearLogo: 10,
}

export const ROLE_LABELS: Record<string, string> = {
  poster: 'Poster',
  fanart: 'Fanart',
  season_poster: 'Season Poster',
  banner: 'Banner',
  clearLogo: 'Clear Logo',
}

export const ROLE_ORDER = ['poster', 'fanart', 'season_poster', 'banner', 'clearLogo'] as const
