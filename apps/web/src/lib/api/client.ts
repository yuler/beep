import { CORE_URL } from "@/config";

export class ApiError extends Error {
	status: number;
	code?: string;

	constructor(status: number, message: string, code?: string) {
		super(message);
		this.name = "ApiError";
		this.status = status;
		this.code = code;
	}
}

type ApiOptions = Omit<RequestInit, "body"> & {
	body?: unknown;
};

/** Origin for API calls. Mode A: absolute VITE_CORE_URL. Mode B browser: "". */
function resolveApiOrigin(): string {
	if (CORE_URL) return CORE_URL;
	// During SSR, relative `/api` must hit Rails (via Nitro proxy or direct).
	if (import.meta.env.SSR) {
		const internal = process.env.CORE_INTERNAL_URL;
		if (!internal) {
			throw new Error(
				"CORE_INTERNAL_URL is required. Copy .env.example to .env.",
			);
		}
		return internal;
	}
	return "";
}

async function serverCookieHeader(): Promise<string | undefined> {
	if (!import.meta.env.SSR) return undefined;
	const { getRequestHeader } = await import("@tanstack/react-start/server");
	return getRequestHeader("cookie");
}

async function request(
	path: string,
	{ body, headers, method = "GET", ...init }: ApiOptions = {},
): Promise<Response> {
	const requestHeaders = new Headers(headers);
	if (body !== undefined) {
		requestHeaders.set("Content-Type", "application/json");
	}
	requestHeaders.set("Accept", "application/json");

	const cookie = await serverCookieHeader();
	if (cookie && !requestHeaders.has("Cookie")) {
		requestHeaders.set("Cookie", cookie);
	}

	// Mode A (split): CORE_URL is absolute origin. Mode B (proxy): browser uses
	// same-origin relative `/api/v1/...`; SSR uses CORE_INTERNAL_URL.
	const url = `${resolveApiOrigin()}${path}`;
	const response = await fetch(url, {
		...init,
		method,
		credentials: "include",
		headers: requestHeaders,
		body: body === undefined ? undefined : JSON.stringify(body),
	});

	if (!response.ok) {
		let message = response.statusText || "Request failed";
		let code: string | undefined;
		try {
			const payload = (await response.json()) as {
				message?: string;
				error?: string;
				code?: string;
			};
			message = payload.message ?? payload.error ?? message;
			code = payload.code;
		} catch {
			// ignore JSON parse errors
		}
		throw new ApiError(response.status, message, code);
	}

	return response;
}

export async function apiFetch<T>(
	path: string,
	options: ApiOptions = {},
): Promise<T> {
	const response = await request(path, options);

	if (response.status === 204) {
		return undefined as T;
	}

	return (await response.json()) as T;
}

/** Like apiFetch, but also returns response headers (e.g. X-Magic-Link-Code). */
export async function apiFetchWithHeaders<T>(
	path: string,
	options: ApiOptions = {},
): Promise<{ data: T; headers: Headers }> {
	const response = await request(path, options);

	if (response.status === 204) {
		return { data: undefined as T, headers: response.headers };
	}

	return {
		data: (await response.json()) as T,
		headers: response.headers,
	};
}
