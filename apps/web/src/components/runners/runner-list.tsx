import { Link } from "@tanstack/react-router";
import {
	Clock,
	Edit,
	KeyRound,
	MoreHorizontal,
	Server,
	ShieldCheck,
	Terminal,
	Trash2,
} from "lucide-react";
import { useState } from "react";
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
		if (!window.confirm(m.runners_delete_confirm({ name: runner.name }))) {
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
			!window.confirm(m.runners_regenerate_token_confirm({ name: runner.name }))
		) {
			return;
		}

		setBusyRunnerId(runner.id);
		setActionError(null);
		try {
			const res = await regenerateRunnerToken(slug, runner.id);
			onTokenRevealed(res.runner);
			onRefresh();
		} catch (err) {
			setActionError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_regenerate_failed(),
			);
		} finally {
			setBusyRunnerId(null);
		}
	}

	function getStatusBadge(runner: Runner) {
		const isHealthyOnline = runner.is_online ?? runner.status !== "offline";
		if (!isHealthyOnline || runner.status === "offline") {
			return (
				<Badge
					variant="outline"
					className="gap-1.5 border-zinc-500/30 text-muted-foreground"
				>
					<span className="size-1.5 rounded-full bg-zinc-400" />
					{m.runners_status_offline()}
				</Badge>
			);
		}

		if (runner.status === "online") {
			return (
				<Badge
					variant="outline"
					className="gap-1.5 border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
				>
					<span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
					{m.runners_status_online()}
				</Badge>
			);
		}

		return (
			<Badge
				variant="outline"
				className="gap-1.5 border-sky-500/30 bg-sky-500/10 text-sky-600 dark:text-sky-400"
			>
				<span className="size-1.5 rounded-full bg-sky-500" />
				{m.runners_status_idle()}
			</Badge>
		);
	}

	function formatLastSeen(lastSeenAt: string | null | undefined) {
		if (!lastSeenAt) return m.runners_never_seen();
		return m.runners_last_seen({
			time: new Date(lastSeenAt).toLocaleString(),
		});
	}

	return (
		<div className="flex flex-col gap-4">
			{actionError ? (
				<p className="text-sm text-destructive" role="alert">
					{actionError}
				</p>
			) : null}

			<div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
				{runners.map((runner) => {
					const isBusy = busyRunnerId === runner.id;
					return (
						<Card
							key={runner.id}
							className={cn(
								"relative flex flex-col justify-between transition-all duration-200 hover:ring-foreground/15 hover:shadow-xs",
								isBusy && "opacity-60 pointer-events-none",
							)}
						>
							<CardHeader className="pb-3">
								<div className="flex items-start justify-between gap-2">
									<Link
										to="/$account_slug/runners/$runnerId"
										params={{ account_slug: slug, runnerId: runner.id }}
										className="flex items-center gap-2.5 min-w-0"
										<div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
											<Server className="size-4.5" />
										</div>
										<div className="min-w-0 flex-1">
											<CardTitle className="truncate text-base font-semibold">
												{runner.name}
											</CardTitle>
											<CardDescription className="font-mono text-xs text-muted-foreground truncate">
												{runner.token_prefix}••••
											</CardDescription>
										</div>
									</Link>

									<div className="flex items-center gap-1.5">
										{getStatusBadge(runner)}
										<DropdownMenu>
											<DropdownMenuTrigger
												render={
													<Button
														variant="ghost"
														size="icon"
														className="size-8"
														aria-label={`Actions for runner ${runner.name}`}
													>
														<MoreHorizontal className="size-4" />
													</Button>
												}
											/>
											<DropdownMenuContent align="end" className="w-44">
												<DropdownMenuItem onClick={() => onEdit(runner)}>
													<Edit data-icon="inline-start" />
													<span>{m.runners_edit_runner()}</span>
												</DropdownMenuItem>
												<DropdownMenuItem
													onClick={() => void handleRegenerateToken(runner)}
												>
													<KeyRound data-icon="inline-start" />
													<span>{m.runners_regenerate_token()}</span>
												</DropdownMenuItem>
												<DropdownMenuSeparator />
												<DropdownMenuItem
													variant="destructive"
													onClick={() => void handleDelete(runner)}
												>
													<Trash2 data-icon="inline-start" />
													<span>{m.common_delete()}</span>
												</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									</div>
								</div>
							</CardHeader>

							<CardContent className="flex flex-col gap-3.5 text-xs text-muted-foreground pt-0">
								{/* Tags & Execution Badge */}
								<div className="flex flex-wrap items-center gap-1.5 min-h-[1.5rem]">
									{runner.allow_exec ? (
										<Badge
											variant="secondary"
											className="gap-1 text-[10px] font-normal"
										>
											<Terminal className="size-3 text-amber-500" />
											<span>{m.runners_exec_enabled()}</span>
										</Badge>
									) : (
										<Badge
											variant="outline"
											className="gap-1 text-[10px] font-normal text-muted-foreground"
										>
											<ShieldCheck className="size-3 text-emerald-500" />
											<span>{m.runners_scripts_locked()}</span>
										</Badge>
									)}

									{runner.tags && runner.tags.length > 0
										? runner.tags.map((tag) => (
												<Badge
													key={tag}
													variant="outline"
													className="font-mono text-[10px] bg-muted/30"
												>
													{tag}
												</Badge>
											))
										: null}
								</div>

								{/* Metadata Grid */}
								<div className="grid grid-cols-2 gap-2 rounded-lg bg-muted/40 p-2.5 font-mono text-[11px]">
									<div>
										<span className="text-muted-foreground block text-[10px]">
											{m.runners_version()}
										</span>
										<span className="text-foreground truncate block">
											{runner.version || m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="text-muted-foreground block text-[10px]">
											{m.runners_os_arch()}
										</span>
										<span className="text-foreground truncate block">
											{runner.os && runner.arch
												? `${runner.os}/${runner.arch}`
												: m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="text-muted-foreground block text-[10px]">
											{m.runners_hostname()}
										</span>
										<span className="text-foreground truncate block">
											{runner.hostname || m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="text-muted-foreground block text-[10px]">
											{m.runners_jobs_count()}
										</span>
										<span className="text-foreground block font-medium">
											{runner.jobs_count ?? 0}
										</span>
									</div>
								</div>

								{/* Last seen footer */}
								<div className="flex items-center gap-1.5 text-[11px]">
									<Clock className="size-3.5 shrink-0" />
									<span className="truncate">
										{formatLastSeen(runner.last_seen_at)}
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
