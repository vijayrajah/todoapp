import { el } from "../lib/dom.js";
import { taskRow, enableTaskReorder } from "./task-row.js";

// A list of task rows; pass reorder:true to enable drag ordering.
export function taskList(tasks, { refresh, reorder = false, showProject = true } = {}) {
  const list = el("div", { class: "task-list" }, ...tasks.map((t) => taskRow(t, { refresh, showProject })));
  if (reorder) enableTaskReorder(list);
  return list;
}

// A titled group of tasks.
export function taskGroup(title, tasks, opts = {}, cls = "") {
  return el("section", { class: "task-group" },
    el("h2", { class: `group-title ${cls}` }, title, " ", el("span", { class: "count" }, String(tasks.length))),
    taskList(tasks, opts));
}
