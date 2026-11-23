import { el, readOnly } from "../lib/dom.js";
import { post } from "../lib/api.js";
import { tomorrowNineISO } from "../lib/format.js";
import { toast, toastError } from "./toast.js";

// Snooze + Dismiss controls for a reminder. done() is called after a successful action.
export function reminderActions(reminder, done) {
  if (readOnly()) return el("span");
  const act = async (url, body, msg) => {
    try {
      await post(url, body);
      if (msg) toast(msg);
      done();
    } catch (e) { toastError(e); }
  };
  const snooze = (body, label) => act(`/api/reminders/${reminder.id}/snooze`, body, `Snoozed ${label}`);

  const minutes = el("input", { type: "number", min: "1", value: "30", "aria-label": "Custom snooze minutes", style: "width:5rem" });
  const menu = el("div", { class: "menu", hidden: true },
    el("button", { type: "button", class: "btn btn-sm", onclick: () => snooze({ minutes: 10 }, "10 minutes") }, "10 minutes"),
    el("button", { type: "button", class: "btn btn-sm", onclick: () => snooze({ minutes: 60 }, "1 hour") }, "1 hour"),
    el("button", { type: "button", class: "btn btn-sm", onclick: () => snooze({ until: tomorrowNineISO() }, "until tomorrow 9:00") }, "Tomorrow 9:00"),
    el("div", { class: "row" }, minutes,
      el("button", { type: "button", class: "btn btn-sm", onclick: () => {
        const n = parseInt(minutes.value, 10);
        if (n > 0) snooze({ minutes: n }, `${n} minutes`);
      } }, "Snooze min"))
  );
  const toggle = el("button", { type: "button", class: "btn btn-sm", "aria-expanded": "false",
    onclick: () => { menu.hidden = !menu.hidden; toggle.setAttribute("aria-expanded", String(!menu.hidden)); } }, "Snooze ▾");
  const dismiss = el("button", { type: "button", class: "btn btn-sm", onclick: () => act(`/api/reminders/${reminder.id}/dismiss`, {}) }, "Dismiss");
  return el("div", {}, el("div", { class: "toast-actions" }, toggle, dismiss), menu);
}
