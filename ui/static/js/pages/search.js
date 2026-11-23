import { $, el, clear, emptyState, params, setTrustedHTML } from "../lib/dom.js";
import { get } from "../lib/api.js";
import { toastError } from "../components/toast.js";

const out = $("#search-results");
const q = (params().get("q") || "").trim();
const input = $("#search-input");
if (input) input.value = q;

const LINKS = { task: (id) => `/tasks?open=${id}`, note: (id) => `/notes?open=${id}`, project: (id) => `/projects/${id}` };
const GROUPS = [["task", "Tasks"], ["note", "Notes"], ["project", "Projects"]];

async function run() {
  if (!q) return out.append(emptyState("🔎", "Type something in the search box."));
  let res;
  try { res = (await get(`/api/search?q=${encodeURIComponent(q)}`)) || []; } catch (e) { return toastError(e); }
  clear(out);
  if (!res.length) return out.append(emptyState("🤷", `No results for “${q}”.`));
  for (const [kind, label] of GROUPS) {
    const items = res.filter((r) => r.kind === kind);
    if (!items.length) continue;
    out.append(el("section", { class: "section" },
      el("h2", { class: "group-title" }, label, " ", el("span", { class: "count" }, String(items.length))),
      ...items.map((r) => {
        const snip = el("div", { class: "snippet" });
        setTrustedHTML(snip, r.snippet);
        return el("div", { class: "result" }, el("a", { class: "title", href: LINKS[kind](r.id) }, r.title), snip);
      })));
  }
}
run();
