import { type ReactNode, useEffect, useRef, useState } from "react";
import { StatusPill, type StatusPillTone } from "@/components/ui/status-pill";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

function useRevealOnScroll<T extends HTMLElement>() {
	const ref = useRef<T | null>(null);
	const [visible, setVisible] = useState(false);

	useEffect(() => {
		const el = ref.current;
		if (!el) return;
		if (typeof IntersectionObserver === "undefined") {
			setVisible(true);
			return;
		}
		const observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (entry.isIntersecting) {
						setVisible(true);
						observer.disconnect();
					}
				}
			},
			{ threshold: 0.18, rootMargin: "0px 0px -40px 0px" },
		);
		observer.observe(el);
		return () => observer.disconnect();
	}, []);

	return { ref, visible };
}

/* -------------------------------------------------------------------------- */
/* Story A Illustrations (1.5px monochrome line art + semantic accents)       */
/* -------------------------------------------------------------------------- */

function OutageSceneIllustration() {
	return (
		<svg
			viewBox="0 0 360 220"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Subtle grid backdrop */}
			<line
				x1="24"
				y1="52"
				x2="336"
				y2="52"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			<line
				x1="24"
				y1="170"
				x2="336"
				y2="170"
				stroke="var(--line)"
				strokeWidth="1"
			/>

			{/* 03:00 AM digital clock frame */}
			<rect
				x="24"
				y="22"
				width="86"
				height="22"
				rx="2"
				fill="var(--wash)"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			<circle
				cx="35"
				cy="33"
				r="3"
				fill="var(--danger)"
				className="story-pulse-danger"
			/>
			<text
				x="45"
				y="37"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				03:00 AM
			</text>

			{/* Server rack silhouette (left) */}
			<rect
				x="28"
				y="68"
				width="112"
				height="88"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			{/* Rack blades */}
			<line
				x1="28"
				y1="96"
				x2="140"
				y2="96"
				stroke="var(--line)"
				strokeWidth="1.2"
			/>
			<line
				x1="28"
				y1="126"
				x2="140"
				y2="126"
				stroke="var(--line)"
				strokeWidth="1.2"
			/>
			{/* Blade 1 healthy */}
			<circle cx="42" cy="82" r="3" fill="var(--success)" />
			<line
				x1="54"
				y1="82"
				x2="96"
				y2="82"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			{/* Blade 2 warning */}
			<circle cx="42" cy="111" r="3" fill="var(--warning)" />
			<line
				x1="54"
				y1="111"
				x2="108"
				y2="111"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			{/* Blade 3 failing */}
			<circle
				cx="42"
				cy="141"
				r="3.5"
				fill="var(--danger)"
				className="story-pulse-danger"
			/>
			<line
				x1="54"
				y1="141"
				x2="118"
				y2="141"
				stroke="var(--danger)"
				strokeWidth="1.5"
			/>

			{/* Heartbeat / latency waveform (right) */}
			<rect
				x="156"
				y="68"
				width="176"
				height="88"
				rx="2"
				fill="var(--ground)"
				stroke="var(--line)"
				strokeWidth="1.2"
			/>
			{/* Baseline */}
			<line
				x1="168"
				y1="132"
				x2="320"
				y2="132"
				stroke="var(--line)"
				strokeDasharray="2 3"
				strokeWidth="1"
			/>
			{/* Normal pulse transitioning to spike + flatline */}
			<polyline
				points="168,132 186,132 194,118 202,138 210,132 226,132"
				stroke="var(--gray)"
				strokeWidth="1.5"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
			<polyline
				points="226,132 238,82 248,146 258,90 268,132 318,132"
				stroke="var(--danger)"
				strokeWidth="1.75"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
			{/* Red fault marker */}
			<circle
				cx="238"
				cy="82"
				r="4"
				fill="var(--danger)"
				className="story-pulse-danger"
			/>
			<rect
				x="248"
				y="74"
				width="68"
				height="18"
				rx="2"
				fill="var(--danger-soft)"
			/>
			<text
				x="282"
				y="86"
				textAnchor="middle"
				fill="var(--danger)"
				fontSize="10"
				fontFamily="monospace"
			>
				HTTP 502
			</text>

			{/* Caption bar */}
			<text
				x="28"
				y="192"
				fill="var(--gray)"
				fontSize="11"
				fontFamily="monospace"
			>
				https://yuler.dev · https://beep.yuler.cc/up (5000ms)
			</text>
		</svg>
	);
}

