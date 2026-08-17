import { createFileRoute, getRouteApi, notFound } from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beeps_/$beepId")({
	loader: withAuthRedirects(async ({ params }) => {
		try {
			return await fetchBeep(params.account_slug ?? "", params.beepId ?? "");
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: BeepDetailPage,
});

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

function formatWhen(value: string | null) {
	if (!value) return "—";
	return new Date(value).toLocaleString();
}

function BeepDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const beep = Route.useLoaderData();

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{
						label: "Beeps",
						to: "/$account_slug/beeps",
						params: { account_slug: slug },
					},
					{ label: beep.message, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{beep.message}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							{beep.kind} · {beep.timezone}
						</p>
					</div>
					<Badge variant={STATUS_VARIANT[beep.status] ?? "secondary"}>
						{beep.status}
					</Badge>
				</div>

				<Card className="max-w-lg">
					<CardHeader>
						<CardTitle>Schedule</CardTitle>
					</CardHeader>
					<CardContent className="flex flex-col gap-3 text-sm">
						<DetailRow label="Run at" value={formatWhen(beep.run_at)} />
						<DetailRow label="Next" value={formatWhen(beep.next_run_at)} />
						<DetailRow label="Last" value={formatWhen(beep.last_run_at)} />
						<DetailRow label="Created" value={formatWhen(beep.created_at)} />
					</CardContent>
				</Card>
			</div>
		</>
	);
}

function DetailRow({ label, value }: { label: string; value: string }) {
	return (
		<div className="flex justify-between gap-4">
			<span className="text-muted-foreground">{label}</span>
			<span className="text-right tabular-nums">{value}</span>
		</div>
	);
}
