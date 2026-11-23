import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get, post, del } from "../lib/api.js";
import { quickAdd } from "../components/quick-add.js";
import { openProjectDialog } from "../components/project-dialog.js";
import { openTaskDialog } from "../components/task-dialog.js";
import { taskList } from "../components/task-list.js";
import { confirmDialog } from "../lib/modal.js";
import { toast, toastError } from "../components/toast.js";

const id = parseInt($("#project-root").dataset.projectId, 10);
const header = $("#project-header");
const listEl = $("#task-list");

function renderHeader(p) {
  document.title = `${p.name} · todoapp`;
  const pct = p.total_tasks ? Math.round((p.done_tasks / p.total_tasks) * 100) : 0;
  const actions = el("div", { class: "page-actions" });
  if (!readOnly()) {
    actions.append(
      el("button", { class: "btn", type: "button", onclick: () => openProjectDialog({ project: p, onSaved: load }) }, "Edit"),
      el("button", { class: "btn", type: "button", onclick: async () => {
        try { await post(`/api/projects/${id}/${p.archived ? "unarchive" : "archive"}`); load(); } catch (e) { toastError(e); }
      } }, p.archived ? "Unarchive" : "Archive"),
      el("button", { class: "btn btn-danger", type: "button", onclick: async () => {
        if (!(await confirmDialog("Move this project and its tasks to trash?", "Delete"))) return;
        try { await del(`/api/projects/${id}`); toast("Moved to trash"); location.href = "/projects"; } catch (e) { toastError(e); }
      } }, "Delete"));
  }
  clear(header).append(
    el("div", { style: "flex:1;min-width:240px" },
      el("h1", {}, p.name, p.archived ? el("span", { class: "chip", style: "margin-left:.5rem" }, "archived") : null),
      p.description ? el("p", { class: "muted" }, p.description) : null,
      el("div", { class: "progress", style: "margin-top:.6rem", role: "progressbar", "aria-valuenow": pct, "aria-valuemin": "0", "aria-valuemax": "100" }, el("span", { style: `width:${pct}%` })),
      el("p", { class: "hint" }, `${p.done_tasks}/${p.total_tasks} done`)),
    actions);
}

async function load() {
  try {
    const [p, tasks] = await Promise.all([get(`/api/projects/${id}`), get(`/api/tasks?view=all&project=${id}`)]);
    renderHeader(p);
    clear(listEl);
    if (!(tasks || []).length) listEl.append(emptyState("🗂️", "No open tasks in this project."));
    else listEl.append(taskList(tasks, { refresh: load, reorder: true, showProject: false }));
  } catch (e) { toastError(e); }
}

if (!readOnly()) $("#quick-add-slot").append(quickAdd({ projectId: id, onAdd: load }));
load();
