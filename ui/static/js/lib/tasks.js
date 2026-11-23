// Build a create/update body from an existing Task (PUT replaces all fields).
export function taskBody(t, overrides = {}) {
  return {
    title: t.title,
    description: t.description || "",
    priority: t.priority || 0,
    due_date: t.due_date || "",
    project_id: t.project_id || 0,
    parent_id: t.parent_id || 0,
    recurrence: t.recurrence || "",
    tags: t.tags || [],
    ...overrides,
  };
}
