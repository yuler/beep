import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import { Trash2 } from "lucide-react";
import { useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { BeepRuns } from "@/components/beeps/beep-runs";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { deleteBeep, fetchBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beeps_/$beepId")({
	loader: withAuthRedirects(async ({ params }) => {
		try {
			return await fetchBeep(params?.account_slug ?? "", params?.beepId ?? "");
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
	const router = useRouter();
	const beep = Route.useLoaderData();
	const [deleting, setDeleting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function handleDelete() {
		if (!window.confirm(`Delete "${beep.title}"? This cannot be undone.`)) {
			return;
		}

		setDeleting(true);
		setError(null);
		try {
			await deleteBeep(slug, beep.id);
			await router.navigate({
				to: "/$account_slug/beeps",
				params: { account_slug: slug },
			});
			await router.invalidate();
		} catch (err) {
			setDeleting(false);
			setError(
				err instanceof ApiError ? err.message : "Failed to delete beep.",
			);
		}
	}

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
					{ label: beep.title, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{beep.title}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							{beep.kind} · {beep.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						<Badge variant={STATUS_VARIANT[beep.status] ?? "secondary"}>
							{beep.status}
						</Badge>
						<Button
							variant="destructive"
							size="sm"
							disabled={deleting}
							aria-label={`Delete beep ${beep.title}`}
							onClick={() => void handleDelete()}
						>
							<Trash2 data-icon="inline-start" />
							{deleting ? "Deleting…" : "Delete"}
						</Button>
					</div>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				{beep.body ? (
					<Card className="max-w-lg">
						<CardHeader>
							<CardTitle>Body</CardTitle>
						</CardHeader>
						<CardContent>
							<BeepMarkdown source={beep.body} />
						</CardContent>
					</Card>
				) : null}

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

				<div className="max-w-lg">
					<BeepRuns runs={beep.runs} />
				</div>
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
