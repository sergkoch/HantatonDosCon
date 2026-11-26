import { svelte } from "@sveltejs/vite-plugin-svelte";
import type { Plugin } from "vite";
import { defineConfig } from "vitest/config";

function inlineCss(): Plugin {
  return {
    name: "inline-css",
    apply: "build",
    enforce: "post",
    generateBundle(_options, bundle) {
      const html = Object.values(bundle).find(
        (file) => file.type === "asset" && file.fileName.endsWith(".html"),
      );

      if (!html || html.type !== "asset") return;

      let source = String(html.source);

      for (const [key, file] of Object.entries(bundle)) {
        if (file.type !== "asset" || !file.fileName.endsWith(".css")) continue;

        const href = `/${file.fileName}`;
        const escaped = href.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        const link = new RegExp(`<link[^>]*href="${escaped}"[^>]*>`);

        if (!link.test(source)) continue;

        source = source.replace(link, `<style>${String(file.source)}</style>`);
        delete bundle[key];
      }

      const entry = Object.values(bundle).find(
        (file) => file.type === "chunk" && file.isEntry,
      );

      if (entry && entry.type === "chunk") {
        const href = `/${entry.fileName}`;
        if (!source.includes(`href="${href}"`)) {
          source = source.replace(
            "</head>",
            `  <link rel="modulepreload" href="${href}">\n</head>`,
          );
        }
      }

      html.source = source;
    },
  };
}

export default defineConfig({
  plugins: [svelte(), inlineCss()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  test: {
    environment: "node",
  },
});
