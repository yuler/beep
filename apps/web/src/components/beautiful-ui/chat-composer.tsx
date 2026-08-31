import { type HTMLAttributes, useEffect, useMemo, useRef, useState } from "react";

import { useTranslation } from "@/lib/i18n";

type Phase = "idle" | "sent" | "reply1" | "reply2" | "done";

function Section({
	label,
	sub,
	time,
	body,
	resolving,
}: {
	label: string;
	sub: string;
	time: string;
	body: string;
	resolving?: boolean;
}) {
	return (
		<div
			className="flex w-full flex-col gap-1.5 transition-[opacity,filter,transform] duration-400"
			style={{
				opacity: resolving ? 0.55 : 1,
				filter: resolving ? "blur(0.5px)" : "blur(0)",
				transform: resolving ? "scale(0.985)" : "scale(1)",
				transformOrigin: "top left",
				transitionTimingFunction: "cubic-bezier(0.23, 1, 0.32, 1)",
				animation: "bui-fade-up 400ms cubic-bezier(0.23,1,0.32,1) both",
			}}
		>
			<div className="flex items-center gap-1 text-[12px] leading-[1.3]">
				<span className="font-medium text-ink">{label}</span>
				<span className="text-ink-2">{sub}</span>
				<span className="text-ink">for {time}</span>
			</div>
			<p className="text-[13px] leading-normal text-ink">{body}</p>
		</div>
	);
}

