import { el } from "../lib/dom.js";
import { post, put } from "../lib/api.js";
import { taskBody } from "../lib/tasks.js";
import { toastError } from "./toast.js";

const HINT = "Try: Pay rent tomorrow !high #home @Finance every month  ·  dates: today, fri, next week, in 3 days, 2026-01-31";

// Single-input quick add. projectId (optional) assigns created tasks to that project.
export function quickAdd({ projectId = 0, onAdd } = {}) {
  const input = el("input", { type: "text", id: "quick-add-input", placeholder: "Add a task…  (press n)", "aria-label": "Quick add task", autocomplete: "off", "aria-describedby": "quick-add-hint" });
  const form = el("form", { class: "quick-add-wrap", onsubmit: async (e) => {
    e.preventDefault();
    const text = input.value.trim();
    if (!text) return;
    input.disabled = true;
    try {
      let task = await post("/api/tasks/quick", { text });
      if (projectId && task.project_id !== projectId) task = await put(`/api/tasks/${task.id}`, taskBody(task, { project_id: projectId }));
      input.value = "";
      if (onAdd) onAdd(task);
    } catch (x) { toastError(x); }
    input.disabled = false;
    input.focus();
  } },
    el("div", { class: "quick-add" }, input, el("button", { type: "submit", class: "btn-primary" }, "Add")),
    el("p", { class: "hint", id: "quick-add-hint" }, HINT));
  return form;
}
