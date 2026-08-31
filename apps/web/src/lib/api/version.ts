import type { BuildInfo } from "@/lib/build-info";

function isBuildInfo(value: unknown): value is BuildInfo {
	if (typeof value !== "object" || value === null) return false;
	const record = value as Record<string, unknown>;
	return (
		typeof record.version === "string" &&
		record.version.length > 0 &&
		typeof record.gitHash === "string" &&
		record.gitHash.length > 0 &&
		typeof record.buildTime === "string" &&
		record.buildTime.length > 0
	);
}

export async function fetchDeployedVersion(): Promise<BuildInfo> {
	const res = await fetch("/version.json", { cache: "no-store" });
	if (!res.ok) throw new Error("version fetch failed");
	const data: unknown = await res.json();
	if (!isBuildInfo(data)) throw new Error("invalid version.json");
	return data;
}

export function isNewerDeploy(
	deployed: BuildInfo,
	current: BuildInfo,
): boolean {
	return (
		deployed.gitHash !== current.gitHash || deployed.version !== current.version
	);
}
