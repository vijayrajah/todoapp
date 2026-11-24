package model

import (
	"path/filepath"
	"strings"
)

// MaxAttachmentBytes is the size limit of one attachment.
const MaxAttachmentBytes = 10 << 20

// MaxAttachmentsPerOwner caps attachments per task or note.
const MaxAttachmentsPerOwner = 50

// Attachment describes a file stored as a blob in the database.
type Attachment struct {
	ID          int64  `json:"id"`
	TaskID      int64  `json:"task_id"`
	NoteID      int64  `json:"note_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
}

// AttachmentData is an attachment with its content (base64 in JSON), used for upload and backup.
type AttachmentData struct {
	Attachment
	Data []byte `json:"data"`
}

// CleanFilename keeps only a safe base name.
func CleanFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" {
		name = "file"
	}
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
