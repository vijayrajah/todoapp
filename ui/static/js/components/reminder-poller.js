import { get } from "../lib/api.js";
import { el } from "../lib/dom.js";
import { show } from "./toast.js";
import { reminderActions } from "./snooze-menu.js";

const POLL_MS = 30000;
const shown = new Map(); // reminder id -> key of the occurrence already shown

const keyOf = (r) => `${r.id}|${r.snoozed_until || r.remind_at}`;

function notify(r) {
  if (!("Notification" in window) || Notification.permission !== "granted") return;
  const seen = JSON.parse(sessionStorage.getItem("notified") || "[]");
  const k = keyOf(r);
  if (seen.includes(k)) return;
  sessionStorage.setItem("notified", JSON.stringify([...seen.slice(-50), k]));
  try { new Notification("Reminder", { body: r.message, tag: k }); } catch { /* ignore */ }
}

function showReminder(r) {
  let handle;
  const body = el("div", {},
    el("div", { class: "toast-head" },
      el("span", { class: "toast-msg" }, "⏰ " + r.message),
      el("button", { type: "button", class: "icon-btn", "aria-label": "Close", onclick: () => handle.close() }, "×")),
    reminderActions(r, () => handle.close()));
  handle = show({ node: body, kind: "reminder", persistent: true });
}

async function poll() {
  try {
    const due = (await get("/api/reminders/due")) || [];
    for (const r of due) {
      if (shown.get(r.id) === keyOf(r)) continue;
      shown.set(r.id, keyOf(r));
      showReminder(r);
      notify(r);
    }
  } catch { /* network hiccup; try again next tick */ }
}

export function startReminderPoller() {
  poll();
  setInterval(poll, POLL_MS);
}
