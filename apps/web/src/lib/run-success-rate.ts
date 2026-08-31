export function runSuccessRate(runs: Array<{ status: string }>) {
	if (runs.length === 0) return 0;
	const succeeded = runs.filter((run) => run.status === "succeeded").length;
	return Math.round((succeeded / runs.length) * 100);
}
