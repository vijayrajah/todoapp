package service

import (
	"context"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// NoteService adds rendered markdown to notes.
type NoteService struct {
	store *store.Store
}

// NewNoteService returns a NoteService.
func NewNoteService(st *store.Store) *NoteService {
	return &NoteService{store: st}
}

// List returns notes, optionally filtered by tag.
func (s *NoteService) List(ctx context.Context, tag string) ([]model.Note, error) {
	notes, err := s.store.ListNotes(ctx, tag)
	if err != nil {
		return nil, err
	}
	for i := range notes {
		notes[i].BodyHTML = RenderMarkdown(notes[i].Body)
	}
	return notes, nil
}

// Get returns one note.
func (s *NoteService) Get(ctx context.Context, id int64) (model.Note, error) {
	n, err := s.store.GetNote(ctx, id)
	if err != nil {
		return model.Note{}, err
	}
	n.BodyHTML = RenderMarkdown(n.Body)
	return n, nil
}

// Create validates and stores a note.
func (s *NoteService) Create(ctx context.Context, n model.Note) (model.Note, error) {
	if err := n.Clean(); err != nil {
		return model.Note{}, err
	}
	id, err := s.store.CreateNote(ctx, &n)
	if err != nil {
		return model.Note{}, err
	}
	return s.Get(ctx, id)
}

// Update validates and saves a note.
func (s *NoteService) Update(ctx context.Context, n model.Note) (model.Note, error) {
	if err := n.Clean(); err != nil {
		return model.Note{}, err
	}
	if err := s.store.UpdateNote(ctx, &n); err != nil {
		return model.Note{}, err
	}
	return s.Get(ctx, n.ID)
}
