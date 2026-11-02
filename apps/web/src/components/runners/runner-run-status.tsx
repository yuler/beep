import { Badge } from "@/components/ui/badge";
import { StatusPill } from "@/components/ui/status-pill";
import type { RunnerJobStatus, RunnerRun } from "@/lib/api/runner-jobs";
import { beepRunStatusLabel, healthStatusLabel } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

export function isActiveRunnerRun(run: RunnerRun) {
	return run.status === "pending" || run.status === "running";
}

export function RunnerJobStatusBadge({ status }: { status: RunnerJobStatus }) {
	if (status === "firing") {
		return <StatusPill label={m.status_beep_firing()} tone="live" />;
	}
	if (status === "paused") {
		return <Badge variant="secondary">{m.status_beep_paused()}</Badge>;
	}
	return <Badge variant="outline">{m.status_beep_active()}</Badge>;
}

export function RunnerRunStatusBadge({ run }: { run: RunnerRun }) {
	if (run.status === "running") {
		return <StatusPill label={beepRunStatusLabel(run.status)} tone="live" />;
	}

	const key = run.result_status || run.status;
	const variant =
		key === "succeeded" || key === "ok"
			? "success"
			: key === "failed" || key === "error" || key === "alerting"
				? "destructive"
				: key === "expired"
					? "warning"
					: "outline";
	const label = run.result_status
		? healthStatusLabel(run.result_status)
		: beepRunStatusLabel(run.status);

	return <Badge variant={variant}>{label}</Badge>;
}
