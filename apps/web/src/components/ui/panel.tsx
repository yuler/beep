import type * as React from "react";

import { cn } from "@/lib/utils";

export interface PanelProps extends React.ComponentProps<"div"> {
	borderAll?: boolean;
}

export function Panel({
	className,
	borderAll,
	children,
	...props
}: PanelProps) {
	return (
		<div
			className={cn(
				"bg-transparent text-ink transition-colors",
				borderAll
					? "border border-line"
					: "border-t border-line md:border md:border-line",
				className,
			)}
			{...props}
		>
			{children}
		</div>
	);
}

export function PanelHeader({
	className,
	children,
	...props
}: React.ComponentProps<"div">) {
	return (
		<div
			className={cn(
				"flex items-center justify-between p-4 border-b border-line",
				className,
			)}
			{...props}
		>
			{children}
		</div>
	);
}

export function PanelContent({
	className,
	children,
	...props
}: React.ComponentProps<"div">) {
	return (
		<div className={cn("p-4", className)} {...props}>
			{children}
		</div>
	);
}
