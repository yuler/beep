import { ChevronRight } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import type { BeepRun } from "@/lib/api/beeps";

const RUN_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	pending: "secondary",
	running: "default",
	succeeded: "outline",
	failed: "destructive",
	skipped: "secondary",
	expired: "destructive",
};

const CHECK_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	ok: "outline",
	alerting: "destructive",
	error: "destructive",
};

function formatWhen(value: string) {
	return new Date(value).toLocaleString();
}

export function BeepRuns({ runs }: { runs: BeepRun[] }) {
	return (
		<details className="group/runs rounded-lg border bg-muted/20 text-sm">
			<summary className="flex cursor-pointer list-none items-center gap-2 px-3 py-2 marker:hidden [&::-webkit-details-marker]:hidden">
				<ChevronRight className="size-3.5 shrink-0 text-muted-foreground transition-transform group-open/runs:rotate-90" />
				<span className="font-medium">Execution Runs</span>
				<span className="text-muted-foreground">{runs.length}</span>
			</summary>
			{runs.length === 0 ? (
				<p className="border-t px-3 py-2 text-xs text-muted-foreground">
					No runs yet.
				</p>
			) : (
				<ul className="flex flex-col gap-2 border-t px-3 py-2">
					{runs.map((run) => (
						<li
							key={run.id}
							className="flex flex-col gap-2 rounded-md bg-background p-2.5 ring-1 ring-foreground/10"
						>
							<div className="flex flex-wrap items-center justify-between gap-2">
								<span className="tabular-nums text-xs text-muted-foreground">
									{formatWhen(run.scheduled_for)}
								</span>
								<div className="flex items-center gap-1.5">
									{run.check_status ? (
										<Badge
											variant={
												CHECK_STATUS_VARIANT[run.check_status] ?? "secondary"
											}
										>
											Check: {run.check_status.toUpperCase()}
										</Badge>
									) : null}
									<Badge
										variant={RUN_STATUS_VARIANT[run.status] ?? "secondary"}
									>
										Delivery: {run.status}
									</Badge>
								</div>
							</div>

							{run.check_result ? (
								<div className="rounded bg-muted/40 p-2 text-xs">
									{run.check_result.title ? (
										<div className="font-medium text-foreground">
											{run.check_result.title}
										</div>
									) : null}
									{run.check_result.message ? (
										<div className="text-muted-foreground mt-0.5">
											{run.check_result.message}
										</div>
									) : null}
									{run.check_result.metrics ? (
										<div className="mt-1 flex flex-wrap gap-2 text-[11px] font-mono text-muted-foreground">
											{Object.entries(run.check_result.metrics).map(
												([k, v]) => (
													<span
														key={k}
														className="rounded bg-background px-1.5 py-0.5 border"
													>
														{k}: {String(v)}
													</span>
												),
											)}
										</div>
									) : null}
								</div>
							) : null}

							{run.result && Object.keys(run.result).length > 0 ? (
								<pre className="max-h-36 overflow-auto rounded bg-muted/50 p-2 text-[11px] leading-snug whitespace-pre-wrap">
									{JSON.stringify(run.result, null, 2)}
								</pre>
							) : null}
						</li>
					))}
				</ul>
			)}
		</details>
	);
}
