import { el } from "../lib/dom.js";

// Chip-style tag editor. Returns {node, get()}; get() includes any pending text.
export function tagInput(initial = []) {
  const tags = [...initial];
  const input = el("input", { type: "text", placeholder: "Add tag…", "aria-label": "Add tag", autocomplete: "off" });
  const node = el("div", { class: "tag-input" });

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
      input
    );
  };
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === ",") { e.preventDefault(); commit(); }
    else if (e.key === "Backspace" && !input.value && tags.length) { tags.pop(); render(); input.focus(); }
  });
  input.addEventListener("blur", commit);
  render();
  return { node, get: () => { commit(); return [...tags]; } };
}
