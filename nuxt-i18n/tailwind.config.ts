import type { Config } from 'tailwindcss'

const learnGripeGreen = {
  50: '#F5FFE8',
  100: '#E9FFD1',
  200: '#D9FFB2',
  300: '#C8FF8F',
  400: '#B5FF6D',
  500: '#B5FF6D',
  600: '#A6F05F',
  700: '#86C948',
  800: '#5F9138',
  900: '#365621',
  950: '#1B2E10',
} as const

export default {
  content: [
    './app/components/**/*.{vue,js,ts}',
    './app/layouts/**/*.{vue,js,ts}',
    './app/pages/**/*.{vue,js,ts}',
    './app/composables/**/*.{js,ts}',
    './app/plugins/**/*.{js,ts}',
    './app/App.{vue,js,ts}',
    './app/app.{vue,js,ts}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          'StorefrontSystem',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'sans-serif',
        ],
      },
      colors: {
        green: learnGripeGreen,
        emerald: learnGripeGreen,
        lime: learnGripeGreen,
        teal: learnGripeGreen,
      },
      screens: {
        // More granular mobile breakpoints for precise layout tuning.
        'phone-375': '375px',
        'phone-390': '390px',
        'phone-414': '414px',
        'phone-430': '430px',
        'tablet-768': '768px',
        'tablet-820': '820px',
      },
    },
  },
  plugins: [],
} satisfies Config
