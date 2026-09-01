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
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export const Route = createFileRoute("/")({ component: Home });

const features = [
	{
		id: "feature-1",
		icon: Box,
		title: m.marketing_feature_1_title,
		description: m.marketing_feature_1_description,
	},
	{
		id: "feature-2",
		icon: Layers,
		title: m.marketing_feature_2_title,
		description: m.marketing_feature_2_description,
	},
	{
		id: "feature-3",
		icon: Rocket,
		title: m.marketing_feature_3_title,
		description: m.marketing_feature_3_description,
	},
] as const;

function Home() {
	return (
		<SiteLayout>
			<main className="flex-1">
				<section className="mx-auto max-w-5xl px-4 py-16 sm:px-6 sm:py-24">
					<div className="mx-auto flex max-w-2xl flex-col items-center text-center">
						<Badge variant="secondary" className="mb-4">
							{m.marketing_badge()}
						</Badge>
						<h1 className="font-heading text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
							{m.marketing_hero_title()}
						</h1>
						<p className="mt-4 text-lg text-muted-foreground text-pretty">
							{m.marketing_hero_description()}
						</p>
						<div className="mt-8 flex flex-wrap items-center justify-center gap-3">
							<SiteAuthButton
								signInLabel={m.common_get_started()}
								dashboardLabel={m.common_open_dashboard()}
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
								{m.common_view_docs()}
							</a>
						</div>
					</div>
				</section>

				<section className="border-t border-border bg-muted/30">
					<div className="mx-auto max-w-5xl px-4 py-16 sm:px-6">
						<div className="mb-10 max-w-xl">
							<h2 className="font-heading text-2xl font-semibold tracking-tight">
								{m.marketing_features_title()}
							</h2>
							<p className="mt-2 text-muted-foreground">
								{m.marketing_features_description()}
							</p>
						</div>

						<div className="grid gap-4 sm:grid-cols-3">
							{features.map(({ id, icon: Icon, title, description }) => (
								<Card key={id} size="sm">
									<CardHeader>
										<div className="mb-2 inline-flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
											<Icon className="size-4" />
										</div>
										<CardTitle>{title()}</CardTitle>
										<CardDescription>{description()}</CardDescription>
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
								{m.marketing_cta_title()}
							</CardTitle>
							<CardDescription className="text-primary-foreground/80">
								{m.marketing_cta_description()}
							</CardDescription>
						</CardHeader>
						<CardContent>
							<SiteAuthButton
								signInLabel={m.common_open_app()}
								dashboardLabel={m.common_dashboard()}
								variant="secondary"
							/>
						</CardContent>
					</Card>
				</section>
			</main>
		</SiteLayout>
	);
}
