import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm"],
  target: "node20",
  platform: "node",
  outDir: "dist",
  clean: true,
  dts: true,
  sourcemap: true,
  // Prepend a shebang so the built file is directly runnable via the `bin` entry / npx.
  banner: {
    js: "#!/usr/bin/env node",
  },
});
