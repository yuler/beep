import {
	CheckCircle2,
	Flame,
	type LucideIcon,
	PauseCircle,
	XCircle,
} from "lucide-react";

import type { Beep } from "@/lib/api/beeps";

type BadgeVariant =
	| "default"
	| "secondary"
	| "outline"
	| "destructive"
	| "ghost";

/**
 * Single source of truth for how a beep status is displayed across the app
 * (list indicators and detail badges).
 *
 * Strict monochrome per DESIGN:
 * active = solid/ink, firing = outline/ink, paused = secondary/gray, completed = outline/gray, cancelled = ghost/gray.
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
		colorClass: "text-ink",
		badgeVariant: "default",
	},
	firing: {
		icon: Flame,
		colorClass: "text-ink",
		badgeVariant: "outline",
	},
	paused: {
		icon: PauseCircle,
		colorClass: "text-gray",
		badgeVariant: "secondary",
	},
	completed: {
		icon: CheckCircle2,
		colorClass: "text-gray",
		badgeVariant: "outline",
	},
	cancelled: {
		icon: XCircle,
		colorClass: "text-gray",
		badgeVariant: "ghost",
	},
};
