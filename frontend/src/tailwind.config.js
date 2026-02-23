import typography from "@tailwindcss/typography";
import daisyui from "daisyui";
import catppuccin from "@catppuccin/daisyui/legacy";

/** @type {import('tailwindcss').Config} */
export default {
  content: ["./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [typography, daisyui],
  daisyui: {
    themes: [catppuccin("mocha")],
  },
};
