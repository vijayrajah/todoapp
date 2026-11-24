import { $, readOnly } from "./lib/dom.js";
import { initials } from "./lib/format.js";
import { startReminderPoller } from "./components/reminder-poller.js";
import { showDigestIfDue } from "./components/digest.js";
import { get, put } from "./lib/api.js";
import { toast } from "./components/toast.js";

const root = document.documentElement;
const body = document.body;

if (readOnly()) body.classList.add("read-only");

// Theme: stored on the server; the button cycles through the available themes.
const THEMES = ["auto", "light", "dark", "midnight", "nord", "solarized", "sepia"];
$("#theme-toggle")?.addEventListener("click", async () => {
  const cur = body.dataset.themeName || "auto";
  const next = THEMES[(THEMES.indexOf(cur) + 1) % THEMES.length];
  body.dataset.themeName = next;
  if (next === "auto") root.removeAttribute("data-theme"); else root.dataset.theme = next;
  toast(`Theme: ${next}`);
  if (readOnly()) return;
  try {
    const p = await get("/api/preferences");
    await put("/api/preferences/appearance", { theme: next, accent: p.accent });
  } catch (e) { toast(e.message, "error"); }
});

// Mobile drawer
const menuBtn = $("#menu-btn");
const setMenu = (open) => {
  body.classList.toggle("menu-open", open);
  menuBtn?.setAttribute("aria-expanded", String(open));
};
menuBtn?.addEventListener("click", () => setMenu(!body.classList.contains("menu-open")));
$("#scrim")?.addEventListener("click", () => setMenu(false));
$("#sidebar")?.addEventListener("click", (e) => { if (e.target.closest("a")) setMenu(false); });

// Avatar initials (name is set via textContent)
const avatar = $("#avatar");
if (avatar) avatar.textContent = initials(avatar.dataset.name);

// Shortcuts: "/" search, "n" new task
document.addEventListener("keydown", (e) => {
  if (e.metaKey || e.ctrlKey || e.altKey) return;
  const t = e.target;
  if (t.closest && t.closest("input, textarea, select, dialog, [contenteditable]")) return;
  if (document.querySelector("dialog[open]")) return;
  if (e.key === "/") {
    e.preventDefault();
    $("#search-input")?.focus();
  } else if (e.key === "n") {
    e.preventDefault();
    const qa = $("#quick-add-input");
    if (qa) qa.focus();
    else location.href = "/tasks?new=1";
  }
});

startReminderPoller();
if (body.dataset.digest === "true" && !readOnly()) showDigestIfDue();
