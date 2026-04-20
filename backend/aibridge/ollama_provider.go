package aibridge

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/coder/aibridge/config"
	"github.com/coder/aibridge/intercept"
	"github.com/coder/aibridge/intercept/chatcompletions"
	"github.com/coder/aibridge/provider"
	"github.com/coder/aibridge/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OllamaProvider is an aibridge Provider for the Ollama local inference server.
// It reuses the OpenAI chat-completions interceptors (Ollama is OpenAI-compatible)
// but always uses centralized auth so the caller's PAT is never forwarded upstream.
type OllamaProvider struct {
	baseURL string
	key     string
	name    string
}

var _ provider.Provider = (*OllamaProvider)(nil)

// NewOllamaProvider creates an Ollama provider pointing at baseURL.
// key is optional — Ollama does not require authentication.
func NewOllamaProvider(baseURL, key string) *OllamaProvider {
	return NewNamedOllamaProvider("ollama", baseURL, key)
}

// NewNamedOllamaProvider creates an Ollama provider with a custom name and route prefix.
func NewNamedOllamaProvider(name, baseURL, key string) *OllamaProvider {
	baseURL = strings.TrimRight(baseURL, "/") + "/"
	return &OllamaProvider{
		baseURL: baseURL,
		key:     key,
		name:    name,
	}
}

func (*OllamaProvider) Type() string {
	return "ollama"
}

func (p *OllamaProvider) Name() string {
	return p.name
}

func (p *OllamaProvider) BaseURL() string                              { return p.baseURL }
func (p *OllamaProvider) APIDumpDir() string                           { return "" }
func (p *OllamaProvider) CircuitBreakerConfig() *config.CircuitBreaker { return nil }
func (p *OllamaProvider) AuthHeader() string                           { return "Authorization" }

func (p *OllamaProvider) RoutePrefix() string {
	return fmt.Sprintf("/%s/v1", p.name)
}

func (p *OllamaProvider) BridgedRoutes() []string {

	return []string{"/chat/completions"}
}

func (p *OllamaProvider) PassthroughRoutes() []string {
	return []string{"/models", "/models/"}
}

// InjectAuthHeader sets the Authorization header for upstream Ollama requests.
// Always uses the centralized key — never the caller's PAT.
func (p *OllamaProvider) InjectAuthHeader(h *http.Header) {
	if h == nil {
		return
	}
	if p.key != "" {
		h.Set("Authorization", "Bearer "+p.key)
	} else {
		h.Set("Authorization", "Bearer ollama")
	}
}

func (p *OllamaProvider) CreateInterceptor(_ http.ResponseWriter, r *http.Request, tracer trace.Tracer) (_ intercept.Interceptor, outErr error) {
	id := uuid.New()

	_, span := tracer.Start(r.Context(), "Intercept.CreateInterceptor")
	defer tracing.EndSpanErr(span, &outErr)

	path := strings.TrimPrefix(r.URL.Path, p.RoutePrefix())
	if path != "/chat/completions" {
		span.SetStatus(codes.Error, "unknown route: "+r.URL.Path)
		return nil, provider.ErrUnknownRoute
	}

	// Always centralized: never forward the caller's PAT to Ollama.
	cfg := config.OpenAI{
		Name:    p.name,
		BaseURL: p.baseURL,
		Key:     p.key,
	}
	cred := intercept.NewCredentialInfo(intercept.CredentialKindCentralized, cfg.Key)

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	var req chatcompletions.ChatCompletionNewParamsWrapper
	if err := req.UnmarshalJSON(raw); err != nil {
		return nil, fmt.Errorf("unmarshal request body: %w", err)
	}

	var interceptor intercept.Interceptor
	if req.Stream {
		interceptor = chatcompletions.NewStreamingInterceptor(id, &req, p.name, cfg, r.Header, p.AuthHeader(), tracer, cred)
	} else {
		interceptor = chatcompletions.NewBlockingInterceptor(id, &req, p.name, cfg, r.Header, p.AuthHeader(), tracer, cred)
	}

	span.SetAttributes(interceptor.TraceAttributes(r)...)
	return interceptor, nil
}
