/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./cmd/web/**/*.html", // Or wherever your Go templates are
    "./cmd/web/**/*.templ", // If you're using templ files
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
