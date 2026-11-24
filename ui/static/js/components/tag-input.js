import { el } from "../lib/dom.js";
import { get } from "../lib/api.js";

let uid = 0;

async function loadKnownTags() {
  try { return ((await get("/api/tags")) || []).map((t) => t.name); } catch { return []; }
}

// Chip-style tag editor. Returns {node, get()}; get() includes any pending text.
export function tagInput(initial = []) {
  const tags = [...initial];
  const input = el("input", { type: "text", placeholder: "Add tag…", "aria-label": "Add tag", autocomplete: "off" });
  const node = el("div", { class: "tag-input" });
  const list = el("datalist", { id: `tag-suggestions-${++uid}` });
  input.setAttribute("list", list.id);
  loadKnownTags().then((names) => list.replaceChildren(...names.map((n) => el("option", { value: n }))));

  const norm = (s) => s.trim().replace(/^#+/, "").toLowerCase();
  const commit = () => {
    for (const part of input.value.split(",")) {
      const t = norm(part);
      if (t && !tags.includes(t)) tags.push(t);
    }
    input.value = "";
    render();
  };
  const render = () => {
    node.replaceChildren(
      ...tags.map((t) =>
        el("span", { class: "chip" }, "#" + t,
          el("button", { type: "button", class: "x", "aria-label": `Remove ${t}`, onclick: () => { tags.splice(tags.indexOf(t), 1); render(); } }, "×"))
      ),
      input,
      list
    );
  };
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === ",") { e.preventDefault(); commit(); }
    else if (e.key === "Backspace" && !input.value && tags.length) { tags.pop(); render(); input.focus(); }
  });
  // Picking a suggestion fires "input" without a keyboard insert type.
  input.addEventListener("input", (e) => {
    if (!e.inputType || e.inputType === "insertReplacementText") commit(), input.focus();
  });
  input.addEventListener("blur", commit);
  render();
  return { node, get: () => { commit(); return [...tags]; } };
}
