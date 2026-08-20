import { Link } from "@tanstack/react-router";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { BeepRuns } from "@/components/beeps/beep-runs";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardAction,
	CardContent,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import type { Beep } from "@/lib/api/beeps";
import { beepRunAt } from "@/lib/beep-stats";

const STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	active: "default",
	paused: "secondary",
	completed: "outline",
	cancelled: "destructive",
	firing: "default",
};

export function BeepList({
	beeps,
	slug,
	variant = "full",
}: {
	beeps: Beep[];
	slug: string;
	variant?: "compact" | "full";
}) {
	if (beeps.length === 0) {
		return <p className="text-sm text-muted-foreground">No beeps to show.</p>;
	}

	return (
		<ul className="flex flex-col gap-3">
			{beeps.map((beep) => (
				<li key={beep.id}>
					<BeepListCard beep={beep} slug={slug} variant={variant} />
				</li>
			))}
		</ul>
	);
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
	const nextRunAt = beepRunAt(beep);

	return (
		<Card size="sm">
			<Link
				to="/$account_slug/beeps/$beepId"
				params={{
					account_slug: slug,
					beepId: beep.id,
				}}
				className="block rounded-xl transition-colors hover:bg-muted/30"
			>
				<CardHeader>
					<CardTitle className="truncate">{beep.title}</CardTitle>
					<CardAction>
						<Badge variant={STATUS_VARIANT[beep.status] ?? "secondary"}>
							{beep.status}
						</Badge>
					</CardAction>
				</CardHeader>
				<CardContent className="flex flex-wrap gap-x-4 gap-y-1 pt-0 text-xs text-muted-foreground">
					{nextRunAt ? (
						<span className="tabular-nums">
							Next: {new Date(nextRunAt).toLocaleString()}
						</span>
					) : null}
					<span>{beep.timezone}</span>
				</CardContent>
			</Link>
			{variant === "full" && beep.body ? (
				<CardContent>
					<BeepMarkdown source={beep.body} />
				</CardContent>
			) : null}
			{variant === "full" ? (
				<CardContent>
					<BeepRuns runs={beep.runs} />
				</CardContent>
			) : null}
		</Card>
	);
}
