import { useRef, useState } from "react";

import { mockFineTuneTypes } from "@/lib/mock/ai-chat";

export type FineTuneValues = {
	segment: number;
	width: number;
	height: number;
	radius: number;
	opacity: number;
	typeValue: string;
};

export const defaultFineTuneValues: FineTuneValues = {
	segment: 0,
	width: 324,
	height: 96,
	radius: 28,
	opacity: 100,
	typeValue: "Select type",
};

function ScrubField({
	label,
	value,
	onChange,
	min,
	max,
	step = 1,
	suffix = "",
	active,
}: {
	label: string;
	value: number;
	onChange: (value: number) => void;
	min: number;
	max: number;
	step?: number;
	suffix?: string;
	active?: boolean;
}) {
	const drag = useRef<{ x: number; v: number } | null>(null);
	const clamp = (next: number) =>
		Math.min(max, Math.max(min, Math.round(next)));

	return (
		<label
			className="flex h-6.5 min-w-0 items-center gap-1 rounded-chip py-1 pr-1 pl-0.5 transition-[background-color,box-shadow] duration-200"
			style={{
				background: active ? "var(--accent-tint)" : "var(--field)",
				boxShadow: active ? "0 0 0 1px var(--accent)" : "none",
			}}
		>
			<span
				role="slider"
				aria-label={label}
				aria-valuenow={value}
				aria-valuemin={min}
				aria-valuemax={max}
				tabIndex={0}
				onPointerDown={(event) => {
					(event.target as HTMLElement).setPointerCapture(event.pointerId);
					drag.current = { x: event.clientX, v: value };
				}}
				onPointerMove={(event) => {
					if (!drag.current) return;
					onChange(
						clamp(
							drag.current.v + ((event.clientX - drag.current.x) / 2) * step,
						),
					);
				}}
				onPointerUp={() => {
					drag.current = null;
				}}
				onKeyDown={(event) => {
					const mult = event.shiftKey ? 10 : 1;
					if (event.key === "ArrowUp" || event.key === "ArrowRight") {
						event.preventDefault();
						onChange(clamp(value + step * mult));
					} else if (event.key === "ArrowDown" || event.key === "ArrowLeft") {
						event.preventDefault();
						onChange(clamp(value - step * mult));
					}
				}}
				className="flex h-full shrink-0 cursor-ew-resize touch-none items-center rounded-[4px] px-0.5 text-[12px] text-ink-3 select-none hover:text-ink-2 focus-visible:text-accent-ink focus-visible:outline-none"
			>
				{label}
			</span>
			<input
				inputMode="numeric"
				value={value}
				onChange={(event) => {
					const next = Number(event.target.value.replace(/[^\d-]/g, ""));
					if (!Number.isNaN(next)) onChange(clamp(next));
				}}
				aria-label={`${label} value`}
				className="min-w-0 flex-1 bg-transparent text-[12px] text-ink tabular-nums outline-none"
			/>
			{suffix ? (
				<span className="shrink-0 pr-0.5 text-[11.5px] text-ink-3">
					{suffix}
				</span>
			) : null}
		</label>
	);
}

const SEGMENTS = ["row", "col", "grid"] as const;

function SegmentIcon({ kind }: { kind: string }) {
	const dot = "size-1.5 rounded-[2px] border-[1.2px] border-current";
	if (kind === "row") {
		return (
			<span className="flex gap-0.5">
				{[0, 1, 2].map((index) => (
					<span key={index} className={dot} />
				))}
			</span>
		);
	}
	if (kind === "col") {
		return (
			<span className="flex flex-col gap-0.5">
				{[0, 1].map((index) => (
					<span key={index} className={dot} />
				))}
			</span>
		);
	}
	return (
		<span className="grid grid-cols-2 gap-0.5">
			{[0, 1, 2, 3].map((index) => (
				<span key={index} className={dot} />
			))}
		</span>
	);
}