function ProbeCatchIllustration() {
	return (
		<svg
			viewBox="0 0 360 220"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Radar / Probe target rings (left) */}
			<circle cx="88" cy="110" r="58" stroke="var(--line)" strokeWidth="1.2" />
			<circle
				cx="88"
				cy="110"
				r="38"
				stroke="var(--line)"
				strokeDasharray="3 3"
				strokeWidth="1.2"
			/>
			<circle cx="88" cy="110" r="18" stroke="currentColor" strokeWidth="1.5" />
			{/* Crosshair */}
			<line
				x1="22"
				y1="110"
				x2="154"
				y2="110"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			<line
				x1="88"
				y1="44"
				x2="88"
				y2="176"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			{/* Locked anomaly blip */}
			<circle
				cx="112"
				cy="84"
				r="9"
				fill="var(--danger-soft)"
				stroke="var(--danger)"
				strokeWidth="1.2"
			/>
			<circle
				cx="112"
				cy="84"
				r="4"
				fill="var(--danger)"
				className="story-pulse-danger"
			/>

			{/* Connector arrow */}
			<line
				x1="152"
				y1="110"
				x2="180"
				y2="110"
				stroke="currentColor"
				strokeWidth="1.5"
				className="story-dash-flow"
			/>

			{/* Beeper inspector card (right) */}
			<rect
				x="182"
				y="50"
				width="154"
				height="120"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			<text
				x="196"
				y="72"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				BEEPER · site-uptime
			</text>
			{/* Status transition: OK -> ALERTING */}
			<rect
				x="196"
				y="84"
				width="34"
				height="18"
				rx="2"
				fill="var(--ground)"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			<text
				x="213"
				y="96"
				textAnchor="middle"
				fill="var(--gray)"
				fontSize="9"
				fontFamily="monospace"
			>
				OK
			</text>
			<line
				x1="236"
				y1="93"
				x2="250"
				y2="93"
				stroke="var(--gray)"
				strokeWidth="1.2"
			/>
			<rect
				x="254"
				y="84"
				width="68"
				height="18"
				rx="2"
				fill="var(--danger-soft)"
				stroke="var(--danger)"
				strokeWidth="1"
			/>
			<text
				x="288"
				y="96"
				textAnchor="middle"
				fill="var(--danger)"
				fontSize="9"
				fontFamily="monospace"
			>
				ALERTING
			</text>

			{/* Consecutive run dots (2 errors) */}
			<text
				x="196"
				y="124"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				yuler.dev · */5 * * * *
			</text>
			<circle cx="199" cy="138" r="4" fill="var(--success)" />
			<circle cx="213" cy="138" r="4" fill="var(--success)" />
			<circle cx="227" cy="138" r="4" fill="var(--success)" />
			<circle cx="241" cy="138" r="4" fill="var(--warning)" />
			<circle cx="255" cy="138" r="4" fill="var(--danger)" />
			<circle
				cx="269"
				cy="138"
				r="4.5"
				fill="var(--danger)"
				className="story-pulse-danger"
			/>

			<text
				x="196"
				y="158"
				fill="currentColor"
				fontSize="10"
				fontFamily="monospace"
			>
				2/2 failed → trigger alert
			</text>
		</svg>
	);
}

