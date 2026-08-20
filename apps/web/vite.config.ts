import path from "node:path";
import { fileURLToPath, URL } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { nitro } from "nitro/vite";
import { defineConfig, type Plugin } from "vite";

// Vite's own host check hard-codes *.localhost as allowed (they resolve to
// loopback), so allowedHosts cannot reject a wrong *.localhost that hits our
// port. This middleware locks the Host header to the canonical web host.
// Plain localhost is deliberately excluded: Mode A login is cookie/CSRF-unsafe
// from there, so only web.<APP_HOST> should serve this app.
function hostAllowlist(webHost: string): Plugin {
	const allowed = new Set([webHost]);
	return {
		name: "host-allowlist",
		configureServer(server) {
			server.middlewares.use((req, res, next) => {
				let hostname = "";
				try {
					hostname = new URL(`http://${req.headers.host}`).hostname;
				} catch {
					// fall through to the block below
				}
				if (hostname && allowed.has(hostname)) return next();
				const port = server.config.server.port;
				const canonical = `http://${webHost}:${port}`;
				res.statusCode = 403;
				res.setHeader("Content-Type", "text/html; charset=utf-8");
				res.end(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Blocked host</title>
<style>
body { font-family: -apple-system, "Segoe UI", sans-serif; padding: 48px; line-height: 1.5; }
h1 { font-size: 20px; }
code, a { font-size: 15px; }
code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; }
a { color: #0b57d0; }
</style>
</head>
<body>
<h1>Blocked hosts: ${req.headers.host ?? "unknown"}</h1>
<p>This is the <strong>beep</strong> development server. Wrong host or plain
loopback is refused so a different project&rsquo;s <code>*.localhost</code> can&rsquo;t
load it. Use the canonical URL:</p>
<p><a href="${canonical}">${canonical}</a></p>
</body>
</html>
`);
			});
		},
	};
}

export default defineConfig(({ command }) => {
	// Dev server must run under mise (`_.file` loads `.env` / `.env.local`
	// into process.env). Builds (Docker/CI) run without mise, so skip them.
	if (command === "serve" && !process.env.MISE_TASK_NAME) {
		throw new Error(
			"web must be started via `mise dev` (or `mise run web:dev`). Direct `pnpm run dev` does not load .env/.env.local.",
		);
	}

	const requireEnv = (name: string): string => {
		const value = process.env[name];
		if (!value) {
			throw new Error(`${name} is required. Copy .env.example to .env.`);
		}
		return value;
	};

	const coreProxy = requireEnv("CORE_INTERNAL_URL");
	const isDevServer = command === "serve";
	const appHost = isDevServer ? requireEnv("APP_HOST") : undefined;
	const webPort = isDevServer ? Number(requireEnv("WEB_PORT")) : undefined;

	return {
		server: isDevServer
			? {
					port: webPort,
					allowedHosts: [`web.${appHost}`],
				}
			: undefined,
		resolve: {
			tsconfigPaths: true,
			alias: {
				"@": path.resolve(fileURLToPath(new URL(".", import.meta.url)), "src"),
			},
		},
		plugins: [
			...(appHost ? [hostAllowlist(`web.${appHost}`)] : []),
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
					"/service-worker.js": {
						headers: {
							"Cache-Control": "no-cache",
							"Service-Worker-Allowed": "/",
						},
					},
				},
			}),
		],
	};
});