export function FineTuneCard({
	values,
	onChange,
}: {
	values: FineTuneValues;
	onChange: (values: FineTuneValues) => void;
}) {
	const [menuOpen, setMenuOpen] = useState(false);
	const done =
		values.segment !== 0 ||
		values.width !== defaultFineTuneValues.width ||
		values.height !== defaultFineTuneValues.height ||
		values.radius !== defaultFineTuneValues.radius ||
		values.opacity !== defaultFineTuneValues.opacity ||
		values.typeValue !== defaultFineTuneValues.typeValue;

	const patch = (partial: Partial<FineTuneValues>) =>
		onChange({ ...values, ...partial });

	return (
		<div className="relative w-full max-w-60 rounded-card bg-surface shadow-raised">
			<div className="primitive-card-bar flex items-center justify-between border-b border-line">
				<span className="text-[13px] font-medium text-ink">Beep card</span>
				{done ? (
					<span
						className="flex items-center gap-1.5 text-[12px] font-medium text-green"
						style={{
							animation: "bui-pop-in 250ms cubic-bezier(0.23,1,0.32,1) both",
						}}
					>
						<svg
							width="10"
							height="10"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							strokeWidth="3"
							strokeLinecap="round"
							strokeLinejoin="round"
							aria-hidden="true"
						>
							<path d="M20 6L9 17l-5-5" />
						</svg>
						Edited
					</span>
				) : (
					<span className="flex items-center gap-1.5">
						<span className="flex size-4.5 items-center justify-center rounded-[5px] border border-accent/30 bg-accent-tint">
							<svg
								width="9"
								height="9"
								viewBox="0 0 24 24"
								fill="var(--accent)"
								aria-hidden="true"
							>
								<path d="M12 2l2.4 7.2L22 12l-7.6 2.8L12 22l-2.4-7.2L2 12l7.6-2.8z" />
							</svg>
						</span>
						<span
							className="bg-clip-text text-[12px] font-medium text-transparent"
							style={{
								backgroundImage:
									"linear-gradient(90deg, var(--accent) 35%, var(--accent-ink) 50%, var(--accent) 65%)",
								backgroundSize: "200% 100%",
								animation: "bui-shimmer-text 1.4s linear infinite",
							}}
						>
							Adjust
						</span>
					</span>
				)}
			</div>

			<div className="primitive-card-pad flex flex-col gap-2 border-b border-line">
				<p className="text-[12.5px] font-medium text-ink">Layout</p>
				<div className="relative grid grid-cols-3 rounded-control bg-field p-0.5">
					<span
						aria-hidden
						className="absolute inset-y-0.5 rounded-[6px] bg-surface shadow-btn transition-transform duration-300"
						style={{
							width: "calc((100% - 4px) / 3)",
							left: 2,
							transform: `translateX(${values.segment * 100}%)`,
							transitionTimingFunction: "cubic-bezier(0.23, 1, 0.32, 1)",
						}}
					/>
					{SEGMENTS.map((segment, index) => (
						<button
							key={segment}
							type="button"
							aria-label={`${segment} layout`}
							aria-pressed={index === values.segment}
							onClick={() => patch({ segment: index })}
							className={`relative z-10 flex h-6 items-center justify-center transition-colors duration-200 ${index === values.segment ? "text-accent" : "text-ink-3"}`}
						>
							<SegmentIcon kind={segment} />
						</button>
					))}
				</div>
				<div className="grid min-w-0 grid-cols-2 gap-2">
					<ScrubField
						label="W"
						value={values.width}
						onChange={(width) => patch({ width })}
						min={40}
						max={999}
						active={values.width !== defaultFineTuneValues.width}
					/>
					<ScrubField
						label="H"
						value={values.height}
						onChange={(height) => patch({ height })}
						min={24}
						max={999}
						active={values.height !== defaultFineTuneValues.height}
					/>
				</div>
				<div className="grid min-w-0 grid-cols-2 gap-2">
					<ScrubField
						label="Radius"
						value={values.radius}
						onChange={(radius) => patch({ radius })}
						min={0}
						max={64}
						active={values.radius !== defaultFineTuneValues.radius}
					/>
					<ScrubField
						label="Opacity"
						value={values.opacity}
						onChange={(opacity) => patch({ opacity })}
						min={0}
						max={100}
						suffix="%"
						active={values.opacity !== defaultFineTuneValues.opacity}
					/>
				</div>
			</div>

			<div className="primitive-card-footer flex items-center justify-between">
				<span className="text-[12px] text-ink-3">Type</span>
				<div className="relative -mr-0.5 w-30">
					<button
						type="button"
						aria-expanded={menuOpen}
						onClick={() => setMenuOpen((current) => !current)}
						className="flex h-6.5 w-full items-center justify-between rounded-chip bg-inset py-1 pr-1 pl-2 shadow-hairline transition-shadow duration-200 focus-visible:outline-none"
						style={{
							boxShadow: menuOpen ? "0 0 0 1px var(--accent)" : undefined,
						}}
					>
						<span
							className={`text-[12px] ${values.typeValue !== defaultFineTuneValues.typeValue ? "text-ink" : "text-ink-3"}`}
						>
							{values.typeValue}
						</span>
						<svg
							width="11"
							height="11"
							viewBox="0 0 24 24"
							fill="none"
							stroke="var(--ink-3)"
							strokeWidth="2.5"
							strokeLinecap="round"
							strokeLinejoin="round"
							className="transition-transform duration-200"
							style={{
								transform: menuOpen ? "rotate(180deg)" : "rotate(0)",
							}}
							aria-hidden="true"
						>
							<path d="M6 9l6 6 6-6" />
						</svg>
					</button>

					{menuOpen ? (
						<div
							className="absolute right-0 bottom-8 z-10 w-30 rounded-[10px] bg-surface p-1 shadow-raised"
							style={{
								animation: "bui-pop-in 200ms cubic-bezier(0.23,1,0.32,1) both",
								transformOrigin: "bottom right",
							}}
						>
							{mockFineTuneTypes.map((item) => (
								<button
									key={item}
									type="button"
									onClick={() => {
										patch({ typeValue: item });
										setMenuOpen(false);
									}}
									className="flex h-6.5 w-full items-center rounded-[6px] px-2 text-left text-[12.5px] text-ink transition-colors duration-150 hover:bg-field"
									style={{
										background:
											item === values.typeValue
												? "var(--field)"
												: "transparent",
									}}
								>
									{item}
								</button>
							))}
						</div>
					) : null}
				</div>
			</div>
		</div>
	);
}

