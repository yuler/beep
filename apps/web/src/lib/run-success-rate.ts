import type { Beeper } from "@/lib/api/beepers";
import type { Beep } from "@/lib/api/beeps";

function successRateFromCounts(total: number, succeeded: number) {
	if (total === 0) return 0;
	return Math.round((succeeded / total) * 100);
}

export function runSuccessRate(runs: Array<{ status: string }>) {
	return successRateFromCounts(
		runs.length,
		runs.filter((run) => run.status === "succeeded").length,
	);
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
	return successRateFromCounts(total, succeeded);
}

export function beeperRunCount(beeper: Pick<Beeper, "runs" | "run_stats">) {
	return beeperRunStats(beeper).total;
}

export function beepRunStats(beep: Pick<Beep, "runs" | "run_stats">) {
	if (beep.run_stats) return beep.run_stats;
	const runs = beep.runs ?? [];
	return {
		total: runs.length,
		succeeded: runs.filter((run) => run.status === "succeeded").length,
	};
}

export function beepRunSuccessRate(beep: Pick<Beep, "runs" | "run_stats">) {
	const { total, succeeded } = beepRunStats(beep);
	return successRateFromCounts(total, succeeded);
}

export function beepRunCount(beep: Pick<Beep, "runs" | "run_stats">) {
	return beepRunStats(beep).total;
}
