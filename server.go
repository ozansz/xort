package xort

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// LoginPath is the path to the login page.
	LoginPath = "login"

	// DefaultPort is the default port to listen on. 8027 looks like XORT.
	DefaultPort = ":8027"
)

type Server interface {
	// Handle registers a new route with a function name and an HTTP handler.
	// The name is unique and is used to identify the route when building the URL.
	// If the name is used more than once, the last one overrides the previous one.
	Handle(name string, handler http.HandlerFunc)

	// Routes returns a list of all the registered routes.
	// The list is used in login handler to build the session.
	Routes() []string

	// ServeHTTP is the HTTP handler for the server.
	// It handles the authentication and session management.
	// It also handles the routing of the request to the correct handler.
	ServeHTTP(w http.ResponseWriter, req *http.Request)

	// LoginHandler is the HTTP handler for the login page.
	// It handles the authentication and session management.
	LoginHandler() http.HandlerFunc
}

type DefaultServer struct {
	// middlewares     []http.HandlerFunc
	routes          map[string]http.HandlerFunc
	sessionRegistry SessionRegistry
	authHandler     AuthHandler
}

type ServerOption func(*DefaultServer)

func NewServer(sessionRegistry SessionRegistry, authHandler AuthHandler, opts ...ServerOption) *DefaultServer {
	s := &DefaultServer{
		routes:          make(map[string]http.HandlerFunc),
		sessionRegistry: sessionRegistry,
		authHandler:     authHandler,
	}
	return s
}

func (s *DefaultServer) Handle(name string, handler http.HandlerFunc) {
	s.routes[name] = handler
}

func (s *DefaultServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var realPath string

	inner := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		urlPath := r.URL.Path
		urlPath = strings.TrimLeft(urlPath, "/")

		if urlPath == LoginPath {
			s.LoginHandler()(w, r)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, ErrBodyRead)
			return
		}

		var request Request
		err = msgpack.Unmarshal(b, &request)
		if err != nil {
			writeErrorResponse(w, ErrMsgpackUnmarshal)
			return
		}

		session, err := s.sessionRegistry.Get(ctx, request.Token)
		if err != nil {
			writeErrorResponse(w, ErrSessionRegistryGet)
			return
		}

		if session == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path, ok := session[urlPath]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		realPath = path

		handler, ok := s.routes[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(request.Body))
		handler(w, r)
	}

	lrw := loggingResponseWriter{w, http.StatusOK}
	http.HandlerFunc(inner).ServeHTTP(lrw, r)

	statusCode := lrw.statusCode

	if realPath != "" {
		log.Printf("%s %s %s (%s) %d (%s)", r.RemoteAddr, r.Method, r.URL.Path, realPath, statusCode, http.StatusText(statusCode))
	} else {
		log.Printf("%s %s %s %d (%s)", r.RemoteAddr, r.Method, r.URL.Path, statusCode, http.StatusText(statusCode))
	}
}

func (s *DefaultServer) Routes() []string {
	routes := make([]string, 0, len(s.routes))
	for name := range s.routes {
		routes = append(routes, name)
	}
	return routes
}

func (s *DefaultServer) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		b, err := io.ReadAll(req.Body)
		if err != nil {
			writeErrorResponse(w, ErrBodyRead)
			return
		}

		var request LoginRequest
		err = msgpack.Unmarshal(b, &request)
		if err != nil {
			writeErrorResponse(w, ErrMsgpackUnmarshal)
			return
		}
		if request.Creds == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ok, err := s.authHandler.Authenticate(ctx, &UserCredentials{
			Username: request.Creds.Username,
			Password: request.Creds.Password,
		})
		if err != nil {
			writeErrorResponse(w, ErrAuthHandlerAuthenticate)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sessionID := generateLongUUID()
		session := Session{}
		for _, name := range s.Routes() {
			session[generateLongUUID()] = name
		}

		err = s.sessionRegistry.Set(ctx, sessionID, session)
		if err != nil {
			writeErrorResponse(w, ErrSessionRegistrySet)
			return
		}

		reverseSession := make(Session)
		for k, v := range session {
			reverseSession[v] = k
		}

		err = json.NewEncoder(w).Encode(LoginResponse{
			Token:   sessionID,
			Session: reverseSession,
		})
		if err != nil {
			writeErrorResponse(w, ErrJSONMarshal)
			return
		}
	}
}

func generateLongUUID() string {
	return uuid.NewString() + "-" + uuid.NewString()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
