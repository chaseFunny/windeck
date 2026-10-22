import { defineConfig } from "cf/config";

// Windeck's website: the built static files, served by an assets-only
// Worker at windeck.islumi.com.
export default defineConfig({
  worker: {
    name: "windeck",
    compatibilityDate: "2026-10-06",
    assets: { htmlHandling: "auto-trailing-slash", notFoundHandling: "none" },
    domains: ["windeck.islumi.com"],
    workersDev: false,
  },
});
