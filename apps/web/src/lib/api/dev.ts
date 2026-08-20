import {
	clearDevLetters as serverClearDevLetters,
	deleteDevLetter as serverDeleteDevLetter,
	fetchDevLetters as serverFetchDevLetters,
} from "@/server/dev";

export type { DevLetter, DevLettersResponse } from "@/server/dev";

export function fetchDevLetters() {
	return serverFetchDevLetters({});
}

export function deleteDevLetter(id: string) {
	return serverDeleteDevLetter({ data: { id } });
}

export function clearDevLetters() {
	return serverClearDevLetters({});
}
