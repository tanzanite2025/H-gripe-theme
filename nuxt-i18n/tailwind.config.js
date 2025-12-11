/** @type {import('tailwindcss').Config} */
module.exports = {
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
      screens: {
        // 更细粒度的移动端宽度段，用于精细调整布局
        // iPhone SE / 8 等 375 宽（同时涵盖 375x667、375x812 等）
        'phone-375': '375px',
        // 390 宽（如部分较新 iPhone），用于替代 min-max 组合防止警告
        'phone-390': '390px',
        // 414 宽（例如 iPhone 11 / 12 Pro）
        'phone-414': '414px',
        // 430 宽等更宽的手机，但仍小于平板
        'phone-430': '430px',
        // iPad / 小平板 768 宽（768x1024 等）
        'tablet-768': '768px',
        // iPad Air 等 820 宽（820x1180 等），到桌面断点之前
        'tablet-820': '820px',
      },
    },
  },
  plugins: [],
}
