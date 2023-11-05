package xort

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type httpError struct {
	StatusCode int
	Message    string
}

type httpErrorResponse struct {
	Code   HTTPErrorCode `json:"code"`
	Reason string        `json:"reason"`
	Detail string        `json:"detail,omitempty"`
}

type HTTPErrorCode int

const (
	ErrNoSuchErrorCode HTTPErrorCode = iota
	ErrJSONMarshal
	ErrMsgpackUnmarshal
	ErrBodyRead
	ErrSessionRegistryGet
	ErrSessionRegistrySet
	ErrAuthHandlerAuthenticate
)

var (
	httpErrors = map[HTTPErrorCode]*httpError{
		ErrNoSuchErrorCode:         {500, "no such error code"},
		ErrJSONMarshal:             {500, "failed to encode response body to JSON"},
		ErrBodyRead:                {500, "failed to read request body"},
		ErrMsgpackUnmarshal:        {400, "failed to decode request body from MessagePack"},
		ErrSessionRegistryGet:      {500, "failed to get session info from session registry"},
		ErrSessionRegistrySet:      {500, "failed to set session info to session registry"},
		ErrAuthHandlerAuthenticate: {500, "failed to handle authentication through auth handler"},
	}
)

func writeErrorResponse(w http.ResponseWriter, code HTTPErrorCode) {
	err, ok := httpErrors[code]
	if !ok {
		code = ErrNoSuchErrorCode
		err = httpErrors[ErrNoSuchErrorCode]
	}

	w.WriteHeader(err.StatusCode)
	if err := json.NewEncoder(w).Encode(httpErrorResponse{
		Code:   code,
		Reason: err.Message,
	}); err != nil {
		// Write the internal server error (JSON marshal error) manually
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(
			`{"code":%d,"reason":"%s","detail":"%s"}`,
			ErrJSONMarshal, httpErrors[ErrJSONMarshal].Message, err.Error())))
	}
}
