import handler from "@tanstack/react-start/server-entry";
import { localeMiddleware } from "@/locale/middleware";

export default {
	fetch(request: Request) {
		return localeMiddleware(request, () => handler.fetch(request));
	},
};
