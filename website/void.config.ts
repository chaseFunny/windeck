import { defineConfig } from "void/config";

export default defineConfig({
  // A static site: every page is prerendered at build time.
  output: "static",
  head: {
    titleTemplate: "%s | Windeck",
    htmlAttrs: { lang: "en" },
    meta: [
      { charset: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { name: "theme-color", content: "#f7f8fa" },
    ],
    link: [{ rel: "icon", type: "image/png", href: "/favicon.png" }],
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
