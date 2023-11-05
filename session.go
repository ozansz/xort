package xort

import (
	"context"
	"log"
	"time"
)

// TODO: Implement SQL-based session registry for common databases and
// 		 a KV-based session registry for Redis

type SessionRegistry interface {
	// Get returns the session for the given ID.
	// If the session does not exist, it returns nil.
	Get(ctx context.Context, id string) (Session, error)

	// Set sets the session for the given ID.
	// If the session already exists, it is overwritten.
	Set(ctx context.Context, id string, session Session) error

	// Delete deletes the session for the given ID.
	// If the session does not exist, it returns nil.
	Delete(ctx context.Context, id string) error
}

const (
	inMemorySessionValidityDuration = 5 * time.Minute
)

type InMemorySessionRegistry struct {
	sessions map[string]Session
}

func NewInMemorySessionRegistry() *InMemorySessionRegistry {
	return &InMemorySessionRegistry{
		sessions: make(map[string]Session),
	}
}

func (r *InMemorySessionRegistry) trackSessionTimeout(ctx context.Context, id string) {
	go func() {
		select {
		case <-time.After(inMemorySessionValidityDuration):
			log.Printf("Invalidating session %s after %.2f seconds", id, inMemorySessionValidityDuration.Seconds())
			r.Delete(ctx, id)
		case <-ctx.Done():
			return
		}
	}()
}

func (r *InMemorySessionRegistry) Get(ctx context.Context, id string) (Session, error) {
	return r.sessions[id], nil
}

func (r *InMemorySessionRegistry) Set(ctx context.Context, id string, session Session) error {
	r.sessions[id] = session
	r.trackSessionTimeout(ctx, id)
	log.Printf("Session %s has been set", id)
	return nil
}

func (r *InMemorySessionRegistry) Delete(ctx context.Context, id string) error {
	delete(r.sessions, id)
	log.Printf("Session %s has been deleted", id)
	return nil
}
