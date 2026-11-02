import { cn } from "@/lib/utils";

/**
 * Semantic tones only express outcome / health. Lifecycle states
 * (active, paused, completed…) stay `outline` / `muted`; `live` marks
 * "executing now" with a pulsing monochrome dot.
 */
const STATUS_PILL_STYLES = {
	success: "bg-success-soft text-success",
	warning: "bg-warning-soft text-warning",
	danger: "bg-danger-soft text-danger",
	live: "border border-ink bg-transparent text-ink",
	muted: "border border-line bg-transparent text-gray",
	outline: "border border-line bg-transparent text-ink",
} as const;

export type StatusPillTone = keyof typeof STATUS_PILL_STYLES;

export function StatusPill({
	label,
	tone = "muted",
	className,
}: {
	label: string;
	tone?: StatusPillTone;
	className?: string;
}) {
	return (
		<span
			className={cn(
				"inline-flex h-5 items-center gap-1.5 rounded-sm px-2 text-xs font-normal whitespace-nowrap",
				STATUS_PILL_STYLES[tone],
				className,
			)}
		>
			{tone === "live" ? <StatusDot tone="live" /> : null}
			{label}
		</span>
	);
}

const STATUS_DOT_STYLES = {
	success: "bg-success",
	warning: "bg-warning",
	danger: "bg-danger",
	live: "bg-ink status-dot-pulse",
	muted: "bg-gray-subtle",
} as const;

export type StatusDotTone = keyof typeof STATUS_DOT_STYLES;

export function StatusDot({
	tone = "muted",
	className,
}: {
	tone?: StatusDotTone;
	className?: string;
}) {
	return (
		<span
			aria-hidden="true"
			className={cn(
				"inline-block size-1.5 shrink-0 rounded-full",
				STATUS_DOT_STYLES[tone],
				className,
			)}
		/>
	);
}

export function ProgressBar({
	value,
	className,
}: {
	value: number;
	className?: string;
}) {
	const clamped = Math.max(0, Math.min(100, value));
	return (
		<div className={cn("flex min-w-24 items-center gap-2", className)}>
			<span className="w-8 text-right text-xs tabular-nums text-gray">
				{clamped}%
			</span>
			<div className="h-1 flex-1 overflow-hidden rounded-full bg-line">
				<div
					className="h-full rounded-full bg-ink transition-[width]"
					style={{ width: `${clamped}%` }}
				/>
			</div>
		</div>
	);
}
