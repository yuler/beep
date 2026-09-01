import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import {
	Check,
	Copy,
	KeyRound,
	Loader2,
	Plus,
	ShieldCheck,
	Trash2,
} from "lucide-react";
import { useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	type AccessToken,
	type AccessTokenPermission,
	createAccessToken,
	deleteAccessToken,
	fetchAccessTokens,
} from "@/lib/api/access-tokens";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import * as m from "@/locale/paraglide/messages";

const myRoute = getRouteApi("/my");

export const Route = createFileRoute("/my/access_tokens")({
	loader: withAuthRedirects(() => fetchAccessTokens()),
	component: AccessTokensPage,
});

function AccessTokensPage() {
	const { account } = myRoute.useRouteContext();
	const data = Route.useLoaderData();
	const router = useRouter();

	const [isCreateOpen, setIsCreateOpen] = useState(false);
	const [description, setDescription] = useState("");
	const [permission, setPermission] = useState<AccessTokenPermission>("write");
	const [isCreating, setIsCreating] = useState(false);
	const [createError, setCreateError] = useState<string | null>(null);

	const [createdToken, setCreatedToken] = useState<string | null>(null);
	const [hasCopied, setHasCopied] = useState(false);

	const [deletingId, setDeletingId] = useState<string | null>(null);

	async function handleCreate(e: React.FormEvent) {
		e.preventDefault();
		setIsCreating(true);
		setCreateError(null);

		try {
			const res = await createAccessToken({
				description: description.trim(),
				permission,
			});
			setCreatedToken(res.access_token.token);
			setDescription("");
			setPermission("write");
			await router.invalidate();
		} catch (err) {
			setCreateError(
				err instanceof ApiError ? err.message : m.errors_create_token_failed(),
			);
		} finally {
			setIsCreating(false);
		}
	}

	async function handleDelete(id: string) {
		if (!confirm(m.my_revoke_confirm())) {
			return;
		}

		setDeletingId(id);
		try {
			await deleteAccessToken(id);
			await router.invalidate();
		} catch (err) {
			alert(
				err instanceof ApiError ? err.message : m.errors_revoke_token_failed(),
			);
		} finally {
			setDeletingId(null);
		}
	}

	function handleCopy() {
		if (!createdToken) return;
		navigator.clipboard.writeText(createdToken);
		setHasCopied(true);
		setTimeout(() => setHasCopied(false), 2000);
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: m.my_personal_settings() },
					{ label: m.my_api_tokens_breadcrumb(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{m.my_api_tokens_title()}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							{m.my_api_tokens_description()}
						</p>
					</div>

					<Dialog
						open={isCreateOpen}
						onOpenChange={(open) => {
							setIsCreateOpen(open);
							if (!open) {
								setCreatedToken(null);
								setCreateError(null);
							}
						}}
					>
						<DialogTrigger render={<Button className="gap-2" />}>
							<Plus className="size-4" />
							{m.my_generate_token()}
						</DialogTrigger>

						<DialogContent className="sm:max-w-md">
							{createdToken ? (
								<>
									<DialogHeader>
										<DialogTitle className="flex items-center gap-2 text-emerald-600 dark:text-emerald-500">
											<ShieldCheck className="size-5" />
											{m.my_token_generated()}
										</DialogTitle>
										<DialogDescription>
											{m.my_token_copy_warning()}
										</DialogDescription>
									</DialogHeader>

									<div className="my-2 flex items-center gap-2">
										<Input
											readOnly
											value={createdToken}
											className="font-mono text-sm"
										/>
										<Button
											type="button"
											variant="secondary"
											onClick={handleCopy}
											className="shrink-0 gap-1.5"
										>
											{hasCopied ? (
												<>
													<Check className="size-4 text-emerald-600" />
													{m.common_copied()}
												</>
											) : (
												<>
													<Copy className="size-4" />
													{m.common_copy()}
												</>
											)}
										</Button>
									</div>

									<DialogFooter>
										<Button
											type="button"
											className="w-full sm:w-auto"
											onClick={() => {
												setIsCreateOpen(false);
												setCreatedToken(null);
											}}
										>
											{m.common_done()}
										</Button>
									</DialogFooter>
								</>
							) : (
								<form onSubmit={handleCreate} className="space-y-4">
									<DialogHeader>
										<DialogTitle>{m.my_generate_new_token()}</DialogTitle>
										<DialogDescription>
											{m.my_token_auth_description()}
										</DialogDescription>
									</DialogHeader>

									{createError ? (
										<div
											className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
											role="alert"
										>
											{createError}
										</div>
									) : null}

									<div className="space-y-2">
										<Label htmlFor="token-description">
											{m.common_description()}
										</Label>
										<Input
											id="token-description"
											placeholder={m.my_token_description_placeholder()}
											value={description}
											onChange={(e) => setDescription(e.target.value)}
											required
											autoFocus
										/>
									</div>

									<div className="space-y-2">
										<Label htmlFor="token-permission">
											{m.common_permission()}
										</Label>
										<select
											id="token-permission"
											className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-base shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
											value={permission}
											onChange={(e) =>
												setPermission(e.target.value as AccessTokenPermission)
											}
										>
											<option value="write">{m.my_permission_write()}</option>
											<option value="read">{m.my_permission_read()}</option>
										</select>
									</div>

									<DialogFooter>
										<Button
											type="button"
											variant="outline"
											onClick={() => setIsCreateOpen(false)}
										>
											{m.common_cancel()}
										</Button>
										<Button type="submit" disabled={isCreating}>
											{isCreating ? (
												<>
													<Loader2 className="mr-2 size-4 animate-spin" />
													{m.common_generating()}
												</>
											) : (
												m.my_generate_token()
											)}
										</Button>
									</DialogFooter>
								</form>
							)}
						</DialogContent>
					</Dialog>
				</div>

				<Card>
					<CardHeader>
						<CardTitle className="text-lg">{m.my_active_tokens()}</CardTitle>
						<CardDescription>
							{m.my_active_tokens_description()}
						</CardDescription>
					</CardHeader>
					<CardContent className="p-0">
						{data.access_tokens.length === 0 ? (
							<div className="flex flex-col items-center justify-center p-8 text-center text-muted-foreground">
								<KeyRound className="mb-2 size-8 text-muted-foreground/50" />
								<p className="font-medium text-foreground">
									{m.my_no_tokens()}
								</p>
								<p className="text-sm">{m.my_no_tokens_description()}</p>
							</div>
						) : (
							<div className="divide-y divide-border">
								{data.access_tokens.map((token: AccessToken) => (
									<div
										key={token.id}
										className="flex flex-col gap-2 p-4 sm:flex-row sm:items-center sm:justify-between"
									>
										<div className="space-y-1">
											<div className="flex items-center gap-2">
												<span className="font-medium">
													{token.description || m.my_personal_access_token()}
												</span>
												<Badge
													variant={
														token.permission === "write"
															? "default"
															: "secondary"
													}
													className="text-xs"
												>
													{token.permission === "write"
														? m.my_read_write()
														: m.my_read_only()}
												</Badge>
											</div>
											<div className="text-xs text-muted-foreground">
												{m.admin_jobs_created()}{" "}
												{new Date(token.created_at).toLocaleDateString()} ·{" "}
												{token.last_used_at
													? m.common_last_used({
															date: new Date(
																token.last_used_at,
															).toLocaleDateString(),
														})
													: m.common_never_used()}
											</div>
										</div>

										<Button
											variant="ghost"
											size="sm"
											className="self-end text-destructive hover:bg-destructive/10 hover:text-destructive sm:self-center"
											disabled={deletingId === token.id}
											onClick={() => handleDelete(token.id)}
										>
											{deletingId === token.id ? (
												<Loader2 className="size-4 animate-spin" />
											) : (
												<Trash2 className="size-4" />
											)}
											<span className="ml-1 sm:hidden">
												{m.common_revoke()}
											</span>
										</Button>
									</div>
								))}
							</div>
						)}
					</CardContent>
					<CardFooter className="border-t bg-muted/30 px-6 py-3">
						<p className="text-xs text-muted-foreground">
							{m.my_api_header_hint()}
						</p>
					</CardFooter>
				</Card>

				{data.access_tokens.length > 0 ? (
					<Card>
						<CardHeader>
							<CardTitle className="text-lg">
								{m.my_api_example_title()}
							</CardTitle>
							<CardDescription>
								{m.my_api_example_description()}
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-3">
							<div className="relative rounded-lg bg-muted/50 p-4 font-mono text-xs">
								<pre className="overflow-x-auto text-foreground">
									{`curl -H "Authorization: Bearer <YOUR_TOKEN>" \\
  ${typeof window !== "undefined" ? window.location.origin : ""}/api/v1/me`}
								</pre>
							</div>
							<p className="text-xs text-muted-foreground">
								{m.my_api_example_beeps_hint({ slug: account.slug })}
							</p>
						</CardContent>
					</Card>
				) : null}
			</div>
		</>
	);
}
