import { el } from "../lib/dom.js";
import { post, put, del } from "../lib/api.js";
import { modal, field, confirmDialog } from "../lib/modal.js";
import { tagInput } from "./tag-input.js";
import { toast, toastError } from "./toast.js";

// Create (note = null) or edit a note. onSaved() runs after save/delete.
export function openNoteDialog({ note = null, onSaved } = {}) {
  const n = note || {};
  const title = el("input", { type: "text", value: n.title || "", required: true, maxlength: "200" });
  const bodyText = el("textarea", { rows: "12", placeholder: "Write in Markdown…" }, n.body || "");
  const pinned = el("input", { type: "checkbox", checked: !!n.pinned });
  const tags = tagInput(n.tags || []);

  const extra = [];
  if (n.id) {
    extra.push(el("button", { type: "button", class: "btn btn-danger", onclick: async () => {
      if (!(await confirmDialog("Move this note to trash?", "Delete"))) return;
      try {
        await del(`/api/notes/${n.id}`);
        toast("Moved to trash");
        dlg.close();
        if (onSaved) onSaved();
      } catch (e) { toastError(e); }
    } }, "Delete"));
  }

  const dlg = modal({
    title: n.id ? "Edit note" : "New note",
    body: [
      field("Title", title), field("Body", bodyText),
      el("div", { class: "field" }, el("label", {}, "Tags"), tags.node),
      el("label", { class: "inline field" }, pinned, "Pin to top"),
    ],
    extra,
    onSubmit: async () => {
      const payload = { title: title.value.trim(), body: bodyText.value, pinned: pinned.checked, tags: tags.get() };
      if (n.id) await put(`/api/notes/${n.id}`, payload);
      else await post("/api/notes", payload);
      if (onSaved) onSaved();
    },
  });
  return dlg;
}
