import { createFileRoute, getRouteApi, notFound } from "@tanstack/react-router";
import { RunnerJobScreen } from "@/components/runners/runner-job-screen";
import { ApiError } from "@/lib/api/client";
import {
	fetchRunnerJob,
	fetchRunnerJobRun,
	fetchRunnerJobRuns,
} from "@/lib/api/runner-jobs";
import { fetchRunner } from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute(
	"/$account_slug/runners_/$runnerId_/jobs/$jobId_/runs/$runId",
)({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const runnerId = params?.runnerId ?? "";
		const jobId = params?.jobId ?? "";
		const runId = params?.runId ?? "";
		try {
			const [{ runner }, job, { runs }, run] = await Promise.all([
				fetchRunner(slug, runnerId),
				fetchRunnerJob(slug, runnerId, jobId),
				fetchRunnerJobRuns(slug, runnerId, jobId),
				fetchRunnerJobRun(slug, runnerId, jobId, runId),
			]);
			return { runner, job, runs, run };
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: RunnerRunPage,
});

function RunnerRunPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const { runner, job, runs, run } = Route.useLoaderData();

	return (
		<RunnerJobScreen
			slug={slug}
			runner={runner}
			job={job}
			runs={runs}
			selectedRun={run}
		/>
	);
}
