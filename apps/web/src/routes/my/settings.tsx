import { createFileRoute, getRouteApi } from "@tanstack/react-router";
import { ExternalLink, Globe, Mail } from "lucide-react";
import { useEffect, useState } from "react";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { fetchMe } from "@/lib/api/session";
import { withAuthRedirects } from "@/lib/auth/guards";
import { getGravatarUrl } from "@/lib/gravatar";

const myRoute = getRouteApi("/my");

export const Route = createFileRoute("/my/settings")({
	loader: withAuthRedirects(() => fetchMe({ force: true })),
	component: MySettingsPage,
});

function MySettingsPage() {
	const { account } = myRoute.useRouteContext();
	const me = Route.useLoaderData();
	const identity = me.identity;

	const [avatarUrl, setAvatarUrl] = useState<string | null>(null);

	useEffect(() => {
		if (identity.email) {
			void getGravatarUrl(identity.email, 200).then(setAvatarUrl);
		}
	}, [identity.email]);

	const initials = identity.email.charAt(0).toUpperCase() || "U";

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: "Personal Settings" },
					{ label: "Profile & Preferences", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						Personal Settings
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						Manage your profile, email, and personal preferences.
					</p>
				</div>

				<div className="grid max-w-2xl gap-6">
					<Card>
						<CardHeader>
							<CardTitle className="text-lg">Profile Information</CardTitle>
							<CardDescription>
								Your public avatar and identity details.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-6">
							<div className="flex items-center gap-4">
								<a
									href="https://gravatar.com"
									target="_blank"
									rel="noreferrer"
									className="group/avatar-link transition-opacity hover:opacity-85"
									title="Change avatar on Gravatar"
								>
									<Avatar size="lg" className="size-16 border border-border">
										{avatarUrl ? (
											<AvatarImage src={avatarUrl} alt={identity.email} />
										) : null}
										<AvatarFallback className="text-lg">
											{initials}
										</AvatarFallback>
									</Avatar>
								</a>
								<div className="space-y-1">
									<p className="font-medium leading-none">Avatar</p>
									<p className="text-xs text-muted-foreground">
										Avatar is automatically fetched from{" "}
										<a
											href="https://gravatar.com"
											target="_blank"
											rel="noreferrer"
											className="inline-flex items-center gap-0.5 font-medium text-foreground underline underline-offset-2 hover:text-primary"
										>
											Gravatar
											<ExternalLink className="size-3" />
										</a>{" "}
										using your email address.
									</p>
								</div>
							</div>

							<div className="space-y-2">
								<Label htmlFor="email" className="flex items-center gap-2">
									<Mail className="size-4 text-muted-foreground" />
									Email Address
								</Label>
								<Input
									id="email"
									value={identity.email}
									disabled
									className="bg-muted/40 font-mono text-sm"
								/>
								<p className="text-xs text-muted-foreground">
									This email is associated with your global Beep identity.
								</p>
							</div>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle className="text-lg">Preferences</CardTitle>
							<CardDescription>
								Regional and display preferences for your account.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="space-y-2">
								<div className="flex items-center justify-between">
									<Label htmlFor="language" className="flex items-center gap-2">
										<Globe className="size-4 text-muted-foreground" />
										Language
									</Label>
									<Badge variant="secondary" className="text-xs font-normal">
										English only
									</Badge>
								</div>
								<select
									id="language"
									disabled
									className="flex h-9 w-full rounded-md border border-input bg-muted/40 px-3 py-1 text-sm shadow-xs disabled:cursor-not-allowed disabled:opacity-75"
									value="en"
								>
									<option value="en">English (US)</option>
								</select>
								<p className="text-xs text-muted-foreground">
									Additional languages will be available in future releases.
								</p>
							</div>
						</CardContent>
					</Card>
				</div>
			</div>
		</>
	);
}
