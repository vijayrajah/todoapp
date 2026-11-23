// Safe DOM helpers. User text always goes through textContent / text nodes.
export function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (v === null || v === undefined || v === false) continue;
    if (k.startsWith("on") && typeof v === "function") node.addEventListener(k.slice(2), v);
    else if (k === "class") node.className = v;
    else if (k === "dataset") Object.assign(node.dataset, v);
    else if (k === "style") node.style.cssText = v;
    else if (["value", "checked", "disabled", "selected", "hidden"].includes(k)) node[k] = v;
    else node.setAttribute(k, v === true ? "" : v);
  }
  append(node, children);
  return node;
}

export function append(node, children) {
  for (const c of children.flat(Infinity)) {
    if (c === null || c === undefined || c === false) continue;
    node.append(c instanceof Node ? c : document.createTextNode(String(c)));
  }
  return node;
}

// Only for server-rendered, sanitised fields: description_html, body_html, snippet.
export function setTrustedHTML(node, html) {
  node.innerHTML = html || "";
  return node;
}

export function clear(node) {
  node.replaceChildren();
  return node;
}

export const $ = (sel, root = document) => root.querySelector(sel);
export const readOnly = () => document.body.dataset.readOnly === "true";

export function emptyState(icon, text) {
  return el("div", { class: "empty" }, el("span", { class: "big", "aria-hidden": "true" }, icon), text);
}

export function params() {
  return new URLSearchParams(location.search);
}

// Remove a one-shot query param (e.g. ?open=ID) without reloading.
export function dropParam(name) {
  const p = params();
  p.delete(name);
  const q = p.toString();
  history.replaceState(null, "", location.pathname + (q ? "?" + q : ""));
}
