package model

// SearchResult is one full-text search hit. Snippet is safe HTML with <mark> highlights.
type SearchResult struct {
	Kind    string `json:"kind"`
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// Trash lists soft-deleted items that can be restored.
type Trash struct {
	Tasks    []Task    `json:"tasks"`
	Notes    []Note    `json:"notes"`
	Projects []Project `json:"projects"`
}
