import { readOnly } from "./dom.js";

// HTML5 drag-and-drop reordering of direct children of `container` matching `itemSelector`.
// Dragging starts only from a `.drag-handle`. Calls onReorder(idsInOrder) from data-id if order changed.
export function makeSortable(container, itemSelector, onReorder) {
  if (readOnly()) return;
  let dragged = null;
  let before = [];
  const items = () => [...container.children].filter((c) => c.matches(itemSelector));
  const ids = () => items().map((i) => i.dataset.id);

  const directItem = (target) => {
    let n = target.closest ? target.closest(itemSelector) : null;
    while (n && n.parentElement !== container) n = n.parentElement && n.parentElement.closest(itemSelector);
    return n;
  };

  container.addEventListener("mousedown", (e) => {
    const handle = e.target.closest(".drag-handle");
    const item = handle && handle.closest(itemSelector);
    if (item && item.parentElement === container) item.draggable = true;
  });

  container.addEventListener("dragstart", (e) => {
    const item = e.target.closest ? e.target.closest(itemSelector) : null;
    if (!item || item.parentElement !== container || !item.draggable) return;
    e.stopPropagation();
    dragged = item;
    before = ids();
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", item.dataset.id || "");
    item.classList.add("dragging");
  });

  container.addEventListener("dragover", (e) => {
    if (!dragged) return;
    e.preventDefault();
    const over = directItem(e.target);
    if (!over || over === dragged) return;
    const rect = over.getBoundingClientRect();
    const after = e.clientY > rect.top + rect.height / 2;
    container.insertBefore(dragged, after ? over.nextSibling : over);
  });

  container.addEventListener("drop", (e) => { if (dragged) e.preventDefault(); });

  container.addEventListener("dragend", () => {
    if (!dragged) return;
    dragged.classList.remove("dragging");
    dragged.draggable = false;
    dragged = null;
    const after = ids();
    if (after.join() !== before.join()) onReorder(after.map(Number));
  });
}
