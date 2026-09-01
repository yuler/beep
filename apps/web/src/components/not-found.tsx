import { Link } from "@tanstack/react-router";
import { SiteLayout } from "@/components/layout";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import * as m from "@/locale/paraglide/messages";

export function NotFound() {
	return (
		<SiteLayout>
			<main className="mx-auto flex w-full max-w-5xl flex-1 flex-col items-center justify-center px-4 py-16 text-center sm:px-6">
				<p className="text-sm font-medium text-muted-foreground">
					{m.errors_not_found_code()}
				</p>
				<h1 className="mt-2 font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
					{m.errors_not_found()}
				</h1>
				<p className="mt-3 max-w-md text-muted-foreground text-pretty">
					{m.errors_not_found_description()}
				</p>
				<Link to="/" className={cn(buttonVariants({ size: "lg" }), "mt-8")}>
					{m.common_back()}
				</Link>
			</main>
		</SiteLayout>
	);
}
