import { MessageCircle } from "lucide-react";
import { type PointerEvent, useRef, useState } from "react";

import { ChatComposer } from "@/components/beautiful-ui/chat-composer";
import { cn } from "@/lib/utils";

const PANEL_CLASS =
	"flex h-[min(520px,calc(100vh-6rem))] min-h-[420px] w-[min(calc(var(--spacing)*95),calc(100vw-3rem))] flex-col overflow-hidden rounded-[14px] bg-surface shadow-card";

const DEFAULT_OFFSET = { x: 0, y: 0 };

function useWidgetDrag(initial = DEFAULT_OFFSET) {
	const [offset, setOffset] = useState(initial);
	const drag = useRef<{
		pointerId: number;
		startX: number;
		startY: number;
		originX: number;
		originY: number;
		moved: boolean;
	} | null>(null);

	const bind = {
		onPointerDown: (event: PointerEvent<HTMLElement>) => {
			event.currentTarget.setPointerCapture(event.pointerId);
			drag.current = {
				pointerId: event.pointerId,
				startX: event.clientX,
				startY: event.clientY,
				originX: offset.x,
				originY: offset.y,
				moved: false,
			};
		},
		onPointerMove: (event: PointerEvent<HTMLElement>) => {
			if (!drag.current || drag.current.pointerId !== event.pointerId) return;
			const dx = event.clientX - drag.current.startX;
			const dy = event.clientY - drag.current.startY;
			if (Math.abs(dx) > 3 || Math.abs(dy) > 3) drag.current.moved = true;
			setOffset({
				x: drag.current.originX + dx,
				y: drag.current.originY + dy,
			});
		},
		onPointerUp: (event: PointerEvent<HTMLElement>) => {
			if (!drag.current || drag.current.pointerId !== event.pointerId) return;
			const moved = drag.current.moved;
			drag.current = null;
			event.currentTarget.releasePointerCapture(event.pointerId);
			return moved;
		},
	};

	return { offset, bind };
}

export function ChatWidget() {
	const [collapsed, setCollapsed] = useState(true);
	const { offset, bind } = useWidgetDrag();

	const positionStyle = {
		transform: `translate(${offset.x}px, ${offset.y}px)`,
	};

	if (collapsed) {
		return (
			<div
				className="pointer-events-none fixed right-6 bottom-6 z-50"
				style={positionStyle}
			>
				<button
					type="button"
					aria-label="Open chat"
					aria-expanded={false}
					className={cn(
						"pointer-events-auto flex size-12 cursor-grab touch-none items-center justify-center rounded-full bg-surface text-ink shadow-card transition-[transform,box-shadow] duration-200",
						"hover:shadow-raised active:cursor-grabbing active:scale-[0.97]",
					)}
					{...bind}
					onPointerUp={(event) => {
						const moved = bind.onPointerUp(event);
						if (!moved) setCollapsed(false);
					}}
				>
					<MessageCircle
						className="size-5"
						strokeWidth={2}
						aria-hidden="true"
					/>
				</button>
			</div>
		);
	}

	return (
		<div
			className="pointer-events-none fixed right-6 bottom-6 z-50 max-w-[calc(100vw-3rem)]"
			style={positionStyle}
		>
			<div className={cn("pointer-events-auto", PANEL_CLASS)}>
				<ChatComposer
					onCollapse={() => setCollapsed(true)}
					dragHandleProps={{
						...bind,
						onPointerUp: bind.onPointerUp,
					}}
				/>
			</div>
		</div>
	);
}