export function ChatComposer({
	onCollapse,
	dragHandleProps,
}: {
	onCollapse?: () => void;
	dragHandleProps?: HTMLAttributes<HTMLButtonElement>;
} = {}) {
	const { t } = useTranslation();
	const mockChatReplies = useMemo(
		() => [
			{
				label: t("dev.mock_reply_1_label"),
				sub: t("dev.mock_reply_1_sub"),
				time: t("dev.mock_reply_1_time"),
				body: t("dev.mock_reply_1_body"),
			},
			{
				label: t("dev.mock_reply_2_label"),
				sub: t("dev.mock_reply_2_sub"),
				time: t("dev.mock_reply_2_time"),
				body: t("dev.mock_reply_2_body"),
			},
		],
		[t],
	);
	const [phase, setPhase] = useState<Phase>("done");
	const [draft, setDraft] = useState("");
	const [submitted, setSubmitted] = useState(() => t("dev.mock_starter_message"));
	const [tab, setTab] = useState<"beeps" | "history">("beeps");
	const chatTabs = useMemo(
		() => [
			{ id: "beeps" as const, label: t("dev.chat_tab_beeps") },
			{ id: "history" as const, label: t("dev.chat_tab_history") },
		],
		[t],
	);
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		let timer: ReturnType<typeof setTimeout> | undefined;
		if (phase === "sent") timer = setTimeout(() => setPhase("reply1"), 500);
		else if (phase === "reply1")
			timer = setTimeout(() => setPhase("reply2"), 1400);
		else if (phase === "reply2")
			timer = setTimeout(() => setPhase("done"), 1200);
		else return;
		return () => {
			if (timer) clearTimeout(timer);
		};
	}, [phase]);

	const sent = phase !== "idle";
	const canSend = draft.trim().length > 0;

	const send = () => {
		if (!canSend) return;
		setSubmitted(draft.trim());
		setDraft("");
		setPhase("sent");
	};

	return (
		<div className="flex min-h-0 flex-1 flex-col overflow-hidden">
			<div className="flex shrink-0 items-center justify-between border-b border-line p-1.5">
				<div className="flex min-w-0 items-center gap-0.5">
					{dragHandleProps ? (
						<button
							type="button"
							aria-label="Drag chat"
							className="flex size-6 shrink-0 cursor-grab touch-none items-center justify-center rounded-[6px] text-ink-3 active:cursor-grabbing hover:bg-hover hover:text-ink-2"
							{...dragHandleProps}
						>
							<svg
								width="14"
								height="14"
								viewBox="0 0 24 24"
								fill="currentColor"
								aria-hidden="true"
							>
								<circle cx="9" cy="6" r="1.5" />
								<circle cx="15" cy="6" r="1.5" />
								<circle cx="9" cy="12" r="1.5" />
								<circle cx="15" cy="12" r="1.5" />
								<circle cx="9" cy="18" r="1.5" />
								<circle cx="15" cy="18" r="1.5" />
							</svg>
						</button>
					) : null}
					<div className="flex items-center">
						{chatTabs.map((item) => (
							<button
								key={item.id}
								type="button"
								aria-pressed={tab === item.id}
								onClick={() => setTab(item.id)}
								className={`rounded-[6px] px-2 py-[3px] text-[13px] text-ink transition-[background-color,opacity] duration-100 ${tab === item.id ? "bg-field" : "opacity-50 hover:opacity-75"}`}
							>
								{item.label}
							</button>
						))}
					</div>
				</div>
				<div className="flex items-center gap-1">
					{onCollapse ? (
						<button
							type="button"
							aria-label="Collapse chat"
							onClick={onCollapse}
							className="flex size-6 items-center justify-center rounded-full text-ink-3 transition-colors duration-100 hover:bg-hover hover:text-ink-2"
						>
							<svg
								width="15"
								height="15"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
								strokeLinejoin="round"
								aria-hidden="true"
							>
								<path d="M5 12h14" />
							</svg>
						</button>
					) : null}
					{(
						[
							{ key: "add", label: t("common.add"), icon: <path d="M12 5v14M5 12h14" /> },
							{
								key: "history",
								label: t("common.history"),
								icon: (
									<g>
										<circle cx="12" cy="12" r="9" />
										<path d="M12 7v5l3 2" />
									</g>
								),
							},
							{
								key: "menu",
								label: t("common.more"),
								icon: (
									<g fill="currentColor" stroke="none">
										<circle cx="5" cy="12" r="1.8" />
										<circle cx="12" cy="12" r="1.8" />
										<circle cx="19" cy="12" r="1.8" />
									</g>
								),
							},
						] as const
					).map(({ key, label, icon }) => (
						<button
							key={key}
							type="button"
							aria-label={label}
							className="flex size-6 items-center justify-center rounded-[6px] text-ink-3 transition-colors duration-100 hover:bg-hover hover:text-ink-2"
						>
							<svg
								width="15"
								height="15"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
								strokeLinejoin="round"
								aria-hidden="true"
							>
								{icon}
							</svg>
						</button>
					))}
				</div>
			</div>

			<div className="flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto px-3 pt-2.5 pb-1">
				<div className="flex justify-end pl-14">
					<div
						className="rounded-xl bg-field px-3 py-1.5 text-[13px] leading-[1.4] text-ink transition-[opacity,transform] duration-300"
						style={{
							opacity: sent ? 1 : 0,
							transform: sent ? "translateY(0)" : "translateY(10px)",
							transitionTimingFunction: "cubic-bezier(0.23, 1, 0.32, 1)",
						}}
					>
						{submitted}
					</div>
				</div>

				{phase === "reply1" || phase === "reply2" || phase === "done" ? (
					<Section {...mockChatReplies[0]} />
				) : null}
				{phase === "reply2" || phase === "done" ? (
					<Section {...mockChatReplies[1]} resolving={phase === "reply2"} />
				) : null}
			</div>

			<div className="mt-auto shrink-0 p-1.5">
				<label
					htmlFor="beep-chat-prompt"
					className="flex cursor-text flex-col gap-2 rounded-control border border-line bg-field p-2.5 shadow-[0_1px_2px_rgba(0,0,0,0.035)] transition-[border-color,box-shadow] duration-150 focus-within:border-line-strong focus-within:shadow-[0_1px_2px_rgba(0,0,0,0.025)]"
				>
					<input
						ref={inputRef}
						id="beep-chat-prompt"
						value={draft}
						onChange={(event) => setDraft(event.target.value)}
						onKeyDown={(event) => {
							if (event.key === "Enter") send();
						}}
						placeholder={t("dev.chat_placeholder")}
						aria-label={t("dev.chat_placeholder")}
						className="min-h-4.5 bg-transparent text-[13px] leading-[1.4] text-ink outline-none placeholder:text-ink-3"
					/>
					<div className="flex items-center justify-end">
						<button
							type="button"
							aria-label="Send"
							disabled={!canSend}
							onClick={send}
							className="flex size-7 items-center justify-center rounded-[8px] transition-[background-color,color,transform] duration-200 enabled:active:scale-[0.96]"
							style={{
								background: canSend ? "var(--ink)" : "var(--line-strong)",
								color: canSend ? "var(--surface)" : "var(--ink-2)",
							}}
						>
							<svg
								width="16"
								height="16"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2.4"
								strokeLinecap="round"
								strokeLinejoin="round"
								aria-hidden="true"
							>
								<path d="M12 19V5M5 12l7-7 7 7" />
							</svg>
						</button>
					</div>
				</label>
			</div>
		</div>
	);
}
