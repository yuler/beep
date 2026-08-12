import type { MeResponse } from "@/lib/api/session";
import { safeReturnTo } from "@/lib/auth/return-to";

export type AccountSummary = MeResponse["accounts"][number];

export type PostAuthTarget =
	| { kind: "href"; href: string }
	| { kind: "account"; slug: string }
	| { kind: "picker" }
	| { kind: "sign" };

/**
 * After login: honor safe return_to; one membership → that account; many → picker.
 * Matches Core `after_authentication_url` + ACCOUNT.md picker rules.
 */
export function resolvePostAuthTarget(
	accounts: AccountSummary[],
	returnTo?: string | null,
): PostAuthTarget {
	const safe = safeReturnTo(returnTo);
	if (safe) return { kind: "href", href: safe };
	if (accounts.length === 0) return { kind: "sign" };
	if (accounts.length === 1) {
		return { kind: "account", slug: accounts[0].slug };
	}
	return { kind: "picker" };
}

/** Shell account hint: last-used membership, else personal, else first. */
export function resolveShellAccount(me: MeResponse): AccountSummary | null {
	if (me.accounts.length === 0) return null;
	if (me.last_account_slug) {
		const last = me.accounts.find(
			(account) => account.slug === me.last_account_slug,
		);
		if (last) return last;
	}
	return (
		me.accounts.find((account) => account.personal) ?? me.accounts[0] ?? null
	);
}

/** Dashboard CTA: one account → home; many → picker (never invent a tenant). */
export function resolveDashboardTarget(
	accounts: AccountSummary[],
): PostAuthTarget {
	if (accounts.length === 0) return { kind: "sign" };
	if (accounts.length === 1) {
		return { kind: "account", slug: accounts[0].slug };
	}
	return { kind: "picker" };
}
