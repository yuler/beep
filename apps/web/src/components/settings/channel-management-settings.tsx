import { Check, Loader2, Plus, Send, Terminal, Trash2 } from "lucide-react";
import { type FormEvent, useCallback, useEffect, useState } from "react";
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

export function ChannelManagementSettings({ slug }: { slug: string }) {
	const [channels, setChannels] = useState<Channel[]>([]);
	const [loading, setLoading] = useState(true);
	const [name, setName] = useState("");
	const [creating, setCreating] = useState(false);
	const [createdChannel, setCreatedChannel] = useState<Channel | null>(null);
	const [testingId, setTestingId] = useState<string | null>(null);
	const [testSentId, setTestSentId] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const { copied, copy } = useCopyToClipboard();

	const load = useCallback(async () => {
		try {
			const data = await fetchChannels(slug, "cli");
			setChannels(data.filter((ch) => ch.kind === "cli"));
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setLoading(false);
		}
	}, [slug]);

	useEffect(() => {
		void load();
	}, [load]);

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
			await load();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setCreating(false);
		}
	}

	async function handleDelete(id: string) {
		if (!confirm("Are you sure you want to remove this channel?")) return;
		try {
			await deleteChannel(slug, id);
			await load();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		}
	}

	async function handleTest(id: string) {
		setError(null);
		setTestingId(id);
		try {
			await testChannel(slug, id);
			setTestSentId(id);
			setTimeout(() => {
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
					CLI Channels
				</CardTitle>
				<CardDescription>
					Connect your Beep CLI daemons to receive notifications and trigger
					local actions. Run{" "}
					<code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
						beep channel connect
					</code>{" "}
					in your terminal to connect instantly.
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-6">
				<form onSubmit={handleCreate} className="flex items-end gap-3">
					<div className="flex-1">
						<Label htmlFor="channel-name" className="text-xs">
							CLI Channel Name
						</Label>
						<Input
							id="channel-name"
							placeholder="e.g. work-laptop"
							value={name}
							onChange={(e) => setName(e.target.value)}
							disabled={creating}
						/>
					</div>
					<Button type="submit" disabled={creating || !name.trim()}>
						<Plus className="mr-1 size-4" />
						Add CLI
					</Button>
				</form>

				{createdChannel?.token ? (
					<div className="rounded-lg border border-primary/40 bg-primary/5 p-4 text-sm">
						<div className="font-semibold text-foreground">
							CLI Channel Created: {createdChannel.name}
						</div>
						<p className="mt-1 text-xs text-muted-foreground">
							Copy this token and configure it on your machine using the Beep
							CLI:
						</p>
						<div className="mt-2">
							<CopyableCode
								code={createdChannel.token}
								copied={copied}
								onCopy={() => copy(createdChannel.token ?? "")}
								label="Copy channel token"
							/>
						</div>
						<div className="mt-2 text-xs text-muted-foreground">
							Run:{" "}
							<code className="rounded bg-muted px-1">
								beep config set channel_token {createdChannel.token}
							</code>
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
							Loading channels...
						</div>
					) : channels.length === 0 ? (
						<div className="p-4 text-center text-sm text-muted-foreground">
							No CLI channels registered yet.
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
												Active
											</Badge>
										) : (
											<Badge
												variant="outline"
												className="gap-1.5 border-zinc-500/30 text-muted-foreground text-[10px]"
											>
												<span className="size-1.5 rounded-full bg-zinc-400" />
												Offline
											</Badge>
										)}
									</div>
									<div className="flex flex-wrap items-center gap-x-2 text-xs text-muted-foreground">
										<span>Token: {ch.masked_token}</span>
										<span>•</span>
										<span>
											{ch.last_seen_at
												? `Last active: ${new Date(ch.last_seen_at).toLocaleString()}`
												: "Never connected"}
										</span>
									</div>
								</div>
								<div className="flex items-center gap-2">
									<Button
										type="button"
										size="sm"
										variant="outline"
										className="h-8 gap-1.5 px-2.5 text-xs"
										disabled={testingId === ch.id}
										onClick={() => handleTest(ch.id)}
									>
										{testingId === ch.id ? (
											<Loader2 className="size-3.5 animate-spin" />
										) : testSentId === ch.id ? (
											<Check className="size-3.5 text-emerald-500" />
										) : (
											<Send className="size-3.5" />
										)}
										<span>{testSentId === ch.id ? "Sent" : "Test"}</span>
									</Button>
									<Button
										size="icon-sm"
										variant="ghost"
										className="text-muted-foreground hover:text-destructive"
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
