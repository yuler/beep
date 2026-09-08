import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import { Plus, Server } from "lucide-react";
import { useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { RunnerFormDialog } from "@/components/runners/runner-form-dialog";
import { RunnerList } from "@/components/runners/runner-list";
import { RunnerTokenModal } from "@/components/runners/runner-token-modal";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardTitle } from "@/components/ui/card";
import {
	fetchRunners,
	type Runner,
	type RunnerWithToken,
} from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/runners")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		return await fetchRunners(slug);
	}),
	component: RunnersPage,
});

function RunnersPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { runners } = Route.useLoaderData();

	const [isFormOpen, setIsFormOpen] = useState(false);
	const [editingRunner, setEditingRunner] = useState<Runner | null>(null);
	const [revealedRunner, setRevealedRunner] = useState<RunnerWithToken | null>(
		null,
	);

	function handleOpenAdd() {
		setEditingRunner(null);
		setIsFormOpen(true);
	}

	function handleEdit(runner: Runner) {
		setEditingRunner(runner);
		setIsFormOpen(true);
	}

	function handleFormSuccess(saved: Runner | RunnerWithToken) {
		void router.invalidate();
		if ("token" in saved && saved.token) {
			setRevealedRunner(saved as RunnerWithToken);
		}
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: m.nav_runners(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-center justify-between gap-4">
					<div className="flex flex-col gap-1">
						<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
							{m.runners_title()}
						</h1>
						<p className="text-sm text-muted-foreground">
							{m.runners_description()}
						</p>
					</div>

					<Button size="sm" onClick={handleOpenAdd}>
						<Plus data-icon="inline-start" />
						{m.runners_add_runner()}
					</Button>
				</div>

				{runners.length === 0 ? (
					<Card className="flex flex-col items-center justify-center p-8 text-center border-dashed">
						<div className="flex size-12 items-center justify-center rounded-xl bg-primary/10 text-primary mb-3">
							<Server className="size-6" />
						</div>
						<CardTitle className="text-base font-semibold">
							{m.runners_no_runners()}
						</CardTitle>
						<CardDescription className="mt-1 max-w-sm text-xs leading-relaxed">
							{m.runners_no_runners_hint()}
						</CardDescription>
						<Button size="sm" className="mt-4" onClick={handleOpenAdd}>
							<Plus data-icon="inline-start" />
							{m.runners_add_runner()}
						</Button>
					</Card>
				) : (
					<RunnerList
						slug={slug}
						runners={runners}
						onEdit={handleEdit}
						onTokenRevealed={(runnerWithToken) => {
							setRevealedRunner(runnerWithToken);
						}}
						onRefresh={() => {
							void router.invalidate();
						}}
					/>
				)}

				{/* Add / Edit Runner Dialog */}
				<RunnerFormDialog
					slug={slug}
					runner={editingRunner}
					open={isFormOpen}
					onOpenChange={setIsFormOpen}
					onSuccess={handleFormSuccess}
				/>

				{/* Token Setup & Reveal Modal */}
				<RunnerTokenModal
					runner={revealedRunner}
					open={revealedRunner !== null}
					onOpenChange={(open) => {
						if (!open) {
							setRevealedRunner(null);
						}
					}}
				/>
			</div>
		</>
	);
}
