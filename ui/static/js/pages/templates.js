import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { openTemplateDialog, openUseTemplateDialog } from "../components/template-dialog.js";
import { toastError } from "../components/toast.js";

const listEl = $("#template-list");

async function load() {
  try {
    const items = (await get("/api/templates")) || [];
    clear(listEl);
    if (!items.length) return listEl.append(emptyState("📋", "No templates yet."));
    for (const t of items) {
      const parts = [];
      if ((t.checklist || []).length) parts.push(`${t.checklist.length} checklist items`);
      if ((t.subtasks || []).length) parts.push(`${t.subtasks.length} subtasks`);
      listEl.append(el("div", { class: "card-row" },
        el("div", { class: "grow" }, el("strong", {}, t.name), el("div", { class: "muted" }, [t.title, ...parts].join(" · "))),
        readOnly() ? null : el("button", { type: "button", class: "btn", onclick: () => openUseTemplateDialog({ template: t, onCreated: (task) => { location.href = `/tasks?open=${task.id}`; } }) }, "Use"),
        el("button", { type: "button", class: "btn", onclick: () => openTemplateDialog({ template: t, onSaved: load }) }, "Edit")));
    }
  } catch (e) { toastError(e); }
}

if (!readOnly()) {
  $("#page-actions").append(el("button", { class: "btn", type: "button", onclick: () => openTemplateDialog({ onSaved: load }) }, "New template…"));
}
load();
