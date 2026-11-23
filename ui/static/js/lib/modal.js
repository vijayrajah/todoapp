import { el, readOnly } from "../lib/dom.js";

// Generic modal form. onSubmit may throw; its message is shown inline.
// extra: nodes placed at the left of the action row (e.g. a Delete button).
export function modal({ title, body, submitLabel = "Save", onSubmit, extra = [] }) {
  const dlg = el("dialog", { "aria-label": title });
  const err = el("p", { class: "error-text", role: "alert" });
  const submit = el("button", { type: "submit", class: "btn-primary", disabled: readOnly() }, submitLabel);
  const cancel = el("button", { type: "button", class: "btn", onclick: () => dlg.close() }, "Cancel");
  const form = el(
    "form",
    {
      onsubmit: async (e) => {
        e.preventDefault();
        err.textContent = "";
        submit.disabled = true;
        try {
          await onSubmit();
          dlg.close();
        } catch (x) {
          err.textContent = x.message;
          submit.disabled = readOnly();
        }
      },
    },
    el("h2", {}, title),
    ...body,
    err,
    el("div", { class: "dialog-actions" }, el("span", { class: "spacer" }, ...extra), cancel, submit)
  );
  dlg.append(form);
  dlg.addEventListener("close", () => dlg.remove());
  dlg.addEventListener("mousedown", (e) => { if (e.target === dlg) dlg.close(); });
  document.body.append(dlg);
  dlg.showModal();
  return dlg;
}

export function field(label, input, hint) {
  const id = "f" + Math.random().toString(36).slice(2, 8);
  input.id = id;
  return el("div", { class: "field" }, el("label", { for: id }, label), input, hint ? el("p", { class: "hint" }, hint) : null);
}

export function confirmDialog(message, label = "Confirm") {
  return new Promise((resolve) => {
    let ok = false;
    const dlg = modal({
      title: "Are you sure?",
      body: [el("p", {}, message)],
      submitLabel: label,
      onSubmit: async () => { ok = true; },
    });
    dlg.addEventListener("close", () => resolve(ok));
  });
}
