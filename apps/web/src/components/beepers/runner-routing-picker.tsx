import { Cloud, Server, Tag } from "lucide-react";
import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { Runner } from "@/lib/api/runners";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export type RoutingMode = "core" | "runner" | "tag";

interface RunnerRoutingPickerProps {
	runners: Runner[];
	runnerId: string | null;
	runnerTag: string | null;
	onChange: (val: {
		runner_id: string | null;
		runner_tag: string | null;
	}) => void;
	disabled?: boolean;
}

export function RunnerRoutingPicker({
	runners,
	runnerId,
	runnerTag,
	onChange,
	disabled = false,
}: RunnerRoutingPickerProps) {
	const initialMode: RoutingMode = runnerId
		? "runner"
		: runnerTag
			? "tag"
			: "core";
	const [mode, setMode] = useState<RoutingMode>(initialMode);
	const [selectedRunnerId, setSelectedRunnerId] = useState<string>(
		runnerId || (runners[0]?.id ?? ""),
	);
	const [customTag, setCustomTag] = useState<string>(runnerTag || "");

	useEffect(() => {
		if (runnerId) {
			setMode("runner");
			setSelectedRunnerId(runnerId);
		} else if (runnerTag) {
			setMode("tag");
			setCustomTag(runnerTag);
		} else {
			setMode("core");
		}
	}, [runnerId, runnerTag]);

	function handleModeChange(newMode: RoutingMode) {
		setMode(newMode);
		if (newMode === "core") {
			onChange({ runner_id: null, runner_tag: null });
		} else if (newMode === "runner") {
			const targetId = selectedRunnerId || runners[0]?.id || null;
			onChange({ runner_id: targetId, runner_tag: null });
		} else if (newMode === "tag") {
			onChange({ runner_id: null, runner_tag: customTag.trim() || null });
		}
	}

	function handleRunnerSelect(id: string) {
		setSelectedRunnerId(id);
		onChange({ runner_id: id || null, runner_tag: null });
	}

	function handleTagChange(tag: string) {
		setCustomTag(tag);
		onChange({ runner_id: null, runner_tag: tag.trim() || null });
	}

	return (
		<div className="flex flex-col gap-3">
			<Label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
				{m.beepers_routing_title()}
			</Label>

			<div className="grid gap-2.5 sm:grid-cols-3">
				{/* 1. Cloud Core Option */}
				<button
					type="button"
					disabled={disabled}
					onClick={() => handleModeChange("core")}
					className={cn(
						"flex flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors cursor-pointer",
						mode === "core"
							? "border-primary bg-primary/5 text-primary"
							: "border-border hover:bg-muted/40 text-muted-foreground",
						disabled && "pointer-events-none opacity-50",
					)}
				>
					<div className="flex items-center gap-1.5 font-medium text-xs text-foreground">
						<Cloud className="size-3.5 text-sky-500" />
						<span>{m.beepers_routing_core()}</span>
					</div>
					<p className="text-[11px] leading-tight text-muted-foreground">
						{m.beepers_routing_core_desc()}
					</p>
				</button>

				{/* 2. Specific Runner Option */}
				<button
					type="button"
					disabled={disabled}
					onClick={() => handleModeChange("runner")}
					className={cn(
						"flex flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors cursor-pointer",
						mode === "runner"
							? "border-primary bg-primary/5 text-primary"
							: "border-border hover:bg-muted/40 text-muted-foreground",
						disabled && "pointer-events-none opacity-50",
					)}
				>
					<div className="flex items-center gap-1.5 font-medium text-xs text-foreground">
						<Server className="size-3.5 text-amber-500" />
						<span>{m.beepers_routing_node()}</span>
					</div>
					<p className="text-[11px] leading-tight text-muted-foreground">
						{m.beepers_routing_node_desc()}
					</p>
				</button>

				{/* 3. Tag Route Option */}
				<button
					type="button"
					disabled={disabled}
					onClick={() => handleModeChange("tag")}
					className={cn(
						"flex flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors cursor-pointer",
						mode === "tag"
							? "border-primary bg-primary/5 text-primary"
							: "border-border hover:bg-muted/40 text-muted-foreground",
						disabled && "pointer-events-none opacity-50",
					)}
				>
					<div className="flex items-center gap-1.5 font-medium text-xs text-foreground">
						<Tag className="size-3.5 text-emerald-500" />
						<span>{m.beepers_routing_tag()}</span>
					</div>
					<p className="text-[11px] leading-tight text-muted-foreground">
						{m.beepers_routing_tag_desc()}
					</p>
				</button>
			</div>

			{/* Sub-inputs based on mode */}
			{mode === "runner" ? (
				<div className="flex flex-col gap-1.5 rounded-lg border bg-muted/30 p-3">
					<Label
						htmlFor="runner-select"
						className="text-xs font-medium text-foreground"
					>
						{m.beepers_routing_node()}
					</Label>
					{runners.length === 0 ? (
						<p className="text-xs text-destructive">
							{m.runners_no_runners_hint()}
						</p>
					) : (
						<select
							id="runner-select"
							disabled={disabled}
							value={selectedRunnerId}
							onChange={(e) => handleRunnerSelect(e.target.value)}
							className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-xs text-foreground shadow-xs outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
						>
							<option value="">{m.beepers_select_runner_placeholder()}</option>
							{runners.map((r) => (
								<option key={r.id} value={r.id}>
									{r.name} ({r.status})
									{r.tags && r.tags.length > 0 ? ` [${r.tags.join(", ")}]` : ""}
								</option>
							))}
						</select>
					)}
				</div>
			) : null}

			{mode === "tag" ? (
				<div className="flex flex-col gap-1.5 rounded-lg border bg-muted/30 p-3">
					<Label
						htmlFor="runner-tag-input"
						className="text-xs font-medium text-foreground"
					>
						{m.beepers_routing_tag_label()}
					</Label>
					<Input
						id="runner-tag-input"
						disabled={disabled}
						placeholder={m.beepers_tag_placeholder()}
						value={customTag}
						onChange={(e) => handleTagChange(e.target.value)}
						className="text-xs font-mono"
					/>
				</div>
			) : null}
		</div>
	);
}