function MultiChannelDispatchIllustration() {
	return (
		<svg
			viewBox="0 0 360 220"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Left: Ringing Beep Bell Hub */}
			<rect
				x="26"
				y="76"
				width="68"
				height="68"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			{/* Signal waves */}
			<path
				d="M18 90 C12 102, 12 118, 18 130"
				stroke="var(--warning)"
				strokeWidth="1.5"
				strokeLinecap="round"
			/>
			<path
				d="M102 90 C108 102, 108 118, 102 130"
				stroke="var(--warning)"
				strokeWidth="1.5"
				strokeLinecap="round"
			/>
			{/* Bell icon */}
			<g className="story-bell-swing">
				<path
					d="M60 92 C51 92 46 98 46 107 L46 116 L41 121 L79 121 L74 116 L74 107 C74 98 69 92 60 92 Z"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinejoin="round"
				/>
				<path
					d="M56 125 C57 128 63 128 64 125"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
				/>
				<circle cx="69" cy="93" r="3.5" fill="var(--danger)" />
			</g>

			{/* Fan-out lines */}
			<path
				d="M94 110 L146 52"
				stroke="currentColor"
				strokeWidth="1.4"
				className="story-dash-flow"
			/>
			<path
				d="M94 110 L146 110"
				stroke="currentColor"
				strokeWidth="1.4"
				className="story-dash-flow"
			/>
			<path
				d="M94 110 L146 168"
				stroke="currentColor"
				strokeWidth="1.4"
				className="story-dash-flow"
			/>

			{/* Channel 1: Web Push */}
			<rect
				x="148"
				y="30"
				width="186"
				height="44"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.4"
			/>
			<circle cx="164" cy="52" r="4" fill="var(--danger)" />
			<text
				x="176"
				y="48"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				Web Push · iPhone / Desktop
			</text>
			<text
				x="176"
				y="63"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				[ALERT] https://yuler.dev 502
			</text>

			{/* Channel 2: Email */}
			<rect
				x="148"
				y="88"
				width="186"
				height="44"
				rx="2"
				fill="var(--wash)"
				stroke="var(--line)"
				strokeWidth="1.4"
			/>
			<circle cx="164" cy="110" r="4" fill="var(--warning)" />
			<text
				x="176"
				y="106"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				Email · is.yuler@gmail.com
			</text>
			<text
				x="176"
				y="121"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				site-uptime · yuler.dev alerting
			</text>

			{/* Channel 3: Webhook */}
			<rect
				x="148"
				y="146"
				width="186"
				height="44"
				rx="2"
				fill="var(--wash)"
				stroke="var(--line)"
				strokeWidth="1.4"
			/>
			<circle cx="164" cy="168" r="4" fill="var(--success)" />
			<text
				x="176"
				y="164"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				Webhook · beep.yuler.cc
			</text>
			<text
				x="176"
				y="179"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				200 OK · delivered in 24ms
			</text>
		</svg>
	);
}

function RecoverySceneIllustration() {
	return (
		<svg
			viewBox="0 0 360 220"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Outer recovery panel */}
			<rect
				x="24"
				y="32"
				width="312"
				height="156"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>

			{/* Header row */}
			<circle cx="44" cy="56" r="5" fill="var(--success)" />
			<text
				x="56"
				y="60"
				fill="currentColor"
				fontSize="12"
				fontFamily="monospace"
			>
				https://yuler.dev
			</text>
			<rect
				x="236"
				y="45"
				width="84"
				height="20"
				rx="2"
				fill="var(--success-soft)"
			/>
			<text
				x="278"
				y="58"
				textAnchor="middle"
				fill="var(--success)"
				fontSize="10"
				fontFamily="monospace"
			>
				200 OK · 38ms
			</text>

			{/* Uptime bars (red outage dip -> all green recovery) */}
			<g>
				{[
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--warning)",
					"var(--danger)",
					"var(--danger)",
					"var(--warning)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
					"var(--success)",
				].map((color, idx) => (
					<rect
						// biome-ignore lint/suspicious/noArrayIndexKey: static bars
						key={idx}
						x={40 + idx * 15.5}
						y={82}
						width="10"
						height="34"
						rx="1"
						fill={color}
					/>
				))}
			</g>

			{/* Incident closed summary */}
			<line
				x1="40"
				y1="134"
				x2="320"
				y2="134"
				stroke="var(--line)"
				strokeWidth="1"
			/>
			<path
				d="M42 156 L47 161 L56 151"
				stroke="var(--success)"
				strokeWidth="1.8"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
			<text
				x="64"
				y="160"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				Resolved in 4m 20s · 99.98% 30d SLA
			</text>
		</svg>
	);
}

/* -------------------------------------------------------------------------- */
/* Story B Illustrations (Natural Language -> On-Time Delivery)               */
/* -------------------------------------------------------------------------- */

