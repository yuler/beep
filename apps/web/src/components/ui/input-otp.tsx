import { type ClipboardEvent, type KeyboardEvent, useRef } from "react";

import { cn } from "@/lib/utils";

const OTP_SLOT_KEYS = ["0", "1", "2", "3", "4", "5"] as const;

function sanitizeCode(input: string, maxLength: number) {
	return input
		.toUpperCase()
		.replace(/O/g, "0")
		.replace(/[IL]/g, "1")
		.replace(/[^A-Z0-9]/g, "")
		.slice(0, maxLength);
}

function InputOTP({
	value,
	onChange,
	length = 6,
	id,
	autoFocus,
	disabled,
	className,
	"aria-invalid": ariaInvalid,
}: {
	value: string;
	onChange: (value: string) => void;
	length?: number;
	id?: string;
	autoFocus?: boolean;
	disabled?: boolean;
	className?: string;
	"aria-invalid"?: boolean;
}) {
	const inputsRef = useRef<(HTMLInputElement | null)[]>([]);

	function focusAt(index: number) {
		const input = inputsRef.current[index];
		input?.focus();
		input?.select();
	}

	function applyCode(next: string) {
		onChange(sanitizeCode(next, length));
	}

	function handlePaste(index: number, event: ClipboardEvent<HTMLInputElement>) {
		event.preventDefault();
		const pasted = sanitizeCode(
			`${value.slice(0, index)}${event.clipboardData.getData("text")}`,
			length,
		);
		applyCode(pasted);
		focusAt(Math.min(pasted.length, length - 1));
	}

	function handleChange(index: number, nextValue: string) {
		const sanitized = sanitizeCode(nextValue, length);

		if (sanitized.length > 1) {
			applyCode(sanitized);
			focusAt(Math.min(sanitized.length, length - 1));
			return;
		}

		const chars = value.split("");
		while (chars.length < length) chars.push("");

		if (sanitized) {
			chars[index] = sanitized;
			applyCode(chars.join("").trimEnd());
			if (index < length - 1) focusAt(index + 1);
			return;
		}

		chars[index] = "";
		applyCode(chars.join("").trimEnd());
	}

	function handleKeyDown(
		index: number,
		event: KeyboardEvent<HTMLInputElement>,
	) {
		if (event.key === "Backspace") {
			event.preventDefault();
			const chars = value.split("");
			while (chars.length < length) chars.push("");

			if (chars[index]) {
				chars[index] = "";
				applyCode(chars.join("").trimEnd());
				return;
			}

			if (index > 0) {
				chars[index - 1] = "";
				applyCode(chars.join("").trimEnd());
				focusAt(index - 1);
			}
			return;
		}

		if (event.key === "ArrowLeft" && index > 0) {
			event.preventDefault();
			focusAt(index - 1);
			return;
		}

		if (event.key === "ArrowRight" && index < length - 1) {
			event.preventDefault();
			focusAt(index + 1);
		}
	}

	return (
		<div
			className={cn("flex justify-center gap-2", className)}
			role="group"
			aria-label="One-time code"
		>
			{OTP_SLOT_KEYS.slice(0, length).map((slotKey, index) => (
				<input
					key={slotKey}
					ref={(element) => {
						inputsRef.current[index] = element;
					}}
					id={index === 0 ? id : undefined}
					type="text"
					inputMode="text"
					autoComplete={index === 0 ? "one-time-code" : "off"}
					maxLength={index === 0 ? length : 1}
					value={value[index] ?? ""}
					disabled={disabled}
					autoFocus={autoFocus && index === 0}
					aria-invalid={ariaInvalid}
					className={cn(
						"h-11 w-10 rounded-lg border border-input bg-transparent text-center font-mono text-lg uppercase transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40",
					)}
					onChange={(event) => handleChange(index, event.target.value)}
					onKeyDown={(event) => handleKeyDown(index, event)}
					onPaste={(event) => handlePaste(index, event)}
					onFocus={(event) => event.target.select()}
				/>
			))}
		</div>
	);
}

export { InputOTP };
