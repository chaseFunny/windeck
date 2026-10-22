import { defineConfig } from "void/config";

export default defineConfig({
  // A static site: every page is prerendered at build time.
  output: "static",
  head: {
    htmlAttrs: { lang: "en" },
    meta: [
      { charset: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { name: "theme-color", content: "#ffffff" },
    ],
    // The hz type: Inter and Noto Sans SC for text, Nunito for titles.
    link: [
      { rel: "icon", type: "image/png", href: "/favicon.png" },
      { rel: "preconnect", href: "https://fonts.googleapis.com" },
      { rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: "" },
      {
        rel: "stylesheet",
        href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&family=Nunito:wght@500;700&family=Noto+Sans+SC:wght@400;500;700&family=JetBrains+Mono&display=swap",
      },
    ],
  },
  worker: {
    compatibility_date: "2026-02-24",
  },
  cloudflare: {
    name: "windeck",
    workers_dev: true,
    preview_urls: true,
  },
});
