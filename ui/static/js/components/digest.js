import { el } from "../lib/dom.js";
import { get, post } from "../lib/api.js";
import { modal } from "../lib/modal.js";

// Shows the daily digest once per day (when enabled, due and non-empty).
export async function showDigestIfDue() {
  let d;
  try { d = await get("/api/digest"); } catch { return; }
  if (!d.show) return;
  try { await post("/api/digest/seen"); } catch { return; }

  const list = (title, tasks) => tasks.length
    ? [el("h3", {}, `${title} (${tasks.length})`),
       el("ul", { class: "digest-list" }, ...tasks.map((t) => el("li", {}, el("a", { href: `/tasks?open=${t.id}` }, t.title))))]
    : [];
  const dlg = modal({
    title: "Daily digest",
    body: [...list("Overdue", d.overdue), ...list("Due today", d.due_today)],
    submitLabel: "Open tasks",
    onSubmit: async () => { location.href = "/today"; },
  });
  dlg.setAttribute("aria-label", "Daily digest");

  if ("Notification" in window && Notification.permission === "granted") {
    new Notification("todoapp daily digest", { body: d.text });
  }
}
