import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

import { CreateBeepForm } from "@/components/beeps/create-beep-form";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { fetchBeeps } from "@/lib/api/beeps";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/")({
	loader: withAuthRedirects(({ params }) =>
		fetchBeeps(params?.account_slug ?? ""),
	),
	component: AccountHomePage,
});

function AccountHomePage() {
	const router = useRouter();
	const { me } = accountRoute.useRouteContext();
	const { account_slug: slug } = accountRoute.useParams();
	const account = me.accounts.find((item) => item.slug === slug);
	const { beeps } = Route.useLoaderData();

	return (
		<>
			<DashboardHeader breadcrumbs={[{ label: "Home", isCurrentPage: true }]} />

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{account?.name ?? "Account"}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						One-time reminders for /{slug}.
					</p>
				</div>

				{beeps.length === 0 ? (
					<Card>
						<CardHeader>
							<CardTitle>hi, create your first beep</CardTitle>
							<CardDescription>
								A short message and a time. We’ll remind you once.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<CreateBeepForm
								slug={slug}
								onCreated={() => router.invalidate()}
							/>
						</CardContent>
					</Card>
				) : (
					<>
						<Card>
							<CardHeader>
								<CardTitle>New beep</CardTitle>
								<CardDescription>
									Message and a time for a one-time reminder.
								</CardDescription>
							</CardHeader>
							<CardContent>
								<CreateBeepForm
									slug={slug}
									onCreated={() => router.invalidate()}
								/>
							</CardContent>
						</Card>

						<Card>
							<CardHeader>
								<CardTitle>Beeps</CardTitle>
								<CardDescription>Newest first.</CardDescription>
							</CardHeader>
							<CardContent className="flex flex-col gap-3">
								{beeps.map((beep) => (
									<div
										key={beep.id}
										className="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2"
									>
										<div className="min-w-0">
											<p className="truncate font-medium">{beep.message}</p>
											<p className="truncate text-xs text-muted-foreground">
												{beep.run_at
													? new Date(beep.run_at).toLocaleString()
													: "No time"}
											</p>
										</div>
										<Badge variant="secondary">{beep.status}</Badge>
									</div>
								))}
							</CardContent>
						</Card>
					</>
				)}
			</div>
		</>
	);
}
