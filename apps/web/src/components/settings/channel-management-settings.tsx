import { Monitor, Trash2, Plus } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { CopyButton } from "@/components/ui/copy-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	type Channel,
	createChannel,
	deleteChannel,
	fetchChannels,
} from "@/lib/api/channels";
import { ApiError } from "@/lib/api/client";
import { translateError } from "@/lib/i18n-labels";

export function ChannelManagementSettings({ slug }: { slug: string }) {
	const [channels, setChannels] = useState<Channel[]>([]);
	const [loading, setLoading] = useState(true);
	const [name, setName] = useState("");
	const [kind, setKind] = useState("device");
	const [creating, setCreating] = useState(false);
	const [createdChannel, setCreatedChannel] = useState<Channel | null>(null);
	const [error, setError] = useState<string | null>(null);

	async function load() {
		try {
			const data = await fetchChannels(slug);
			setChannels(data);
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setLoading(false);
		}
	}

	useEffect(() => {
		void load();
	}, [slug]);

	async function handleCreate(e: FormEvent) {
		e.preventDefault();
		if (!name.trim()) return;

		setError(null);
		setCreating(true);
		try {
			const ch = await createChannel(slug, { name: name.trim(), kind });
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

	return (
		<Card className="max-w-2xl">
			<CardHeader>
				<CardTitle className="flex items-center gap-2">
					<Monitor className="size-5" />
					Device Channels
				</CardTitle>
				<CardDescription>
					Connect your desktop machines and CLI daemons to receive notifications and trigger local actions.
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-6">
				<form onSubmit={handleCreate} className="flex items-end gap-3">
					<div className="flex-1">
						<Label htmlFor="channel-name" className="text-xs">
							Device / Channel Name
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
						Add Device
					</Button>
				</form>

				{createdChannel?.token ? (
					<div className="rounded-lg border border-primary/40 bg-primary/5 p-4 text-sm">
						<div className="font-semibold text-foreground">
							Device Channel Created: {createdChannel.name}
						</div>
						<p className="mt-1 text-xs text-muted-foreground">
							Copy this token and configure it on your machine using the Beep CLI:
						</p>
						<div className="mt-2 flex items-center justify-between rounded bg-muted p-2 font-mono text-xs">
							<span className="truncate">{createdChannel.token}</span>
							<CopyButton text={createdChannel.token} />
						</div>
						<div className="mt-2 text-xs text-muted-foreground">
							Run: <code className="rounded bg-muted px-1">beep config set device_token {createdChannel.token}</code>
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
							Loading devices...
						</div>
					) : channels.length === 0 ? (
						<div className="p-4 text-center text-sm text-muted-foreground">
							No devices or custom channels registered yet.
						</div>
					) : (
						channels.map((ch) => (
							<div
								key={ch.id}
								className="flex items-center justify-between p-3.5"
							>
								<div className="flex flex-col gap-0.5">
									<div className="flex items-center gap-2 font-medium text-sm">
										{ch.name}
										<Badge variant="outline" className="text-[10px] uppercase">
											{ch.kind}
										</Badge>
										{ch.status === "active" ? (
											<Badge variant="secondary" className="text-[10px]">
												Active
											</Badge>
										) : null}
									</div>
									<div className="text-xs text-muted-foreground">
										Token: {ch.masked_token}
										{ch.last_seen_at ? (
											<span> • Seen: {new Date(ch.last_seen_at).toLocaleString()}</span>
										) : null}
									</div>
								</div>
								<Button
									size="icon"
									variant="ghost"
									className="text-muted-foreground hover:text-destructive"
									onClick={() => handleDelete(ch.id)}
								>
									<Trash2 className="size-4" />
									<span className="sr-only">Delete</span>
								</Button>
							</div>
						))
					)}
				</div>
			</CardContent>
		</Card>
	);
}