function NaturalLanguageParseIllustration() {
	return (
		<svg
			viewBox="0 0 360 210"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Prompt input box */}
			<rect
				x="20"
				y="24"
				width="320"
				height="52"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			<text
				x="34"
				y="44"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				https://beep.yuler.cc/yuler/beeps &gt;
			</text>
			<text x="34" y="62" fill="currentColor" fontSize="12">
				{m.marketing_story_b_scene1_input()}
			</text>

			{/* Parser flow lines */}
			<path
				d="M180 76 L180 112"
				stroke="currentColor"
				strokeWidth="1.5"
				className="story-dash-flow"
			/>
			<path d="M180 96 L88 112" stroke="var(--line)" strokeWidth="1.2" />
			<path d="M180 96 L272 112" stroke="var(--line)" strokeWidth="1.2" />

			{/* Parsed structured tokens */}
			<rect
				x="20"
				y="116"
				width="152"
				height="64"
				rx="2"
				fill="var(--ground)"
				stroke="var(--line)"
				strokeWidth="1.2"
			/>
			<circle cx="34" cy="134" r="3.5" fill="var(--warning)" />
			<text
				x="44"
				y="138"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				SCHEDULE AT
			</text>
			<text
				x="34"
				y="162"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				{m.marketing_story_b_scene1_parsed_time()}
			</text>

			<rect
				x="188"
				y="116"
				width="152"
				height="64"
				rx="2"
				fill="var(--ground)"
				stroke="var(--line)"
				strokeWidth="1.2"
			/>
			<circle cx="202" cy="134" r="3.5" fill="var(--success)" />
			<text
				x="212"
				y="138"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				DISPATCH ENGINE
			</text>
			<text
				x="202"
				y="162"
				fill="currentColor"
				fontSize="11"
				fontFamily="monospace"
			>
				{m.marketing_story_b_scene1_parsed_mode()}
			</text>
		</svg>
	);
}

function OnTimeDeliveryIllustration() {
	return (
		<svg
			viewBox="0 0 360 210"
			fill="none"
			aria-hidden="true"
			className="w-full h-auto text-ink"
		>
			{/* Precision analog/digital clock (left) */}
			<circle
				cx="74"
				cy="105"
				r="46"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			{/* Clock ticks */}
			<line
				x1="74"
				y1="64"
				x2="74"
				y2="70"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			<line
				x1="115"
				y1="105"
				x2="109"
				y2="105"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			<line
				x1="74"
				y1="146"
				x2="74"
				y2="140"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			<line
				x1="33"
				y1="105"
				x2="39"
				y2="105"
				stroke="var(--gray)"
				strokeWidth="1.5"
			/>
			{/* 09:00 hands */}
			<line
				x1="74"
				y1="105"
				x2="74"
				y2="76"
				stroke="currentColor"
				strokeWidth="1.8"
				strokeLinecap="round"
			/>
			<line
				x1="74"
				y1="105"
				x2="52"
				y2="105"
				stroke="var(--success)"
				strokeWidth="2"
				strokeLinecap="round"
			/>
			<circle cx="74" cy="105" r="3" fill="var(--success)" />

			{/* Connector */}
			<line
				x1="124"
				y1="105"
				x2="152"
				y2="105"
				stroke="var(--success)"
				strokeWidth="1.5"
				className="story-dash-flow"
			/>

			{/* Delivered Toast Card (right) */}
			<rect
				x="154"
				y="58"
				width="186"
				height="94"
				rx="2"
				fill="var(--wash)"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
			<rect x="154" y="58" width="4" height="94" fill="var(--success)" />
			<circle cx="174" cy="82" r="6" fill="var(--success-soft)" />
			<path
				d="M171 82 L173.5 84.5 L178 79.5"
				stroke="var(--success)"
				strokeWidth="1.5"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
			<text
				x="188"
				y="85"
				fill="var(--success)"
				fontSize="10"
				fontFamily="monospace"
			>
				DELIVERED · 09:00:00
			</text>
			<text x="170" y="110" fill="currentColor" fontSize="12">
				{m.marketing_story_b_scene2_toast_title()}
			</text>
			<text
				x="170"
				y="132"
				fill="var(--gray)"
				fontSize="10"
				fontFamily="monospace"
			>
				{m.marketing_story_b_scene2_toast_meta()}
			</text>
		</svg>
	);
}

/* -------------------------------------------------------------------------- */
/* Story Scene Row                                                            */
/* -------------------------------------------------------------------------- */

interface StorySceneItem {
	step: string;
	timeOrTag: () => string;
	statusLabel?: () => string;
	tone: StatusPillTone;
	title: () => string;
	description: () => string;
	illustration: ReactNode;
}

