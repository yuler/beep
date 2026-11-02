import type * as React from "react";

import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export type NoticeTone = "default" | "success" | "warning" | "danger";

export interface NoticeProps extends React.ComponentProps<"div"> {
	label?: string;
	error?: boolean;
	tone?: NoticeTone;
}

const NOTICE_BORDER: Record<NoticeTone, string> = {
	default: "border-l border-line",
	success: "border-l-2 border-success",
	warning: "border-l-2 border-warning",
	danger: "border-l-2 border-danger",
};

const NOTICE_LABEL: Record<NoticeTone, string> = {
	default: "text-gray",
	success: "text-success",
	warning: "text-warning",
	danger: "text-danger",
};

export function Notice({
	className,
	label,
	error,
	tone,
	children,
	...props
}: NoticeProps) {
	const resolvedTone: NoticeTone = tone ?? (error ? "danger" : "default");
	const resolvedLabel = label ?? (error ? m.common_error() : undefined);
	return (
		<div
			className={cn(
				"flex items-start gap-2.5 py-2 pl-3 text-sm text-ink",
				NOTICE_BORDER[resolvedTone],
				className,
			)}
			{...props}
		>
			{resolvedLabel && (
				<span
					className={cn(
						"type-kicker shrink-0 select-none",
						NOTICE_LABEL[resolvedTone],
					)}
				>
					{resolvedLabel}
				</span>
			)}
			<div className="flex-1 min-w-0">{children}</div>
		</div>
	);
}
