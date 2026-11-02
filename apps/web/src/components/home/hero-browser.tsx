import { Bell, Lock } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { StatusPill } from "@/components/ui/status-pill";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

type Tab = "beeps" | "beepers";
type BeepPhase = "compose" | "scheduling" | "scheduled";
type ProbePhase = "idle" | "checking" | "live";

const BEEPS_URL = "app.beep.dev/acme";
const BEEPERS_URL = "app.beep.dev/acme/beepers";

function usePrefersReducedMotion() {
	const [reduced, setReduced] = useState(false);

	useEffect(() => {
		const media = window.matchMedia("(prefers-reduced-motion: reduce)");
		const sync = () => setReduced(media.matches);
		sync();
		media.addEventListener("change", sync);
		return () => media.removeEventListener("change", sync);
	}, []);

	return reduced;
}

export function HeroBrowser() {
	const sample = m.marketing_hero_prompt_input();
	const inputId = useId();
	const rootRef = useRef<HTMLDivElement>(null);
	const userTimers = useRef<number[]>([]);
	const reducedMotion = usePrefersReducedMotion();

	const [inView, setInView] = useState(false);
	const [hidden, setHidden] = useState(false);
	const [manual, setManual] = useState(false);
	const [tab, setTab] = useState<Tab>("beeps");
	const [typed, setTyped] = useState("");
	const [beepPhase, setBeepPhase] = useState<BeepPhase>("compose");
	const [probePhase, setProbePhase] = useState<ProbePhase>("idle");
	const [toast, setToast] = useState(false);

	useEffect(() => {
		const node = rootRef.current;
		if (!node) {
			return;
		}

		const observer = new IntersectionObserver(
			([entry]) => setInView(entry.isIntersecting),
			{ threshold: 0.35 },
		);
		observer.observe(node);
		return () => observer.disconnect();
	}, []);

	useEffect(() => {
		const sync = () => setHidden(document.visibilityState === "hidden");
		sync();
		document.addEventListener("visibilitychange", sync);
		return () => document.removeEventListener("visibilitychange", sync);
	}, []);

	useEffect(() => {
		return () => {
			for (const timer of userTimers.current) {
				window.clearTimeout(timer);
			}
		};
	}, []);

	useEffect(() => {
		if (manual) {
			return;
		}

		if (reducedMotion) {
			setTyped(sample);
			setBeepPhase("scheduled");
			setProbePhase("live");
			setTab("beeps");
			setToast(false);
			return;
		}

		if (!inView || hidden) {
			return;
		}

		const timers = new Set<number>();
		let cancelled = false;
		const wait = (ms: number) =>
			new Promise<void>((resolve) => {
				timers.add(window.setTimeout(resolve, ms));
			});
		const alive = () => !cancelled;

		async function play() {
			while (alive()) {
				setTab("beeps");
				setTyped("");
				setBeepPhase("compose");
				setProbePhase("idle");
				setToast(false);
				await wait(240);
				if (!alive()) {
					return;
				}

				for (let index = 1; index <= sample.length; index += 1) {
					if (!alive()) {
						return;
					}
					setTyped(sample.slice(0, index));
					await wait(index === sample.length ? 420 : 26);
				}

				setBeepPhase("scheduling");
				await wait(780);
				if (!alive()) {
					return;
				}
				setBeepPhase("scheduled");
				setToast(true);
				await wait(1600);
				if (!alive()) {
					return;
				}

				setToast(false);
				setTab("beepers");
				setProbePhase("checking");
				await wait(900);
				if (!alive()) {
					return;
				}
				setProbePhase("live");
				await wait(2200);
			}
		}

		void play();

		return () => {
			cancelled = true;
			for (const timer of timers) {
				window.clearTimeout(timer);
			}
		};
	}, [hidden, inView, manual, reducedMotion, sample]);

	function takeOver() {
		setManual(true);
	}

	function selectTab(next: Tab) {
		takeOver();
		setTab(next);
		setToast(false);
		if (next === "beeps") {
			setTyped((current) => current || sample);
		} else {
			setProbePhase("live");
		}
	}

	function handleSchedule() {
		if (beepPhase === "scheduling") {
			return;
		}
		takeOver();
		setTab("beeps");
		if (!typed.trim()) {
			setTyped(sample);
		}
		setBeepPhase("scheduling");
		const scheduleAt = reducedMotion ? 0 : 720;
		const toastAt = reducedMotion ? 0 : 2400;
		const scheduledTimer = window.setTimeout(() => {
			setBeepPhase("scheduled");
			setToast(true);
			const toastTimer = window.setTimeout(() => setToast(false), toastAt);
			userTimers.current.push(toastTimer);
		}, scheduleAt);
		userTimers.current.push(scheduledTimer);
	}

	const url = tab === "beeps" ? BEEPS_URL : BEEPERS_URL;
	const canSubmit = typed.trim().length > 0 && beepPhase !== "scheduling";

	return (
		<section
			ref={rootRef}
			className="relative mx-auto mt-14 max-w-3xl text-left"
			aria-label={m.marketing_hero_mock_title()}
		>
			<div className="border border-line bg-ground">
				<div className="flex h-10 items-center gap-3 border-b border-line px-3">
					<div className="flex items-center gap-1.5" aria-hidden="true">
						<span className="size-2.5 border border-line" />
						<span className="size-2.5 border border-line" />
						<span className="size-2.5 border border-line" />
					</div>
					<div className="flex min-w-0 flex-1 items-stretch gap-0 self-end h-8">
						<TabButton
							active={tab === "beeps"}
							onClick={() => selectTab("beeps")}
						>
							{m.nav_beeps()}
						</TabButton>
						<TabButton
							active={tab === "beepers"}
							onClick={() => selectTab("beepers")}
						>
							{m.nav_beepers()}
						</TabButton>
					</div>
				</div>

				<div className="flex h-9 items-center gap-2 border-b border-line px-3">
					<Lock className="size-3 text-gray" aria-hidden="true" />
					<div className="min-w-0 flex-1 truncate font-mono type-kicker text-gray">
						https://{url}
					</div>
				</div>

				<div className="relative min-h-[280px] md:min-h-[300px] p-3 md:p-4">
					{tab === "beeps" ? (
						<BeepsPage
							inputId={inputId}
							typed={typed}
							phase={beepPhase}
							canSubmit={canSubmit}
							onTyped={(value) => {
								takeOver();
								setTyped(value);
								if (beepPhase === "scheduled") {
									setBeepPhase("compose");
								}
							}}
							onSchedule={handleSchedule}
						/>
					) : (
						<BeepersPage phase={probePhase} />
					)}

					{toast ? (
						<button
							type="button"
							onClick={() => {
								takeOver();
								setToast(false);
							}}
							className="absolute right-3 bottom-3 z-10 w-[min(100%-1.5rem,16.5rem)] rounded-sm border border-line border-l-2 border-l-success bg-ground p-3 text-left animate-in fade-in slide-in-from-bottom-2 duration-200"
						>
							<div className="flex items-center justify-between gap-2">
								<span className="inline-flex items-center gap-1.5 type-kicker text-success">
									<span className="size-1.5 rounded-full bg-success" />
									{m.term_beep()}
								</span>
								<span className="type-kicker text-gray">
									{m.marketing_hero_activity_now()}
								</span>
							</div>
							<div className="mt-1.5 text-sm text-ink">
								{m.marketing_hero_toast_title()}
							</div>
							<div className="type-meta mt-0.5 text-xs">
								{m.marketing_hero_toast_body()}
							</div>
						</button>
					) : null}
				</div>
			</div>
		</section>
	);
}

