import { paraglideMiddleware } from "@/locale/paraglide/server";

export function localeMiddleware(
	request: Request,
	resolve: () => Response | Promise<Response>,
) {
	return paraglideMiddleware(request, () => resolve());
}
