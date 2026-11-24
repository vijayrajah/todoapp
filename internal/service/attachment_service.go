package service

import (
	"context"
	"mime"
	"net/http"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// AttachmentService validates uploads before they are stored as blobs in the database.
type AttachmentService struct {
	store *store.Store
}

// NewAttachmentService returns an AttachmentService.
func NewAttachmentService(st *store.Store) *AttachmentService {
	return &AttachmentService{store: st}
}

// Add stores a file on a task or note (kind "task" or "note").
func (s *AttachmentService) Add(ctx context.Context, kind string, ownerID int64, filename string, data []byte) (model.Attachment, error) {
	if len(data) == 0 {
		return model.Attachment{}, model.ValidationError("file is empty")
	}
	if len(data) > model.MaxAttachmentBytes {
		return model.Attachment{}, model.ValidationError("file is too large (max 10 MB)")
	}
	// The type is sniffed from the content; the client-supplied one is never trusted.
	contentType, _, err := mime.ParseMediaType(http.DetectContentType(data))
	if err != nil {
		contentType = "application/octet-stream"
	}
	return s.store.AddAttachment(ctx, kind, ownerID, model.CleanFilename(filename), contentType, data)
}
