package services

import (
	"context"
	"fmt"
	"regexp"

	"cdr.dev/slog/v3"
	aibmcp "github.com/coder/aibridge/mcp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

type CreateMCPServerRequest struct {
	Name         string            `json:"name"`
	DisplayName  string            `json:"display_name"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	AllowPattern string            `json:"allow_pattern"`
	DenyPattern  string            `json:"deny_pattern"`
	Enabled      bool              `json:"enabled"`
}

type UpdateMCPServerRequest struct {
	DisplayName  *string           `json:"display_name"`
	URL          *string           `json:"url"`
	Headers      map[string]string `json:"headers"`
	AllowPattern *string           `json:"allow_pattern"`
	DenyPattern  *string           `json:"deny_pattern"`
	Enabled      *bool             `json:"enabled"`
}

func ListMCPServers(search string) ([]models.MCPServer, error) {
	var servers []models.MCPServer
	q := database.DB.Order("created_at desc")
	if search != "" {
		q = q.Where("name ILIKE ? OR display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func GetMCPServer(id uuid.UUID) (*models.MCPServer, error) {
	var s models.MCPServer
	if err := database.DB.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateMCPServer(req CreateMCPServerRequest) (*models.MCPServer, error) {
	if err := validateProviderName(req.Name); err != nil {
		return nil, err
	}
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
	if req.AllowPattern != "" {
		if _, err := regexp.Compile(req.AllowPattern); err != nil {
			return nil, fmt.Errorf("invalid allow_pattern: %w", err)
		}
	}
	if req.DenyPattern != "" {
		if _, err := regexp.Compile(req.DenyPattern); err != nil {
			return nil, fmt.Errorf("invalid deny_pattern: %w", err)
		}
	}
	s := &models.MCPServer{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		URL:          req.URL,
		Headers:      models.MCPHeaders(req.Headers),
		AllowPattern: req.AllowPattern,
		DenyPattern:  req.DenyPattern,
		Enabled:      req.Enabled,
	}
	if err := database.DB.Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func UpdateMCPServer(id uuid.UUID, req UpdateMCPServerRequest) (*models.MCPServer, error) {
	var s models.MCPServer
	if err := database.DB.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.URL != nil {
		if *req.URL == "" {
			return nil, fmt.Errorf("url cannot be empty")
		}
		updates["url"] = *req.URL
	}
	if req.Headers != nil {
		updates["headers"] = models.MCPHeaders(req.Headers)
	}
	if req.AllowPattern != nil {
		if *req.AllowPattern != "" {
			if _, err := regexp.Compile(*req.AllowPattern); err != nil {
				return nil, fmt.Errorf("invalid allow_pattern: %w", err)
			}
		}
		updates["allow_pattern"] = *req.AllowPattern
	}
	if req.DenyPattern != nil {
		if *req.DenyPattern != "" {
			if _, err := regexp.Compile(*req.DenyPattern); err != nil {
				return nil, fmt.Errorf("invalid deny_pattern: %w", err)
			}
		}
		updates["deny_pattern"] = *req.DenyPattern
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := database.DB.Model(&s).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func DeleteMCPServer(id uuid.UUID) error {
	result := database.DB.Delete(&models.MCPServer{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BuildMCPProxy loads all enabled MCP servers from the DB and constructs a ServerProxier.
// Returns (nil, nil) when no servers are enabled — the bridge accepts a nil proxy.
func BuildMCPProxy(ctx context.Context, logger slog.Logger, tracer trace.Tracer) (aibmcp.ServerProxier, error) {
	var servers []models.MCPServer
	if err := database.DB.Where("enabled = true").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("load mcp servers: %w", err)
	}
	if len(servers) == 0 {
		return nil, nil
	}

	proxiers := make(map[string]aibmcp.ServerProxier, len(servers))
	for _, s := range servers {
		var allowlist, denylist *regexp.Regexp
		var err error
		if s.AllowPattern != "" {
			if allowlist, err = regexp.Compile(s.AllowPattern); err != nil {
				return nil, fmt.Errorf("mcp server %q: invalid allow_pattern: %w", s.Name, err)
			}
		}
		if s.DenyPattern != "" {
			if denylist, err = regexp.Compile(s.DenyPattern); err != nil {
				return nil, fmt.Errorf("mcp server %q: invalid deny_pattern: %w", s.Name, err)
			}
		}

		var headers map[string]string
		if len(s.Headers) > 0 {
			headers = map[string]string(s.Headers)
		}

		proxy, err := aibmcp.NewStreamableHTTPServerProxy(
			s.Name,
			s.URL,
			headers,
			allowlist,
			denylist,
			logger.Named("mcp."+s.Name),
			tracer,
		)
		if err != nil {
			return nil, fmt.Errorf("create mcp proxy %q: %w", s.Name, err)
		}
		proxiers[s.Name] = proxy
	}

	return aibmcp.NewServerProxyManager(proxiers, tracer), nil
}
