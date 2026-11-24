import { $, el, readOnly } from "../lib/dom.js";
import { get, put, post } from "../lib/api.js";
import { field, confirmDialog } from "../lib/modal.js";
import { PROJECT_COLORS } from "../components/project-dialog.js";
import { fmtDateTime } from "../lib/format.js";
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

const FREQUENCIES = [[1, "Every hour"], [6, "Every 6 hours"], [12, "Every 12 hours"], [24, "Daily"], [168, "Weekly"]];

function autoBackupCard(s, files) {
  const ro = readOnly();
  const enabled = el("input", { type: "checkbox", checked: s.enabled, disabled: ro });
  const dir = el("input", { type: "text", value: s.directory || "", placeholder: s.effective_directory, disabled: ro });
  const known = FREQUENCIES.some(([h]) => h === s.interval_hours);
  const freq = el("select", { disabled: ro },
    ...FREQUENCIES.map(([h, label]) => el("option", { value: String(h), selected: h === s.interval_hours }, label)),
    el("option", { value: "custom", selected: !known }, "Custom (hours)..."));
  const hours = el("input", { type: "number", min: "1", max: "8760", value: String(s.interval_hours), disabled: ro });
  const syncHours = () => { hours.classList.toggle("hidden", freq.value !== "custom"); if (freq.value !== "custom") hours.value = freq.value; };
  freq.addEventListener("change", syncHours);
  syncHours();
  const keep = el("input", { type: "number", min: "1", max: "1000", value: String(s.keep), disabled: ro });
  const err = el("p", { class: "error-text", role: "alert" });
  const status = el("p", { class: "muted" });
  const list = el("ul", { class: "muted" });

  const render = (cur, fs) => {
    status.textContent = cur.last_error ? `Last backup failed: ${cur.last_error}`
      : cur.last_backup_at ? `Last backup: ${fmtDateTime(cur.last_backup_at)}` : "No backup made yet.";
    list.replaceChildren(...fs.map((f) => el("li", {}, `${f.name} (${Math.max(1, Math.round(f.size / 1024))} KB)`)));
  };
  render(s, files);

  const form = el("form", { class: "card", onsubmit: async (e) => {
    e.preventDefault();
    err.textContent = "";
    try {
      const cur = await put("/api/backup/settings", { enabled: enabled.checked, directory: dir.value.trim(),
        interval_hours: Number(hours.value), keep: Number(keep.value) });
      dir.placeholder = cur.effective_directory;
      render(cur, await get("/api/backup/files"));
      toast("Backup settings saved");
    } catch (x) { err.textContent = x.message; }
  } },
    el("h2", {}, "Automatic backups"),
    el("label", { class: "checkbox" }, enabled, " Enable automatic backups"),
    field("Location", dir, "Absolute folder path on the server. Leave empty to use the default."),
    field("Frequency", freq), el("div", { class: "field" }, hours),
    field("Backups to keep", keep, "Older backups beyond this number are deleted."),
    err, status,
    el("div", { class: "page-actions" },
      el("button", { class: "btn-primary", type: "submit", disabled: ro }, "Save settings"),
      el("button", { class: "btn", type: "button", disabled: ro, onclick: async () => {
        try {
          await post("/api/backup/run");
          const [cur, fs] = await Promise.all([get("/api/backup/settings"), get("/api/backup/files")]);
          render(cur, fs);
          toast("Backup created");
        } catch (x) { toastError(x); }
      } }, "Back up now")),
    list);
  return form;
}

const THEME_NAMES = ["auto", "light", "dark", "midnight", "nord", "solarized", "sepia"];
const ACCENTS = ["#5b6cf0", "#0ea5e9", "#10b981", "#f59e0b", "#ef4444", "#ec4899", "#8b5cf6", "#64748b"];

function appearanceCard(p) {
  const ro = readOnly();
  const theme = el("select", { disabled: ro }, ...THEME_NAMES.map((t) => el("option", { value: t, selected: t === p.theme }, t[0].toUpperCase() + t.slice(1))));
  let accent = p.accent || "";
  const custom = el("input", { type: "color", value: accent || "#5b6cf0", disabled: ro, "aria-label": "Custom accent colour" });
  const sw = el("div", { class: "swatches", role: "radiogroup", "aria-label": "Accent colour" });
  const renderSw = () => sw.replaceChildren(...ACCENTS.map((c) =>
    el("button", { type: "button", class: "swatch", role: "radio", "aria-checked": String(c === accent), "aria-label": c, style: `background:${c}`, disabled: ro,
      onclick: () => { accent = c; custom.value = c; renderSw(); } })));
  custom.addEventListener("input", () => { accent = custom.value; renderSw(); });
  renderSw();
  const err = el("p", { class: "error-text", role: "alert" });
  return el("form", { class: "card", onsubmit: async (e) => {
    e.preventDefault();
    err.textContent = "";
    try {
      await put("/api/preferences/appearance", { theme: theme.value, accent });
      toast("Appearance saved");
      setTimeout(() => location.reload(), 400);
    } catch (x) { err.textContent = x.message; }
  } },
    el("h2", {}, "Appearance"),
    field("Theme", theme),
    el("div", { class: "field" }, el("label", {}, "Accent colour"), el("div", { class: "swatches" }, sw, custom,
      el("button", { type: "button", class: "btn", disabled: ro, onclick: () => { accent = ""; renderSw(); } }, "Theme default"))),
    err, el("button", { class: "btn-primary", type: "submit", disabled: ro }, "Save appearance"));
}

function digestCard(p) {
  const ro = readOnly();
  const enabled = el("input", { type: "checkbox", checked: p.digest_enabled, disabled: ro });
  const time = el("input", { type: "time", value: p.digest_time, disabled: ro });
  const hook = el("input", { type: "url", value: p.digest_webhook || "", placeholder: "https://ntfy.sh/my-topic", disabled: ro });
  const err = el("p", { class: "error-text", role: "alert" });
  const save = async () => {
    await put("/api/preferences/digest", { digest_enabled: enabled.checked, digest_time: time.value, digest_webhook: hook.value.trim() });
  };
  return el("form", { class: "card", onsubmit: async (e) => {
    e.preventDefault();
    err.textContent = "";
    try {
      await save();
      toast("Digest settings saved");
      setTimeout(() => location.reload(), 400); // the page shell knows whether the digest is enabled
    } catch (x) { err.textContent = x.message; }
  } },
    el("h2", {}, "Daily digest"),
    el("p", { class: "muted" }, "A summary of overdue and due-today tasks, shown once a day when you open the app, and optionally posted to a webhook."),
    el("label", { class: "checkbox" }, enabled, " Enable daily digest"),
    field("Time", time, "In your profile time zone."),
    field("Webhook URL (optional)", hook, "The digest is POSTed as plain text (works with ntfy.sh and similar services)."),
    err,
    el("div", { class: "page-actions" },
      el("button", { class: "btn-primary", type: "submit", disabled: ro }, "Save digest settings"),
      el("button", { class: "btn", type: "button", disabled: ro, onclick: async () => {
        try { await save(); await post("/api/digest/send"); toast("Digest sent"); } catch (x) { toastError(x); }
      } }, "Send now")));
}

try {
  const [settings, files, prefs] = await Promise.all([get("/api/backup/settings"), get("/api/backup/files"), get("/api/preferences")]);
  body.append(profileCard(await get("/api/profile")), appearanceCard(prefs), digestCard(prefs), dataCard(), autoBackupCard(settings, files));
} catch (e) { toastError(e); }
