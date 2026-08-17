/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // The Darkroom — warm near-black ground
        base: '#0d0b0a',
        // Ink — near-black text on accent/success/warning surfaces (avoids the
        // `text-base` collision with Tailwind's built-in font-size utility)
        ink: '#0d0b0a',
        surface: '#16120e',
        elevated: '#1f1a14',
        highlight: '#2a2218',
        border: {
          DEFAULT: '#2a2218',
          strong: '#3a3024',
        },
        // Safelight — the single red accent that illuminates the darkroom
        accent: {
          DEFAULT: '#c8412a',
          hover: '#e85d3a',
          muted: 'rgba(200, 65, 42, 0.12)',
        },
        // Zone system status — tonal ramp, not generic semantic colors
        success: '#8a9a6b',
        warning: '#c4943a',
        danger: '#a02828',
        // Paper — warm photographic white through warm gray
        text: {
          primary: '#f5f0e8',
          secondary: '#9a9082',
          tertiary: '#6b6358',
        },
      },
      fontFamily: {
        sans: ['Saira', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'pulse-soft': 'pulse-soft 2s ease-in-out infinite',
        'cascade': 'cascade 0.4s ease-out',
        'expose': 'expose 1.5s ease-in-out',
      },
      keyframes: {
        'pulse-soft': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.6' },
        },
        'cascade': {
          '0%': { transform: 'translateY(-100%)', opacity: '0' },
          '50%': { opacity: '1' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        'expose': {
          '0%': { filter: 'brightness(0.3) sepia(1) hue-rotate(-20deg)', opacity: '0' },
          '100%': { filter: 'brightness(1) sepia(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}