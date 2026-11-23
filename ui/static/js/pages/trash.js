import { $, el, clear, emptyState, readOnly } from "../lib/dom.js";
import { get, post, del } from "../lib/api.js";
import { confirmDialog } from "../lib/modal.js";
import { toast, toastError } from "../components/toast.js";

const list = $("#trash-list");
const KINDS = [["tasks", "task", "Tasks", "title"], ["notes", "note", "Notes", "title"], ["projects", "project", "Projects", "name"]];

async function act(fn, msg) {
  try { await fn(); toast(msg); load(); } catch (e) { toastError(e); }
}

function row(kind, label, item) {
  return el("div", { class: "trash-item" },
    el("div", { class: "grow" }, item[label]),
    readOnly() ? null : [
      el("button", { class: "btn btn-sm", type: "button", onclick: () => act(() => post(`/api/trash/${kind}/${item.id}/restore`), "Restored") }, "Restore"),
      el("button", { class: "btn btn-sm btn-danger", type: "button", onclick: async () => {
        if (await confirmDialog("Delete this item forever? This cannot be undone.", "Delete forever")) act(() => del(`/api/trash/${kind}/${item.id}`), "Deleted forever");
      } }, "Delete forever"),
    ]);
}

async function load() {
  let data;
  try { data = (await get("/api/trash")) || {}; } catch (e) { return toastError(e); }
  clear(list);
  let total = 0;
  for (const [key, kind, title, label] of KINDS) {
    const items = data[key] || [];
    total += items.length;
    if (!items.length) continue;
    list.append(el("section", { class: "section" }, el("h2", { class: "group-title" }, title, " ", el("span", { class: "count" }, String(items.length))),
      ...items.map((i) => row(kind, label, i))));
  }
  if (!total) list.append(emptyState("🗑️", "Trash is empty."));
  empty.disabled = total === 0;
}

const empty = el("button", { class: "btn btn-danger", type: "button", onclick: async () => {
  if (await confirmDialog("Permanently delete everything in the trash?", "Empty trash")) act(() => del("/api/trash"), "Trash emptied");
} }, "Empty trash");
if (!readOnly()) $("#page-actions").append(empty);
load();
