import { el, $ } from "../lib/dom.js";
import { get, post, put, del } from "../lib/api.js";
import { PRIORITIES, parseRecurrence, buildRecurrence, DAY_CODES, DAY_NAMES } from "../lib/format.js";
import { modal, field, confirmDialog } from "../lib/modal.js";
import { tagInput } from "./tag-input.js";
import { checklistEditor } from "./checklist.js";
import { attachmentsSection } from "./attachments.js";
import { toast, toastError } from "./toast.js";

const option = (value, label, selected) => el("option", { value, selected: !!selected }, label);

// Create (task = null) or edit a task. onSaved() runs after save/delete. dueDate pre-fills a new task.
export async function openTaskDialog({ task = null, projectId = 0, parentId = 0, dueDate = "", onSaved } = {}) {
  let projects = [];
  try { projects = (await get("/api/projects")) || []; } catch (e) { toastError(e); }
  const t = task || {};
  const isSub = !!(t.parent_id || parentId);

  const title = el("input", { type: "text", value: t.title || "", required: true, maxlength: "300" });
  const desc = el("textarea", { rows: "5", placeholder: "Markdown supported" }, t.description || "");
  const prio = el("select", {}, ...[0, 1, 2, 3].map((p) => option(p, PRIORITIES[p].name, (t.priority || 0) === p)));
  const due = el("input", { type: "date", value: t.due_date || dueDate });
  const proj = el("select", {}, option(0, "No project", !t.project_id && !projectId),
    ...projects.map((p) => option(p.id, p.name, (t.project_id || projectId) === p.id)));
  if (isSub) proj.disabled = true;

  const rec = parseRecurrence(t.recurrence);
  const unit = el("select", {}, option("", "Does not repeat", !rec.unit),
    ...["daily", "weekly", "monthly", "yearly"].map((u) => option(u, u[0].toUpperCase() + u.slice(1), rec.unit === u)));
  const interval = el("input", { type: "number", min: "1", value: rec.interval, "aria-label": "Repeat every N" });
  const after = el("input", { type: "checkbox", checked: rec.after });
  const recRow = el("div", { class: "row" },
    el("div", { class: "field" }, el("label", {}, "Repeat"), unit),
    el("div", { class: "field" }, el("label", {}, "Every (interval)"), interval));
  const afterRow = el("label", { class: "inline field" }, after, "Repeat from completion date");
  const syncRec = () => {
    interval.disabled = after.disabled = !unit.value;
    dayRow.hidden = unit.value !== "weekly";
  };
  const dayRow = el("div", { class: "field weekdays-field" }, el("label", {}, "On days"),
    el("div", { class: "weekdays" }, ...DAY_CODES.map((c, i) =>
      el("label", {}, el("input", { type: "checkbox", value: c, checked: rec.days.includes(c) }), DAY_NAMES[i]))));
  const pickedDays = () => [...dayRow.querySelectorAll("input:checked")].map((i) => i.value);
  unit.addEventListener("change", syncRec);
  syncRec();

  const tags = tagInput(t.tags || []);
  const checklist = checklistEditor(t.checklist || []);

  const body = [
    field("Title", title), field("Description", desc),
    el("div", { class: "row" }, field("Priority", prio), field("Due date", due)),
  ];
  if (!isSub) body.push(field("Project", proj), recRow, dayRow, afterRow);
  body.push(el("div", { class: "field" }, el("label", {}, "Tags"), tags.node), checklist.node);
  if (t.id) body.push(attachmentsSection("task", t.id));

  const extra = [];
  if (t.id && !isSub) {
    extra.push(el("button", { type: "button", class: "btn", onclick: async () => {
      try {
        await post("/api/templates", {
          name: t.title, title: t.title, description: t.description || "", priority: t.priority || 0,
          tags: t.tags || [], checklist: (t.checklist || []).map((c) => c.text),
          subtasks: (t.subtasks || []).map((s) => s.title),
        });
        toast("Saved as template");
      } catch (e) { toastError(e); }
    } }, "Save as template"));
  }
  if (t.id) {
    extra.push(el("button", { type: "button", class: "btn btn-danger", onclick: async () => {
      if (!(await confirmDialog("Move this task to trash?", "Delete"))) return;
      try {
        await del(`/api/tasks/${t.id}`);
        toast("Moved to trash");
        dlg.close();
        if (onSaved) onSaved();
      } catch (e) { toastError(e); }
    } }, "Delete"));
  }

  const dlg = modal({
    title: t.id ? "Edit task" : isSub ? "New subtask" : "New task",
    body, extra,
    onSubmit: async () => {
      const payload = {
        title: title.value.trim(),
        description: desc.value,
        priority: parseInt(prio.value, 10),
        due_date: due.value,
        project_id: isSub ? 0 : parseInt(proj.value, 10),
        parent_id: t.parent_id || parentId || 0,
        recurrence: isSub ? "" : buildRecurrence({ unit: unit.value, interval: parseInt(interval.value, 10), after: after.checked, days: pickedDays() }),
        tags: tags.get(),
        checklist: checklist.get(),
      };
      if (t.id) await put(`/api/tasks/${t.id}`, payload);
      else await post("/api/tasks", payload);
      if (onSaved) onSaved();
    },
  });
  return dlg;
}
