import {
	AlertCircle,
	CheckCircle2,
	Flame,
	type LucideIcon,
	PauseCircle,
} from "lucide-react";

import type { Beep } from "@/lib/api/beeps";

type BadgeVariant = "default" | "secondary" | "outline" | "destructive";

/**
 * Single source of truth for how a beep status is displayed across the app
 * (list indicators and detail badges).
 */
export const BEEP_STATUS_META: Record<
	Beep["status"],
	{
		label: string;
		icon: LucideIcon;
		colorClass: string;
		badgeVariant: BadgeVariant;
	}
> = {
	active: {
		label: "Active",
		icon: CheckCircle2,
		colorClass: "text-emerald-600 dark:text-emerald-400",
		badgeVariant: "default",
	},
	firing: {
		label: "Firing",
		icon: Flame,
		colorClass: "text-amber-600 dark:text-amber-400",
		badgeVariant: "default",
	},
	paused: {
		label: "Paused",
		icon: PauseCircle,
		colorClass: "text-muted-foreground",
		badgeVariant: "secondary",
	},
	completed: {
		label: "Completed",
		icon: CheckCircle2,
		colorClass: "text-muted-foreground",
		badgeVariant: "outline",
	},
	cancelled: {
		label: "Cancelled",
		icon: AlertCircle,
		colorClass: "text-destructive",
		badgeVariant: "destructive",
	},
};
