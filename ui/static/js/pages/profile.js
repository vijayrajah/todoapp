import { $, el, readOnly } from "../lib/dom.js";
import { get, put, post } from "../lib/api.js";
import { field, confirmDialog } from "../lib/modal.js";
import { PROJECT_COLORS } from "../components/project-dialog.js";
import { toast, toastError } from "../components/toast.js";

const body = $("#profile-body");

function profileCard(p) {
  let color = p.avatar_color || PROJECT_COLORS[0];
  const name = el("input", { type: "text", value: p.name || "", maxlength: "100" });
  const email = el("input", { type: "email", value: p.email || "" });
  const tz = el("input", { type: "text", value: p.timezone || "", list: "tz-list", placeholder: "e.g. Europe/Berlin" });
  const zones = typeof Intl.supportedValuesOf === "function" ? Intl.supportedValuesOf("timeZone") : [];
  const dl = el("datalist", { id: "tz-list" }, ...zones.map((z) => el("option", { value: z })));
  const bio = el("textarea", { rows: "3" }, p.bio || "");
  const sw = el("div", { class: "swatches", role: "radiogroup", "aria-label": "Avatar colour" });
  const renderSw = () => sw.replaceChildren(...PROJECT_COLORS.map((c) =>
    el("button", { type: "button", class: "swatch", role: "radio", "aria-checked": String(c === color), "aria-label": c, style: `background:${c}`,
      onclick: () => { color = c; renderSw(); } })));
  renderSw();
  const err = el("p", { class: "error-text", role: "alert" });

  return el("form", { class: "card", onsubmit: async (e) => {
    e.preventDefault();
    err.textContent = "";
    try {
      await put("/api/profile", { name: name.value.trim(), email: email.value.trim(), timezone: tz.value.trim(), bio: bio.value, avatar_color: color });
      toast("Profile saved");
      setTimeout(() => location.reload(), 600); // refresh server-rendered name, avatar and "today"
    } catch (x) { err.textContent = x.message; }
  } },
    el("h2", {}, "About you"),
    field("Name", name), field("Email", email), field("Timezone", tz, "Used to decide what “today” means."), dl,
    field("Bio", bio), el("div", { class: "field" }, el("label", {}, "Avatar colour"), sw), err,
    el("button", { class: "btn-primary", type: "submit", disabled: readOnly() }, "Save profile"));
}

function dataCard() {
  const file = el("input", { type: "file", accept: "application/json,.json", class: "hidden", "aria-label": "Import JSON file", onchange: async () => {
    const f = file.files[0];
    if (!f) return;
    let data;
    try { data = JSON.parse(await f.text()); } catch { file.value = ""; return toastError(new Error("That file is not valid JSON")); }
    if (!(await confirmDialog("Importing replaces ALL current data (tasks, notes, projects, reminders, profile). This cannot be undone.", "Replace everything"))) { file.value = ""; return; }
    try { await post("/api/backup/import", data); toast("Import complete"); setTimeout(() => location.reload(), 600); } catch (e) { toastError(e); }
    file.value = "";
  } });
  return el("section", { class: "card" },
    el("h2", {}, "Data"),
    el("p", { class: "muted", style: "margin-bottom:.9rem" }, "Back up everything, or move it to another instance."),
    el("div", { class: "page-actions" },
      el("a", { class: "btn", href: "/api/backup/export", download: "" }, "Export JSON"),
      el("a", { class: "btn", href: "/api/backup/database", download: "" }, "Download database"),
      readOnly() ? null : el("button", { class: "btn", type: "button", onclick: () => file.click() }, "Import JSON…"), file));
}

try {
  body.append(profileCard(await get("/api/profile")), dataCard());
} catch (e) { toastError(e); }
