import { el, readOnly } from "../lib/dom.js";
import { post, put, del } from "../lib/api.js";
import { taskBody } from "../lib/tasks.js";
import { dueLabel, isOverdue, recurrenceLabel, todayStr } from "../lib/format.js";
import { makeSortable } from "../lib/sortable.js";
import { toast, toastError } from "./toast.js";
import { openTaskDialog } from "./task-dialog.js";
import { openReminderDialog } from "./reminder-dialog.js";

const reorderTasks = (ids) => post("/api/tasks/reorder", { ids }).catch(toastError);

// Make direct `.task` children of container drag-reorderable.
export function enableTaskReorder(container) {
  makeSortable(container, ".task", reorderTasks);
}

// Renders one task (and its subtasks). refresh() reloads the surrounding list.
export function taskRow(task, { refresh, sub = false, showProject = true } = {}) {
  const ro = readOnly();
  const done = task.status === "done";

  const check = el("input", {
    type: "checkbox", class: `check p${task.priority || 0}`, checked: done, disabled: ro,
    "aria-label": done ? `Reopen ${task.title}` : `Complete ${task.title}`,
    onchange: async () => {
      try {
        if (check.checked) {
          const res = await post(`/api/tasks/${task.id}/complete`);
          if (res && res.next) toast(`Next occurrence: ${dueLabel(res.next.due_date) || "scheduled"}`);
        } else {
          await post(`/api/tasks/${task.id}/reopen`);
        }
        refresh();
      } catch (e) { check.checked = !check.checked; toastError(e); }
    },
  });

  const titleBtn = el("button", { type: "button", class: "task-title", onclick: () => openTaskDialog({ task, onSaved: refresh }) }, task.title);

  const meta = el("div", { class: "task-meta" });
  if (task.priority > 0) meta.append(el("span", { class: `dot p${task.priority}`, title: `Priority ${task.priority}` }));
  if (task.due_date) {
    const cls = !done && isOverdue(task.due_date) ? "overdue" : task.due_date === todayStr() ? "today" : "";
    meta.append(el("span", { class: `due ${cls}` }, dueLabel(task.due_date)));
  }
  if (task.recurrence) meta.append(el("span", { title: recurrenceLabel(task.recurrence) }, "↻ ", recurrenceLabel(task.recurrence)));
  if (showProject && task.project_id) {
    meta.append(el("a", { class: "chip project-chip", href: `/projects/${task.project_id}` },
      el("i", { style: `background:${CSS.supports("color", task.project_color || "") ? task.project_color : "var(--accent)"}` }), task.project_name));
  }
  for (const t of task.tags || []) meta.append(el("a", { class: "chip", href: `/tags?name=${encodeURIComponent(t)}` }, "#" + t));
  const checks = task.checklist || [];
  let checkBox = null;
  if (checks.length) {
    const n = checks.filter((c) => c.done).length;
    checkBox = el("ul", { class: "task-checks", hidden: true });
    for (const [i, c] of checks.entries()) {
      checkBox.append(el("li", {}, el("label", {},
        el("input", { type: "checkbox", checked: c.done, disabled: ro, onchange: async (e) => {
          const items = checks.map((x, j) => ({ text: x.text, done: j === i ? e.target.checked : x.done }));
          try { await put(`/api/tasks/${task.id}`, taskBody(task, { checklist: items })); refresh(); } catch (x) { e.target.checked = !e.target.checked; toastError(x); }
        } }),
        el("span", { class: c.done ? "muted" : "" }, c.text))));
    }
    meta.append(el("button", { type: "button", class: "chip-progress", title: "Show checklist",
      onclick: () => { checkBox.hidden = !checkBox.hidden; } }, `☑ ${n}/${checks.length}`));
  }
  const subs = task.subtasks || [];
  if (subs.length) {
    const n = subs.filter((s) => s.status === "done").length;
    meta.append(el("span", { class: "sub-progress", title: `${n} of ${subs.length} subtasks done` },
      `${n}/${subs.length}`, el("span", { class: "progress" }, el("span", { style: `width:${(n / subs.length) * 100}%` }))));
  }

  const row = el("div", { class: `task${done ? " done" : ""}`, dataset: { id: task.id } });
  const subBox = el("div", { class: "subtasks", hidden: !subs.length });
  const addBox = el("div", { class: "add-sub", hidden: true });

  const actions = el("div", { class: "task-actions" });
  const btn = (label, text, fn) => el("button", { type: "button", class: "icon-btn", title: label, "aria-label": label, onclick: fn }, text);
  actions.append(btn("Edit", "✎", () => openTaskDialog({ task, onSaved: refresh })));
  if (!ro) {
    if (!sub && !done) actions.append(btn("Add subtask", "＋", () => { addBox.hidden = !addBox.hidden; if (!addBox.hidden) addInput.focus(); }));
    if (!done) actions.append(btn("Remind me", "⏰", () => openReminderDialog({ taskId: task.id, message: task.title })));
    actions.append(btn("Delete", "🗑", async () => {
      try { await del(`/api/tasks/${task.id}`); toast("Moved to trash"); refresh(); } catch (e) { toastError(e); }
    }));
  }

  const addInput = el("input", { type: "text", placeholder: "Add subtask…", "aria-label": "Subtask title", class: "sub-input",
    onkeydown: async (e) => {
      if (e.key === "Escape") { addBox.hidden = true; return; }
      if (e.key !== "Enter" || !addInput.value.trim()) return;
      e.preventDefault();
      try { await post("/api/tasks", { title: addInput.value.trim(), parent_id: task.id }); refresh(); } catch (x) { toastError(x); }
    } });
  addBox.append(addInput);

  row.append(
    el("div", { class: "task-main" },
      ro ? null : el("span", { class: "drag-handle", title: "Drag to reorder", "aria-hidden": "true" }, "⋮⋮"),
      check, el("div", { class: "task-body" }, titleBtn, meta), actions),
    subBox, addBox
  );
  if (checkBox) row.querySelector(".task-main").after(checkBox);
  for (const s of subs) subBox.append(taskRow(s, { refresh, sub: true, showProject: false }));
  if (subs.length) makeSortable(subBox, ".task", reorderTasks);
  return row;
}
