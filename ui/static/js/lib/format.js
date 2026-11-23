const DAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

export const PRIORITIES = {
  3: { name: "High", color: "var(--p3)" },
  2: { name: "Medium", color: "var(--p2)" },
  1: { name: "Low", color: "var(--p1)" },
  0: { name: "None", color: "var(--p0)" },
};

const pad = (n) => String(n).padStart(2, "0");
export const toISODate = (d) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
export const todayStr = () => document.body.dataset.today || toISODate(new Date());

export function parseDate(s) {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}

export function dayDiff(s) {
  return Math.round((parseDate(s) - parseDate(todayStr())) / 86400000);
}

export const isOverdue = (due) => !!due && due < todayStr();

export function dueLabel(due) {
  if (!due) return "";
  const diff = dayDiff(due);
  if (diff === 0) return "Today";
  if (diff === 1) return "Tomorrow";
  if (diff === -1) return "Yesterday";
  const d = parseDate(due);
  const year = d.getFullYear() !== parseDate(todayStr()).getFullYear() ? " " + d.getFullYear() : "";
  return `${DAYS[d.getDay()]} ${d.getDate()} ${MONTHS[d.getMonth()]}${year}`;
}

export function parseRecurrence(r) {
  if (!r) return { unit: "", interval: 1, after: false };
  const [unit, interval, after] = r.split(":");
  return { unit, interval: parseInt(interval, 10) || 1, after: after === "after" };
}

export function buildRecurrence({ unit, interval, after }) {
  if (!unit) return "";
  return `${unit}:${Math.max(1, interval | 0)}${after ? ":after" : ""}`;
}

export function recurrenceLabel(r) {
  const { unit, interval, after } = parseRecurrence(r);
  if (!unit) return "";
  const word = { daily: "day", weekly: "week", monthly: "month", yearly: "year" }[unit] || unit;
  const base = interval === 1 ? `Every ${word}` : `Every ${interval} ${word}s`;
  return after ? `${base} after completion` : base;
}

export function relTime(iso) {
  const diff = new Date(iso).getTime() - Date.now();
  const abs = Math.abs(diff);
  const units = [["day", 86400000], ["hour", 3600000], ["minute", 60000]];
  let text = "just now";
  for (const [name, ms] of units) {
    if (abs >= ms) {
      const n = Math.round(abs / ms);
      text = `${n} ${name}${n === 1 ? "" : "s"}`;
      return diff > 0 ? `in ${text}` : `${text} ago`;
    }
  }
  return text;
}

export function fmtDateTime(iso) {
  const d = new Date(iso);
  return `${DAYS[d.getDay()]} ${d.getDate()} ${MONTHS[d.getMonth()]}, ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function greeting(hour = new Date().getHours()) {
  if (hour < 5) return "Good night";
  if (hour < 12) return "Good morning";
  if (hour < 18) return "Good afternoon";
  return "Good evening";
}

export function tomorrowNineISO() {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  d.setHours(9, 0, 0, 0);
  return d.toISOString();
}

// Value for <input type="datetime-local"> from an ISO string.
export function toLocalInput(iso) {
  const d = iso ? new Date(iso) : new Date(Date.now() + 3600000);
  return `${toISODate(d)}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function initials(name) {
  const parts = (name || "").trim().split(/\s+/).filter(Boolean);
  if (!parts.length) return "?";
  return (parts[0][0] + (parts.length > 1 ? parts[parts.length - 1][0] : "")).toUpperCase();
}
