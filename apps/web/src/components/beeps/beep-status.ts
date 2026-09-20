import {
	CheckCircle2,
	Flame,
	type LucideIcon,
	PauseCircle,
	XCircle,
} from "lucide-react";

import type { Beep } from "@/lib/api/beeps";

type BadgeVariant = "default" | "secondary" | "outline" | "destructive";

/**
 * Single source of truth for how a beep status is displayed across the app
 * (list indicators and detail badges).
 *
 * Beep status represents the schedule lifecycle (active, firing, paused, completed, cancelled).
 * Execution results (succeeded, failed, skipped, expired) belong exclusively to Beep::Run.
 */
export const BEEP_STATUS_META: Record<
	Beep["status"],
	{
		icon: LucideIcon;
		colorClass: string;
		badgeVariant: BadgeVariant;
	}
> = {
	active: {
		icon: CheckCircle2,
		colorClass: "text-emerald-600 dark:text-emerald-400",
		badgeVariant: "default",
	},
	firing: {
		icon: Flame,
		colorClass: "text-amber-600 dark:text-amber-400",
		badgeVariant: "default",
	},
	paused: {
		icon: PauseCircle,
		colorClass: "text-muted-foreground",
		badgeVariant: "secondary",
	},
	completed: {
		icon: CheckCircle2,
		colorClass: "text-muted-foreground",
		badgeVariant: "outline",
	},
	cancelled: {
		icon: XCircle,
		colorClass: "text-muted-foreground",
		badgeVariant: "secondary",
	},
};
