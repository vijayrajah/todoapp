async function request(method, url, body) {
  const opts = { method, headers: { "Content-Type": "application/json" } };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const res = await fetch(url, opts);
  const text = await res.text();
  let data = null;
  if (text) {
    try { data = JSON.parse(text); } catch { data = null; }
  }
  if (!res.ok) throw new Error((data && data.error) || res.statusText || "Request failed");
  return data;
}

export const get = (url) => request("GET", url);
export const post = (url, body = {}) => request("POST", url, body);
export const put = (url, body = {}) => request("PUT", url, body);
export const del = (url) => request("DELETE", url);

export function qs(obj) {
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(obj)) if (v !== "" && v !== null && v !== undefined && v !== 0) p.set(k, v);
  const s = p.toString();
  return s ? "?" + s : "";
}
