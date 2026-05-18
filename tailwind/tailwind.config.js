/** @type {import('tailwindcss').Config} */
const defaultTheme = require("tailwindcss/defaultTheme");

module.exports = {
  content: [
    "./internal/view/**/*.templ",
  ],
  theme: {
    extend: {
      colors: {
        ff: {
          dark:    "#0f1117",
          dark2:   "#161a24",
          panel:   "#13161f",
          panel2:  "#1a1e2b",
          dusk:    "#5b7ec4",
          "dusk-hover": "#6d8fd4",
          "dusk-mid": "#2a3d6e",
          "dusk-lo": "#1e2d52",
          ice:     "#a8c0e8",
          moon:    "#c8d5ed",
          paper:   "#edf0f7",
          cream:   "#dde2ed",
          ash:     "#7a8099",
          stone:   "#3d4459",
          border:  "rgba(255,255,255,0.07)",
          border2: "rgba(255,255,255,0.12)",
        },
      },
      fontFamily: {
        display: ['"Cormorant Garamond"', ...defaultTheme.fontFamily.serif],
        body:    ['"DM Sans"', ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [],
}
