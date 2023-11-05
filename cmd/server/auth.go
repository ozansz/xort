package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.sazak.io/xort"
)

type fileAuthHandler struct {
	creds []xort.UserCredentials
}

func newFileAuthHandler(path string) (*fileAuthHandler, error) {
	var creds []xort.UserCredentials
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %v", err)
	}
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %v", err)
	}
	return &fileAuthHandler{creds: creds}, nil
}

// fileAuthHandler implements the AuthHandler interface.
func (h *fileAuthHandler) Authenticate(ctx context.Context, creds *xort.UserCredentials) (bool, error) {
	for _, c := range h.creds {
		if c.Username == creds.Username && c.Password == creds.Password {
			return true, nil
		}
	}
	return false, nil
}
