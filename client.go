package xort

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Client interface {
	// Login logs in the user with the given credentials
	Login(ctx context.Context, creds *UserCredentials) error

	// Request sends a request with a JSON object body to the server.
	Request(ctx context.Context, path string, body any) ([]byte, error)

	// RawRequest sends a request with a raw bytes body to the server.
	RawRequest(ctx context.Context, path string, body []byte) ([]byte, error)
}

type DefaultClient struct {
	cl        *http.Client
	serverURL string
	token     string
	session   Session
}

func NewClient(serverURL string) *DefaultClient {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	cl := &http.Client{
		Transport: t,
		Timeout:   30 * time.Second,
	}
	return &DefaultClient{
		cl:        cl,
		serverURL: serverURL,
	}
}

func (c *DefaultClient) Login(ctx context.Context, creds *UserCredentials) error {
	lr := &LoginRequest{
		Creds: creds,
	}
	b, err := msgpack.Marshal(lr)
	if err != nil {
		return fmt.Errorf("msgpack.Marshal: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/"+LoginPath, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %v", err)
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client.Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("json.NewDecoder.Decode: %v", err)
	}
	c.token = loginResp.Token
	c.session = loginResp.Session

	return nil
}

func (c *DefaultClient) Request(ctx context.Context, path string, body any) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}
	return c.RawRequest(ctx, path, b)
}

func (c *DefaultClient) RawRequest(ctx context.Context, path string, body []byte) ([]byte, error) {
	request := &Request{
		Token: c.token,
		Body:  body,
	}
	b, err := msgpack.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("msgpack.Marshal: %v", err)
	}

	realPath, ok := c.session[path]
	if !ok {
		return nil, fmt.Errorf("path %s not found in session", path)
	}
	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/"+realPath, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %v", err)
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}
	return b, nil
}
