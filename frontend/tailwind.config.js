/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#0f172a",
        panel: "#111827",
        accent: "#f59e0b",
        mist: "#cbd5e1"
      }
    },
  },
  plugins: [],
};
