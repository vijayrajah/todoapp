import { el, readOnly } from "../lib/dom.js";
import { get, post, put, del } from "../lib/api.js";
import { PRIORITIES } from "../lib/format.js";
import { modal, field, confirmDialog } from "../lib/modal.js";
import { tagInput } from "./tag-input.js";
import { toast, toastError } from "./toast.js";

const lines = (s) => s.split("\n").map((x) => x.trim()).filter(Boolean);

// Create or edit a template. onSaved() runs afterwards.
export function openTemplateDialog({ template = null, onSaved } = {}) {
  const t = template || {};
  const name = el("input", { type: "text", value: t.name || "", required: true, maxlength: "100" });
  const title = el("input", { type: "text", value: t.title || "", required: true, maxlength: "300" });
  const desc = el("textarea", { rows: "3", placeholder: "Markdown supported" }, t.description || "");
  const prio = el("select", {}, ...[0, 1, 2, 3].map((p) => el("option", { value: p, selected: (t.priority || 0) === p }, PRIORITIES[p].name)));
  const tags = tagInput(t.tags || []);
  const checklist = el("textarea", { rows: "5", placeholder: "One item per line" }, (t.checklist || []).join("\n"));
  const subtasks = el("textarea", { rows: "4", placeholder: "One subtask per line" }, (t.subtasks || []).join("\n"));

  const extra = [];
  if (t.id) {
    extra.push(el("button", { type: "button", class: "btn btn-danger", onclick: async () => {
      if (!(await confirmDialog("Delete this template?", "Delete"))) return;
      try { await del(`/api/templates/${t.id}`); dlg.close(); if (onSaved) onSaved(); } catch (e) { toastError(e); }
    } }, "Delete"));
  }
  const dlg = modal({
    title: t.id ? "Edit template" : "New template",
    body: [field("Template name", name), field("Task title", title), field("Description", desc), field("Priority", prio),
      el("div", { class: "field" }, el("label", {}, "Tags"), tags.node),
      field("Checklist items", checklist), field("Subtasks", subtasks)],
    extra,
    onSubmit: async () => {
      const payload = {
        name: name.value.trim(), title: title.value.trim(), description: desc.value, priority: parseInt(prio.value, 10),
        tags: tags.get(), checklist: lines(checklist.value), subtasks: lines(subtasks.value),
      };
      if (t.id) await put(`/api/templates/${t.id}`, payload);
      else await post("/api/templates", payload);
      if (onSaved) onSaved();
    },
  });
  return dlg;
}

// Create a task from a template, with optional project and due date. onCreated(task) runs afterwards.
export async function openUseTemplateDialog({ template, projectId = 0, onCreated } = {}) {
  let projects = [];
  try { projects = (await get("/api/projects")) || []; } catch (e) { toastError(e); }
  const proj = el("select", {}, el("option", { value: 0 }, "No project"),
    ...projects.map((p) => el("option", { value: p.id, selected: p.id === projectId }, p.name)));
  const due = el("input", { type: "date" });
  return modal({
    title: `New task from “${template.name}”`,
    submitLabel: "Create",
    body: [field("Project", proj), field("Due date", due)],
    onSubmit: async () => {
      const task = await post(`/api/templates/${template.id}/use`, { project_id: parseInt(proj.value, 10), due_date: due.value });
      toast("Task created");
      if (onCreated) onCreated(task);
    },
  });
}

// Pick a template, then create a task from it.
export async function chooseTemplate({ projectId = 0, onCreated } = {}) {
  let list = [];
  try { list = (await get("/api/templates")) || []; } catch (e) { return toastError(e); }
  if (!list.length) return toast("No templates yet. Create one on the Templates page.");
  const sel = el("select", {}, ...list.map((t) => el("option", { value: t.id }, t.name)));
  modal({
    title: "New task from template", submitLabel: "Next", body: [field("Template", sel)],
    onSubmit: async () => {
      const template = list.find((t) => String(t.id) === sel.value);
      setTimeout(() => openUseTemplateDialog({ template, projectId, onCreated }), 0);
    },
  });
}

export const canEdit = () => !readOnly();
