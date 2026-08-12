import { type NavigateOptions, redirect } from "@tanstack/react-router";

import { ApiError } from "@/lib/api/client";
import type { MeResponse } from "@/lib/api/session";
import {
	type AccountSummary,
	type PostAuthTarget,
	resolveDashboardTarget,
	resolvePostAuthTarget,
} from "@/lib/auth/account";
import { safeReturnTo } from "@/lib/auth/return-to";

export function redirectForTarget(target: PostAuthTarget): never {
	switch (target.kind) {
		case "href":
			throw redirect({ href: target.href });
		case "account":
			throw redirect({
				to: "/$account_slug",
				params: { account_slug: target.slug },
			});
		case "picker":
			throw redirect({ to: "/accounts" });
		case "sign":
			throw redirect({ to: "/sign" });
	}
}

export function redirectToSign(returnTo?: string | null): never {
	const safe = safeReturnTo(returnTo);
	throw redirect({
		to: "/sign",
		search: safe ? { return_to: safe } : {},
	});
}

export type RootRouteContext = {
	me: MeResponse | null;
};

/**
 * Like core `Authentication#require_authentication` after `resume_session`:
 * root `beforeLoad` already probed the cookie; no identity → redirect to sign.
 */
export function requireSession({
	context,
	location,
}: {
	context: RootRouteContext;
	location: { pathname: string; searchStr: string };
}): MeResponse {
	if (!context.me) {
		redirectToSign(`${location.pathname}${location.searchStr}`);
	}
	return context.me;
}

/**
 * Like core `redirect_authenticated_user`: root already resumed the session;
 * signed-in users should not see the sign-in flow.
 */
export function requireGuest({
	context,
	search,
}: {
	context: RootRouteContext;
	search: { return_to?: string };
}) {
	if (!context.me) return;
	redirectForTarget(
		resolvePostAuthTarget(context.me.accounts, search.return_to),
	);
}

export function requireStaff(me: {
	identity: { staff: boolean };
	accounts: AccountSummary[];
}) {
	if (!me.identity.staff) {
		redirectForTarget(resolveDashboardTarget(me.accounts));
	}
}

type LoaderContext = { location: { pathname: string; searchStr: string } };

/**
 * Wrap a route loader so a 401 (stale/revoked session) or 403 (role change)
 * redirects like the removed `useAdminResource` hook did, instead of dumping a
 * raw error page through the default error boundary.
 */
export function withAuthRedirects<R>(load: () => Promise<R>) {
	return async ({ location }: LoaderContext): Promise<R> => {
		try {
			return await load();
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				redirectToSign(`${location.pathname}${location.searchStr}`);
			}
			if (err instanceof ApiError && err.status === 403) {
				throw redirect({ to: "/accounts" });
			}
			throw err;
		}
	};
}

type NavigateFn = (opts: NavigateOptions) => Promise<void> | void;

export function navigateForTarget(
	navigate: NavigateFn,
	target: PostAuthTarget,
) {
	switch (target.kind) {
		case "href":
			return navigate({ href: target.href });
		case "account":
			return navigate({
				to: "/$account_slug",
				params: { account_slug: target.slug },
			});
		case "picker":
			return navigate({ to: "/accounts" });
		case "sign":
			return navigate({ to: "/sign" });
	}
}

export type { AccountSummary };
