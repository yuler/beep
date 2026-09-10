import type { Beeper } from "@/lib/api/beepers";

export function runSuccessRate(runs: Array<{ status: string }>) {
	if (runs.length === 0) return 0;
	const succeeded = runs.filter((run) => run.status === "succeeded").length;
	return Math.round((succeeded / runs.length) * 100);
}

export function beeperRunStats(beeper: Pick<Beeper, "runs" | "run_stats">) {
	if (beeper.run_stats) return beeper.run_stats;
	const runs = beeper.runs ?? [];
	return {
		total: runs.length,
		succeeded: runs.filter((run) => run.status === "succeeded").length,
	};
}

export function beeperRunSuccessRate(
	beeper: Pick<Beeper, "runs" | "run_stats">,
) {
	const { total, succeeded } = beeperRunStats(beeper);
	if (total === 0) return 0;
	return Math.round((succeeded / total) * 100);
}

export function beeperRunCount(beeper: Pick<Beeper, "runs" | "run_stats">) {
	return beeperRunStats(beeper).total;
}
