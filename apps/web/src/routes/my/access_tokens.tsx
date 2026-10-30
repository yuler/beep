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
import { useMemo, useState } from "react";
import { confirm } from "@/components/confirm-dialog";
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
import { CopyableCode, useCopyToClipboard } from "@/components/ui/copy-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	ResponsiveDialog,
	ResponsiveDialogBody,
	ResponsiveDialogContent,
	ResponsiveDialogDescription,
	ResponsiveDialogFooter,
	ResponsiveDialogHeader,
	ResponsiveDialogTitle,
	ResponsiveDialogTrigger,
} from "@/components/ui/responsive-dialog";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { publicApiOrigin } from "@/config";
import {
	type AccessToken,
	type AccessTokenPermission,
	createAccessToken,
	deleteAccessToken,
	fetchAccessTokens,
} from "@/lib/api/access-tokens";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

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
		if (!(await confirm(m.my_revoke_confirm()))) {
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

					<ResponsiveDialog
						open={isCreateOpen}
						onOpenChange={(open) => {
							setIsCreateOpen(open);
							if (!open) {
								setCreatedToken(null);
								setCreateError(null);
							}
						}}
					>
						<ResponsiveDialogTrigger
							render={<Button className="gap-2 w-full sm:w-auto" />}
						>
							<Plus className="size-4" />
							{m.my_generate_token()}
						</ResponsiveDialogTrigger>

						<ResponsiveDialogContent className="sm:max-w-md">
							{createdToken ? (
								<div className="flex flex-col flex-1 min-h-0 overflow-hidden">
									<ResponsiveDialogHeader>
										<ResponsiveDialogTitle className="flex items-center gap-2 text-emerald-600 dark:text-emerald-500">
											<ShieldCheck className="size-5" />
											{m.my_token_generated()}
										</ResponsiveDialogTitle>
										<ResponsiveDialogDescription>
											{m.my_token_copy_warning()}
										</ResponsiveDialogDescription>
									</ResponsiveDialogHeader>

									<ResponsiveDialogBody>
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
									</ResponsiveDialogBody>

									<ResponsiveDialogFooter>
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
									</ResponsiveDialogFooter>
								</div>
							) : (
								<form onSubmit={handleCreate}>
									<ResponsiveDialogHeader>
										<ResponsiveDialogTitle>
											{m.my_generate_new_token()}
										</ResponsiveDialogTitle>
										<ResponsiveDialogDescription>
											{m.my_token_auth_description()}
										</ResponsiveDialogDescription>
									</ResponsiveDialogHeader>

									<ResponsiveDialogBody>
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
											<Select
												value={permission}
												onValueChange={(val) => {
													if (val) {
														setPermission(val as AccessTokenPermission);
													}
												}}
											>
												<SelectTrigger id="token-permission" className="w-full">
													<SelectValue />
												</SelectTrigger>
												<SelectContent>
													<SelectItem value="write">
														{m.my_permission_write()}
													</SelectItem>
													<SelectItem value="read">
														{m.my_permission_read()}
													</SelectItem>
												</SelectContent>
											</Select>
										</div>
									</ResponsiveDialogBody>

									<ResponsiveDialogFooter>
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
									</ResponsiveDialogFooter>
								</form>
							)}
						</ResponsiveDialogContent>
					</ResponsiveDialog>
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
												{m.common_created()}{" "}
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

				{data.access_tokens.length > 0 ? <ApiCodeExamples /> : null}
			</div>
		</>
	);
}

type ExampleLanguage = "curl" | "python" | "javascript" | "go";

const LANGUAGE_TABS: { id: ExampleLanguage; label: string }[] = [
	{ id: "curl", label: "cURL" },
	{ id: "python", label: "Python" },
	{ id: "javascript", label: "JavaScript" },
	{ id: "go", label: "Go" },
];

