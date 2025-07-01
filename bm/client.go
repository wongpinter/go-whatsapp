package bm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/wongpinter/go-whatsapp"
)

// Client handles interactions with the WhatsApp Business Management API.
type Client struct {
	restyClient *resty.Client
	logger      *zerolog.Logger
	accessToken string
	wabaID      string
	apiVersion  string
}

// NewClient creates a new Business Management API client.
func NewClient(accessToken string, opts ...Option) *Client {
	nopLogger := zerolog.Nop()
	c := &Client{
		accessToken: accessToken,
		apiVersion:  "v19.0",    // Default API version
		logger:      &nopLogger, // Default to no-op logger
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

// Option is a functional option for configuring the Business Management client.
type Option func(*Client)

// WithLogger sets a custom logger for the client.
func WithLogger(logger zerolog.Logger) Option {
	return func(c *Client) {
		c.logger = &logger
	}
}

// WithWABAID sets the WhatsApp Business Account ID.
func WithWABAID(wabaID string) Option {
	return func(c *Client) {
		c.wabaID = wabaID
	}
}

// WithAPIVersion sets a custom API version.
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}

// GetBusinessAccount retrieves comprehensive details of a WhatsApp Business Account.
func (c *Client) GetBusinessAccount(ctx context.Context, businessAccountID string) (*BusinessAccount, error) {
	if businessAccountID == "" && c.wabaID != "" {
		businessAccountID = c.wabaID
	}

	if businessAccountID == "" {
		return nil, fmt.Errorf("business account ID is required")
	}

	var result BusinessAccount
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("waba_id", businessAccountID).
		Msg("Retrieving business account details")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		SetQueryParam("fields", "id,name,timezone_id,message_templates,phone_numbers").
		Get(fmt.Sprintf("/%s", businessAccountID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get business account")
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Interface("api_error", apiError).
			Int("status_code", resp.StatusCode()).
			Msg("WhatsApp API returned an error")
		return nil, &apiError
	}

	c.logger.Info().
		Str("account_name", result.Name).
		Str("account_id", result.ID).
		Msg("Business account retrieved successfully")

	return &result, nil
}

// GetPhoneNumbers retrieves all phone numbers associated with a WhatsApp Business Account.
func (c *Client) GetPhoneNumbers(ctx context.Context, businessAccountID string) ([]PhoneNumber, error) {
	if businessAccountID == "" && c.wabaID != "" {
		businessAccountID = c.wabaID
	}

	if businessAccountID == "" {
		return nil, fmt.Errorf("business account ID is required")
	}

	var result struct {
		Data []PhoneNumber `json:"data"`
	}
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("waba_id", businessAccountID).
		Msg("Retrieving phone numbers")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		SetQueryParam("fields", "id,display_phone_number,verified_name,quality_rating,status").
		Get(fmt.Sprintf("/%s/phone_numbers", businessAccountID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get phone numbers")
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Interface("api_error", apiError).
			Int("status_code", resp.StatusCode()).
			Msg("WhatsApp API returned an error")
		return nil, &apiError
	}

	c.logger.Info().
		Int("count", len(result.Data)).
		Msg("Phone numbers retrieved successfully")

	return result.Data, nil
}

// GetBusinessProfile retrieves the business profile for a phone number.
func (c *Client) GetBusinessProfile(ctx context.Context, phoneNumberID string) (*BusinessProfile, error) {
	if phoneNumberID == "" {
		return nil, fmt.Errorf("phone number ID is required")
	}

	var result struct {
		Data []BusinessProfile `json:"data"`
	}
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("phone_number_id", phoneNumberID).
		Msg("Retrieving business profile")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		Get(fmt.Sprintf("/%s/whatsapp_business_profile", phoneNumberID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get business profile")
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Interface("api_error", apiError).
			Int("status_code", resp.StatusCode()).
			Msg("WhatsApp API returned an error")
		return nil, &apiError
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no business profile found for phone number ID: %s", phoneNumberID)
	}

	c.logger.Info().
		Str("about", result.Data[0].About).
		Msg("Business profile retrieved successfully")

	return &result.Data[0], nil
}

// handleResponse processes HTTP responses and converts API errors.
func (c *Client) handleResponse(client *resty.Client, resp *resty.Response) error {
	if resp.IsError() {
		var apiError whatsapp.APIError
		if err := json.Unmarshal(resp.Body(), &apiError); err != nil {
			// If we can't parse the error response, create a generic error
			return whatsapp.NewAPIError(resp.StatusCode(),
				fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status()),
				"HTTPError", "")
		}
		return &apiError
	}
	return nil
}

// GetWABAID returns the configured WhatsApp Business Account ID.
func (c *Client) GetWABAID() string {
	return c.wabaID
}

// SetLogger updates the client's logger.
func (c *Client) SetLogger(logger zerolog.Logger) {
	c.logger = &logger
}

// GetMessageMetrics retrieves message metrics for a phone number.
func (c *Client) GetMessageMetrics(ctx context.Context, phoneNumberID string, start, end time.Time, granularity string) ([]MessageMetrics, error) {
	var result struct {
		Data []MessageMetrics `json:"data"`
	}
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("phone_number_id", phoneNumberID).
		Str("start", start.Format(time.RFC3339)).
		Str("end", end.Format(time.RFC3339)).
		Str("granularity", granularity).
		Msg("Retrieving message metrics")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&apiError).
		SetQueryParam("start", start.Format(time.RFC3339)).
		SetQueryParam("end", end.Format(time.RFC3339)).
		SetQueryParam("granularity", granularity).
		Get(fmt.Sprintf("/%s/analytics", phoneNumberID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get message metrics")
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Interface("api_error", apiError).
			Int("status_code", resp.StatusCode()).
			Msg("WhatsApp API returned an error")
		return nil, &apiError
	}

	c.logger.Info().
		Int("count", len(result.Data)).
		Msg("Message metrics retrieved successfully")

	return result.Data, nil
}

// UpdateBusinessProfile updates the business profile for a phone number.
func (c *Client) UpdateBusinessProfile(ctx context.Context, phoneNumberID string, profile *BusinessProfileUpdateRequest) error {
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("phone_number_id", phoneNumberID).
		Msg("Updating business profile")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(profile).
		SetError(&apiError).
		Post(fmt.Sprintf("/%s/whatsapp_business_profile", phoneNumberID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to update business profile")
		return fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Interface("api_error", apiError).
			Int("status_code", resp.StatusCode()).
			Msg("WhatsApp API returned an error")
		return &apiError
	}

	c.logger.Info().Msg("Business profile updated successfully")
	return nil
}

// Health checks if the client can communicate with the WhatsApp Business Management API.
func (c *Client) Health(ctx context.Context) error {
	if c.wabaID == "" {
		return fmt.Errorf("WABA ID is required for health check")
	}

	// Simple health check by trying to get business account info
	_, err := c.GetBusinessAccount(ctx, c.wabaID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Business Management API health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}

	c.logger.Info().Msg("Business Management API health check successful")
	return nil
}
