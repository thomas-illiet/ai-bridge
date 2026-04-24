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

const defaultOpenAIBaseURL = "https://api.openai.com/v1/"

// NamedOpenAIProvider is an aibridge Provider for OpenAI-compatible APIs.
// Unlike the library's built-in provider, it supports custom names, route prefixes,
// and base URLs, enabling multiple OpenAI-compatible endpoints simultaneously.
// It always uses centralized auth so the caller's PAT is never forwarded upstream.
type NamedOpenAIProvider struct {
	baseURL string
	key     string
	name    string
}

var _ provider.Provider = (*NamedOpenAIProvider)(nil)

// NewNamedOpenAIProvider creates an OpenAI-compatible provider with a custom name.
// baseURL defaults to https://api.openai.com/v1/ if empty.
func NewNamedOpenAIProvider(name, key, baseURL string) *NamedOpenAIProvider {
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/") + "/"
	return &NamedOpenAIProvider{
		baseURL: baseURL,
		key:     key,
		name:    name,
	}
}

func (*NamedOpenAIProvider) Type() string { return "openai" }

func (p *NamedOpenAIProvider) Name() string    { return p.name }
func (p *NamedOpenAIProvider) BaseURL() string { return p.baseURL }

func (p *NamedOpenAIProvider) APIDumpDir() string                           { return "" }
func (p *NamedOpenAIProvider) CircuitBreakerConfig() *config.CircuitBreaker { return nil }
func (p *NamedOpenAIProvider) AuthHeader() string                           { return "Authorization" }

func (p *NamedOpenAIProvider) RoutePrefix() string {
	return fmt.Sprintf("/%s/v1", p.name)
}

func (p *NamedOpenAIProvider) BridgedRoutes() []string {
	return []string{"/chat/completions"}
}

func (p *NamedOpenAIProvider) PassthroughRoutes() []string {
	return []string{"/models", "/models/"}
}

func (p *NamedOpenAIProvider) InjectAuthHeader(h *http.Header) {
	if h == nil {
		return
	}
	h.Set("Authorization", "Bearer "+p.key)
}

func (p *NamedOpenAIProvider) CreateInterceptor(_ http.ResponseWriter, r *http.Request, tracer trace.Tracer) (_ intercept.Interceptor, outErr error) {
	id := uuid.New()

	_, span := tracer.Start(r.Context(), "Intercept.CreateInterceptor")
	defer tracing.EndSpanErr(span, &outErr)

	path := strings.TrimPrefix(r.URL.Path, p.RoutePrefix())
	if path != "/chat/completions" {
		span.SetStatus(codes.Error, "unknown route: "+r.URL.Path)
		return nil, provider.ErrUnknownRoute
	}

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
