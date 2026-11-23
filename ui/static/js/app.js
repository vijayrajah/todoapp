import { $, readOnly } from "./lib/dom.js";
import { initials } from "./lib/format.js";
import { startReminderPoller } from "./components/reminder-poller.js";

const root = document.documentElement;
const body = document.body;

if (readOnly()) body.classList.add("read-only");

// Theme: explicit choice persists; otherwise follows the OS.
$("#theme-toggle")?.addEventListener("click", () => {
  const dark = root.dataset.theme ? root.dataset.theme === "dark" : matchMedia("(prefers-color-scheme: dark)").matches;
  const next = dark ? "light" : "dark";
  root.dataset.theme = next;
  try { localStorage.setItem("theme", next); } catch { /* storage unavailable */ }
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
