package xort

import "context"

type AuthHandler interface {
	// Authenticate returns true if the  and password are valid.
	Authenticate(ctx context.Context, creds *UserCredentials) (bool, error)
}

// TODO: Implement SQL-based auth handler for common databases
