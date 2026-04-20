package aibridge

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"cdr.dev/slog/v3"
	"github.com/coder/aibridge"
	aibrecorder "github.com/coder/aibridge/recorder"
	"go.opentelemetry.io/otel/trace"
)

// BridgeManager is a hot-swappable wrapper around aibridge.RequestBridge.
// It implements http.Handler and is safe for concurrent use.
// When Reload is called, it atomically replaces the active bridge and
// gracefully drains in-flight requests on the old one.
type BridgeManager struct {
	current  atomic.Pointer[aibridge.RequestBridge]
	ctx      context.Context
	recorder aibrecorder.Recorder
	logger   slog.Logger
	metrics  *aibridge.Metrics
	tracer   trace.Tracer
}

func NewBridgeManager(
	ctx context.Context,
	recorder aibrecorder.Recorder,
	logger slog.Logger,
	metrics *aibridge.Metrics,
	tracer trace.Tracer,
) *BridgeManager {
	return &BridgeManager{
		ctx:      ctx,
		recorder: recorder,
		logger:   logger,
		metrics:  metrics,
		tracer:   tracer,
	}
}

// ServeHTTP forwards to the active bridge, returning 503 if no providers are loaded.
func (bm *BridgeManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b := bm.current.Load()
	if b == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"no AI providers configured"}`))
		return
	}
	b.ServeHTTP(w, r)
}

// Reload builds a new bridge from the provided providers, swaps it in atomically,
// then gracefully shuts down the previous bridge with a 30-second timeout.
func (bm *BridgeManager) Reload(providers []aibridge.Provider) error {
	if len(providers) == 0 {
		old := bm.current.Swap(nil)
		if old != nil {
			go bm.shutdownBridge(old)
		}
		return nil
	}

	newBridge, err := aibridge.NewRequestBridge(bm.ctx, providers, bm.recorder, nil, bm.logger, bm.metrics, bm.tracer)
	if err != nil {
		return fmt.Errorf("build bridge: %w", err)
	}

	old := bm.current.Swap(newBridge)
	if old != nil {
		go bm.shutdownBridge(old)
	}
	return nil
}

func (bm *BridgeManager) shutdownBridge(b *aibridge.RequestBridge) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = b.Shutdown(ctx)
}

// IsReady reports whether at least one provider is currently active.
func (bm *BridgeManager) IsReady() bool {
	return bm.current.Load() != nil
}
