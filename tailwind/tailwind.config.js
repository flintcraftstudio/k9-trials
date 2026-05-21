/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/view/**/*.templ",
  ],
  theme: {
    extend: {
      colors: {
        // K9 Elements design tokens. All values reference CSS vars defined in
        // K9 Elements Design System/colors_and_type.css, imported through
        // tailwind/input.css. Tailwind opacity modifiers (bg-mist-100/50) are
        // not supported on these — use rgb(from var(--mist-N) r g b / X) in
        // custom CSS when you need transparency.
        mist: {
          50:  "var(--mist-50)",
          100: "var(--mist-100)",
          200: "var(--mist-200)",
          300: "var(--mist-300)",
          400: "var(--mist-400)",
          500: "var(--mist-500)",
          600: "var(--mist-600)",
          700: "var(--mist-700)",
          800: "var(--mist-800)",
          900: "var(--mist-900)",
          950: "var(--mist-950)",
        },
        obedience: {
          50:  "var(--blue-50)",  100: "var(--blue-100)", 200: "var(--blue-200)",
          300: "var(--blue-300)", 400: "var(--blue-400)", 500: "var(--blue-500)",
          600: "var(--blue-600)", 700: "var(--blue-700)", 800: "var(--blue-800)",
          900: "var(--blue-900)", 950: "var(--blue-950)",
        },
        protection: {
          50:  "var(--char-50)",  100: "var(--char-100)", 200: "var(--char-200)",
          300: "var(--char-300)", 400: "var(--char-400)", 500: "var(--char-500)",
          600: "var(--char-600)", 700: "var(--char-700)", 800: "var(--char-800)",
          900: "var(--char-900)", 950: "var(--char-950)",
        },
        tracking: {
          50:  "var(--green-50)",  100: "var(--green-100)", 200: "var(--green-200)",
          300: "var(--green-300)", 400: "var(--green-400)", 500: "var(--green-500)",
          600: "var(--green-600)", 700: "var(--green-700)", 800: "var(--green-800)",
          900: "var(--green-900)", 950: "var(--green-950)",
        },
        detection: {
          50:  "var(--orange-50)",  100: "var(--orange-100)", 200: "var(--orange-200)",
          300: "var(--orange-300)", 400: "var(--orange-400)", 500: "var(--orange-500)",
          600: "var(--orange-600)", 700: "var(--orange-700)", 800: "var(--orange-800)",
          900: "var(--orange-900)", 950: "var(--orange-950)",
        },
        // `event-*` resolves to whichever event hue is in scope via
        // [data-event="..."]. Use inside a data-event container.
        event: {
          50:  "var(--event-50)",
          100: "var(--event-100)",
          600: "var(--event-600)",
          700: "var(--event-700)",
        },
      },
      fontFamily: {
        // Match var(--font-display) and var(--font-sans) from
        // colors_and_type.css.
        "instrument-display": ['"Instrument Serif"', "ui-serif", "Georgia", '"Times New Roman"', "serif"],
        "instrument-sans":    ['"Instrument Sans"', "ui-sans-serif", "system-ui", "-apple-system", '"Segoe UI"', "Roboto", "sans-serif"],
      },
    },
  },
  plugins: [],
}
