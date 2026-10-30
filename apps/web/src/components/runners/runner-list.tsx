import { Link } from "@tanstack/react-router";
import {
	Clock,
	Edit,
	KeyRound,
	MoreHorizontal,
	Server,
	Trash2,
} from "lucide-react";
import { useState } from "react";
import { confirm } from "@/components/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Notice } from "@/components/ui/notice";
import { ApiError } from "@/lib/api/client";
import {
	deleteRunner,
	type Runner,
	type RunnerWithToken,
	regenerateRunnerToken,
} from "@/lib/api/runners";
import { translateError } from "@/lib/i18n-labels";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

interface RunnerListProps {
	slug: string;
	runners: Runner[];
	onEdit: (runner: Runner) => void;
	onTokenRevealed: (runnerWithToken: RunnerWithToken) => void;
	onRefresh: () => void;
}

export function getRunnerStatusBadge(runner: Runner) {
	const isHealthyOnline = runner.is_online ?? runner.status === "online";
	if (isHealthyOnline && runner.status === "online") {
		return (
			<Badge variant="default" className="gap-1.5">
				<span className="size-1.5 bg-solid-ink" />
				{m.runners_status_online()}
			</Badge>
		);
	}

	return (
		<Badge variant="outline" className="gap-1.5 text-gray">
			<span className="size-1.5 bg-gray" />
			{m.runners_status_offline()}
		</Badge>
	);
}

export function formatRunnerLastSeen(lastSeenAt: string | null | undefined) {
	if (!lastSeenAt) return m.runners_never_seen();
	return m.runners_last_seen({
		time: new Date(lastSeenAt).toLocaleString(),
	});
}

export function RunnerList({
	slug,
	runners,
	onEdit,
	onTokenRevealed,
	onRefresh,
}: RunnerListProps) {
	const [actionError, setActionError] = useState<string | null>(null);
	const [busyRunnerId, setBusyRunnerId] = useState<string | null>(null);

	async function handleDelete(runner: Runner) {
		if (!(await confirm(m.runners_delete_confirm({ name: runner.name })))) {
			return;
		}

		setBusyRunnerId(runner.id);
		setActionError(null);
		try {
			await deleteRunner(slug, runner.id);
			onRefresh();
		} catch (err) {
			setActionError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_delete_failed(),
			);
		} finally {
			setBusyRunnerId(null);
		}
	}

	async function handleRegenerateToken(runner: Runner) {
		if (
			!(await confirm(
				m.runners_regenerate_token_confirm({ name: runner.name }),
			))
		) {
			return;
		}

		setBusyRunnerId(runner.id);
		setActionError(null);
		try {
			const res = await regenerateRunnerToken(slug, runner.id);
			onTokenRevealed({ ...runner, ...res });
			onRefresh();
		} catch (err) {
			setActionError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_token_reveal_failed(),
			);
		} finally {
			setBusyRunnerId(null);
		}
	}

	if (runners.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center p-8 border border-line text-center text-gray">
				<Server className="size-8 mb-2 text-gray" />
				<p className="type-app-title text-ink font-medium">
					{m.runners_no_runners()}
				</p>
				<p className="type-meta text-gray mt-1 max-w-sm">
					{m.runners_empty_state_description()}
				</p>
			</div>
		);
	}

	return (
		<div className="flex flex-col gap-4">
			{actionError ? (
				<Notice error label="出错" role="alert">
					{actionError}
				</Notice>
			) : null}

			<div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
				{runners.map((runner) => {
					const isBusy = busyRunnerId === runner.id;
					return (
						<Card
							key={runner.id}
							className={cn(
								"relative flex flex-col justify-between border-t border-line sm:border sm:border-line bg-transparent transition-colors hover:bg-wash",
								isBusy && "opacity-60 pointer-events-none",
							)}
						>
							<CardHeader className="pb-3">
								<div className="flex items-start justify-between gap-2">
									<Link
										to="/$account_slug/runners/$runnerId"
										params={{ account_slug: slug, runnerId: runner.id }}
										className="flex items-center gap-2.5 min-w-0"
									>
										<div className="flex size-8 shrink-0 items-center justify-center border border-line text-ink">
											<Server className="size-4" />
										</div>
										<div className="min-w-0 flex-1">
											<CardTitle className="truncate text-base font-medium text-ink">
												{runner.name}
											</CardTitle>
											<CardDescription className="font-mono text-xs text-gray truncate">
												{runner.masked_token}
											</CardDescription>
										</div>
									</Link>

									<div className="flex items-center gap-1.5">
										{getRunnerStatusBadge(runner)}
										<DropdownMenu>
											<DropdownMenuTrigger
												render={
													<Button
														variant="ghost"
														size="icon"
														className="size-8 min-h-0 sm:min-h-0"
														aria-label={`Actions for runner ${runner.name}`}
													>
														<MoreHorizontal className="size-4" />
													</Button>
												}
											/>
											<DropdownMenuContent align="end" className="w-44">
												<DropdownMenuItem onClick={() => onEdit(runner)}>
													<Edit className="size-3.5 text-gray shrink-0" />
													<span>{m.runners_edit_runner()}</span>
												</DropdownMenuItem>
												<DropdownMenuItem
													onClick={() => void handleRegenerateToken(runner)}
												>
													<KeyRound className="size-3.5 text-gray shrink-0" />
													<span>{m.runners_regenerate_token()}</span>
												</DropdownMenuItem>
												<DropdownMenuSeparator />
												<DropdownMenuItem
													variant="destructive"
													onClick={() => void handleDelete(runner)}
												>
													<Trash2 className="size-3.5 shrink-0" />
													<span>{m.common_delete()}</span>
												</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									</div>
								</div>
							</CardHeader>

							<CardContent className="flex flex-col gap-3.5 type-meta text-gray pt-0">
								{/* Tags */}
								{runner.tags && runner.tags.length > 0 ? (
									<div className="flex flex-wrap items-center gap-1.5 min-h-[1.5rem]">
										{runner.tags.map((tag) => (
											<Badge
												key={tag}
												variant="outline"
												className="font-mono text-[10px]"
											>
												{tag}
											</Badge>
										))}
									</div>
								) : null}

								{/* Metadata Grid */}
								<div className="grid grid-cols-2 gap-2 border border-line p-2.5 font-mono text-[11px] text-gray">
									<div>
										<span className="type-kicker text-gray block text-[10px]">
											{m.runners_version()}
										</span>
										<span className="text-ink truncate block">
											{runner.version || m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="type-kicker text-gray block text-[10px]">
											{m.runners_os_arch()}
										</span>
										<span className="text-ink truncate block">
											{runner.os && runner.arch
												? `${runner.os}/${runner.arch}`
												: m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="type-kicker text-gray block text-[10px]">
											{m.runners_hostname()}
										</span>
										<span className="text-ink truncate block">
											{runner.hostname || m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="type-kicker text-gray block text-[10px]">
											{m.runners_jobs_count()}
										</span>
										<span className="text-ink block font-medium">
											{runner.jobs_count ?? 0}
										</span>
									</div>
								</div>

								{/* Last seen footer */}
								<div className="flex items-center gap-1.5 text-[11px] text-gray">
									<Clock className="size-3.5 shrink-0" />
									<span className="truncate">
										{formatRunnerLastSeen(runner.last_seen_at)}
									</span>
								</div>
							</CardContent>
						</Card>
					);
				})}
			</div>
		</div>
	);
}
