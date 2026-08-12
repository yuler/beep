import {
	Card,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export function StatCard({ label, value }: { label: string; value: number }) {
	return (
		<Card size="sm">
			<CardHeader>
				<CardDescription>{label}</CardDescription>
				<CardTitle className="text-2xl tabular-nums">{value}</CardTitle>
			</CardHeader>
		</Card>
	);
}
