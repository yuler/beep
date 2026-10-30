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
				? "muted"
				: "rose";

	const beeperTone = beepers.alerting > 0 ? "rose" : "emerald";

	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{/* Executions & Success Rate */}
			<Card
				size="sm"
				className="bg-transparent border-t border-line sm:border sm:border-line"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4">
					<div className="flex items-center justify-between">
						<span className="type-kicker text-gray">Total Executions</span>
						<div className="flex size-7 items-center justify-center border border-line text-ink">
							<Activity className="size-3.5" />
						</div>
					</div>
					<div className="flex flex-col gap-1.5">
						<span className="type-title font-medium tabular-nums text-ink">
							{executions.total.toLocaleString()}
						</span>
						<div className="flex items-center gap-2">
							<StatusPill
								label={`${executions.success_rate}% success`}
								tone={healthTone}
							/>
							<span className="type-meta text-gray">
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
				className="bg-transparent border-t border-line sm:border sm:border-line"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4">
					<div className="flex items-center justify-between">
						<span className="type-kicker text-gray">Active Beeps</span>
						<div className="flex size-7 items-center justify-center border border-line text-ink">
							<Bell className="size-3.5" />
						</div>
					</div>
					<div className="flex flex-col gap-1.5">
						<span className="type-title font-medium tabular-nums text-ink">
							{beeps.active}
						</span>
						<span className="type-meta text-gray">
							{beeps.due_today} due today • {beeps.firing} firing
						</span>
					</div>
				</CardContent>
			</Card>

			{/* Beepers Health */}
			<Card
				size="sm"
				className="bg-transparent border-t border-line sm:border sm:border-line"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4">
					<div className="flex items-center justify-between">
						<span className="type-kicker text-gray">Beepers Probes</span>
						<div className="flex size-7 items-center justify-center border border-line text-ink">
							<Radio className="size-3.5" />
						</div>
					</div>
					<div className="flex flex-col gap-1.5">
						<div className="flex items-baseline gap-2">
							<span className="type-title font-medium tabular-nums text-ink">
								{beepers.healthy} / {beepers.total}
							</span>
							<span className="type-meta text-gray">healthy</span>
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
				className="bg-transparent border-t border-line sm:border sm:border-line"
			>
				<CardContent className="flex flex-col justify-between gap-3 p-4">
					<div className="flex items-center justify-between">
						<span className="type-kicker text-gray">Runners & Nodes</span>
						<div className="flex size-7 items-center justify-center border border-line text-ink">
							<Server className="size-3.5" />
						</div>
					</div>
					<div className="flex flex-col gap-1.5">
						<div className="flex items-baseline gap-2">
							<span className="type-title font-medium tabular-nums text-ink">
								{runners.online} / {runners.total}
							</span>
							<span className="type-meta text-gray">online</span>
						</div>
						<span className="type-meta text-gray">
							{runners.active_jobs} configured job
							{runners.active_jobs === 1 ? "" : "s"}
						</span>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
