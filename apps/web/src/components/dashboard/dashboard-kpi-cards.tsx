import { Activity, Bell, Radio, Server } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { StatusPill } from "@/components/ui/status-pill";
import type { DashboardStats } from "@/lib/api/dashboard";

export function DashboardKpiCards({ stats }: { stats: DashboardStats }) {
	const { executions, beeps, beepers, runners } = stats;

	const healthTone =
		executions.success_rate >= 99
			? "emerald"
			: executions.success_rate >= 90
				? "amber"
				: "rose";

	const beeperTone = beepers.alerting > 0 ? "amber" : "emerald";

	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{/* Executions & Success Rate */}
			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4 sm:p-5">
					<div className="flex items-center justify-between">
						<span className="text-xs font-medium text-muted-foreground">
							Total Executions
						</span>
						<div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
							<Activity className="size-4" />
						</div>
					</div>
					<div className="flex flex-col gap-1">
						<span className="font-heading text-2xl font-bold tabular-nums tracking-tight">
							{executions.total.toLocaleString()}
						</span>
						<div className="flex items-center gap-2">
							<StatusPill
								label={`${executions.success_rate}% success`}
								tone={healthTone}
							/>
							<span className="text-xs text-muted-foreground">
								{executions.failure_count > 0
									? `${executions.failure_count} failed`
									: "0 failures"}
							</span>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* Beeps Status */}
			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4 sm:p-5">
					<div className="flex items-center justify-between">
						<span className="text-xs font-medium text-muted-foreground">
							Active Beeps
						</span>
						<div className="flex size-8 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
							<Bell className="size-4" />
						</div>
					</div>
					<div className="flex flex-col gap-1">
						<span className="font-heading text-2xl font-bold tabular-nums tracking-tight">
							{beeps.active}
						</span>
						<span className="text-xs text-muted-foreground">
							{beeps.due_today} due today • {beeps.firing} firing
						</span>
					</div>
				</CardContent>
			</Card>

			{/* Beepers Health */}
			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4 sm:p-5">
					<div className="flex items-center justify-between">
						<span className="text-xs font-medium text-muted-foreground">
							Beepers Probes
						</span>
						<div className="flex size-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
							<Radio className="size-4" />
						</div>
					</div>
					<div className="flex flex-col gap-1">
						<div className="flex items-baseline gap-2">
							<span className="font-heading text-2xl font-bold tabular-nums tracking-tight">
								{beepers.healthy} / {beepers.total}
							</span>
							<span className="text-xs text-muted-foreground">healthy</span>
						</div>
						<div className="flex items-center gap-1.5">
							<StatusPill
								label={
									beepers.alerting > 0
										? `${beepers.alerting} alerting`
										: "All healthy"
								}
								tone={beeperTone}
							/>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* Runners */}
			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4 sm:p-5">
					<div className="flex items-center justify-between">
						<span className="text-xs font-medium text-muted-foreground">
							Runners & Nodes
						</span>
						<div className="flex size-8 items-center justify-center rounded-lg bg-violet-500/10 text-violet-600 dark:text-violet-400">
							<Server className="size-4" />
						</div>
					</div>
					<div className="flex flex-col gap-1">
						<div className="flex items-baseline gap-2">
							<span className="font-heading text-2xl font-bold tabular-nums tracking-tight">
								{runners.online} / {runners.total}
							</span>
							<span className="text-xs text-muted-foreground">online</span>
						</div>
						<span className="text-xs text-muted-foreground">
							{runners.active_jobs} configured job
							{runners.active_jobs === 1 ? "" : "s"}
						</span>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