function TabButton({
	active,
	onClick,
	children,
}: {
	active: boolean;
	onClick: () => void;
	children: string;
}) {
	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"type-kicker -mb-px min-h-0 border border-b-0 px-3 text-xs uppercase tracking-[0.18em]",
				active
					? "border-line bg-ground text-ink"
					: "border-transparent text-gray hover:text-ink",
			)}
		>
			{children}
		</button>
	);
}

function BeepsPage({
	inputId,
	typed,
	phase,
	canSubmit,
	onTyped,
	onSchedule,
}: {
	inputId: string;
	typed: string;
	phase: BeepPhase;
	canSubmit: boolean;
	onTyped: (value: string) => void;
	onSchedule: () => void;
}) {
	const scheduled = phase === "scheduled";

	return (
		<div className="space-y-4">
			<div className="border border-line p-4">
				<label htmlFor={inputId} className="type-kicker text-gray block">
					{m.marketing_hero_mock_title()}
				</label>
				<form
					className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center"
					onSubmit={(event) => {
						event.preventDefault();
						onSchedule();
					}}
				>
					<Input
						id={inputId}
						value={typed}
						placeholder={m.marketing_hero_prompt_placeholder()}
						autoComplete="off"
						spellCheck={false}
						enterKeyHint="send"
						onChange={(event) => onTyped(event.target.value)}
						className="flex-1 px-0"
					/>
					<Button
						type="submit"
						variant="solid"
						size="sm"
						className="shrink-0"
						loading={phase === "scheduling"}
						loadingText={m.marketing_hero_scheduling()}
						disabled={!canSubmit && phase !== "scheduling"}
					>
						{m.marketing_hero_mock_schedule()}
					</Button>
				</form>
				<div className="mt-2.5 flex items-center gap-2 type-meta text-gray text-xs">
					<span>{m.marketing_hero_channels_title()}</span>
					<span className="font-mono text-ink">
						{m.marketing_hero_dispatch_status()}
					</span>
				</div>
			</div>

			{scheduled ? (
				<div className="border-t border-line animate-in fade-in slide-in-from-bottom-2 duration-300 motion-reduce:animate-none">
					<div className="flex items-start gap-3 px-1 py-2.5">
						<Bell className="mt-0.5 size-3.5 text-ink" aria-hidden="true" />
						<div className="min-w-0 flex-1">
							<div className="flex items-baseline justify-between gap-2">
								<span className="truncate text-sm text-ink">
									{typed.trim() || m.marketing_hero_mock_title()}
								</span>
								<span className="type-kicker shrink-0 text-gray">
									{m.marketing_hero_activity_now()}
								</span>
							</div>
							<p className="type-meta mt-0.5 truncate text-xs">
								{m.marketing_hero_prompt_scheduled()}
							</p>
						</div>
					</div>
				</div>
			) : null}
		</div>
	);
}

function BeepersPage({ phase }: { phase: ProbePhase }) {
	const live = phase === "live";

	return (
		<div className="border border-line p-4">
			<div className="flex items-start justify-between gap-3">
				<div>
					<div className="type-kicker text-gray">{m.nav_beepers()}</div>
					<h3 className="type-app-title mt-1 text-base font-medium text-ink">
						{m.marketing_hero_probe_name()}
					</h3>
					<p className="type-meta mt-1 text-xs">
						{m.marketing_hero_probe_cron()}
					</p>
				</div>
				<StatusPill
					label={
						live
							? m.marketing_hero_live_status()
							: m.marketing_hero_probe_checking()
					}
					tone={live ? "success" : "muted"}
				/>
			</div>
			<div className="mt-4 flex items-center gap-2 border-t border-line pt-3 text-sm text-ink">
				{live ? (
					<span className="font-mono text-xs sm:text-sm">
						{m.marketing_hero_probe_status()}
					</span>
				) : (
					<>
						<span className="is-loading-spin" aria-hidden="true" />
						<span>{m.marketing_hero_probe_checking()}</span>
					</>
				)}
			</div>
		</div>
	);
}
