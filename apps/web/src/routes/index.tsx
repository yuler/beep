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

export const Route = createFileRoute("/")({ component: Home });

const features = [
	{
		icon: Box,
		title: "Monolith first",
		description:
			"Rails 8.1 core with Solid Queue, Cache, and Cable — one backend, one deploy.",
	},
	{
		icon: Layers,
		title: "Full stack, one repo",
		description:
			"Web, admin, mobile, and desktop apps share types, utilities, and API clients.",
	},
	{
		icon: Rocket,
		title: "Solo-optimized",
		description:
			"Opinionated defaults and minimal ceremony so you can ship without a team.",
	},
] as const;

function Home() {
	return (
		<SiteLayout>
			<main className="flex-1">
				<section className="mx-auto max-w-5xl px-4 py-16 sm:px-6 sm:py-24">
					<div className="mx-auto flex max-w-2xl flex-col items-center text-center">
						<Badge variant="secondary" className="mb-4">
							Monolith + Solo
						</Badge>
						<h1 className="font-heading text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
							One repository for your entire product
						</h1>
						<p className="mt-4 text-lg text-muted-foreground text-pretty">
							beep helps solo developers and one-person companies run a
							full-stack product from a single codebase — backend, web, and
							clients together.
						</p>
						<div className="mt-8 flex flex-wrap items-center justify-center gap-3">
							<SiteAuthButton
								signInLabel="Get started"
								dashboardLabel="Open dashboard"
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
								View docs
							</a>
						</div>
					</div>
				</section>

				<section className="border-t border-border bg-muted/30">
					<div className="mx-auto max-w-5xl px-4 py-16 sm:px-6">
						<div className="mb-10 max-w-xl">
							<h2 className="font-heading text-2xl font-semibold tracking-tight">
								Everything in one place
							</h2>
							<p className="mt-2 text-muted-foreground">
								A minimal, modern stack with semantic design tokens and light /
								dark themes.
							</p>
						</div>

						<div className="grid gap-4 sm:grid-cols-3">
							{features.map(({ icon: Icon, title, description }) => (
								<Card key={title} size="sm">
									<CardHeader>
										<div className="mb-2 inline-flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
											<Icon className="size-4" />
										</div>
										<CardTitle>{title}</CardTitle>
										<CardDescription>{description}</CardDescription>
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
								Ready to build?
							</CardTitle>
							<CardDescription className="text-primary-foreground/80">
								Run{" "}
								<code className="rounded bg-primary-foreground/10 px-1.5 py-0.5 text-sm">
									mise dev
								</code>{" "}
								to start the Rails core and this web app together.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<SiteAuthButton
								signInLabel="Open app"
								dashboardLabel="Dashboard"
								variant="secondary"
							/>
						</CardContent>
					</Card>
				</section>
			</main>
		</SiteLayout>
	);
}
