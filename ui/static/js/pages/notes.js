import { $, el, clear, emptyState, params, dropParam, readOnly, setTrustedHTML } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { openNoteDialog } from "../components/note-dialog.js";
import { toastError } from "../components/toast.js";

const grid = $("#note-grid");
const filter = $("#tag-filter");
let activeTag = params().get("tag") || "";
let allNotes = [];

function card(n) {
  const open = () => openNoteDialog({ note: n, onSaved: load });
  const body = el("div", { class: "md" });
  setTrustedHTML(body, n.body_html);
  return el("article", { class: "note-card", tabindex: "0", role: "button", "aria-label": `Open note ${n.title}`, onclick: open,
    onkeydown: (e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); open(); } } },
    n.pinned ? el("span", { class: "pin", title: "Pinned" }, "📌") : null,
    el("h3", {}, n.title), body,
    (n.tags || []).length ? el("div", { class: "chip-row" }, ...n.tags.map((t) => el("span", { class: "chip" }, "#" + t))) : null);
}

function renderFilter() {
  clear(filter);
  const tags = [...new Set(allNotes.flatMap((n) => n.tags || []))].sort();
  if (!tags.length && !activeTag) return;
  const chip = (label, value) => el("button", { type: "button", class: `chip${activeTag === value ? " active" : ""}`, "aria-pressed": String(activeTag === value),
    onclick: () => { activeTag = value; load(); } }, label);
  filter.append(chip("All", ""), ...tags.map((t) => chip("#" + t, t)));
}

async function load() {
  try {
    allNotes = (await get("/api/notes" + (activeTag ? `?tag=${encodeURIComponent(activeTag)}` : ""))) || [];
  } catch (e) { return toastError(e); }
  clear(grid);
  if (!activeTag) renderFilter();
  if (!allNotes.length) return grid.append(emptyState("📝", "No notes yet."));
  grid.append(...allNotes.map(card));
}

if (!readOnly()) $("#page-actions").append(el("button", { class: "btn-primary", type: "button", onclick: () => openNoteDialog({ onSaved: load }) }, "New note"));

async function openFromQuery() {
  const id = parseInt(params().get("open"), 10);
  if (!id) return;
  dropParam("open");
  try { openNoteDialog({ note: await get(`/api/notes/${id}`), onSaved: load }); } catch (e) { toastError(e); }
}
load().then(openFromQuery);
