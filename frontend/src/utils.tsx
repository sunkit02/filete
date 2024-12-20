import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import type { ColorTheme } from "@/types";

export function cn(...inputs: ClassValue[]): string {
	return twMerge(clsx(inputs));
}

export function getPreferedTheme(): ColorTheme {
	const systemPreference = window.matchMedia("(prefers-color-scheme: dark)")
		.matches
		? "dark"
		: "light";

	let userPreference = localStorage.getItem("theme");
	if (userPreference !== "light" && userPreference !== "dark") {
		userPreference = systemPreference;
	}
	return userPreference as ColorTheme;
}

const rootNode = document.querySelector("html") as HTMLElement;

export function setPreferedTheme(theme: ColorTheme) {
	localStorage.setItem("theme", theme);

	if (theme === "dark") {
		rootNode.classList.add("dark");
	} else {
		rootNode.classList.remove("dark");
	}
}

export function clearPreferedTheme() {
	localStorage.removeItem("theme");
	rootNode.classList.add(getPreferedTheme());
}
