import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	type ChartConfig,
	ChartContainer,
	ChartTooltip,
	ChartTooltipContent,
} from "@/components/ui/chart";
import type { DashboardChart } from "@/lib/api/dashboard";

const chartConfig = {
	success: {
		label: "Successful",
		color: "var(--ink)",
	},
	failure: {
		label: "Failed / Alert",
		color: "var(--gray)",
	},
} satisfies ChartConfig;

export function DashboardTrendChart({
	chart,
	onRangeChange,
	isLoading,
}: {
	chart: DashboardChart;
	onRangeChange: (range: "24h" | "7d") => void;
	isLoading?: boolean;
}) {
	const hasData = chart.points.some((p) => p.total > 0);

	return (
		<Card className="bg-transparent border-t border-line sm:border sm:border-line">
			<CardHeader className="flex flex-row items-center justify-between gap-4 pb-2">
				<div>
					<CardTitle className="type-app-title font-medium text-ink">
						Execution & Health Trends
					</CardTitle>
					<CardDescription className="type-meta text-gray">
						{chart.range === "24h"
							? "Hourly performance and failure rate over the past 24 hours"
							: "Daily execution volume and system stability over the last 7 days"}
					</CardDescription>
				</div>
				<div className="flex items-center border border-line">
					<Button
						variant={chart.range === "24h" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0 border-0"
						onClick={() => onRangeChange("24h")}
						disabled={isLoading}
					>
						24h
					</Button>
					<Button
						variant={chart.range === "7d" ? "solid" : "ghost"}
						size="xs"
						className="h-7 text-xs px-2.5 min-h-0 sm:min-h-0 border-0 border-l border-line"
						onClick={() => onRangeChange("7d")}
						disabled={isLoading}
					>
						7d
					</Button>
				</div>
			</CardHeader>
			<CardContent className="pt-4">
				{!hasData ? (
					<div className="flex h-56 flex-col items-center justify-center border border-line text-gray">
						<p className="text-sm font-normal text-ink">
							No activity recorded yet
						</p>
						<p className="type-meta text-gray">
							Triggers, probe health checks, and runner runs will show here.
						</p>
					</div>
				) : (
					<ChartContainer
						config={chartConfig}
						className="h-56 w-full aspect-auto"
					>
						<AreaChart
							data={chart.points}
							margin={{ top: 8, right: 12, left: -20, bottom: 0 }}
						>
							<defs>
								<linearGradient id="fillSuccess" x1="0" y1="0" x2="0" y2="1">
									<stop offset="5%" stopColor="var(--ink)" stopOpacity={0.2} />
									<stop offset="95%" stopColor="var(--ink)" stopOpacity={0.0} />
								</linearGradient>
								<linearGradient id="fillFailure" x1="0" y1="0" x2="0" y2="1">
									<stop offset="5%" stopColor="var(--gray)" stopOpacity={0.2} />
									<stop
										offset="95%"
										stopColor="var(--gray)"
										stopOpacity={0.0}
									/>
								</linearGradient>
							</defs>
							<CartesianGrid
								strokeDasharray="2 2"
								vertical={false}
								stroke="var(--line)"
							/>
							<XAxis
								dataKey="label"
								tickLine={false}
								axisLine={false}
								tickMargin={8}
								fontSize={11}
								stroke="var(--gray)"
							/>
							<YAxis
								tickLine={false}
								axisLine={false}
								fontSize={11}
								allowDecimals={false}
								stroke="var(--gray)"
							/>
							<ChartTooltip
								content={
									<ChartTooltipContent
										indicator="line"
										labelFormatter={(_, payload) => {
											const point = payload?.[0]?.payload as
												| (typeof chart.points)[number]
												| undefined;
											return point
												? `${point.label} (${point.total} total)`
												: "";
										}}
									/>
								}
							/>
							<Area
								type="monotone"
								dataKey="success"
								name="Successful"
								stroke="var(--ink)"
								strokeWidth={1.5}
								fill="url(#fillSuccess)"
							/>
							<Area
								type="monotone"
								dataKey="failure"
								name="Failed / Alert"
								stroke="var(--gray)"
								strokeDasharray="3 3"
								strokeWidth={1.5}
								fill="url(#fillFailure)"
							/>
						</AreaChart>
					</ChartContainer>
				)}
			</CardContent>
		</Card>
	);
}
