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
	is_online?: boolean;
	token?: string;
	masked_token: string;
	user?: ChannelUser;
	last_seen_at?: string | null;
	created_at: string;
}

export async function fetchChannels(
	accountSlug: string,
	kind?: string,
): Promise<Channel[]> {
	const url = kind
		? `/api/v1/${accountSlug}/channels?kind=${encodeURIComponent(kind)}`
		: `/api/v1/${accountSlug}/channels`;
	const res = await apiFetch<{ channels: Channel[] }>(url, {
		method: "GET",
	});
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

export async function testChannel(
	accountSlug: string,
	channelId: string,
): Promise<void> {
	await apiFetch(`/api/v1/${accountSlug}/channels/${channelId}/test`, {
		method: "POST",
	});
}

export interface DeviceAuthInfo {
	user_code: string;
	channel_name: string | null;
	status: string;
	expires_in: number;
}

export async function verifyDeviceCode(code: string): Promise<DeviceAuthInfo> {
	return apiFetch<DeviceAuthInfo>(
		`/api/v1/channels/cli/authorizations/${encodeURIComponent(code)}`,
		{
			method: "GET",
		},
	);
}

export async function approveDeviceAuth(data: {
	user_code: string;
	channel_name?: string;
}): Promise<{
	status: string;
	channel: { id: string; name: string; kind: string; status: string };
}> {
	return apiFetch<{
		status: string;
		channel: { id: string; name: string; kind: string; status: string };
	}>(
		`/api/v1/channels/cli/authorizations/${encodeURIComponent(data.user_code)}`,
		{
			method: "PATCH",
			body: { channel_name: data.channel_name },
		},
	);
}

export async function denyDeviceAuth(user_code: string): Promise<void> {
	await apiFetch(
		`/api/v1/channels/cli/authorizations/${encodeURIComponent(user_code)}`,
		{
			method: "DELETE",
		},
	);
}
