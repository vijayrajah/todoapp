import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get, post } from "../lib/api.js";
import { makeSortable } from "../lib/sortable.js";
import { openProjectDialog } from "../components/project-dialog.js";
import { toastError } from "../components/toast.js";

const grid = $("#project-grid");
let showArchived = false;

const safeColor = (c) => (CSS.supports("color", c || "") ? c : "var(--accent)");

function card(p) {
  const pct = p.total_tasks ? Math.round((p.done_tasks / p.total_tasks) * 100) : 0;
  return el("article", { class: `card project-card${p.archived ? " archived" : ""}`, dataset: { id: p.id }, style: `border-top-color:${safeColor(p.color)}` },
    el("div", { class: "task-main" },
      readOnly() ? null : el("span", { class: "drag-handle", title: "Drag to reorder", "aria-hidden": "true" }, "⋮⋮"),
      el("a", { class: "title", href: `/projects/${p.id}` }, p.name)),
    p.description ? el("p", { class: "muted" }, p.description) : null,
    el("div", { class: "progress", role: "progressbar", "aria-valuenow": pct, "aria-valuemin": "0", "aria-valuemax": "100" }, el("span", { style: `width:${pct}%` })),
    el("p", { class: "muted" }, `${p.done_tasks}/${p.total_tasks} done${p.archived ? " · archived" : ""}`));
}

async function load() {
  let projects;
  try { projects = (await get("/api/projects" + (showArchived ? "?archived=1" : ""))) || []; } catch (e) { return toastError(e); }
  clear(grid);
  if (!projects.length) return grid.append(emptyState("📁", "No projects yet. Create one to group your tasks."));
  grid.append(...projects.map(card));
}

makeSortable(grid, ".project-card", (ids) => post("/api/projects/reorder", { ids }).catch(toastError));

const actions = $("#page-actions");
const toggle = el("label", { class: "inline" }, el("input", { type: "checkbox", onchange: (e) => { showArchived = e.target.checked; load(); } }), "Show archived");
actions.append(toggle);
if (!readOnly()) actions.append(el("button", { class: "btn-primary", type: "button", onclick: () => openProjectDialog({ onSaved: load }) }, "New project"));
load();
