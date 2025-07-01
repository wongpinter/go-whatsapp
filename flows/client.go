package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// Client provides Flow management operations via the WhatsApp Graph API.
type Client struct {
	wabaID      string
	accessToken string
	apiVersion  string
	logger      *zerolog.Logger
	restyClient *resty.Client
}

// Option configures the Flow client.
type Option func(*Client)

// WithLogger sets a custom logger for the client.
func WithLogger(logger *zerolog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithAPIVersion sets a custom API version.
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}

// NewClient creates a new Flow management client.
func NewClient(wabaID, accessToken string, opts ...Option) *Client {
	nopLogger := zerolog.Nop()
	c := &Client{
		wabaID:      wabaID,
		accessToken: accessToken,
		apiVersion:  "v18.0", // Default API version for Flows
		logger:      &nopLogger,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize Resty client
	c.restyClient = resty.New().
		SetBaseURL(fmt.Sprintf("https://graph.facebook.com/%s", c.apiVersion)).
		SetAuthToken(c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	// Set up response handling
	c.restyClient.OnAfterResponse(c.handleResponse)

	return c
}

// CreateFlowRequest represents a request to create a new Flow.
type CreateFlowRequest struct {
	Name        string   `json:"name"`
	Categories  []string `json:"categories"`
	EndpointURI string   `json:"endpoint_uri,omitempty"`
}

// CreateFlowResponse represents the response from creating a Flow.
type CreateFlowResponse struct {
	ID string `json:"id"`
}

// UpdateFlowRequest represents a request to update Flow metadata.
type UpdateFlowRequest struct {
	Name        string   `json:"name,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	EndpointURI string   `json:"endpoint_uri,omitempty"`
}

// ListFlowsResponse represents the response from listing Flows.
type ListFlowsResponse struct {
	Data   []FlowInfo `json:"data"`
	Paging *Paging    `json:"paging,omitempty"`
}

// Paging represents pagination information.
type Paging struct {
	Cursors  *Cursors `json:"cursors,omitempty"`
	Next     string   `json:"next,omitempty"`
	Previous string   `json:"previous,omitempty"`
}

// Cursors represents pagination cursors.
type Cursors struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// PreviewResponse represents the response from getting a Flow preview URL.
type PreviewResponse struct {
	PreviewURL string `json:"preview_url"`
	ExpiresAt  string `json:"expires_at"`
}

// CreateFlow creates a new Flow.
func (c *Client) CreateFlow(ctx context.Context, req *CreateFlowRequest) (*CreateFlowResponse, error) {
	c.logger.Info().
		Str("name", req.Name).
		Strs("categories", req.Categories).
		Msg("Creating Flow")

	var result CreateFlowResponse
	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		SetError(&apiError).
		Post(fmt.Sprintf("/%s/flows", c.wabaID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create Flow")
		return nil, fmt.Errorf("failed to create Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Msg("API error creating Flow")
		return nil, &apiError
	}

	c.logger.Info().Str("flow_id", result.ID).Msg("Flow created successfully")
	return &result, nil
}

// GetFlow retrieves Flow information.
func (c *Client) GetFlow(ctx context.Context, flowID string, fields ...string) (*FlowInfo, error) {
	c.logger.Info().Str("flow_id", flowID).Msg("Getting Flow")

	var result FlowInfo
	var apiError APIError

	req := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError)

	if len(fields) > 0 {
		req.SetQueryParam("fields", joinFields(fields))
	}

	resp, err := req.Get(fmt.Sprintf("/%s", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to get Flow")
		return nil, fmt.Errorf("failed to get Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error getting Flow")
		return nil, &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow retrieved successfully")
	return &result, nil
}

// UpdateFlow updates Flow metadata.
func (c *Client) UpdateFlow(ctx context.Context, flowID string, req *UpdateFlowRequest) error {
	c.logger.Info().
		Str("flow_id", flowID).
		Str("name", req.Name).
		Msg("Updating Flow")

	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(req).
		SetError(&apiError).
		Post(fmt.Sprintf("/%s", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to update Flow")
		return fmt.Errorf("failed to update Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error updating Flow")
		return &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow updated successfully")
	return nil
}

// ListFlows lists all Flows for the WABA.
func (c *Client) ListFlows(ctx context.Context) (*ListFlowsResponse, error) {
	c.logger.Info().Msg("Listing Flows")

	var result ListFlowsResponse
	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		Get(fmt.Sprintf("/%s/flows", c.wabaID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to list Flows")
		return nil, fmt.Errorf("failed to list Flows: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Msg("API error listing Flows")
		return nil, &apiError
	}

	c.logger.Info().Int("count", len(result.Data)).Msg("Flows listed successfully")
	return &result, nil
}

// DeleteFlow deletes a Flow (only works for DRAFT status).
func (c *Client) DeleteFlow(ctx context.Context, flowID string) error {
	c.logger.Info().Str("flow_id", flowID).Msg("Deleting Flow")

	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetError(&apiError).
		Delete(fmt.Sprintf("/%s", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to delete Flow")
		return fmt.Errorf("failed to delete Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error deleting Flow")
		return &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow deleted successfully")
	return nil
}

// GetPreviewURL gets a preview URL for the Flow.
func (c *Client) GetPreviewURL(ctx context.Context, flowID string, invalidate bool) (*PreviewResponse, error) {
	c.logger.Info().
		Str("flow_id", flowID).
		Bool("invalidate", invalidate).
		Msg("Getting Flow preview URL")

	var result PreviewResponse
	var apiError APIError

	fields := fmt.Sprintf("preview.invalidate(%t)", invalidate)

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetQueryParam("fields", fields).
		SetResult(&result).
		SetError(&apiError).
		Get(fmt.Sprintf("/%s", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to get Flow preview URL")
		return nil, fmt.Errorf("failed to get Flow preview URL: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error getting Flow preview URL")
		return nil, &apiError
	}

	c.logger.Info().
		Str("flow_id", flowID).
		Str("preview_url", result.PreviewURL).
		Msg("Flow preview URL retrieved successfully")
	return &result, nil
}

// APIError represents an error response from the Graph API.
type APIError struct {
	Err struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"code"`
		ErrorSubcode int    `json:"error_subcode"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %s (code: %d)", e.Err.Message, e.Err.Code)
}

// handleResponse handles HTTP responses and logs them.
func (c *Client) handleResponse(client *resty.Client, resp *resty.Response) error {
	c.logger.Debug().
		Str("method", resp.Request.Method).
		Str("url", resp.Request.URL).
		Int("status", resp.StatusCode()).
		Dur("duration", resp.Time()).
		Msg("API request completed")

	return nil
}

// PublishFlow publishes a Flow (irreversible operation).
func (c *Client) PublishFlow(ctx context.Context, flowID string) error {
	c.logger.Info().Str("flow_id", flowID).Msg("Publishing Flow")

	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetError(&apiError).
		Post(fmt.Sprintf("/%s/publish", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to publish Flow")
		return fmt.Errorf("failed to publish Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error publishing Flow")
		return &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow published successfully")
	return nil
}

// DeprecateFlow deprecates a published Flow.
func (c *Client) DeprecateFlow(ctx context.Context, flowID string) error {
	c.logger.Info().Str("flow_id", flowID).Msg("Deprecating Flow")

	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetError(&apiError).
		Post(fmt.Sprintf("/%s/deprecate", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to deprecate Flow")
		return fmt.Errorf("failed to deprecate Flow: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error deprecating Flow")
		return &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow deprecated successfully")
	return nil
}

// Helper function to join fields for API requests.
func joinFields(fields []string) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ","
		}
		result += field
	}
	return result
}
