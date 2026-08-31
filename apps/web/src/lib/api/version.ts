import type { BuildInfo } from "@/lib/build-info";

export async function fetchDeployedVersion(): Promise<BuildInfo> {
	const res = await fetch("/version.json", { cache: "no-store" });
	if (!res.ok) throw new Error("version fetch failed");
	return res.json() as Promise<BuildInfo>;
}

export function isNewerDeploy(
	deployed: BuildInfo,
	current: BuildInfo,
): boolean {
	return (
		deployed.gitHash !== current.gitHash || deployed.version !== current.version
	);
}
