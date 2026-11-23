import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get, post, del } from "../lib/api.js";
import { fmtDateTime, relTime } from "../lib/format.js";
import { openReminderDialog } from "../components/reminder-dialog.js";
import { reminderActions } from "../components/snooze-menu.js";
import { toast, toastError } from "../components/toast.js";

const list = $("#reminder-list");

function item(r) {
  const when = r.snoozed_until || r.remind_at;
  const due = new Date(when) <= new Date();
  const links = [];
  if (r.task_id) links.push(el("a", { href: `/tasks?open=${r.task_id}` }, "task"));
  if (r.note_id) links.push(el("a", { href: `/notes?open=${r.note_id}` }, "note"));
  return el("div", { class: `reminder-item${due ? " due" : ""}` },
    el("div", { class: "grow" },
      el("div", {}, r.message),
      el("div", { class: "hint" }, `${fmtDateTime(when)} · ${relTime(when)}${r.snoozed_until ? " · snoozed" : ""}`, links.length ? " · " : "", ...links)),
    reminderActions(r, load),
    readOnly() ? null : el("button", { class: "icon-btn", type: "button", title: "Delete", "aria-label": "Delete reminder", onclick: async () => {
      try { await del(`/api/reminders/${r.id}`); load(); } catch (e) { toastError(e); }
    } }, "🗑"));
}

async function load() {
  let rs;
  try { rs = (await get("/api/reminders")) || []; } catch (e) { return toastError(e); }
  clear(list);
  if (!rs.length) return list.append(emptyState("🔔", "No active reminders."));
  list.append(...rs.map(item));
}

const actions = $("#page-actions");
if ("Notification" in window && Notification.permission === "default") {
  actions.append(el("button", { class: "btn", type: "button", onclick: async (e) => {
    const res = await Notification.requestPermission();
    toast(res === "granted" ? "Browser notifications enabled" : "Notifications not enabled");
    if (res !== "default") e.target.remove();
  } }, "Enable browser notifications"));
}
if (!readOnly()) actions.append(el("button", { class: "btn-primary", type: "button", onclick: () => openReminderDialog({ onSaved: load }) }, "New reminder"));
load();
