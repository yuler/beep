import { createFileRoute } from "@tanstack/react-router";
import { Box, Layers, Rocket } from "lucide-react";

import { SiteAuthButton, SiteLayout } from "@/components/layout";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useTranslation } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/")({ component: Home });

const features = [
	{
		icon: Box,
		titleKey: "marketing.feature_1_title",
		descriptionKey: "marketing.feature_1_description",
	},
	{
		icon: Layers,
		titleKey: "marketing.feature_2_title",
		descriptionKey: "marketing.feature_2_description",
	},
	{
		icon: Rocket,
		titleKey: "marketing.feature_3_title",
		descriptionKey: "marketing.feature_3_description",
	},
] as const;

function Home() {
	const { t } = useTranslation();

	return (
		<SiteLayout>
			<main className="flex-1">
				<section className="mx-auto max-w-5xl px-4 py-16 sm:px-6 sm:py-24">
					<div className="mx-auto flex max-w-2xl flex-col items-center text-center">
						<Badge variant="secondary" className="mb-4">
							{t("marketing.badge")}
						</Badge>
						<h1 className="font-heading text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
							{t("marketing.hero_title")}
						</h1>
						<p className="mt-4 text-lg text-muted-foreground text-pretty">
							{t("marketing.hero_description")}
						</p>
						<div className="mt-8 flex flex-wrap items-center justify-center gap-3">
							<SiteAuthButton
								signInLabel={t("common.get_started")}
								dashboardLabel={t("common.open_dashboard")}
								size="lg"
							/>
							<a
								href="https://github.com"
								target="_blank"
								rel="noreferrer"
								className={cn(
									buttonVariants({ size: "lg", variant: "outline" }),
								)}
							>
								{t("common.view_docs")}
							</a>
						</div>
					</div>
				</section>

				<section className="border-t border-border bg-muted/30">
					<div className="mx-auto max-w-5xl px-4 py-16 sm:px-6">
						<div className="mb-10 max-w-xl">
							<h2 className="font-heading text-2xl font-semibold tracking-tight">
								{t("marketing.features_title")}
							</h2>
							<p className="mt-2 text-muted-foreground">
								{t("marketing.features_description")}
							</p>
						</div>

						<div className="grid gap-4 sm:grid-cols-3">
							{features.map(({ icon: Icon, titleKey, descriptionKey }) => (
								<Card key={titleKey} size="sm">
									<CardHeader>
										<div className="mb-2 inline-flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
											<Icon className="size-4" />
										</div>
										<CardTitle>{t(titleKey)}</CardTitle>
										<CardDescription>{t(descriptionKey)}</CardDescription>
									</CardHeader>
								</Card>
							))}
						</div>
					</div>
				</section>

				<section className="mx-auto max-w-5xl px-4 py-16 sm:px-6">
					<Card className="bg-primary text-primary-foreground ring-primary/20">
						<CardHeader>
							<CardTitle className="text-primary-foreground">
								{t("marketing.cta_title")}
							</CardTitle>
							<CardDescription className="text-primary-foreground/80">
								{t("marketing.cta_description")}
							</CardDescription>
						</CardHeader>
						<CardContent>
							<SiteAuthButton
								signInLabel={t("common.open_app")}
								dashboardLabel={t("common.dashboard")}
								variant="secondary"
							/>
						</CardContent>
					</Card>
				</section>
			</main>
		</SiteLayout>
	);
}
