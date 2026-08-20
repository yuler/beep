import { StatCard } from "@/components/dashboard/stat-card";
import type { Beep } from "@/lib/api/beeps";
import { beepStats } from "@/lib/beep-stats";

export function BeepStats({ beeps }: { beeps: Beep[] }) {
	const stats = beepStats(beeps);

	return (
		<div className="grid grid-cols-3 gap-3">
			<StatCard label="Active" value={stats.active} />
			<StatCard label="Due today" value={stats.dueToday} />
			<StatCard label="Firing" value={stats.firing} />
		</div>
	);
}
