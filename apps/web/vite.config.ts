import path from "node:path";
import { fileURLToPath, URL } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { nitro } from "nitro/vite";
import { defineConfig, loadEnv } from "vite";

const monorepoRoot = path.resolve(
	fileURLToPath(new URL(".", import.meta.url)),
	"../..",
);

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, monorepoRoot, "");
	const coreProxy = env.CORE_INTERNAL_URL || "http://core.beep.localhost:3001";

	return {
		envDir: monorepoRoot,
		server: {
			port: Number(env.WEB_PORT) || 3000,
			// Allow http://web.beep.localhost:<port> (and any *.localhost)
			allowedHosts: [".localhost"],
		},
		resolve: {
			tsconfigPaths: true,
			alias: {
				"@": path.resolve(fileURLToPath(new URL(".", import.meta.url)), "src"),
			},
		},
		plugins: [
			tailwindcss(),
			tanstackStart({
				srcDirectory: "src",
			}),
			viteReact(),
			nitro({
				// Mode B: same-origin /api → Rails core (Compose service `core`, or local).
				routeRules: {
					"/api/**": { proxy: `${coreProxy}/api/**` },
					"/up": { proxy: `${coreProxy}/up` },
				},
			}),
		],
	};
});
