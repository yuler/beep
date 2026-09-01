import { m } from "@/locale/paraglide/messages";

export function LogoMark({ className }: { className?: string }) {
	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			viewBox="-2.0 -2.0 104.0 96.834"
			role="img"
			aria-label={m.term_beep()}
			className={className}
		>
			<path
				fill="currentColor"
				d="M4.886 0.000 L13.681 0.000 L37.622 26.384 L37.622 85.342 A6.840 6.840 0 0 1 30.782 92.182 L6.840 92.182 A6.840 6.840 0 0 1 0.000 85.342 L0.000 4.886 A4.886 4.886 0 0 1 4.886 0.000 Z"
			/>
			<path
				fill="currentColor"
				d="M53.933 37.930 L87.459 0.000 L94.951 0.000 A4.886 4.886 0 0 1 99.837 4.886 L99.837 87.785 A4.886 4.886 0 0 1 94.951 92.671 L91.531 92.671 L53.740 44.314 Q51.183 41.042 53.933 37.930 Z"
			/>
		</svg>
	);
}
