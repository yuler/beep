/**
 * Canonical API error surfaced to the UI. Server functions (`src/server/*`)
 * throw this after a non-OK Rails response, and it crosses the wire to the
 * browser so forms and auth guards can react to status codes.
 */
export class ApiError extends Error {
	status: number;
	code?: string;

	constructor(status: number, message: string, code?: string) {
		super(message);
		this.name = "ApiError";
		this.status = status;
		this.code = code;
	}
}
