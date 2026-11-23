import { el } from "../lib/dom.js";
import { post } from "../lib/api.js";
import { toLocalInput } from "../lib/format.js";
import { modal, field } from "../lib/modal.js";
import { toast } from "./toast.js";

// Create a reminder (optionally attached to a task). onSaved() runs after success.
export function openReminderDialog({ taskId = 0, message = "", onSaved } = {}) {
  const msg = el("input", { type: "text", value: message, required: true, maxlength: "300" });
  const when = el("input", { type: "datetime-local", value: toLocalInput(), required: true });
  modal({
    title: "New reminder",
    body: [field("Message", msg), field("When", when)],
    submitLabel: "Add reminder",
    onSubmit: async () => {
      if (!when.value) throw new Error("Pick a date and time");
      const body = { message: msg.value.trim(), remind_at: new Date(when.value).toISOString() };
      if (taskId) body.task_id = taskId;
      await post("/api/reminders", body);
      toast("Reminder added");
      if (onSaved) onSaved();
    },
  });
}