export function FineTunePreview({ values }: { values: FineTuneValues }) {
	const layoutClass =
		values.segment === 1
			? "flex flex-col gap-2"
			: values.segment === 2
				? "grid grid-cols-2 gap-2"
				: "flex gap-2";

	return (
		<div className="flex min-h-[320px] flex-1 items-center justify-center rounded-window bg-canvas p-8 shadow-hairline">
			<div
				className="overflow-hidden border border-line bg-surface shadow-card transition-[width,height,border-radius,opacity] duration-300"
				style={{
					width: Math.min(values.width, 420),
					height: values.height,
					borderRadius: values.radius,
					opacity: values.opacity / 100,
				}}
			>
				<div className="border-b border-line px-3 py-2 text-[12px] font-medium text-ink">
					{values.typeValue === "Select type" ? "Reminder" : values.typeValue}{" "}
					beep
				</div>
				<div className={`p-3 ${layoutClass}`}>
					<div className="h-2 flex-1 rounded-full bg-field" />
					<div className="h-2 w-2/3 rounded-full bg-field" />
					{values.segment === 2 ? (
						<div className="h-2 flex-1 rounded-full bg-field" />
					) : null}
				</div>
			</div>
		</div>
	);
}

export function FineTuneWorkspace() {
	const [values, setValues] = useState(defaultFineTuneValues);

	return (
		<div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_240px]">
			<FineTunePreview values={values} />
			<FineTuneCard values={values} onChange={setValues} />
		</div>
	);
}
