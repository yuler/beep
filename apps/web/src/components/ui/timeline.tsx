import type * as React from "react";

import { cn } from "@/lib/utils";

export function Timeline({
	className,
	children,
	...props
}: React.ComponentProps<"div">) {
	return (
		<div className={cn("flex flex-col w-full", className)} {...props}>
			{children}
		</div>
	);
}

export interface TimelineItemProps extends React.ComponentProps<"div"> {
	bordered?: boolean;
}

export function TimelineItem({
	className,
	bordered,
	children,
	...props
}: TimelineItemProps) {
	return (
		<div
			className={cn(
				"w-full border-t border-line py-2.5 px-3 text-sm text-ink flex items-start gap-3 transition-colors hover:bg-wash/50",
				bordered && "border-l border-l-line",
				className,
			)}
			{...props}
		>
			{children}
		</div>
	);
}
