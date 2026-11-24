package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const jsonCT = "application/json"

func mustJSON(t *testing.T, rec interface {
	Result() *http.Response
}, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Result().Body).Decode(v); err != nil {
		t.Fatal(err)
	}
}

func TestChecklistTemplateAndFilter(t *testing.T) {
	h := testHandler(t, false)

	rec := do(h, "POST", "/api/tasks", jsonCT, `{"title":"Trip","checklist":[{"text":"Passport"},{"text":"  "},{"text":"Charger","done":true}]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", rec.Code, rec.Body.String())
	}
	var task struct {
		ID        int64 `json:"id"`
		Checklist []struct {
			Text string `json:"text"`
			Done bool   `json:"done"`
		} `json:"checklist"`
	}
	mustJSON(t, rec, &task)
	if len(task.Checklist) != 2 || !task.Checklist[1].Done {
		t.Fatalf("checklist = %+v", task.Checklist)
	}

	// A PUT without the checklist field keeps it.
	if rec := do(h, "PUT", "/api/tasks/1", jsonCT, `{"title":"Trip 2"}`); rec.Code != 200 {
		t.Fatalf("put = %d %s", rec.Code, rec.Body.String())
	}
	rec = do(h, "GET", "/api/tasks/1", "", "")
	if !strings.Contains(rec.Body.String(), "Passport") {
		t.Errorf("checklist lost: %s", rec.Body.String())
	}

	// Templates.
	rec = do(h, "POST", "/api/templates", jsonCT, `{"name":"Packing","title":"Pack","checklist":["Socks","Shirt"],"subtasks":["Book taxi"],"tags":["Travel"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("template = %d %s", rec.Code, rec.Body.String())
	}
	rec = do(h, "POST", "/api/templates/1/use", jsonCT, `{"due_date":"2030-01-02"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("use = %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Socks", "Book taxi", "travel", "2030-01-02"} {
		if !strings.Contains(body, want) {
			t.Errorf("template task missing %q: %s", want, body)
		}
	}

	// Saved filter and range listing.
	rec = do(h, "POST", "/api/filters", jsonCT, `{"name":"Far future","due":"this_month"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("filter = %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "POST", "/api/filters", jsonCT, `{"name":"bad","due":"someday"}`); rec.Code != 400 {
		t.Errorf("bad filter = %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/filters/1/tasks", "", ""); rec.Code != 200 || strings.Contains(rec.Body.String(), "2030-01-02") {
		t.Errorf("filter tasks = %d %s", rec.Code, rec.Body.String())
	}
	rec = do(h, "GET", "/api/tasks?view=range&from=2030-01-01&to=2030-01-31", "", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "2030-01-02") {
		t.Errorf("range = %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "GET", "/api/tasks?view=range&from=2030-01-01&to=2031-01-31", "", ""); rec.Code != 400 {
		t.Errorf("huge range = %d", rec.Code)
	}
}

func TestAttachmentsAndBackup(t *testing.T) {
	h := testHandler(t, false)
	do(h, "POST", "/api/tasks", jsonCT, `{"title":"With file"}`)

	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 32))
	payload, _ := json.Marshal(map[string]string{"filename": "../pic.png", "data": base64.StdEncoding.EncodeToString(png)})
	rec := do(h, "POST", "/api/tasks/1/attachments", jsonCT, string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload = %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "..") {
		t.Errorf("filename not cleaned: %s", rec.Body.String())
	}
	rec = do(h, "GET", "/api/attachments/1", "", "")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("serve = %d %v", rec.Code, rec.Header())
	}

	// Active content is never served inline.
	html, _ := json.Marshal(map[string]string{"filename": "x.html", "data": base64.StdEncoding.EncodeToString([]byte("<html><script>alert(1)</script></html>"))})
	do(h, "POST", "/api/tasks/1/attachments", jsonCT, string(html))
	rec = do(h, "GET", "/api/attachments/2", "", "")
	if ct := rec.Header().Get("Content-Type"); ct != "application/octet-stream" || !strings.HasPrefix(rec.Header().Get("Content-Disposition"), "attachment") {
		t.Errorf("html served as %q / %q", ct, rec.Header().Get("Content-Disposition"))
	}
	if rec := do(h, "POST", "/api/tasks/99/attachments", jsonCT, string(payload)); rec.Code != 404 && rec.Code != 400 {
		t.Errorf("upload to missing task = %d", rec.Code)
	}

	// Export contains the blob and import restores it.
	export := do(h, "GET", "/api/backup/export", "", "").Body.String()
	if !strings.Contains(export, base64.StdEncoding.EncodeToString(png)) {
		t.Fatal("export lacks attachment data")
	}
	if rec := do(h, "POST", "/api/backup/import", jsonCT, export); rec.Code != 200 && rec.Code != 204 {
		t.Fatalf("import = %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "GET", "/api/tasks/1/attachments", "", ""); !strings.Contains(rec.Body.String(), "pic.png") {
		t.Errorf("attachment lost after import: %s", rec.Body.String())
	}
	if rec := do(h, "DELETE", "/api/attachments/1", jsonCT, ""); rec.Code != 204 {
		t.Errorf("delete = %d", rec.Code)
	}
}

func TestPreferencesAndDigest(t *testing.T) {
	h := testHandler(t, false)
	if rec := do(h, "PUT", "/api/preferences/appearance", jsonCT, `{"theme":"nord","accent":"#112233"}`); rec.Code != 200 || !strings.Contains(rec.Body.String(), "nord") {
		t.Fatalf("appearance = %d %s", rec.Code, rec.Body.String())
	}
	for _, bad := range []string{`{"theme":"neon"}`, `{"theme":"dark","accent":"red;x"}`} {
		if rec := do(h, "PUT", "/api/preferences/appearance", jsonCT, bad); rec.Code != 400 {
			t.Errorf("%s = %d", bad, rec.Code)
		}
	}
	if rec := do(h, "PUT", "/api/preferences/digest", jsonCT, `{"digest_enabled":true,"digest_time":"25:00"}`); rec.Code != 400 {
		t.Errorf("bad time = %d", rec.Code)
	}
	if rec := do(h, "PUT", "/api/preferences/digest", jsonCT, `{"digest_enabled":true,"digest_time":"00:00","digest_webhook":"javascript:alert(1)"}`); rec.Code != 400 {
		t.Errorf("bad webhook = %d", rec.Code)
	}
	if rec := do(h, "PUT", "/api/preferences/digest", jsonCT, `{"digest_enabled":true,"digest_time":"00:00"}`); rec.Code != 200 {
		t.Fatalf("digest = %d %s", rec.Code, rec.Body.String())
	}

	do(h, "POST", "/api/tasks", jsonCT, `{"title":"Late","due_date":"2000-01-01"}`)
	rec := do(h, "GET", "/api/digest", "", "")
	if !strings.Contains(rec.Body.String(), `"show":true`) || !strings.Contains(rec.Body.String(), "Late") {
		t.Fatalf("digest = %s", rec.Body.String())
	}
	do(h, "POST", "/api/digest/seen", jsonCT, "")
	if rec := do(h, "GET", "/api/digest", "", ""); !strings.Contains(rec.Body.String(), `"show":false`) {
		t.Errorf("digest still shown: %s", rec.Body.String())
	}
	if rec := do(h, "POST", "/api/digest/send", jsonCT, ""); rec.Code != 400 {
		t.Errorf("send without webhook = %d", rec.Code)
	}
}
