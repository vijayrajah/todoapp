import { $, el, clear, readOnly } from "../lib/dom.js";
import { get, put, qs } from "../lib/api.js";
import { toISODate, parseDate, todayStr } from "../lib/format.js";
import { taskBody } from "../lib/tasks.js";
import { openTaskDialog } from "../components/task-dialog.js";
import { toast, toastError } from "../components/toast.js";

const grid = $("#cal-grid");
const DOW = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const MONTHS = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

let mode = "month";
let anchor = parseDate(todayStr());

const addDays = (d, n) => { const x = new Date(d); x.setDate(x.getDate() + n); return x; };
const mondayOf = (d) => addDays(d, -((d.getDay() + 6) % 7));

function range() {
  if (mode === "week") {
    const start = mondayOf(anchor);
    return { start, end: addDays(start, 6) };
  }
  const first = new Date(anchor.getFullYear(), anchor.getMonth(), 1);
  const last = new Date(anchor.getFullYear(), anchor.getMonth() + 1, 0);
  return { start: mondayOf(first), end: addDays(mondayOf(last), 6) };
}

function chip(t) {
  const overdue = t.status !== "done" && t.due_date < todayStr();
  return el("div", {
    class: `cal-chip${t.status === "done" ? " done" : ""}${overdue ? " overdue" : ""}`,
    draggable: readOnly() ? null : "true",
    dataset: { p: t.priority || 0 },
    title: t.title,
    onclick: (e) => { e.stopPropagation(); openTaskDialog({ task: t, onSaved: load }); },
    ondragstart: (e) => { e.dataTransfer.setData("text/plain", String(t.id)); e.dataTransfer.effectAllowed = "move"; },
  }, t.title);
}

async function reschedule(id, date) {
  try {
    const t = await get(`/api/tasks/${id}`);
    if (t.due_date === date) return;
    await put(`/api/tasks/${id}`, taskBody(t, { due_date: date }));
    toast("Rescheduled");
  } catch (e) { toastError(e); }
  load();
}

async function load() {
  const { start, end } = range();
  $("#cal-title").textContent = mode === "week"
    ? `${start.getDate()} ${MONTHS[start.getMonth()].slice(0, 3)} – ${end.getDate()} ${MONTHS[end.getMonth()].slice(0, 3)} ${end.getFullYear()}`
    : `${MONTHS[anchor.getMonth()]} ${anchor.getFullYear()}`;
  $("#cal-mode").textContent = mode === "week" ? "Month" : "Week";

  let tasks = [];
  try { tasks = (await get("/api/tasks" + qs({ view: "range", from: toISODate(start), to: toISODate(end) }))) || []; } catch (e) { toastError(e); }
  const byDate = new Map();
  for (const t of tasks) byDate.set(t.due_date, [...(byDate.get(t.due_date) || []), t]);

  clear(grid);
  grid.className = `cal-grid ${mode}`;
  for (const d of DOW) grid.append(el("div", { class: "cal-dow" }, d));
  for (let d = start; d <= end; d = addDays(d, 1)) {
    const iso = toISODate(d);
    const cell = el("div", {
      class: `cal-day${iso === todayStr() ? " today" : ""}${mode === "month" && d.getMonth() !== anchor.getMonth() ? " other" : ""}`,
      onclick: () => { if (!readOnly()) openTaskDialog({ dueDate: iso, onSaved: load }); },
      ondragover: (e) => { if (readOnly()) return; e.preventDefault(); cell.classList.add("drop"); },
      ondragleave: () => cell.classList.remove("drop"),
      ondrop: (e) => {
        e.preventDefault();
        cell.classList.remove("drop");
        const id = parseInt(e.dataTransfer.getData("text/plain"), 10);
        if (id) reschedule(id, iso);
      },
    }, el("span", { class: "num" }, d.getDate()), ...(byDate.get(iso) || []).map(chip));
    grid.append(cell);
  }
}

const step = (n) => {
  anchor = mode === "week" ? addDays(anchor, 7 * n) : new Date(anchor.getFullYear(), anchor.getMonth() + n, 1);
  load();
};
$("#cal-prev").addEventListener("click", () => step(-1));
$("#cal-next").addEventListener("click", () => step(1));
$("#cal-today").addEventListener("click", () => { anchor = parseDate(todayStr()); load(); });
$("#cal-mode").addEventListener("click", () => { mode = mode === "week" ? "month" : "week"; load(); });
load();
