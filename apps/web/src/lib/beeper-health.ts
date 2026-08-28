import type { Beeper, BeeperRun } from "@/lib/api/beepers";

const BROKEN_ERROR_STREAK = 3;

export function isBeeperProbeBroken(runs: BeeperRun[] | undefined): boolean {
	if (!runs?.length) return false;

	const recent = runs
		.filter(
			(run) =>
				run.signal_status &&
				run.status !== "pending" &&
				run.status !== "running",
		)
		.slice(0, BROKEN_ERROR_STREAK);

	return (
		recent.length >= BROKEN_ERROR_STREAK &&
		recent.every((run) => run.signal_status === "error")
	);
}

export function beeperHealthLabel(
	beeper: Pick<Beeper, "alert_state" | "runs">,
): string {
	if (isBeeperProbeBroken(beeper.runs)) return "broken";
	return beeper.alert_state;
}

export function beeperHealthIsDestructive(
	beeper: Pick<Beeper, "alert_state" | "runs">,
): boolean {
	return (
		beeperHealthLabel(beeper) === "alerting" || isBeeperProbeBroken(beeper.runs)
	);
}
