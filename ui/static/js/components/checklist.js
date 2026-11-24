import { el } from "../lib/dom.js";

// Editable checklist. get() returns [{text, done}] without blank rows.
export function checklistEditor(items = []) {
  const list = el("ul", { class: "checklist" });
  const add = (item = { text: "", done: false }) => {
    const done = el("input", { type: "checkbox", checked: !!item.done, "aria-label": "Done" });
    const text = el("input", { type: "text", value: item.text, maxlength: "300", placeholder: "Checklist item", "aria-label": "Checklist item",
      onkeydown: (e) => {
        if (e.key !== "Enter") return;
        e.preventDefault(); // Enter adds a new row instead of submitting the dialog
        add().querySelector("input[type=text]").focus();
      } });
    const li = el("li", {}, done, text,
      el("button", { type: "button", class: "icon-btn", title: "Remove", "aria-label": "Remove item", onclick: () => li.remove() }, "×"));
    list.append(li);
    return li;
  };
  items.forEach(add);
  const node = el("div", { class: "field" },
    el("label", {}, "Checklist"), list,
    el("button", { type: "button", class: "btn", onclick: () => add().querySelector("input[type=text]").focus() }, "Add item"));
  const get = () => [...list.children]
    .map((li) => ({ text: li.querySelector("input[type=text]").value.trim(), done: li.querySelector("input[type=checkbox]").checked }))
    .filter((i) => i.text);
  return { node, get };
}
