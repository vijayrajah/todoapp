import { el } from "../lib/dom.js";

const root = () => document.getElementById("toasts");

// show({message|node, kind, persistent, timeout}) -> {close}
export function show({ message = "", node = null, kind = "info", persistent = false, timeout = 4000 } = {}) {
  const t = el("div", { class: `toast ${kind}`, role: kind === "error" ? "alert" : "status" });
  if (node) t.append(node);
  else t.append(el("span", { class: "toast-msg" }, message));
  let closed = false;
  const close = () => {
    if (closed) return;
    closed = true;
    t.classList.add("out");
    setTimeout(() => t.remove(), 200);
  };
  root().append(t);
  if (!persistent) setTimeout(close, timeout);
  return { close, element: t };
}

export const toast = (message, opts = {}) => show({ message, ...opts });
export const toastError = (err) => show({ message: err && err.message ? err.message : String(err), kind: "error", timeout: 6000 });
