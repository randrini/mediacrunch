/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        base: '#141210',
        ink: '#141210',
        surface: '#1d1a17',
        elevated: '#26221e',
        highlight: '#322c26',
        border: {
          DEFAULT: '#2e2924',
          strong: '#3d362f',
        },
        accent: {
          DEFAULT: '#e8a33d',
          hover: '#f0b45a',
          muted: '#3a2d18',
        },
        success: '#7fb069',
        warning: '#e3c14a',
        danger: '#d96c5f',
        text: {
          primary: '#ece5da',
          secondary: '#a89f92',
          tertiary: '#6f675c',
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
