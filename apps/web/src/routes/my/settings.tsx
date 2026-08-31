import { createFileRoute, getRouteApi } from "@tanstack/react-router";
import { Globe, Mail } from "lucide-react";
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
import { Label } from "@/components/ui/label";
import { fetchMe } from "@/lib/api/session";
import { withAuthRedirects } from "@/lib/auth/guards";
import { getGravatarUrl } from "@/lib/gravatar";
import { useTranslation } from "@/lib/i18n";

const myRoute = getRouteApi("/my");

export const Route = createFileRoute("/my/settings")({
	loader: withAuthRedirects(() => fetchMe({ force: true })),
	component: MySettingsPage,
});

function MySettingsPage() {
	const { t } = useTranslation();
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
	const gravatarLabel = t("term.gravatar");
	const avatarHintParts = t("my.avatar_gravatar_hint").split(gravatarLabel);

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: t("nav.home"),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: t("my.personal_settings") },
					{ label: t("my.profile_preferences"), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{t("my.personal_settings")}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						{t("my.profile_subtitle")}
					</p>
				</div>

				<div className="grid max-w-2xl gap-6">
					<Card>
						<CardHeader>
							<CardTitle className="text-lg">
								{t("my.profile_information")}
							</CardTitle>
							<CardDescription>
								{t("my.public_avatar_description")}
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-6">
							<div className="flex items-center gap-4">
								<a
									href="https://gravatar.com"
									target="_blank"
									rel="noreferrer"
									className="group/avatar-link transition-opacity hover:opacity-85"
									title={t("my.avatar_gravatar_title")}
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
									<p className="font-medium leading-none">{t("my.avatar")}</p>
									<p className="text-xs text-muted-foreground">
										{avatarHintParts[0]}
										<a
											href="https://gravatar.com"
											target="_blank"
											rel="noreferrer"
											className="inline-flex items-center gap-0.5 font-medium text-foreground underline underline-offset-2 hover:text-primary"
										>
											{gravatarLabel}
										</a>
										{avatarHintParts[1] ?? ""}
									</p>
								</div>
							</div>

							<div className="space-y-2">
								<Label htmlFor="email" className="flex items-center gap-2">
									<Mail className="size-4 text-muted-foreground" />
									{t("my.email_address")}
								</Label>
								<div className="flex items-center gap-2">
									<div className="flex h-9 flex-1 items-center rounded-md border border-input bg-muted/40 px-3 font-mono text-sm text-foreground select-all">
										{identity.email}
									</div>
								</div>
								<p className="text-xs text-muted-foreground">
									{t("my.email_identity_hint", {
										beep: t("term.beep_capitalized"),
									})}
								</p>
							</div>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle className="text-lg">{t("my.preferences")}</CardTitle>
							<CardDescription>
								{t("my.preferences_description")}
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="space-y-2">
								<div className="flex items-center justify-between">
									<Label htmlFor="language" className="flex items-center gap-2">
										<Globe className="size-4 text-muted-foreground" />
										{t("my.language")}
									</Label>
									<Badge variant="secondary" className="text-xs font-normal">
										{t("my.english_only")}
									</Badge>
								</div>
								<select
									id="language"
									disabled
									className="flex h-9 w-full rounded-md border border-input bg-muted/40 px-3 py-1 text-sm shadow-xs disabled:cursor-not-allowed disabled:opacity-75"
									value="en"
								>
									<option value="en">{t("my.english_us")}</option>
								</select>
								<p className="text-xs text-muted-foreground">
									{t("my.languages_future")}
								</p>
							</div>
						</CardContent>
					</Card>
				</div>
			</div>
		</>
	);
}
