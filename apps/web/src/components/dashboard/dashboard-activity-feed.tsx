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
import { Timeline, TimelineItem } from "@/components/ui/timeline";
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
				return <Bell className="size-3.5 text-ink" />;
			case "beeper":
				return <Radio className="size-3.5 text-ink" />;
			case "runner":
				return <Server className="size-3.5 text-ink" />;
			default:
				return <Activity className="size-3.5 text-gray" />;
		}
	}

	return (
		<Card className="bg-transparent border-t border-line sm:border sm:border-line">
			<CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between pb-3">
				<div>
					<CardTitle className="type-app-title font-medium text-ink">
						Recent Activity
					</CardTitle>
					<CardDescription className="type-meta text-gray">
						Live timeline across task executions, probe checks, and runner runs
					</CardDescription>
				</div>

				<div className="flex flex-wrap items-center gap-1">
					<Button
						variant={filter === "all" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0"
						onClick={() => setFilter("all")}
					>
						All
					</Button>
					<Button
						variant={filter === "beep" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0"
						onClick={() => setFilter("beep")}
					>
						Beeps
					</Button>
					<Button
						variant={filter === "beeper" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0"
						onClick={() => setFilter("beeper")}
					>
						Beepers
					</Button>
					<Button
						variant={filter === "runner" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0"
						onClick={() => setFilter("runner")}
					>
						Runners
					</Button>
					<Button
						variant={filter === "failures" ? "solid" : "ghost"}
						size="xs"
						className={cn(
							"h-7 text-xs px-2.5 min-h-0 sm:min-h-0",
							filter === "failures" && "border-ink",
						)}
						onClick={() => setFilter("failures")}
					>
						Failures Only
					</Button>
				</div>
			</CardHeader>

			<CardContent className="pt-2">
				{filteredActivities.length === 0 ? (
					<div className="flex h-36 flex-col items-center justify-center border border-line text-gray">
						<Filter className="size-5 mb-1.5 opacity-60 text-gray" />
						<p className="text-sm font-normal text-ink">
							No recent activities found
						</p>
						<p className="type-meta text-gray">
							{filter === "all"
								? "Executions will appear here in real time."
								: "No matching activities for the selected filter."}
						</p>
					</div>
				) : (
					<Timeline className="border-b border-line">
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
							const isFailure =
								act.status === "failure" || act.status === "warning";

							return (
								<TimelineItem
									key={act.id}
									bordered={isFailure}
									className="flex items-center justify-between gap-3 py-3 px-2"
								>
									<div className="flex items-center gap-3 min-w-0">
										<div className="flex size-7 shrink-0 items-center justify-center border border-line bg-transparent">
											{getTypeIcon(act.type)}
										</div>

										<div className="flex flex-col min-w-0">
											<div className="flex items-center gap-2">
												<Link
													to={targetHref}
													className="truncate text-sm font-normal text-ink hover:underline"
												>
													{act.title}
												</Link>
												<Badge
													variant="outline"
													className="type-kicker px-1.5 py-0 uppercase text-gray"
												>
													{act.type}
												</Badge>
											</div>
											<span className="truncate type-meta text-gray">
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
											className="type-meta text-gray whitespace-nowrap text-xs"
											title={act.occurred_at}
										>
											{relativeTime}
										</span>
									</div>
								</TimelineItem>
							);
						})}
					</Timeline>
				)}
			</CardContent>
		</Card>
	);
}
