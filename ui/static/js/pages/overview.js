import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { greeting, dueLabel } from "../lib/format.js";
import { quickAdd } from "../components/quick-add.js";
import { taskGroup } from "../components/task-list.js";
import { reminderActions } from "../components/snooze-menu.js";
import { toastError } from "../components/toast.js";

const root = $("#overview");
const name = root.dataset.name;
const body = $("#overview-body");

$("#greeting").textContent = name ? `${greeting()}, ${name.split(" ")[0]}` : greeting();
$("#overview-date").textContent = dueLabel(document.body.dataset.today) + " · " + new Date().toLocaleDateString(undefined, { weekday: "long", day: "numeric", month: "long" });

if (!readOnly()) $("#quick-add-slot").append(quickAdd({ onAdd: load }));

function reminderSection(reminders) {
  const list = el("div", { class: "stack", style: "gap:.5rem" }, ...reminders.map((r) =>
    el("div", { class: "reminder-item due" }, el("div", { class: "grow" }, "⏰ " + r.message), reminderActions(r, load))));
  return el("section", { class: "section" }, el("h2", { class: "group-title" }, "Reminders ", el("span", { class: "count" }, String(reminders.length))), list);
}

async function load() {
  let data;
  try { data = await get("/api/overview"); } catch (e) { return toastError(e); }
  clear(body);
  const opts = { refresh: load };
  const overdue = data.overdue || [], today = data.due_today || [], high = data.high_priority || [], rem = data.reminders || [];
  if (rem.length) body.append(reminderSection(rem));
  if (overdue.length) body.append(taskGroup("Overdue", overdue, opts, "overdue"));
  if (today.length) body.append(taskGroup("Due today", today, opts));
  if (high.length) body.append(taskGroup("High priority", high, opts));
  if (!rem.length && !overdue.length && !today.length && !high.length) {
    body.append(emptyState("🌿", "You're all caught up. Add a task to get started."));
  }
}

load();
