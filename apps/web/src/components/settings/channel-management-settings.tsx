import {
	Check,
	ChevronDown,
	KeyRound,
	Loader2,
	Plus,
	Send,
	Terminal,
	Trash2,
} from "lucide-react";
import { type FormEvent, useEffect, useRef, useState } from "react";
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
import { CopyableCode, useCopyToClipboard } from "@/components/ui/copy-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	type Channel,
	createChannel,
	deleteChannel,
	fetchChannels,
	testChannel,
} from "@/lib/api/channels";
import { ApiError } from "@/lib/api/client";
import { translateError } from "@/lib/i18n-labels";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function ChannelManagementSettings({ slug }: { slug: string }) {
	const [channels, setChannels] = useState<Channel[]>([]);
	const [loading, setLoading] = useState(true);
	const [showManual, setShowManual] = useState(false);
	const [name, setName] = useState("");
	const [creating, setCreating] = useState(false);
	const [createdChannel, setCreatedChannel] = useState<Channel | null>(null);
	const [testingId, setTestingId] = useState<string | null>(null);
	const [deletingId, setDeletingId] = useState<string | null>(null);
	const [testSentId, setTestSentId] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const { copied, copy } = useCopyToClipboard();
	const { copied: connectCopied, copy: copyConnect } = useCopyToClipboard();
	const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	useEffect(() => {
		return () => {
			if (timerRef.current) clearTimeout(timerRef.current);
		};
	}, []);

	useEffect(() => {
		setLoading(true);
		setCreatedChannel(null);
		setError(null);
		let cancelled = false;
		void (async () => {
			try {
				const data = await fetchChannels(slug, "cli");
				if (cancelled) return;
				setChannels(data.filter((ch) => ch.kind === "cli"));
			} catch (err) {
				if (cancelled) return;
				setError(err instanceof ApiError ? err.message : translateError(err));
			} finally {
				if (!cancelled) setLoading(false);
			}
		})();
		return () => {
			cancelled = true;
		};
	}, [slug]);

	async function handleCreate(e: FormEvent) {
		e.preventDefault();
		if (!name.trim()) return;

		setError(null);
		setCreating(true);
		try {
			const ch = await createChannel(slug, {
				name: name.trim(),
				kind: "cli",
			});
			setCreatedChannel(ch);
			setName("");
			const data = await fetchChannels(slug, "cli");
			setChannels(data.filter((c) => c.kind === "cli"));
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setCreating(false);
		}
	}

	async function handleDelete(id: string) {
		if (deletingId || testingId) return;
		if (!(await confirm("Are you sure you want to remove this channel?")))
			return;
		setDeletingId(id);
		try {
			await deleteChannel(slug, id);
			const data = await fetchChannels(slug, "cli");
			setChannels(data.filter((ch) => ch.kind === "cli"));
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setDeletingId(null);
		}
	}

	async function handleTest(id: string) {
		if (testingId || deletingId) return;
		setError(null);
		setTestingId(id);
		try {
			await testChannel(slug, id);
			setTestSentId(id);
			if (timerRef.current) clearTimeout(timerRef.current);
			timerRef.current = setTimeout(() => {
				setTestSentId((current) => (current === id ? null : current));
			}, 3000);
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setTestingId(null);
		}
	}

	return (
		<Card className="max-w-2xl">
			<CardHeader>
				<CardTitle className="flex items-center gap-2">
					<Terminal className="size-5" />
					{m.channels_cli_title()}
				</CardTitle>
				<CardDescription>{m.channels_cli_desc()}</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-6">
				{/* Recommended: CLI Connect */}
				<div className="rounded-lg border bg-muted/20 p-4 shadow-2xs">
					<div className="flex items-center justify-between gap-2">
						<div className="flex items-center gap-2">
							<Terminal className="size-4 text-primary" />
							<span className="font-semibold text-sm">
								{m.channels_connect_title()}
							</span>
						</div>
						<Badge
							variant="outline"
							className="border-primary/30 bg-primary/10 text-primary text-[10px]"
						>
							{m.runners_tab_recommended()}
						</Badge>
					</div>
					<p className="mt-1.5 text-xs text-muted-foreground">
						{m.channels_connect_desc()}
					</p>
					<div className="mt-2.5">
						<CopyableCode
							code="beep channel connect"
							copied={connectCopied}
							onCopy={() => copyConnect("beep channel connect")}
							label={m.runners_copy()}
						/>
					</div>
					<p className="mt-2 text-[11px] text-muted-foreground">
						💡 {m.channels_connect_hint()}
					</p>
				</div>

				{/* Manual Token / Headless Toggle & Form */}
				<div className="flex flex-col gap-3">
					<div className="flex items-center">
						<Button
							type="button"
							variant="ghost"
							size="xs"
							className="h-7 text-xs text-muted-foreground hover:text-foreground gap-1.5 px-2"
							onClick={() => setShowManual((prev) => !prev)}
						>
							<KeyRound className="size-3.5" />
							<span>
								{showManual
									? m.channels_manual_hide()
									: m.channels_manual_button()}
							</span>
							<ChevronDown
								className={cn(
									"size-3.5 transition-transform",
									showManual && "rotate-180",
								)}
							/>
						</Button>
					</div>

					{showManual ? (
						<div className="rounded-lg border border-dashed p-4 bg-muted/10 flex flex-col gap-3">
							<p className="text-xs text-muted-foreground">
								{m.channels_manual_desc()}
							</p>
							<form onSubmit={handleCreate} className="flex items-end gap-3">
								<div className="flex-1">
									<Label htmlFor="channel-name" className="text-xs">
										{m.channels_name_label()}
									</Label>
									<Input
										id="channel-name"
										placeholder={m.channels_name_placeholder()}
										value={name}
										onChange={(e) => setName(e.target.value)}
										disabled={creating}
									/>
								</div>
								<Button
									type="submit"
									disabled={creating || !name.trim()}
									size="sm"
								>
									<Plus className="size-3.5" />
									{m.channels_create_token()}
								</Button>
							</form>
						</div>
					) : null}
				</div>

				{createdChannel?.token ? (
					<div className="rounded-lg border border-primary/40 bg-primary/5 p-4 text-sm">
						<div className="flex items-center justify-between gap-2">
							<div className="font-semibold text-foreground">
								{m.channels_token_banner_title({ name: createdChannel.name })}
							</div>
							<Button
								type="button"
								variant="ghost"
								size="sm"
								onClick={() => setCreatedChannel(null)}
							>
								{m.channels_dismiss()}
							</Button>
						</div>
						<p className="mt-1 text-xs text-muted-foreground">
							{m.channels_token_banner_desc()}
						</p>
						<div className="mt-2">
							<CopyableCode
								code={createdChannel.token}
								copied={copied}
								onCopy={() => copy(createdChannel.token ?? "")}
								label={m.runners_copy()}
							/>
						</div>
						<div className="mt-2 text-xs text-muted-foreground">
							{m.channels_token_banner_run()}
						</div>
					</div>
				) : null}

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				<div className="flex flex-col divide-y rounded-lg border">
					{loading ? (
						<div className="p-4 text-center text-sm text-muted-foreground">
							{m.channels_loading()}
						</div>
					) : channels.length === 0 ? (
						<div className="p-4 text-center text-sm text-muted-foreground">
							{m.channels_empty()}
						</div>
					) : (
						channels.map((ch) => (
							<div
								key={ch.id}
								className="flex items-center justify-between p-3.5"
							>
								<div className="flex flex-col gap-1">
									<div className="flex items-center gap-2 font-medium text-sm">
										{ch.name}
										{ch.is_online ? (
											<Badge
												variant="outline"
												className="gap-1.5 border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[10px]"
											>
												<span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
												{m.channels_active()}
											</Badge>
										) : (
											<Badge
												variant="outline"
												className="gap-1.5 border-zinc-500/30 text-muted-foreground text-[10px]"
											>
												<span className="size-1.5 rounded-full bg-zinc-400" />
												{m.channels_offline()}
											</Badge>
										)}
									</div>
									<div className="flex flex-wrap items-center gap-x-2 text-xs text-muted-foreground">
										<span>Token: {ch.masked_token}</span>
										<span>•</span>
										<span>
											{ch.last_seen_at
												? m.channels_last_active({
														time: new Date(ch.last_seen_at).toLocaleString(),
													})
												: m.channels_never_connected()}
										</span>
									</div>
								</div>
								<div className="flex items-center gap-2">
									<Button
										type="button"
										size="sm"
										variant="outline"
										className="h-8 gap-1.5 px-2.5 text-xs"
										disabled={testingId !== null || deletingId !== null}
										onClick={() => handleTest(ch.id)}
									>
										{testingId === ch.id ? (
											<Loader2 className="size-3.5 animate-spin" />
										) : testSentId === ch.id ? (
											<Check className="size-3.5 text-emerald-500" />
										) : (
											<Send className="size-3.5" />
										)}
										<span>
											{testSentId === ch.id
												? m.channels_sent()
												: m.channels_test()}
										</span>
									</Button>
									<Button
										type="button"
										size="icon-sm"
										variant="ghost"
										className="text-muted-foreground hover:text-destructive"
										disabled={testingId !== null || deletingId !== null}
										onClick={() => handleDelete(ch.id)}
									>
										<Trash2 className="size-4" />
										<span className="sr-only">Delete</span>
									</Button>
								</div>
							</div>
						))
					)}
				</div>
			</CardContent>
		</Card>
	);
}
