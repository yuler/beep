import type { Beep } from "@/lib/api/beeps";
import type { NotificationChannel } from "@/lib/notification-channels";
import { I18nError } from "@/lib/notification-channels";
import * as m from "@/locale/paraglide/messages";

const PUSH_ERROR_MESSAGES: Record<string, () => string> = {
	"push.unsupported": () => m.push_unsupported(),
	"push.network_unreachable": () => m.push_network_unreachable({ host: "" }),
	"push.device_not_subscribed": () => m.push_device_not_subscribed(),
	"push.permission_denied": () => m.push_permission_denied(),
	"push.permission_not_granted": () => m.push_permission_not_granted(),
	"push.subscription_missing_keys": () => m.push_subscription_missing_keys(),
	"push.subscribe_timeout": () => m.push_subscribe_timeout(),
};

const BEEP_STATUS_MESSAGES: Record<Beep["status"], () => string> = {
	active: () => m.status_beep_active(),
	firing: () => m.status_beep_firing(),
	paused: () => m.status_beep_paused(),
	completed: () => m.status_beep_completed(),
	cancelled: () => m.status_beep_cancelled(),
};

const RUN_STATUS_MESSAGES: Record<string, () => string> = {
	pending: () => m.status_run_pending(),
	running: () => m.status_run_running(),
	succeeded: () => m.status_run_succeeded(),
	failed: () => m.status_run_failed(),
	skipped: () => m.status_run_skipped(),
	expired: () => m.status_run_expired(),
};

const HEALTH_STATUS_MESSAGES: Record<string, () => string> = {
	ok: () => m.status_health_ok(),
	alerting: () => m.status_health_alerting(),
	broken: () => m.status_health_broken(),
};

const JOB_STATUS_MESSAGES: Record<string, () => string> = {
	failed: () => m.status_job_failed(),
	finished: () => m.status_job_finished(),
	pending: () => m.status_job_pending(),
};

const BROWSER_MESSAGES: Record<string, () => string> = {
	"Google Chrome": () => m.push_browser_chrome(),
	"Microsoft Edge": () => m.push_browser_edge(),
	Firefox: () => m.push_browser_firefox(),
	Opera: () => m.push_browser_opera(),
	Safari: () => m.push_browser_safari(),
	"this browser": () => m.push_browser_this(),
};

const OS_MESSAGES: Record<string, () => string> = {
	iOS: () => m.push_os_ios(),
	Mac: () => m.push_os_macos(),
	Windows: () => m.push_os_windows(),
	Linux: () => m.push_os_linux(),
	Android: () => m.push_os_android(),
};

export function translateError(err: unknown): string {
	if (err instanceof I18nError) {
		return err.message;
	}
	if (err instanceof Error && err.message in PUSH_ERROR_MESSAGES) {
		return PUSH_ERROR_MESSAGES[err.message]();
	}
	if (err instanceof Error && err.message) {
		return err.message;
	}
	return m.errors_something_went_wrong();
}

export function beepStatusLabel(status: Beep["status"]): string {
	return BEEP_STATUS_MESSAGES[status]();
}

export function beepRunStatusLabel(status: string): string {
	return RUN_STATUS_MESSAGES[status]?.() ?? status;
}

export function healthStatusLabel(status: string): string {
	return HEALTH_STATUS_MESSAGES[status]?.() ?? status;
}

export function jobStatusLabel(status: string): string {
	return JOB_STATUS_MESSAGES[status]?.() ?? status;
}

export function channelLabel(channel: NotificationChannel): string {
	return channel === "email"
		? m.push_channel_email()
		: m.push_channel_web_push();
}

export function browserLabel(browserName: string): string {
	return BROWSER_MESSAGES[browserName]?.() ?? browserName;
}

export function osLabel(os: string): string {
	return OS_MESSAGES[os]?.() ?? os;
}

export function pushPlatformLabel(
	platform: "ios" | "macos" | "windows" | "linux" | "other",
): string {
	switch (platform) {
		case "ios":
			return m.push_help_os_ios();
		case "macos":
			return m.push_help_os_macos();
		case "windows":
			return m.push_help_os_windows();
		case "linux":
			return m.push_help_os_linux();
		default:
			return m.push_help_os_other();
	}
}
