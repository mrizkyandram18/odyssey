/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Core Palette
        background: "rgb(var(--bg-app))",
        "background-realm": "rgb(var(--bg-realm))",
        surface: "rgb(var(--surface))",
        "surface-elevated": "rgb(var(--surface-elevated))",
        "surface-glass": "rgb(var(--surface-glass))",
        "border-subtle": "rgb(var(--border-subtle))",
        
        // Typography
        "text-primary": "rgb(var(--text-primary))",
        "text-secondary": "rgb(var(--text-secondary))",
        
        // Semantic Accents
        "accent-magic": "rgb(var(--accent-magic))",
        "accent-nature": "rgb(var(--accent-nature))",
        "accent-reward": "rgb(var(--accent-reward))",
        "accent-rare": "rgb(var(--accent-rare))",
        "accent-danger": "rgb(var(--accent-danger))",
        
        // Legacy fallbacks for compatibility until fully refactored
        primary: "rgb(var(--accent-nature))",
        secondary: "rgb(var(--accent-magic))",
        accent: "rgb(var(--accent-reward))",
        success: "rgb(var(--accent-nature))",
        error: "rgb(var(--accent-danger))",
      },
      fontFamily: {
        heading: ['Nunito', 'sans-serif'],
        body: ['Inter', 'sans-serif'],
      },
    },
  },
  darkMode: 'class',
  plugins: [],
}
