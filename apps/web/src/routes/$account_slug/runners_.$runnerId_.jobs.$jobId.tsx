import { createFileRoute, getRouteApi, notFound } from "@tanstack/react-router";
import { RunnerJobScreen } from "@/components/runners/runner-job-screen";
import { ApiError } from "@/lib/api/client";
import { fetchRunnerJob, fetchRunnerJobRuns } from "@/lib/api/runner-jobs";
import { fetchRunner } from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute(
	"/$account_slug/runners_/$runnerId_/jobs/$jobId",
)({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const runnerId = params?.runnerId ?? "";
		const jobId = params?.jobId ?? "";
		try {
			const [{ runner }, job, { runs }] = await Promise.all([
				fetchRunner(slug, runnerId),
				fetchRunnerJob(slug, runnerId, jobId),
				fetchRunnerJobRuns(slug, runnerId, jobId),
			]);
			return { runner, job, runs };
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: RunnerJobPage,
});

function RunnerJobPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const { runner, job, runs } = Route.useLoaderData();

	return (
		<RunnerJobScreen
			slug={slug}
			runner={runner}
			job={job}
			runs={runs}
			selectedRun={null}
		/>
	);
}
