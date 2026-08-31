import {
	AlertCircle,
	CheckCircle2,
	Flame,
	type LucideIcon,
	PauseCircle,
} from "lucide-react";

import type { Beep } from "@/lib/api/beeps";
import type { TranslationKey } from "@/lib/i18n";

type BadgeVariant = "default" | "secondary" | "outline" | "destructive";

/**
 * Single source of truth for how a beep status is displayed across the app
 * (list indicators and detail badges).
 */
export const BEEP_STATUS_META: Record<
	Beep["status"],
	{
		labelKey: TranslationKey;
		icon: LucideIcon;
		colorClass: string;
		badgeVariant: BadgeVariant;
	}
> = {
	active: {
		labelKey: "status.beep.active",
		icon: CheckCircle2,
		colorClass: "text-emerald-600 dark:text-emerald-400",
		badgeVariant: "default",
	},
	firing: {
		labelKey: "status.beep.firing",
		icon: Flame,
		colorClass: "text-amber-600 dark:text-amber-400",
		badgeVariant: "default",
	},
	paused: {
		labelKey: "status.beep.paused",
		icon: PauseCircle,
		colorClass: "text-muted-foreground",
		badgeVariant: "secondary",
	},
	completed: {
		labelKey: "status.beep.completed",
		icon: CheckCircle2,
		colorClass: "text-muted-foreground",
		badgeVariant: "outline",
	},
	cancelled: {
		labelKey: "status.beep.cancelled",
		icon: AlertCircle,
		colorClass: "text-destructive",
		badgeVariant: "destructive",
	},
};
