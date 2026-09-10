import { apiFetch } from "./client";

export interface ChannelUser {
	id: string;
	name: string;
}

export interface Channel {
	id: string;
	name: string;
	kind: string;
	status: string;
	token?: string;
	masked_token: string;
	user?: ChannelUser;
	last_seen_at?: string | null;
	created_at: string;
}

export async function fetchChannels(accountSlug: string): Promise<Channel[]> {
	const res = await apiFetch<{ channels: Channel[] }>(
		`/api/v1/${accountSlug}/channels`,
		{ method: "GET" },
	);
	return res.channels;
}

export async function createChannel(
	accountSlug: string,
	data: { name: string; kind?: string },
): Promise<Channel> {
	const res = await apiFetch<{ channel: Channel }>(
		`/api/v1/${accountSlug}/channels`,
		{
			method: "POST",
			body: { channel: data },
		},
	);
	return res.channel;
}

export async function deleteChannel(
	accountSlug: string,
	channelId: string,
): Promise<void> {
	await apiFetch(`/api/v1/${accountSlug}/channels/${channelId}`, {
		method: "DELETE",
	});
}
