import { Button as ButtonPrimitive } from "@base-ui/react/button";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
	"group/button inline-flex shrink-0 items-center justify-center rounded-none border text-sm font-normal whitespace-nowrap transition-colors outline-none select-none disabled:pointer-events-none disabled:opacity-40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 min-h-[44px] sm:min-h-[40px] cursor-pointer",
	{
		variants: {
			variant: {
				default:
					"bg-solid text-solid-ink border-solid hover:opacity-90 active:bg-wash active:text-ink active:border-line",
				solid:
					"bg-solid text-solid-ink border-solid hover:opacity-90 active:bg-wash active:text-ink active:border-line",
				outline:
					"bg-transparent text-ink border-line hover:border-ink active:bg-wash active:border-ink aria-expanded:bg-solid aria-expanded:text-solid-ink",
				secondary:
					"bg-wash text-ink border-line hover:border-ink active:bg-solid active:text-solid-ink",
				ghost:
					"bg-transparent text-ink border-line hover:border-ink active:bg-wash active:border-ink aria-expanded:bg-solid aria-expanded:text-solid-ink aria-selected:bg-solid aria-selected:text-solid-ink data-[state=active]:bg-solid data-[state=active]:text-solid-ink",
				destructive:
					"bg-transparent text-ink border-ink hover:bg-wash active:bg-solid active:text-solid-ink",
				link: "bg-transparent text-ink border-transparent underline-offset-4 hover:underline min-h-0 sm:min-h-0 p-0",
			},
			size: {
				default: "px-3.5 py-2.5 text-sm gap-2",
				xs: "min-h-[32px] sm:min-h-[28px] px-2 py-1 text-xs gap-1",
				sm: "min-h-[36px] sm:min-h-[32px] px-2.5 py-1.5 text-xs gap-1.5",
				lg: "min-h-[48px] sm:min-h-[44px] px-4 py-3 text-base gap-2.5",
				icon: "size-11 sm:size-10 p-0",
				"icon-xs": "size-8 sm:size-7 p-0",
				"icon-sm": "size-9 sm:size-8 p-0",
				"icon-lg": "size-12 sm:size-11 p-0",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	},
);

export interface ButtonProps
	extends ButtonPrimitive.Props,
		VariantProps<typeof buttonVariants> {
	loading?: boolean;
	loadingText?: string;
}

function Button({
	className,
	variant = "default",
	size = "default",
	loading,
	loadingText,
	children,
	disabled,
	...props
}: ButtonProps) {
	return (
		<ButtonPrimitive
			data-slot="button"
			disabled={disabled || loading}
			className={cn(buttonVariants({ variant, size, className }))}
			{...props}
		>
			{loading ? (
				<>
					<span className="is-loading-spin mr-1.5" aria-hidden="true" />
					<span>{loadingText ?? children}</span>
				</>
			) : (
				children
			)}
		</ButtonPrimitive>
	);
}

function ButtonSolid(props: ButtonProps) {
	return <Button variant="solid" {...props} />;
}

function ButtonGhost(props: ButtonProps) {
	return <Button variant="ghost" {...props} />;
}

export { Button, ButtonSolid, ButtonGhost, buttonVariants };
