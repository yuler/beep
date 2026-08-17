import {
	createFileRoute,
	getRouteApi,
	Link,
	useRouter,
} from "@tanstack/react-router";

import { CreateBeepForm } from "@/components/beeps/create-beep-form";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardAction,
	CardContent,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { fetchBeeps } from "@/lib/api/beeps";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beeps")({
	loader: withAuthRedirects(({ params }) =>
		fetchBeeps(params?.account_slug ?? ""),
	),
	component: BeepsPage,
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

function BeepsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { beeps } = Route.useLoaderData();

	async function handleCreated() {
		await router.invalidate();
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
					{ label: "Beeps", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						Beeps
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						Reminders scheduled for this workspace.
					</p>
				</div>

				<Card className="max-w-105">
					<CardHeader>
						<CardTitle>
							{beeps.length === 0 ? "Create your first beep" : "New beep"}
						</CardTitle>
					</CardHeader>
					<CardContent>
						{beeps.length === 0 ? (
							<p className="mb-4 text-sm text-muted-foreground">
								No beeps yet — schedule a one-time reminder to get started.
							</p>
						) : null}
						<CreateBeepForm slug={slug} onCreated={handleCreated} />
					</CardContent>
				</Card>

				{beeps.length > 0 ? (
					<ul className="flex flex-col gap-3">
						{beeps.map((beep) => {
							const nextRunAt = beep.next_run_at ?? beep.run_at;
							return (
								<li key={beep.id}>
									<Link
										to="/$account_slug/beeps/$beepId"
										params={{
											account_slug: slug,
											beepId: beep.id,
										}}
										className="block"
									>
										<Card
											size="sm"
											className="transition-colors hover:bg-muted/30"
										>
											<CardHeader>
												<CardTitle className="truncate">
													{beep.message}
												</CardTitle>
												<CardAction>
													<Badge
														variant={STATUS_VARIANT[beep.status] ?? "secondary"}
													>
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
										</Card>
									</Link>
								</li>
							);
						})}
					</ul>
				) : null}
			</div>
		</>
	);
}
