import { apiFetch } from "@/lib/api/client";

export type AccessTokenPermission = "read" | "write";

export type AccessToken = {
	id: string;
	description: string | null;
	permission: AccessTokenPermission;
	last_used_at: string | null;
	created_at: string;
};

export type CreatedAccessToken = {
	access_token: AccessToken & {
		token: string;
	};
};

export function fetchAccessTokens() {
	return apiFetch<{ access_tokens: AccessToken[] }>(
		"/api/v1/my/access_tokens",
		{
			method: "GET",
		},
	);
}

export function createAccessToken(body: {
	description: string;
	permission: AccessTokenPermission;
}) {
	return apiFetch<CreatedAccessToken>("/api/v1/my/access_tokens", {
		method: "POST",
		body: { access_token: body },
	});
}

export function deleteAccessToken(id: string) {
	return apiFetch<void>(`/api/v1/my/access_tokens/${id}`, {
		method: "DELETE",
	});
}
