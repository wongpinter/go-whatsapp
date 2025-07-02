package flows

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// AssetUploadResponse represents the response from uploading a Flow asset.
type AssetUploadResponse struct {
	Success bool `json:"success"`
}

// UploadFlowJSON uploads Flow JSON to a Flow.
func (c *Client) UploadFlowJSON(ctx context.Context, flowID string, flowJSON string) error {
	c.logger.Info().
		Str("flow_id", flowID).
		Int("json_size", len(flowJSON)).
		Msg("Uploading Flow JSON")

	// Validate Flow JSON before uploading
	if _, err := FromJSON(flowJSON); err != nil {
		c.logger.Error().Err(err).Msg("Invalid Flow JSON")
		return fmt.Errorf("invalid Flow JSON: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", "flow.json")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader([]byte(flowJSON))); err != nil {
		return fmt.Errorf("failed to copy Flow JSON: %w", err)
	}

	// Add asset_type field
	if err := writer.WriteField("asset_type", string(FlowAssetTypeFlowJSON)); err != nil {
		return fmt.Errorf("failed to write asset_type field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	endpoint := fmt.Sprintf("https://graph.facebook.com/%s/%s/assets", c.apiVersion, flowID)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to upload Flow JSON")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("flow_id", flowID).
			Msg("API error uploading Flow JSON")
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow JSON uploaded successfully")
	return nil
}

// UploadFlowFromBuilder uploads a Flow JSON built using the FlowBuilder.
func (c *Client) UploadFlowFromBuilder(ctx context.Context, flowID string, flow *Flow) error {
	// Validate the Flow
	if err := flow.Validate(); err != nil {
		return fmt.Errorf("flow validation failed: %w", err)
	}

	// Convert to JSON
	flowJSON, err := flow.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert flow to JSON: %w", err)
	}

	// Upload the JSON
	return c.UploadFlowJSON(ctx, flowID, flowJSON)
}

// CreateAndUploadFlow creates a new Flow and uploads the Flow JSON in one operation.
func (c *Client) CreateAndUploadFlow(ctx context.Context, metadata *CreateFlowRequest, flow *Flow) (string, error) {
	c.logger.Info().
		Str("name", metadata.Name).
		Strs("categories", metadata.Categories).
		Msg("Creating and uploading Flow")

	// Create the Flow
	createResp, err := c.CreateFlow(ctx, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to create Flow: %w", err)
	}

	flowID := createResp.ID

	// Upload the Flow JSON
	if err := c.UploadFlowFromBuilder(ctx, flowID, flow); err != nil {
		// If upload fails, try to clean up by deleting the created Flow
		if deleteErr := c.DeleteFlow(ctx, flowID); deleteErr != nil {
			c.logger.Error().
				Err(deleteErr).
				Str("flow_id", flowID).
				Msg("Failed to cleanup Flow after upload failure")
		}
		return "", fmt.Errorf("failed to upload Flow JSON: %w", err)
	}

	c.logger.Info().
		Str("flow_id", flowID).
		Str("name", metadata.Name).
		Msg("Flow created and uploaded successfully")

	return flowID, nil
}

// ValidateAndUploadFlow validates a Flow and uploads it if valid.
func (c *Client) ValidateAndUploadFlow(ctx context.Context, flowID string, flow *Flow) error {
	// Perform comprehensive validation
	validator := NewFlowValidator()
	errors := validator.ValidateFlow(flow)

	if len(errors) > 0 {
		c.logger.Error().
			Int("error_count", len(errors)).
			Str("flow_id", flowID).
			Msg("Flow validation failed")

		// Log validation errors
		for _, err := range errors {
			c.logger.Error().
				Str("error", err.Message).
				Int("line", err.Line).
				Msg("Validation error")
		}

		return fmt.Errorf("flow validation failed with %d errors", len(errors))
	}

	// Upload the validated Flow
	return c.UploadFlowFromBuilder(ctx, flowID, flow)
}

// GetFlowAssets retrieves information about Flow assets.
func (c *Client) GetFlowAssets(ctx context.Context, flowID string) (map[string]interface{}, error) {
	c.logger.Info().Str("flow_id", flowID).Msg("Getting Flow assets")

	var result map[string]interface{}
	var apiError APIError

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		Get(fmt.Sprintf("/%s/assets", flowID))

	if err != nil {
		c.logger.Error().Err(err).Str("flow_id", flowID).Msg("Failed to get Flow assets")
		return nil, fmt.Errorf("failed to get Flow assets: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().Interface("error", apiError).Str("flow_id", flowID).Msg("API error getting Flow assets")
		return nil, &apiError
	}

	c.logger.Info().Str("flow_id", flowID).Msg("Flow assets retrieved successfully")
	return result, nil
}

// FlowWorkflow provides a high-level workflow for Flow management.
type FlowWorkflow struct {
	client *Client
}

// NewFlowWorkflow creates a new Flow workflow helper.
func NewFlowWorkflow(client *Client) *FlowWorkflow {
	return &FlowWorkflow{
		client: client,
	}
}

// CreateCompleteFlow creates a Flow, uploads JSON, and optionally publishes it.
func (w *FlowWorkflow) CreateCompleteFlow(ctx context.Context, metadata *CreateFlowRequest, flow *Flow, publish bool) (string, error) {
	// Create and upload the Flow
	flowID, err := w.client.CreateAndUploadFlow(ctx, metadata, flow)
	if err != nil {
		return "", err
	}

	// Publish if requested
	if publish {
		if err := w.client.PublishFlow(ctx, flowID); err != nil {
			w.client.logger.Error().
				Err(err).
				Str("flow_id", flowID).
				Msg("Failed to publish Flow after creation")
			return flowID, fmt.Errorf("failed to publish Flow: %w", err)
		}
		w.client.logger.Info().Str("flow_id", flowID).Msg("Flow published successfully")
	}

	return flowID, nil
}

// UpdateFlowJSON updates an existing Flow's JSON.
func (w *FlowWorkflow) UpdateFlowJSON(ctx context.Context, flowID string, flow *Flow) error {
	// Get current Flow info to check status
	flowInfo, err := w.client.GetFlow(ctx, flowID, "status")
	if err != nil {
		return fmt.Errorf("failed to get Flow info: %w", err)
	}

	// Check if Flow can be updated
	if flowInfo.Status == FlowStatusPublished {
		return fmt.Errorf("cannot update published Flow: %s", flowID)
	}

	// Upload the updated JSON
	return w.client.ValidateAndUploadFlow(ctx, flowID, flow)
}

// CloneFlow creates a copy of an existing Flow with a new name.
func (w *FlowWorkflow) CloneFlow(ctx context.Context, sourceFlowID, newName string, newCategories []string) (string, error) {
	// Get the source Flow info
	sourceFlow, err := w.client.GetFlow(ctx, sourceFlowID)
	if err != nil {
		return "", fmt.Errorf("failed to get source Flow: %w", err)
	}

	// Create metadata for the new Flow
	metadata := &CreateFlowRequest{
		Name:        newName,
		Categories:  newCategories,
		EndpointURI: sourceFlow.EndpointURI,
	}

	// Note: In a real implementation, you would need to get the Flow JSON
	// from the source Flow and use it to create the new Flow.
	// This would require additional API calls to retrieve the Flow JSON.

	// Create the new Flow
	createResp, err := w.client.CreateFlow(ctx, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to create cloned Flow: %w", err)
	}

	w.client.logger.Info().
		Str("source_flow_id", sourceFlowID).
		Str("new_flow_id", createResp.ID).
		Str("new_name", newName).
		Msg("Flow cloned successfully")

	return createResp.ID, nil
}
