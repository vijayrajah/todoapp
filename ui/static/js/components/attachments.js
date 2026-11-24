import { el, readOnly, clear } from "../lib/dom.js";
import { get, post, del } from "../lib/api.js";
import { toastError } from "./toast.js";

const MAX_BYTES = 10 * 1024 * 1024;

const sizeLabel = (n) => (n >= 1048576 ? `${(n / 1048576).toFixed(1)} MB` : `${Math.max(1, Math.round(n / 1024))} KB`);

function toBase64(file) {
  return new Promise((resolve, reject) => {
    const r = new FileReader();
    r.onload = () => resolve(String(r.result).split(",")[1] || "");
    r.onerror = () => reject(r.error);
    r.readAsDataURL(file);
  });
}

// Attachment list + uploader for a saved task or note (kind: "task" | "note").
export function attachmentsSection(kind, ownerId) {
  const box = el("div", { class: "field" }, el("label", {}, "Attachments"));
  if (!ownerId) {
    box.append(el("p", { class: "hint" }, "Save first, then reopen to add attachments."));
    return box;
  }
  const list = el("div", { class: "attachments" });
  const base = `/api/${kind}s/${ownerId}/attachments`;

  const load = async () => {
    try {
      const items = (await get(base)) || [];
      clear(list);
      for (const a of items) {
        const isImg = ["image/png", "image/jpeg", "image/gif", "image/webp"].includes(a.content_type);
        list.append(el("div", { class: "attachment" },
          isImg ? el("img", { src: `/api/attachments/${a.id}`, alt: a.filename, loading: "lazy" }) : null,
          el("a", { class: "name", href: `/api/attachments/${a.id}?download=1`, download: a.filename }, a.filename),
          el("span", { class: "muted" }, sizeLabel(a.size)),
          readOnly() ? null : el("button", { type: "button", class: "icon-btn", title: "Delete", "aria-label": `Delete ${a.filename}`,
            onclick: async () => { try { await del(`/api/attachments/${a.id}`); load(); } catch (e) { toastError(e); } } }, "×")));
      }
    } catch (e) { toastError(e); }
  };

  const input = el("input", { type: "file", multiple: true, hidden: true, onchange: async () => {
    for (const f of input.files) {
      if (f.size > MAX_BYTES) { toastError(new Error(`${f.name} is larger than 10 MB`)); continue; }
      try { await post(base, { filename: f.name, data: await toBase64(f) }); } catch (e) { toastError(e); }
    }
    input.value = "";
    load();
  } });
  box.append(list, input);
  if (!readOnly()) box.append(el("button", { type: "button", class: "btn", onclick: () => input.click() }, "Add file…"));
  load();
  return box;
}
