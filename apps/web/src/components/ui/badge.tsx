import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
	"group/badge inline-flex h-5 w-fit shrink-0 items-center justify-center gap-1 rounded-none border px-1.5 py-0.5 text-xs font-normal whitespace-nowrap transition-colors [&>svg]:pointer-events-none [&>svg]:size-3!",
	{
		variants: {
			variant: {
				default: "bg-solid text-solid-ink border-solid",
				secondary: "bg-wash text-ink border-line",
				destructive: "bg-transparent text-ink border-ink",
				outline: "border-line bg-transparent text-ink",
				ghost: "bg-transparent text-gray border-transparent",
				link: "text-ink underline-offset-4 hover:underline border-transparent",
			},
		},
		defaultVariants: {
			variant: "outline",
		},
	},
);

function Badge({
	className,
	variant = "outline",
	render,
	...props
}: useRender.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
	return useRender({
		defaultTagName: "span",
		props: mergeProps<"span">(
			{
				className: cn(badgeVariants({ variant }), className),
			},
			props,
		),
		render,
		state: {
			slot: "badge",
			variant,
		},
	});
}

export { Badge, badgeVariants };
