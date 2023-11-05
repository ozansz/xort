package xort

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

type Request struct {
	Token string `json:"token" msgpack:"token"`
	Body  []byte `json:"body" msgpack:"body"`
}

type LoginRequest struct {
	Creds *UserCredentials `json:"creds" msgpack:"creds"`
}

type LoginResponse struct {
	Token   string  `json:"token" msgpack:"token"`
	Session Session `json:"session" msgpack:"session"`
}

type UserCredentials struct {
	Username string `json:"username" msgpack:"username"`
	Password string `json:"password" msgpack:"password"`
}

// Session is a mapping from session IDs to a map of unique paths to real endpoint names.
type Session map[string]string

func newJSONRequest(token string, o any) (*Request, error) {
	body, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return &Request{
		Token: token,
		Body:  body,
	}, nil
}

func (r *Request) Marshal() ([]byte, error) {
	return msgpack.Marshal(r)
}
