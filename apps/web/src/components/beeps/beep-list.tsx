import { Link } from "@tanstack/react-router";
import {
	Activity,
	ArrowUpRight,
	Clock,
	Repeat,
	Search,
	Sparkles,
} from "lucide-react";
import { useMemo, useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { BEEP_STATUS_META } from "@/components/beeps/beep-status";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { Beep } from "@/lib/api/beeps";
import { beepRunAt } from "@/lib/beep-stats";
import { cn } from "@/lib/utils";

type FilterStatus = "all" | "active" | "firing" | "recurring" | "completed";

export function BeepList({
	beeps,
	slug,
	variant = "full",
}: {
	beeps: Beep[];
	slug: string;
	variant?: "compact" | "full";
}) {
	const [search, setSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState<FilterStatus>("all");

	const filteredBeeps = useMemo(() => {
		return beeps.filter((beep) => {
			// Status / Kind filter
			if (statusFilter === "active" && beep.status !== "active") return false;
			if (statusFilter === "firing" && beep.status !== "firing") return false;
			if (statusFilter === "completed" && beep.status !== "completed")
				return false;
			if (statusFilter === "recurring" && beep.kind !== "recurring")
				return false;

			// Text search
			if (search.trim()) {
				const query = search.toLowerCase();
				const matchTitle = beep.title.toLowerCase().includes(query);
				const matchBody = beep.body?.toLowerCase().includes(query);
				const matchBeeper = beep.beeper?.name.toLowerCase().includes(query);
				if (!matchTitle && !matchBody && !matchBeeper) return false;
			}

			return true;
		});
	}, [beeps, statusFilter, search]);

	const counts = useMemo(() => {
		return {
			all: beeps.length,
			active: beeps.filter((b) => b.status === "active").length,
			firing: beeps.filter((b) => b.status === "firing").length,
			recurring: beeps.filter((b) => b.kind === "recurring").length,
			completed: beeps.filter((b) => b.status === "completed").length,
		};
	}, [beeps]);

	if (beeps.length === 0) {
		return (
			<Card className="flex flex-col items-center justify-center p-8 text-center">
				<div className="flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
					<Sparkles className="size-6" />
				</div>
				<h3 className="mt-3 font-heading text-base font-semibold">
					No beeps yet
				</h3>
				<p className="mt-1 max-w-sm text-sm text-muted-foreground">
					Create your first reminder above or install a beeper from the gallery.
				</p>
			</Card>
		);
	}

	return (
		<div className="flex flex-col gap-4">
			{variant === "full" ? (
				<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					{/* Status filter tabs */}
					<div className="flex flex-nowrap items-center gap-1 overflow-x-auto rounded-lg border border-input bg-muted/30 p-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
						{(
							[
								{ id: "all", label: "All", count: counts.all },
								{ id: "active", label: "Active", count: counts.active },
								{ id: "firing", label: "Firing", count: counts.firing },
								{
									id: "recurring",
									label: "Recurring",
									count: counts.recurring,
								},
								{
									id: "completed",
									label: "Completed",
									count: counts.completed,
								},
							] as const
						).map((tab) => (
							<Button
								key={tab.id}
								type="button"
								size="xs"
								variant={statusFilter === tab.id ? "default" : "ghost"}
								className={cn(
									"h-7 shrink-0 text-xs font-medium whitespace-nowrap transition-colors",
									statusFilter === tab.id
										? "bg-background text-foreground shadow-sm dark:bg-card dark:text-foreground"
										: "text-muted-foreground hover:text-foreground",
								)}
								onClick={() => setStatusFilter(tab.id)}
							>
								{tab.label}
								<span
									className={cn(
										"ml-1 rounded-full px-1.5 py-0.2 text-[10px] tabular-nums",
										statusFilter === tab.id
											? "bg-primary/10 text-primary dark:bg-primary/20"
											: "bg-muted text-muted-foreground",
									)}
								>
									{tab.count}
								</span>
							</Button>
						))}
					</div>

					{/* Search input */}
					<div className="relative w-full sm:w-64">
						<Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
						<Input
							value={search}
							onChange={(event) => setSearch(event.target.value)}
							placeholder="Search beeps…"
							className="h-8.5 pl-8 text-xs"
						/>
					</div>
				</div>
			) : null}

			{filteredBeeps.length === 0 ? (
				<Card className="flex flex-col items-center justify-center p-8 text-center">
					<p className="text-sm font-medium text-muted-foreground">
						No beeps match your filter.
					</p>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						className="mt-2"
						onClick={() => {
							setStatusFilter("all");
							setSearch("");
						}}
					>
						Clear filters
					</Button>
				</Card>
			) : (
				<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
					{filteredBeeps.map((beep) => (
						<BeepListCard
							key={beep.id}
							beep={beep}
							slug={slug}
							variant={variant}
						/>
					))}
				</div>
			)}
		</div>
	);
}

function StatusIndicator({ status }: { status: Beep["status"] }) {
	const meta = BEEP_STATUS_META[status];
	const Icon = meta.icon;
	return (
		<span
			className={cn(
				"inline-flex items-center gap-1 text-[11px] font-medium",
				meta.colorClass,
			)}
		>
			<Icon className={cn("size-3", status === "firing" && "animate-pulse")} />
			{meta.label}
		</span>
	);
}

function formatSchedule(beep: Beep) {
	if (beep.status === "completed") {
		return { label: "Completed", isRecurring: false };
	}

	if (beep.kind === "recurring") {
		return {
			label: `Cron: ${beep.cron ?? "Scheduled"}`,
			isRecurring: true,
		};
	}

	const nextRun = beepRunAt(beep);
	if (!nextRun) return null;

	const date = new Date(nextRun);
	return {
		label: date.toLocaleString(undefined, {
			month: "short",
			day: "numeric",
			hour: "numeric",
			minute: "2-digit",
		}),
		isRecurring: false,
	};
}

function BeepListCard({
	beep,
	slug,
	variant,
}: {
	beep: Beep;
	slug: string;
	variant: "compact" | "full";
}) {
	const schedule = formatSchedule(beep);
	const lastRun = beep.runs[beep.runs.length - 1];

	return (
		<Card
			size="sm"
			className="group relative overflow-hidden transition-all duration-150 hover:border-primary/40 hover:shadow-xs"
		>
			<Link
				to="/$account_slug/beeps/$beepId"
				params={{
					account_slug: slug,
					beepId: beep.id,
				}}
				className="block p-4 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
			>
				{/* Top metadata row */}
				<div className="flex flex-wrap items-center justify-between gap-2">
					<div className="flex flex-wrap items-center gap-1.5">
						<StatusIndicator status={beep.status} />
						{beep.beeper ? (
							<Badge
								variant="secondary"
								className="gap-1 text-[10px] font-normal bg-primary/10 text-primary dark:bg-primary/20"
							>
								<Activity className="size-2.5" />
								from {beep.beeper.name}
							</Badge>
						) : (
							<Badge
								variant="secondary"
								className="gap-1 text-[11px] font-normal"
							>
								{beep.kind === "recurring" ? (
									<>
										<Repeat className="size-2.5" />
										Recurring
									</>
								) : (
									<>
										<Clock className="size-2.5" />
										Once
									</>
								)}
							</Badge>
						)}
					</div>

					{schedule ? (
						<div className="flex items-center gap-1 text-xs text-muted-foreground">
							<span className="tabular-nums font-medium text-foreground">
								{schedule.label}
							</span>
							<span className="text-[11px]">({beep.timezone})</span>
						</div>
					) : null}
				</div>

				{/* Title and details */}
				<div className="mt-2.5 flex items-start justify-between gap-3">
					<h3 className="font-heading text-base font-semibold text-foreground transition-colors group-hover:text-primary">
						{beep.title}
					</h3>
					<ArrowUpRight className="size-4 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-hover:text-primary" />
				</div>

				{/* Markdown Preview if body present */}
				{variant === "full" && beep.body ? (
					<div className="mt-2 line-clamp-2 text-xs text-muted-foreground">
						<BeepMarkdown source={beep.body} />
					</div>
				) : null}

				{/* Bottom runs summary */}
				{variant === "full" ? (
					<div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-border/60 pt-2.5 text-xs text-muted-foreground">
						<div className="flex items-center gap-2">
							<span className="tabular-nums">
								{beep.runs.length} {beep.runs.length === 1 ? "run" : "runs"}
							</span>
							{lastRun ? (
								<span className="inline-flex items-center gap-1 text-[11px]">
									· Last:{" "}
									<span
										className={cn(
											"font-medium",
											lastRun.status === "succeeded" &&
												"text-emerald-600 dark:text-emerald-400",
											lastRun.status === "failed" && "text-destructive",
											lastRun.status === "running" && "text-primary",
										)}
									>
										{lastRun.status}
									</span>
								</span>
							) : null}
						</div>

						<span className="text-[11px] font-medium text-primary opacity-0 transition-opacity group-hover:opacity-100">
							View details →
						</span>
					</div>
				) : null}
			</Link>
		</Card>
	);
}
