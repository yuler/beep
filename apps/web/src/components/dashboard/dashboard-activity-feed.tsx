import { Link } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import { Activity, Bell, Filter, Radio, Server } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { StatusPill, type StatusPillTone } from "@/components/ui/status-pill";
import type { DashboardActivity } from "@/lib/api/dashboard";
import { cn } from "@/lib/utils";

type FilterType = "all" | "beep" | "beeper" | "runner" | "failures";

export function DashboardActivityFeed({
	activities,
	slug,
}: {
	activities: DashboardActivity[];
	slug: string;
}) {
	const [filter, setFilter] = useState<FilterType>("all");

	const filteredActivities = activities.filter((item) => {
		if (filter === "all") return true;
		if (filter === "failures") {
			return item.status === "failure" || item.status === "warning";
		}
		return item.type === filter;
	});

	function getStatusTone(status: DashboardActivity["status"]): StatusPillTone {
		switch (status) {
			case "success":
				return "emerald";
			case "warning":
				return "amber";
			case "failure":
				return "rose";
			default:
				return "muted";
		}
	}

	function getTypeIcon(type: DashboardActivity["type"]) {
		switch (type) {
			case "beep":
				return <Bell className="size-3.5 text-blue-500" />;
			case "beeper":
				return <Radio className="size-3.5 text-emerald-500" />;
			case "runner":
				return <Server className="size-3.5 text-violet-500" />;
			default:
				return <Activity className="size-3.5 text-muted-foreground" />;
		}
	}

	return (
		<Card className="bg-card/60 transition-colors hover:bg-card">
			<CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between pb-3">
				<div>
					<CardTitle className="font-heading text-lg font-semibold tracking-tight">
						Recent Activity
					</CardTitle>
					<CardDescription className="text-xs">
						Live timeline across task executions, probe checks, and runner runs
					</CardDescription>
				</div>

				<div className="flex flex-wrap items-center gap-1">
					<Button
						variant={filter === "all" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => setFilter("all")}
					>
						All
					</Button>
					<Button
						variant={filter === "beep" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => setFilter("beep")}
					>
						Beeps
					</Button>
					<Button
						variant={filter === "beeper" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => setFilter("beeper")}
					>
						Beepers
					</Button>
					<Button
						variant={filter === "runner" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => setFilter("runner")}
					>
						Runners
					</Button>
					<Button
						variant={filter === "failures" ? "secondary" : "ghost"}
						size="sm"
						className={cn(
							"h-7 text-xs px-2.5",
							filter === "failures" && "text-rose-600 dark:text-rose-400",
						)}
						onClick={() => setFilter("failures")}
					>
						Failures Only
					</Button>
				</div>
			</CardHeader>

			<CardContent className="pt-2">
				{filteredActivities.length === 0 ? (
					<div className="flex h-36 flex-col items-center justify-center rounded-lg border border-dashed border-border/80 text-muted-foreground">
						<Filter className="size-5 mb-1.5 opacity-60" />
						<p className="text-sm font-medium">No recent activities found</p>
						<p className="text-xs text-muted-foreground/80">
							{filter === "all"
								? "Executions will appear here in real time."
								: "No matching activities for the selected filter."}
						</p>
					</div>
				) : (
					<div className="divide-y divide-border/50">
						{filteredActivities.map((act) => {
							const relativeTime = (() => {
								try {
									return formatDistanceToNow(new Date(act.occurred_at), {
										addSuffix: true,
									});
								} catch {
									return act.occurred_at;
								}
							})();

							const targetHref = `/${slug}/${act.target_path}`;

							return (
								<div
									key={act.id}
									className="flex items-center justify-between gap-3 py-3 transition-colors hover:bg-muted/20 px-2 rounded-md"
								>
									<div className="flex items-center gap-3 min-w-0">
										<div className="flex size-7 shrink-0 items-center justify-center rounded-md border border-border/60 bg-background/50">
											{getTypeIcon(act.type)}
										</div>

										<div className="flex flex-col min-w-0">
											<div className="flex items-center gap-2">
												<Link
													to={targetHref}
													className="truncate text-sm font-medium text-foreground hover:underline"
												>
													{act.title}
												</Link>
												<Badge
													variant="outline"
													className="text-[10px] uppercase px-1.5 py-0 font-mono tracking-wider text-muted-foreground"
												>
													{act.type}
												</Badge>
											</div>
											<span className="truncate text-xs text-muted-foreground">
												{act.summary}
											</span>
										</div>
									</div>

									<div className="flex items-center gap-2.5 shrink-0">
										<StatusPill
											label={act.status}
											tone={getStatusTone(act.status)}
											className="capitalize text-[11px] h-5 px-2"
										/>
										<span
											className="text-xs text-muted-foreground whitespace-nowrap"
											title={act.occurred_at}
										>
											{relativeTime}
										</span>
									</div>
								</div>
							);
						})}
					</div>
				)}
			</CardContent>
		</Card>
	);
}
