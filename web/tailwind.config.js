/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "rgb(var(--color-primary))",
        secondary: "rgb(var(--color-secondary))",
        background: "rgb(var(--color-background))",
        surface: "rgb(var(--color-surface))",
        accent: "rgb(var(--color-accent))",
        success: "rgb(var(--color-success))",
        error: "rgb(var(--color-error))",
      },
      fontFamily: {
        body: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  darkMode: 'class',
  plugins: [],
}
