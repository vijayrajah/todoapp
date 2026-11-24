import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { openFilterDialog, DUE_LABELS } from "../components/filter-dialog.js";
import { toastError } from "../components/toast.js";

const listEl = $("#view-list");
const reload = () => location.reload(); // the sidebar also lists views

async function load() {
  try {
    const [items, projects] = await Promise.all([get("/api/filters"), get("/api/projects")]);
    const names = new Map((projects || []).map((p) => [p.id, p.name]));
    clear(listEl);
    if (!(items || []).length) return listEl.append(emptyState("🔎", "No saved views yet."));
    for (const f of items) {
      const parts = [];
      if (f.project_id) parts.push(names.get(f.project_id) || `project ${f.project_id}`);
      if (f.tag) parts.push(`#${f.tag}`);
      if (f.min_priority) parts.push(`priority ≥ ${f.min_priority}`);
      if (f.due) parts.push(DUE_LABELS[f.due] || f.due);
      if (f.status && f.status !== "todo") parts.push(f.status === "done" ? "completed" : "open + completed");
      listEl.append(el("div", { class: "card-row" },
        el("div", { class: "grow" }, el("a", { href: `/views/${f.id}` }, el("strong", {}, f.name)), el("div", { class: "muted" }, parts.join(" · ") || "All open tasks")),
        readOnly() ? null : el("button", { type: "button", class: "btn", onclick: () => openFilterDialog({ filter: f, onSaved: reload }) }, "Edit")));
    }
  } catch (e) { toastError(e); }
}

if (!readOnly()) {
  $("#page-actions").append(el("button", { class: "btn", type: "button", onclick: () => openFilterDialog({ onSaved: reload }) }, "New view…"));
}
load();
