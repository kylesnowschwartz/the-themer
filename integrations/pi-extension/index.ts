/**
 * pi extension: theme switching for pi, driven by the-themer.
 *
 * Provides:
 *   /light   — switch to built-in light theme
 *   /dark    — switch to built-in dark theme
 *   /theme   — toggle (uses last value set by this extension)
 *
 * Plus auto-sync: watches ~/.config/the-themer/pi-variant and calls
 * setTheme(contents) when the file changes. The-themer's `pi` switch
 * handler atomically writes "light" or "dark" to that file, so running
 * `the-themer switch <name>` from your shell propagates light/dark into
 * a running pi session without any direct shell→pi channel.
 *
 * Install (symlink, source-of-truth stays in this repo):
 *   ln -s "$(pwd)/integrations/pi-extension" ~/.pi/agent/extensions/the-themer
 *   # then /reload in pi (or restart the session)
 */

import { readFile } from "node:fs/promises";
import { existsSync, watch } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

const VARIANT_FILE = join(homedir(), ".config", "the-themer", "pi-variant");

async function readVariant(): Promise<"light" | "dark" | null> {
	try {
		const raw = (await readFile(VARIANT_FILE, "utf8")).trim();
		return raw === "light" || raw === "dark" ? raw : null;
	} catch {
		return null;
	}
}

export default function (pi: ExtensionAPI) {
	let lastSet: "light" | "dark" | null = null;
	let watcher: ReturnType<typeof watch> | null = null;

	type CmdCtx = Parameters<Parameters<ExtensionAPI["registerCommand"]>[1]["handler"]>[1];

	const apply = (name: "light" | "dark", ctx: CmdCtx) => {
		const r = ctx.ui.setTheme(name);
		if (r.success) {
			lastSet = name;
			ctx.ui.notify(`theme: ${name}`, "info");
		} else {
			ctx.ui.notify(r.error ?? "setTheme failed", "error");
		}
	};

	pi.registerCommand("light", {
		description: "Switch to light theme",
		handler: async (_args, ctx) => apply("light", ctx),
	});

	pi.registerCommand("dark", {
		description: "Switch to dark theme",
		handler: async (_args, ctx) => apply("dark", ctx),
	});

	pi.registerCommand("theme", {
		description: "Toggle between light and dark",
		handler: async (_args, ctx) => apply(lastSet === "light" ? "dark" : "light", ctx),
	});

	pi.on("session_start", async (_event, ctx) => {
		// Initial sync from whatever the-themer last wrote.
		const initial = await readVariant();
		if (initial && initial !== lastSet) {
			const r = ctx.ui.setTheme(initial);
			if (r.success) lastSet = initial;
		}

		// Watch the directory (not the file) so atomic-rename writes from the-themer
		// trigger reliably on macOS — same reason the tcm/ghostty paths use rename.
		const dir = join(homedir(), ".config", "the-themer");
		if (!existsSync(dir)) return;

		try {
			watcher = watch(dir, async (_event, filename) => {
				if (filename !== "pi-variant") return;
				const v = await readVariant();
				if (!v || v === lastSet) return;
				const r = ctx.ui.setTheme(v);
				if (r.success) lastSet = v;
			});
		} catch {
			// fs.watch can fail on some filesystems; degrade silently.
		}
	});

	pi.on("session_shutdown", () => {
		if (watcher) {
			watcher.close();
			watcher = null;
		}
	});
}
