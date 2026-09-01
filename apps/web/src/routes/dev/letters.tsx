import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import { ExternalLink, Trash2 } from "lucide-react";
import { useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Button, buttonVariants } from "@/components/ui/button";
import {
	Card,
	CardAction,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { coreAppUrl } from "@/config";
import { ApiError } from "@/lib/api/client";
import {
	clearDevLetters,
	deleteDevLetter,
	fetchDevLetters,
} from "@/lib/api/dev";
import { withAuthRedirects } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";
import * as m from "@/locale/paraglide/messages";

const devRoute = getRouteApi("/dev");

export const Route = createFileRoute("/dev/letters")({
	loader: withAuthRedirects(fetchDevLetters),
	component: LettersPage,
});

function LettersPage() {
	const { account } = devRoute.useRouteContext();
	const router = useRouter();
	const { letters } = Route.useLoaderData();
	const [error, setError] = useState<string | null>(null);
	const [pendingId, setPendingId] = useState<string | null>(null);
	const [clearing, setClearing] = useState(false);

	async function handleDelete(id: string) {
		setPendingId(id);
		setError(null);
		try {
			await deleteDevLetter(id);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : m.dev_delete_letter_failed(),
			);
		} finally {
			setPendingId(null);
		}
	}

	async function handleClear() {
		if (!letters.length) return;
		if (!window.confirm(m.dev_clear_letters_confirm())) return;

		setClearing(true);
		setError(null);
		try {
			await clearDevLetters();
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : m.dev_letters_clear_failed(),
			);
		} finally {
			setClearing(false);
		}
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: m.nav_dev() },
					{ label: m.dev_letters(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{m.dev_letters()}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							{m.dev_letters_description()}
						</p>
					</div>
					<div className="flex flex-wrap gap-2">
						{letters.length > 0 ? (
							<Button
								variant="outline"
								disabled={clearing || pendingId !== null}
								onClick={() => void handleClear()}
							>
								<Trash2 data-icon="inline-start" />
								{m.common_clear_all()}
							</Button>
						) : null}
						<a
							href={coreAppUrl("/letter_opener")}
							target="_blank"
							rel="noreferrer"
							className={cn(buttonVariants({ variant: "outline" }))}
						>
							{m.dev_open_letter()}
							<ExternalLink data-icon="inline-end" />
						</a>
					</div>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				{letters.length === 0 ? (
					<p className="text-sm text-muted-foreground">
						{m.dev_letters_empty()}
					</p>
				) : (
					<ul className="flex flex-col gap-3">
						{letters.map((letter) => {
							const subjectLabel = letter.subject ?? m.common_no_subject();
							return (
								<li key={letter.id}>
									<Card
										size="sm"
										className="transition-colors hover:bg-muted/30"
									>
										<CardHeader className="border-b">
											<CardTitle className="truncate">{subjectLabel}</CardTitle>
											<CardDescription className="truncate">
												{letter.to ? `To ${letter.to}` : letter.id}
											</CardDescription>
											<CardAction className="flex items-center gap-2">
												<a
													href={coreAppUrl(`/letter_opener/${letter.id}`)}
													target="_blank"
													rel="noreferrer"
													aria-label={`${m.dev_open_letter()} ${subjectLabel}`}
													className={cn(
														buttonVariants({ variant: "outline", size: "sm" }),
													)}
												>
													{m.common_open()}
													<ExternalLink data-icon="inline-end" />
												</a>
												<Button
													variant="destructive"
													size="sm"
													disabled={pendingId === letter.id || clearing}
													aria-label={`${m.common_delete()} ${subjectLabel}`}
													onClick={() => void handleDelete(letter.id)}
												>
													<Trash2 data-icon="inline-start" />
													{m.common_delete()}
												</Button>
											</CardAction>
										</CardHeader>
										{(letter.from || letter.sent_at) && (
											<CardContent className="flex flex-wrap gap-x-4 gap-y-1 pt-0 text-xs text-muted-foreground">
												{letter.from ? (
													<span className="min-w-0 truncate">
														<span className="text-foreground/60">
															{m.common_from()}
														</span>{" "}
														{letter.from}
													</span>
												) : null}
												{letter.sent_at ? (
													<span className="tabular-nums">
														{new Date(letter.sent_at).toLocaleString()}
													</span>
												) : null}
											</CardContent>
										)}
									</Card>
								</li>
							);
						})}
					</ul>
				)}
			</div>
		</>
	);
}