function ApiCodeExamples() {
	const [activeTab, setActiveTab] = useState<ExampleLanguage>("curl");
	const identityCopy = useCopyToClipboard();
	const createBeepCopy = useCopyToClipboard();

	const baseUrl =
		publicApiOrigin() ||
		(typeof window !== "undefined" ? window.location.origin : "") ||
		"https://api.beep.com";

	const snippets = useMemo(() => {
		return {
			curl: {
				identity: `curl -H "Authorization: Bearer <YOUR_TOKEN>" \\
  ${baseUrl}/api/v1/me`,
				createBeep: `curl -X POST "${baseUrl}/api/v1/beeps" \\
  -H "Authorization: Bearer <YOUR_TOKEN>" \\
  -H "Content-Type: application/json" \\
  -d '{"body": "Hello from API", "kind": "once"}'`,
			},
			python: {
				identity: `import requests

response = requests.get(
    "${baseUrl}/api/v1/me",
    headers={"Authorization": "Bearer <YOUR_TOKEN>"},
)
print(response.json())`,
				createBeep: `import requests

response = requests.post(
    "${baseUrl}/api/v1/beeps",
    headers={
        "Authorization": "Bearer <YOUR_TOKEN>",
        "Content-Type": "application/json",
    },
    json={
        "body": "Hello from API",
        "kind": "once",
    },
)
print(response.json())`,
			},
			javascript: {
				identity: `const response = await fetch("${baseUrl}/api/v1/me", {
  headers: {
    Authorization: "Bearer <YOUR_TOKEN>",
  },
});
const data = await response.json();
console.log(data);`,
				createBeep: `const response = await fetch("${baseUrl}/api/v1/beeps", {
  method: "POST",
  headers: {
    Authorization: "Bearer <YOUR_TOKEN>",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    body: "Hello from API",
    kind: "once",
  }),
});
const data = await response.json();
console.log(data);`,
			},
			go: {
				identity: `package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    req, err := http.NewRequest("GET", "${baseUrl}/api/v1/me", nil)
    if err != nil {
        panic(err)
    }
    req.Header.Set("Authorization", "Bearer <YOUR_TOKEN>")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(body))
}`,
				createBeep: `package main

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
)

func main() {
    payload := []byte(\`{"body": "Hello from API", "kind": "once"}\`)
    req, err := http.NewRequest("POST", "${baseUrl}/api/v1/beeps", bytes.NewBuffer(payload))
    if err != nil {
        panic(err)
    }
    req.Header.Set("Authorization", "Bearer <YOUR_TOKEN>")
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(body))
}`,
			},
		};
	}, [baseUrl]);

	const currentSnippets = snippets[activeTab];

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-lg">{m.my_api_example_title()}</CardTitle>
				<CardDescription>{m.my_api_example_description()}</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="flex items-center gap-1 rounded-lg border border-input bg-muted/40 p-1">
					{LANGUAGE_TABS.map((tab) => (
						<button
							key={tab.id}
							type="button"
							onClick={() => setActiveTab(tab.id)}
							className={cn(
								"flex-1 flex items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-all cursor-pointer",
								activeTab === tab.id
									? "bg-background text-foreground shadow-xs"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							<span>{tab.label}</span>
						</button>
					))}
				</div>

				<div className="flex flex-col gap-2">
					<div className="flex items-center justify-between gap-2">
						<div className="flex items-center gap-2">
							<span className="text-xs font-semibold text-foreground">
								{m.my_api_example_get_me()}
							</span>
							<Badge variant="outline" className="font-mono text-[10px]">
								GET /api/v1/me
							</Badge>
						</div>
						<Button
							variant="ghost"
							size="sm"
							className="h-7 px-2 text-xs gap-1 shrink-0"
							onClick={() => identityCopy.copy(currentSnippets.identity)}
						>
							{identityCopy.copied ? (
								<>
									<Check className="size-3.5 text-emerald-500" />
									<span className="text-emerald-500">{m.common_copied()}</span>
								</>
							) : (
								<>
									<Copy className="size-3.5" />
									<span>{m.common_copy()}</span>
								</>
							)}
						</Button>
					</div>
					<CopyableCode
						code={currentSnippets.identity}
						copied={identityCopy.copied}
						onCopy={() => identityCopy.copy(currentSnippets.identity)}
						label={m.common_copy()}
					/>
				</div>

				<div className="flex flex-col gap-2">
					<div className="flex items-center justify-between gap-2">
						<div className="flex items-center gap-2">
							<span className="text-xs font-semibold text-foreground">
								{m.my_api_example_create_beep()}
							</span>
							<Badge variant="outline" className="font-mono text-[10px]">
								POST /api/v1/beeps
							</Badge>
						</div>
						<Button
							variant="ghost"
							size="sm"
							className="h-7 px-2 text-xs gap-1 shrink-0"
							onClick={() => createBeepCopy.copy(currentSnippets.createBeep)}
						>
							{createBeepCopy.copied ? (
								<>
									<Check className="size-3.5 text-emerald-500" />
									<span className="text-emerald-500">{m.common_copied()}</span>
								</>
							) : (
								<>
									<Copy className="size-3.5" />
									<span>{m.common_copy()}</span>
								</>
							)}
						</Button>
					</div>
					<CopyableCode
						code={currentSnippets.createBeep}
						copied={createBeepCopy.copied}
						onCopy={() => createBeepCopy.copy(currentSnippets.createBeep)}
						label={m.common_copy()}
					/>
				</div>
			</CardContent>
		</Card>
	);
}
