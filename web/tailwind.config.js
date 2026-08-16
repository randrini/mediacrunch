/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        base: '#0a0a0a',
        ink: '#f5f5f5',
        surface: '#111111',
        elevated: '#1a1a1a',
        highlight: '#242424',
        border: {
          DEFAULT: '#222222',
          strong: '#333333',
        },
        accent: {
          DEFAULT: '#10b981',
          hover: '#34d399',
          muted: '#10b9811a',
        },
        success: '#22c55e',
        warning: '#f59e0b',
        danger: '#ef4444',
        text: {
          primary: '#f5f5f5',
          secondary: '#a3a3a3',
          tertiary: '#525252',
        },
      },
      fontFamily: {
        sans: ['DM Sans', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'pulse-soft': 'pulse-soft 2s ease-in-out infinite',
      },
      keyframes: {
        'pulse-soft': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.6' },
        },
      },
    },
  },
  plugins: [],
}
