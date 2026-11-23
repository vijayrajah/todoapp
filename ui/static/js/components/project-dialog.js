import { el } from "../lib/dom.js";
import { post, put } from "../lib/api.js";
import { modal, field } from "../lib/modal.js";

export const PROJECT_COLORS = ["#5b6cf0", "#e5484d", "#f0a020", "#30a46c", "#12a594", "#3b82f6", "#8e4ec6", "#d6409f", "#7c7f93"];

// Create (project = null) or edit a project. onSaved(project) runs after save.
export function openProjectDialog({ project = null, onSaved } = {}) {
  const p = project || {};
  let color = p.color || PROJECT_COLORS[0];
  const name = el("input", { type: "text", value: p.name || "", required: true, maxlength: "100" });
  const desc = el("textarea", { rows: "3" }, p.description || "");

  const swatches = el("div", { class: "swatches", role: "radiogroup", "aria-label": "Colour" });
  const renderSwatches = () => swatches.replaceChildren(...PROJECT_COLORS.map((c) =>
    el("button", { type: "button", class: "swatch", role: "radio", "aria-checked": String(c === color), "aria-label": c,
      style: `background:${c}`, onclick: () => { color = c; renderSwatches(); } })));
  renderSwatches();

  return modal({
    title: p.id ? "Edit project" : "New project",
    body: [field("Name", name), field("Description", desc), el("div", { class: "field" }, el("label", {}, "Colour"), swatches)],
    onSubmit: async () => {
      const payload = { name: name.value.trim(), description: desc.value, color };
      const saved = p.id ? await put(`/api/projects/${p.id}`, payload) : await post("/api/projects", payload);
      if (onSaved) onSaved(saved);
    },
  });
}
