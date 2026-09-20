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
		color: "#10b981",
	},
	failure: {
		label: "Failed / Alert",
		color: "#f43f5e",
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
		<Card className="bg-card/60 transition-colors hover:bg-card">
			<CardHeader className="flex flex-row items-center justify-between gap-4 pb-2">
				<div>
					<CardTitle className="font-heading text-lg font-semibold tracking-tight">
						Execution & Health Trends
					</CardTitle>
					<CardDescription className="text-xs">
						{chart.range === "24h"
							? "Hourly performance and failure rate over the past 24 hours"
							: "Daily execution volume and system stability over the last 7 days"}
					</CardDescription>
				</div>
				<div className="flex items-center gap-1 rounded-lg border border-border p-0.5 bg-muted/40">
					<Button
						variant={chart.range === "24h" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => onRangeChange("24h")}
						disabled={isLoading}
					>
						24h
					</Button>
					<Button
						variant={chart.range === "7d" ? "secondary" : "ghost"}
						size="sm"
						className="h-7 text-xs px-2.5"
						onClick={() => onRangeChange("7d")}
						disabled={isLoading}
					>
						7d
					</Button>
				</div>
			</CardHeader>
			<CardContent className="pt-4">
				{!hasData ? (
					<div className="flex h-56 flex-col items-center justify-center rounded-lg border border-dashed border-border/80 text-muted-foreground">
						<p className="text-sm font-medium">No activity recorded yet</p>
						<p className="text-xs text-muted-foreground/80">
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
									<stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
									<stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
								</linearGradient>
								<linearGradient id="fillFailure" x1="0" y1="0" x2="0" y2="1">
									<stop offset="5%" stopColor="#f43f5e" stopOpacity={0.4} />
									<stop offset="95%" stopColor="#f43f5e" stopOpacity={0.0} />
								</linearGradient>
							</defs>
							<CartesianGrid
								strokeDasharray="3 3"
								vertical={false}
								stroke="currentColor"
								className="stroke-border/40"
							/>
							<XAxis
								dataKey="label"
								tickLine={false}
								axisLine={false}
								tickMargin={8}
								fontSize={11}
							/>
							<YAxis
								tickLine={false}
								axisLine={false}
								fontSize={11}
								allowDecimals={false}
							/>
							<ChartTooltip
								content={
									<ChartTooltipContent
										indicator="dot"
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
								stroke="#10b981"
								strokeWidth={2}
								fill="url(#fillSuccess)"
							/>
							<Area
								type="monotone"
								dataKey="failure"
								name="Failed / Alert"
								stroke="#f43f5e"
								strokeWidth={2}
								fill="url(#fillFailure)"
							/>
						</AreaChart>
					</ChartContainer>
				)}
			</CardContent>
		</Card>
	);
}
