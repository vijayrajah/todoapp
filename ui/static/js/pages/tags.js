import { $, el, clear, emptyState, params, setTrustedHTML } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { taskList } from "../components/task-list.js";
import { toastError } from "../components/toast.js";

const cloud = $("#tag-cloud");
const results = $("#tag-results");
const name = params().get("name") || "";

async function showTag() {
  document.querySelector("h1").textContent = `#${name}`;
  try {
    const [tasks, notes] = await Promise.all([
      get(`/api/tasks?view=all&tag=${encodeURIComponent(name)}`),
      get(`/api/notes?tag=${encodeURIComponent(name)}`),
    ]);
    clear(results);
    if (!(tasks || []).length && !(notes || []).length) results.append(emptyState("🏷️", "Nothing carries this tag."));
    if ((tasks || []).length) results.append(el("section", { class: "task-group" }, el("h2", { class: "group-title" }, "Tasks"), taskList(tasks, { refresh: showTag })));
    if ((notes || []).length) {
      results.append(el("section", { class: "task-group" }, el("h2", { class: "group-title" }, "Notes"),
        el("div", { class: "note-grid" }, ...notes.map((n) => {
          const md = el("div", { class: "md" });
          setTrustedHTML(md, n.body_html);
          return el("a", { class: "note-card", href: `/notes?open=${n.id}`, style: "display:block;color:inherit;text-decoration:none" }, el("h3", {}, n.title), md);
        }))));
    }
  } catch (e) { toastError(e); }
}

async function showCloud() {
  let tags;
  try { tags = (await get("/api/tags")) || []; } catch (e) { return toastError(e); }
  if (!tags.length) return cloud.append(emptyState("🏷️", "No tags yet. Add #tags to tasks and notes."));
  cloud.append(...tags.map((t) => el("a", { class: `chip tag-cloud-item${t.name === name ? " active" : ""}`, href: `/tags?name=${encodeURIComponent(t.name)}` },
    "#" + t.name, el("span", { class: "count" }, String(t.task_count + t.note_count)))));
}

showCloud();
if (name) showTag();