function StorySceneRow({
	scene,
	reverse,
}: {
	scene: StorySceneItem;
	reverse?: boolean;
}) {
	const { ref, visible } = useRevealOnScroll<HTMLDivElement>();

	return (
		<div
			ref={ref}
			className={cn(
				"grid items-center gap-6 border-t border-line py-8 md:grid-cols-12 md:gap-10 md:py-12 transition-all duration-500",
				visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-3",
			)}
		>
			{/* Illustration column (mobile first: always on top) */}
			<div
				className={cn("md:col-span-7", reverse ? "md:order-2" : "md:order-1")}
			>
				<div className="rounded-md border border-line bg-ground p-3 sm:p-5">
					{scene.illustration}
				</div>
			</div>

			{/* Narrative copy column */}
			<div
				className={cn(
					"flex flex-col gap-3 md:col-span-5",
					reverse ? "md:order-1" : "md:order-2",
				)}
			>
				<div className="flex flex-wrap items-center gap-2">
					<span className="font-mono text-xs text-gray">{scene.step}</span>
					<span className="font-mono text-xs text-ink">
						{scene.timeOrTag()}
					</span>
					{scene.statusLabel ? (
						<StatusPill label={scene.statusLabel()} tone={scene.tone} />
					) : null}
				</div>
				<h3 className="type-app-title text-lg font-medium text-ink sm:text-xl">
					{scene.title()}
				</h3>
				<p className="type-meta text-sm leading-relaxed text-gray sm:text-base">
					{scene.description()}
				</p>
			</div>
		</div>
	);
}

export function StorySection() {
	const incidentScenes: StorySceneItem[] = [
		{
			step: "01 / 04",
			timeOrTag: m.marketing_story_a_scene1_time,
			statusLabel: m.marketing_story_a_scene1_status,
			tone: "danger",
			title: m.marketing_story_a_scene1_title,
			description: m.marketing_story_a_scene1_desc,
			illustration: <OutageSceneIllustration />,
		},
		{
			step: "02 / 04",
			timeOrTag: m.marketing_story_a_scene2_time,
			statusLabel: m.marketing_story_a_scene2_status,
			tone: "danger",
			title: m.marketing_story_a_scene2_title,
			description: m.marketing_story_a_scene2_desc,
			illustration: <ProbeCatchIllustration />,
		},
		{
			step: "03 / 04",
			timeOrTag: m.marketing_story_a_scene3_time,
			statusLabel: m.marketing_story_a_scene3_status,
			tone: "warning",
			title: m.marketing_story_a_scene3_title,
			description: m.marketing_story_a_scene3_desc,
			illustration: <MultiChannelDispatchIllustration />,
		},
		{
			step: "04 / 04",
			timeOrTag: m.marketing_story_a_scene4_time,
			statusLabel: m.marketing_story_a_scene4_status,
			tone: "success",
			title: m.marketing_story_a_scene4_title,
			description: m.marketing_story_a_scene4_desc,
			illustration: <RecoverySceneIllustration />,
		},
	];

	const reminderScenes: StorySceneItem[] = [
		{
			step: "01 / 02",
			timeOrTag: m.marketing_story_b_scene1_tag,
			tone: "outline",
			title: m.marketing_story_b_scene1_title,
			description: m.marketing_story_b_scene1_desc,
			illustration: <NaturalLanguageParseIllustration />,
		},
		{
			step: "02 / 02",
			timeOrTag: m.marketing_story_b_scene2_tag,
			statusLabel: () => "09:00 OK",
			tone: "success",
			title: m.marketing_story_b_scene2_title,
			description: m.marketing_story_b_scene2_desc,
			illustration: <OnTimeDeliveryIllustration />,
		},
	];

	return (
		<>
			{/* Story A: 3 AM Outage -> Recovery */}
			<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
				<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center mb-10 md:mb-14">
						<div className="type-kicker uppercase text-gray mb-2">
							{m.marketing_story_a_badge()}
						</div>
						<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
							{m.marketing_story_a_title()}
						</h2>
						<p className="type-meta mt-3 text-base text-gray">
							{m.marketing_story_a_description()}
						</p>
					</div>

					<div>
						{incidentScenes.map((scene, index) => (
							<StorySceneRow
								key={scene.step}
								scene={scene}
								reverse={index % 2 === 1}
							/>
						))}
					</div>
				</div>
			</section>

			{/* Story B: Natural Language Reminder */}
			<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
				<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center mb-10 md:mb-14">
						<div className="type-kicker uppercase text-gray mb-2">
							{m.marketing_story_b_badge()}
						</div>
						<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
							{m.marketing_story_b_title()}
						</h2>
						<p className="type-meta mt-3 text-base text-gray">
							{m.marketing_story_b_description()}
						</p>
					</div>

					<div>
						{reminderScenes.map((scene, index) => (
							<StorySceneRow
								key={scene.step}
								scene={scene}
								reverse={index % 2 === 1}
							/>
						))}
					</div>
				</div>
			</section>
		</>
	);
}
