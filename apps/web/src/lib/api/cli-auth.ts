import { apiFetch } from "@/lib/api/client";

export type CliDeviceAuthInfo = {
	user_code: string;
	client_name: string | null;
	status: string;
	expires_in: number;
};

export async function verifyCliDeviceCode(
	userCode: string,
): Promise<CliDeviceAuthInfo> {
	return apiFetch<CliDeviceAuthInfo>(
		`/api/v1/cli/authorizations/${encodeURIComponent(userCode)}`,
		{
			method: "GET",
		},
	);
}

export async function approveCliDeviceAuth(data: {
	user_code: string;
	client_name?: string;
}): Promise<{ status: string }> {
	return apiFetch<{ status: string }>(
		`/api/v1/cli/authorizations/${encodeURIComponent(data.user_code)}`,
		{
			method: "PATCH",
			body: {
				client_name: data.client_name,
			},
		},
	);
}

export async function denyCliDeviceAuth(user_code: string): Promise<void> {
	await apiFetch(
		`/api/v1/cli/authorizations/${encodeURIComponent(user_code)}`,
		{
			method: "DELETE",
		},
	);
}
