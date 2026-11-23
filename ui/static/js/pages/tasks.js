import { $, el, clear, emptyState, params, dropParam, readOnly } from "../lib/dom.js";
import { get, qs } from "../lib/api.js";
import { todayStr, dueLabel } from "../lib/format.js";
import { quickAdd } from "../components/quick-add.js";
import { openTaskDialog } from "../components/task-dialog.js";
import { taskList, taskGroup } from "../components/task-list.js";
import { toastError } from "../components/toast.js";

const root = $("#tasks-root");
const view = root.dataset.view || "all";
const q = params();
const project = parseInt(q.get("project"), 10) || 0;
const tag = q.get("tag") || "";
const listEl = $("#task-list");

const EMPTY = {
  today: ["🎉", "Nothing due today. Enjoy the calm."],
  upcoming: ["🌤️", "Nothing coming up in the next 7 days."],
  all: ["✨", "No open tasks. Add one above."],
  completed: ["📭", "Nothing completed yet."],
};

function render(tasks) {
  clear(listEl);
  if (!tasks.length) return listEl.append(emptyState(...EMPTY[view]));
  const opts = { refresh: load };
  if (view === "today") {
    const today = todayStr();
    const overdue = tasks.filter((t) => t.due_date < today);
    const rest = tasks.filter((t) => t.due_date >= today);
    if (overdue.length) listEl.append(taskGroup("Overdue", overdue, opts, "overdue"));
    if (rest.length) listEl.append(taskGroup("Today", rest, opts));
  } else if (view === "upcoming") {
    const byDate = new Map();
    for (const t of tasks) byDate.set(t.due_date, [...(byDate.get(t.due_date) || []), t]);
    for (const [d, items] of byDate) listEl.append(taskGroup(dueLabel(d), items, opts));
  } else {
    listEl.append(taskList(tasks, { ...opts, reorder: view === "all" && !project && !tag }));
  }
}

async function load() {
  try {
    render((await get("/api/tasks" + qs({ view, project, tag }))) || []);
  } catch (e) { toastError(e); }
}

async function filterBar() {
  const bar = $("#filter-bar");
  const chips = [];
  if (project) {
    let name = `#${project}`;
    try { name = (await get(`/api/projects/${project}`)).name; } catch { /* keep id */ }
    chips.push(el("a", { class: "chip active", href: location.pathname + (tag ? `?tag=${encodeURIComponent(tag)}` : "") }, `Project: ${name} ×`));
  }
  if (tag) chips.push(el("a", { class: "chip active", href: location.pathname + (project ? `?project=${project}` : "") }, `#${tag} ×`));
  if (chips.length) bar.append(el("div", { class: "chip-row" }, ...chips));
}

async function openFromQuery() {
  const id = parseInt(q.get("open"), 10);
  if (!id) return;
  dropParam("open");
  try { openTaskDialog({ task: await get(`/api/tasks/${id}`), onSaved: load }); } catch (e) { toastError(e); }
}

if (view !== "completed" && !readOnly()) {
  const qa = quickAdd({ projectId: project, onAdd: load });
  $("#quick-add-slot").append(qa);
  if (q.get("new")) { dropParam("new"); $("#quick-add-input").focus(); }
}
if (!readOnly() && view !== "completed") {
  $("#page-actions").append(el("button", { class: "btn", type: "button", onclick: () => openTaskDialog({ projectId: project, onSaved: load }) }, "New task…"));
}
filterBar();
load().then(openFromQuery);
