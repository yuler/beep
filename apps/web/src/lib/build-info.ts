// Build metadata injected by vite.config.ts (define __BUILD_INFO__).
export interface BuildInfo {
	version: string;
	gitHash: string;
	buildTime: string;
}

declare const __BUILD_INFO__: BuildInfo;

export const buildInfo: BuildInfo = __BUILD_INFO__;

const shortHash = (hash: string) =>
	hash === "unknown" ? hash : hash.slice(0, 7);

const formatBuildTime = (iso: string) => {
	const d = new Date(iso);
	const pad = (n: number) => String(n).padStart(2, "0");
	return (
		`${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
		`${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
	);
};

/** Pretty-print build info to the browser console for quick version checks. */
export function logBuildInfo(): void {
	console.log(
		"%cbeep %cv%s%c  %s%c  built %s",
		"color:#0ea5e9;font-weight:700;font-size:13px",
		"background:#0ea5e9;color:#fff;font-weight:600;font-size:11px;padding:1px 6px;border-radius:9999px",
		buildInfo.version,
		"color:inherit;font-size:12px",
		shortHash(buildInfo.gitHash),
		"color:#94a3b8;font-size:12px",
		formatBuildTime(buildInfo.buildTime),
	);
}
