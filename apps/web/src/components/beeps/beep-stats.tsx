import { Activity, Calendar, Zap } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import type { Beep } from "@/lib/api/beeps";
import { beepStats } from "@/lib/beep-stats";
import * as m from "@/locale/paraglide/messages";

export function BeepStats({ beeps }: { beeps: Beep[] }) {
	const stats = beepStats(beeps);

	return (
		<div className="grid grid-cols-3 gap-3">
			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex items-center justify-between p-3.5 sm:p-4">
					<div className="flex flex-col gap-0.5">
						<span className="text-xs font-medium text-muted-foreground">
							{m.beeps_active()}
						</span>
						<span className="font-heading text-xl font-bold tabular-nums tracking-tight sm:text-2xl">
							{stats.active}
						</span>
					</div>
					<div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary sm:size-9">
						<Activity className="size-4" />
					</div>
				</CardContent>
			</Card>

			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex items-center justify-between p-3.5 sm:p-4">
					<div className="flex flex-col gap-0.5">
						<span className="text-xs font-medium text-muted-foreground">
							{m.beeps_due_today()}
						</span>
						<span className="font-heading text-xl font-bold tabular-nums tracking-tight sm:text-2xl">
							{stats.dueToday}
						</span>
					</div>
					<div className="flex size-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 sm:size-9">
						<Calendar className="size-4" />
					</div>
				</CardContent>
			</Card>

			<Card
				size="sm"
				className="relative overflow-hidden bg-card/60 transition-colors hover:bg-card"
			>
				<CardContent className="flex items-center justify-between p-3.5 sm:p-4">
					<div className="flex flex-col gap-0.5">
						<span className="text-xs font-medium text-muted-foreground">
							{m.beeps_firing()}
						</span>
						<span className="font-heading text-xl font-bold tabular-nums tracking-tight sm:text-2xl">
							{stats.firing}
						</span>
					</div>
					<div className="flex size-8 items-center justify-center rounded-lg bg-violet-500/10 text-violet-600 dark:text-violet-400 sm:size-9">
						<Zap className="size-4" />
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
