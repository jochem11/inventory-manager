import { defineConfig } from "vite";
import { nitro } from "nitro/vite";

import { solidStart } from "@solidjs/start/config";

export default defineConfig({
  plugins: [solidStart(),
    nitro()
  ],
  optimizeDeps: {
    // SolidStart's dev-toolbar error viewer imports these CommonJS packages in
    // the browser. It's loaded lazily, so Vite's scan misses them and serves
    // them unconverted ("does not provide an export named …").
    include: ["source-map-js", "error-stack-parser"],
  },
});
