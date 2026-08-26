import type { Config } from 'tailwindcss'

const themeEmerald = {
  50: '#ECFDF5',
  100: '#D1FAE5',
  200: '#A7F3D0',
  300: '#6EE7B7',
  400: '#34D399',
  500: '#10B981',
  600: '#059669',
  700: '#047857',
  800: '#065F46',
  900: '#064E3B',
  950: '#022C22',
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
        sans: ['MapleUILatin', 'MapleUICJK'],
        mono: ['MapleUILatin', 'MapleUICJK'],
      },
      colors: {
        green: themeEmerald,
        emerald: themeEmerald,
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
