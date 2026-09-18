import {
	createFileRoute,
	getRouteApi,
	useNavigate,
	useRouter,
} from "@tanstack/react-router";
import { Plus } from "lucide-react";
import { useState } from "react";

import {
	BeepList,
	type FilterStatus,
	getFilterOptions,
} from "@/components/beeps/beep-list";
import { CreateBeepDialog } from "@/components/beeps/create-beep-dialog";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Button } from "@/components/ui/button";
import {
	type BeepSortDir,
	type BeepSortField,
	beepSortQuery,
	fetchBeepStats,
	fetchBeeps,
	isBeepSortField,
	parseBeepSort,
} from "@/lib/api/beeps";
import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

type BeepsSearch = {
	tab?: FilterStatus;
	q?: string;
	sort?: BeepSortField;
	dir?: BeepSortDir;
};

export const Route = createFileRoute("/$account_slug/beeps")({
	validateSearch: (search: Record<string, unknown>): BeepsSearch => {
		const validTabs: FilterStatus[] = [
			"active",
			"firing",
			"recurring",
			"completed",
			"all",
		];
		const tab =
			typeof search.tab === "string" &&
			validTabs.includes(search.tab as FilterStatus)
				? (search.tab as FilterStatus)
				: "active";
		const q =
			typeof search.q === "string" && search.q.trim().length > 0
				? search.q.trim()
				: undefined;
		const sort =
			typeof search.sort === "string" && isBeepSortField(search.sort)
				? search.sort
				: undefined;
		const dir =
			search.dir === "asc" || search.dir === "desc" ? search.dir : undefined;
		return { tab, q, sort, dir };
	},
	loaderDeps: ({ search }) => ({
		tab: search.tab,
		q: search.q,
		sort: search.sort,
		dir: search.dir,
	}),
	loader: withAuthRedirects(async ({ params, deps, abortController }) => {
		const slug = params?.account_slug ?? "";
		const beepsDeps = deps as BeepsSearch;
		const statusFilter = beepsDeps?.tab ?? "active";
		const { sort, dir } = parseBeepSort(beepsDeps.sort, beepsDeps.dir);
		const [beepsRes, statsRes, settingsRes] = await Promise.all([
			fetchBeeps(slug, {
				...getFilterOptions(statusFilter),
				q: beepsDeps?.q,
				...beepSortQuery(sort, dir),
				signal: abortController?.signal,
			}),
			fetchBeepStats(slug),
			fetchSettings(slug).catch(() => null),
		]);
		return {
			beeps: beepsRes.beeps,
			pagination: beepsRes.pagination,
			stats: statsRes.stats,
			settings: settingsRes,
			currentTab: statusFilter,
			searchQuery: beepsDeps?.q ?? "",
			sort,
			dir,
		};
	}),
	component: BeepsPage,
});

function BeepsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const navigate = useNavigate({ from: Route.fullPath });
	const {
		beeps,
		pagination,
		stats,
		settings,
		currentTab,
		searchQuery,
		sort,
		dir,
	} = Route.useLoaderData();
	const [isCreateOpen, setIsCreateOpen] = useState(false);

	async function handleCreated() {
		await router.invalidate();
	}

	function handleTabChange(nextTab: FilterStatus) {
		void navigate({
			search: (prev) => ({
				...prev,
				tab: nextTab === "active" ? undefined : nextTab,
			}),
		});
	}

	function handleSearchChange(nextQ: string) {
		void navigate({
			search: (prev) => ({
				...prev,
				q: nextQ.trim() ? nextQ.trim() : undefined,
			}),
			replace: true,
		});
	}

	function handleSortChange(nextSort: BeepSortField, nextDir: BeepSortDir) {
		const query = beepSortQuery(nextSort, nextDir);
		void navigate({
			search: (prev) => ({
				...prev,
				sort: query.sort,
				dir: query.dir,
			}),
			replace: true,
		});
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: m.nav_beeps(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div className="flex flex-col gap-1">
						<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
							{m.beeps_page_title()}
						</h1>
						<p className="text-sm text-muted-foreground">
							{m.beeps_manage_description()}
						</p>
					</div>
					<Button
						onClick={() => setIsCreateOpen(true)}
						className="gap-2 shrink-0 self-start sm:self-auto"
					>
						<Plus className="size-4" />
						{m.beeps_create_beep()}
					</Button>
				</div>

				<BeepList
					beeps={beeps}
					initialPagination={pagination}
					stats={stats}
					slug={slug}
					variant="full"
					onCreateClick={() => setIsCreateOpen(true)}
					currentTab={currentTab}
					onTabChange={handleTabChange}
					searchQuery={searchQuery}
					onSearchChange={handleSearchChange}
					sort={sort}
					dir={dir}
					onSortChange={handleSortChange}
				/>

				<CreateBeepDialog
					slug={slug}
					open={isCreateOpen}
					onOpenChange={setIsCreateOpen}
					onCreated={handleCreated}
					defaultChannels={settings?.notification_channels}
				/>
			</div>
		</>
	);
}
