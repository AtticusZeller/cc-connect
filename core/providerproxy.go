package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ProviderProxy is a lightweight local reverse proxy that rewrites
// incompatible Anthropic API fields for third-party providers.
//
// It performs two rewrites:
//  1. Strips "context-management-*" entries from the anthropic-beta header,
//     since third-party providers (e.g. OpenRouter) don't support Anthropic's
//     context management feature.
//  2. If thinkingOverride is set, rewrites thinking.type "adaptive" to the
//     configured value for providers that don't support adaptive thinking
//     (e.g. SiliconFlow).
type ProviderProxy struct {
	targetURL        string
	thinkingOverride string
	listener         net.Listener
	server           *http.Server
	once             sync.Once
}

// NewProviderProxy creates and starts a local reverse proxy for the
// given upstream URL. thinkingOverride controls what thinking.type to
// rewrite "adaptive" to (e.g. "disabled" or "enabled"); pass "" to skip
// thinking rewrites.
// Returns the local URL to use as ANTHROPIC_BASE_URL.
func NewProviderProxy(targetURL, thinkingOverride string) (*ProviderProxy, string, error) {
	target, err := url.Parse(strings.TrimRight(targetURL, "/"))
	if err != nil {
		return nil, "", fmt.Errorf("providerproxy: parse target: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("providerproxy: listen: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = target.Host
	}
	proxy.FlushInterval = -1 // flush SSE events immediately

	override := thinkingOverride
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stripContextManagement(r)
		if override != "" && r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/messages") {
			rewriteThinkingInRequest(r, override)
		}
		proxy.ServeHTTP(w, r)
	})

	pp := &ProviderProxy{
		targetURL:        targetURL,
		thinkingOverride: thinkingOverride,
		listener:         listener,
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  10 * time.Minute,
			WriteTimeout: 10 * time.Minute,
		},
	}

	go func() {
		if err := pp.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("providerproxy: serve error", "error", err)
		}
	}()

	localURL := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
	slog.Info("providerproxy: started", "target", targetURL, "local", localURL, "thinking", thinkingOverride)
	return pp, localURL, nil
}

// Close shuts down the proxy.
func (pp *ProviderProxy) Close() {
	pp.once.Do(func() {
		pp.server.Close()
	})
}

// stripContextManagement removes context-management entries from the
// anthropic-beta header. Third-party providers (e.g. OpenRouter) don't
// support Anthropic's context management feature and return 400 if the
// header is present.
func stripContextManagement(r *http.Request) {
	values := r.Header.Values("anthropic-beta")
	if len(values) == 0 {
		return
	}

	var kept []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" && !strings.HasPrefix(part, "context-management") {
				kept = append(kept, part)
			}
		}
	}

	r.Header.Del("anthropic-beta")
	if len(kept) > 0 {
		r.Header.Set("anthropic-beta", strings.Join(kept, ","))
	}
}

// rewriteThinkingInRequest reads the request body and rewrites
// thinking.type "adaptive" to the given override value.
func rewriteThinkingInRequest(r *http.Request, override string) {
	if r.Body == nil || override == "" {
		return
	}
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		return
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		return
	}

	modified := false
	if thinking, ok := data["thinking"].(map[string]any); ok {
		if t, ok := thinking["type"].(string); ok && t == "adaptive" {
			thinking["type"] = override
			if override == "disabled" {
				delete(thinking, "budget_tokens")
			}
			modified = true
			slog.Debug("providerproxy: rewrote thinking adaptive →", "override", override)
		}
	}

	if !modified {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		return
	}

	newBody, err := json.Marshal(data)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))
}
