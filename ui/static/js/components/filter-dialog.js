import { el } from "../lib/dom.js";
import { get, post, put, del } from "../lib/api.js";
import { PRIORITIES } from "../lib/format.js";
import { modal, field, confirmDialog } from "../lib/modal.js";
import { toastError } from "./toast.js";

export const DUE_LABELS = {
  "": "Any date", none: "No due date", overdue: "Overdue", today: "Due today", today_overdue: "Today or overdue",
  next7: "Next 7 days", this_week: "This week", this_month: "This month",
};
const STATUS_LABELS = { todo: "Open", done: "Completed", any: "Open and completed" };

const option = (value, label, selected) => el("option", { value, selected: !!selected }, label);

// Create (filter = null or a prefill without id) or edit a saved view. onSaved(saved) runs afterwards.
export async function openFilterDialog({ filter = null, onSaved } = {}) {
  let projects = [];
  try { projects = (await get("/api/projects")) || []; } catch (e) { toastError(e); }
  const f = filter || {};

  const name = el("input", { type: "text", value: f.name || "", required: true, maxlength: "100" });
  const proj = el("select", {}, option(0, "Any project", !f.project_id), ...projects.map((p) => option(p.id, p.name, f.project_id === p.id)));
  const tag = el("input", { type: "text", value: f.tag || "", placeholder: "e.g. home", maxlength: "50" });
  const prio = el("select", {}, option(0, "Any priority", !f.min_priority),
    ...[1, 2, 3].map((p) => option(p, `${PRIORITIES[p].name} or higher`, f.min_priority === p)));
  const due = el("select", {}, ...Object.entries(DUE_LABELS).map(([v, l]) => option(v, l, (f.due || "") === v)));
  const status = el("select", {}, ...Object.entries(STATUS_LABELS).map(([v, l]) => option(v, l, (f.status || "todo") === v)));

  const extra = [];
  if (f.id) {
    extra.push(el("button", { type: "button", class: "btn btn-danger", onclick: async () => {
      if (!(await confirmDialog("Delete this saved view?", "Delete"))) return;
      try { await del(`/api/filters/${f.id}`); dlg.close(); if (onSaved) onSaved(null); } catch (e) { toastError(e); }
    } }, "Delete"));
  }
  const dlg = modal({
    title: f.id ? "Edit saved view" : "Save as view",
    body: [field("Name", name), field("Project", proj), field("Tag", tag),
      el("div", { class: "row" }, field("Priority", prio), field("Due", due)), field("Status", status)],
    extra,
    onSubmit: async () => {
      const payload = {
        name: name.value.trim(), project_id: parseInt(proj.value, 10), tag: tag.value.trim(),
        min_priority: parseInt(prio.value, 10), due: due.value, status: status.value,
      };
      const saved = f.id ? await put(`/api/filters/${f.id}`, payload) : await post("/api/filters", payload);
      if (onSaved) onSaved(saved);
    },
  });
  return dlg;
}
