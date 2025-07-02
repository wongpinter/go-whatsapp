package cloudapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp"
)

// Client handles sending messages via the WhatsApp Cloud API.
type Client struct {
	restyClient   *resty.Client
	logger        *zerolog.Logger
	phoneNumberID string
	accessToken   string
}

// NewClient creates a new CloudAPI client.
func NewClient(phoneNumberID, accessToken string, opts ...Option) *Client {
	nopLogger := zerolog.Nop()
	c := &Client{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		logger:        &nopLogger, // Default to no-op logger
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize Resty client
	c.restyClient = resty.New().
		SetBaseURL("https://graph.facebook.com/v19.0").
		SetAuthToken(c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	// Set up response handling
	c.restyClient.OnAfterResponse(c.handleResponse)

	return c
}

// Option is a functional option for configuring the CloudAPI client.
type Option func(*Client)

// WithLogger sets a custom logger for the client.
func WithLogger(logger zerolog.Logger) Option {
	return func(c *Client) {
		c.logger = &logger
	}
}

// Send dispatches any message that conforms to the Message interface.
func (c *Client) Send(ctx context.Context, msg Message) (*SendMessageResponse, error) {
	// Validate the message
	if err := msg.Validate(); err != nil {
		c.logger.Error().Err(err).Msg("Message validation failed")
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Construct the API path
	apiPath := fmt.Sprintf("/%s/messages", c.phoneNumberID)

	// Create the request payload
	payload := &SendMessageRequest{
		MessagingProduct: "whatsapp",
		To:               msg.GetTo(),
		RecipientType:    "individual",
		Type:             msg.MessageType(),
		MessageBody:      msg,
	}

	var result SendMessageResponse
	var apiError whatsapp.APIError

	c.logger.Info().
		Str("to", msg.GetTo()).
		Str("type", msg.MessageType()).
		Msg("Sending message")

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		SetError(&apiError).
		Post(apiPath)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to execute send message request")
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
		Str("message_id", result.GetMessageID()).
		Str("wa_id", result.GetWAID()).
		Msg("Message sent successfully")

	return &result, nil
}

// SendText is a convenience method for sending text messages.
func (c *Client) SendText(ctx context.Context, to, body string) (*SendMessageResponse, error) {
	msg := NewTextMessage(to, body)
	return c.Send(ctx, msg)
}

// SendTextWithPreview is a convenience method for sending text messages with URL preview.
func (c *Client) SendTextWithPreview(ctx context.Context, to, body string, previewURL bool) (*SendMessageResponse, error) {
	msg := NewTextMessage(to, body).WithPreviewURL(previewURL)
	return c.Send(ctx, msg)
}

// SendImageFromURL is a convenience method for sending image messages from a URL.
func (c *Client) SendImageFromURL(ctx context.Context, to, imageURL, caption string) (*SendMessageResponse, error) {
	msg := NewImageMessageFromURL(to, imageURL)
	if caption != "" {
		msg.WithCaption(caption)
	}
	return c.Send(ctx, msg)
}

// SendImageFromID is a convenience method for sending image messages using a media ID.
func (c *Client) SendImageFromID(ctx context.Context, to, mediaID, caption string) (*SendMessageResponse, error) {
	msg := NewImageMessageFromID(to, mediaID)
	if caption != "" {
		msg.WithCaption(caption)
	}
	return c.Send(ctx, msg)
}

// SendDocumentFromURL is a convenience method for sending document messages from a URL.
func (c *Client) SendDocumentFromURL(ctx context.Context, to, documentURL, caption, filename string) (*SendMessageResponse, error) {
	msg := NewDocumentMessageFromURL(to, documentURL)
	if caption != "" {
		msg.WithCaption(caption)
	}
	if filename != "" {
		msg.WithFilename(filename)
	}
	return c.Send(ctx, msg)
}

// SendDocumentFromID is a convenience method for sending document messages using a media ID.
func (c *Client) SendDocumentFromID(ctx context.Context, to, mediaID, caption, filename string) (*SendMessageResponse, error) {
	msg := NewDocumentMessageFromID(to, mediaID)
	if caption != "" {
		msg.WithCaption(caption)
	}
	if filename != "" {
		msg.WithFilename(filename)
	}
	return c.Send(ctx, msg)
}

// SendTemplate is a convenience method for sending template messages.
func (c *Client) SendTemplate(ctx context.Context, to, templateName, languageCode string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)
	return c.Send(ctx, msg)
}

// SendTemplateWithParams is a convenience method for sending template messages with text parameters.
func (c *Client) SendTemplateWithParams(ctx context.Context, to, templateName, languageCode string, params ...string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)
	for _, param := range params {
		msg.WithTextParameter(param)
	}
	return c.Send(ctx, msg)
}

// SendInteractiveButtons is a convenience method for sending interactive button messages.
func (c *Client) SendInteractiveButtons(ctx context.Context, to, bodyText string, buttons map[string]string) (*SendMessageResponse, error) {
	msg := NewInteractiveButtonMessage(to, bodyText)
	for id, title := range buttons {
		msg.AddButton(id, title)
	}
	return c.Send(ctx, msg)
}

// SendFlow is a convenience method for sending Flow interactive messages.
func (c *Client) SendFlow(ctx context.Context, to, bodyText, flowID, flowToken, flowCTA string) (*SendMessageResponse, error) {
	msg := NewInteractiveFlowMessage(to, bodyText).
		WithFlowID(flowID).
		WithFlowToken(flowToken).
		WithFlowCTA(flowCTA)
	return c.Send(ctx, msg)
}

// SendFlowWithScreen is a convenience method for sending Flow messages with initial screen.
func (c *Client) SendFlowWithScreen(ctx context.Context, to, bodyText, flowID, flowToken, flowCTA, screen string) (*SendMessageResponse, error) {
	msg := NewInteractiveFlowMessage(to, bodyText).
		WithFlowID(flowID).
		WithFlowToken(flowToken).
		WithFlowCTA(flowCTA).
		WithFlowScreen(screen)
	return c.Send(ctx, msg)
}

// SendFlowWithData is a convenience method for sending Flow messages with initial data.
func (c *Client) SendFlowWithData(ctx context.Context, to, bodyText, flowID, flowToken, flowCTA string, data map[string]interface{}) (*SendMessageResponse, error) {
	msg := NewInteractiveFlowMessage(to, bodyText).
		WithFlowID(flowID).
		WithFlowToken(flowToken).
		WithFlowCTA(flowCTA).
		WithFlowData(data)
	return c.Send(ctx, msg)
}

// SendInteractiveList is a convenience method for sending interactive list messages.
func (c *Client) SendInteractiveList(ctx context.Context, to, bodyText, buttonText string, rows []Row) (*SendMessageResponse, error) {
	msg := NewInteractiveListMessage(to, bodyText, buttonText)
	msg.AddSection("", rows)
	return c.Send(ctx, msg)
}

// SendLocation is a convenience method for sending location messages.
func (c *Client) SendLocation(ctx context.Context, to string, latitude, longitude float64, name, address string) (*SendMessageResponse, error) {
	msg := NewLocationMessage(to, latitude, longitude)
	if name != "" {
		msg.WithName(name)
	}
	if address != "" {
		msg.WithAddress(address)
	}
	return c.Send(ctx, msg)
}

// Extended Template Message Methods

// SendTemplateWithComponents sends a template message using the Business Management API template structure.
func (c *Client) SendTemplateWithComponents(ctx context.Context, to, templateName, languageCode string, components ...TemplateComponent) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)

	// Add components if provided
	if len(components) > 0 {
		msg.Template.Components = components
	}

	return c.Send(ctx, msg)
}

// SendTemplateWithAdvancedParams sends a template message with text parameters.
func (c *Client) SendTemplateWithAdvancedParams(ctx context.Context, to, templateName, languageCode string, params ...string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)

	// Add text parameters to body component
	if len(params) > 0 {
		var templateParams []TemplateParameter
		for _, param := range params {
			templateParams = append(templateParams, TemplateParameter{
				Type: "text",
				Text: param,
			})
		}

		msg.Template.Components = append(msg.Template.Components, TemplateComponent{
			Type:       "body",
			Parameters: templateParams,
		})
	}

	return c.Send(ctx, msg)
}

// SendTemplateWithHeaderImage sends a template message with a header image.
func (c *Client) SendTemplateWithHeaderImage(ctx context.Context, to, templateName, languageCode, imageURL string, bodyParams ...string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)

	// Add header image component
	msg.Template.Components = append(msg.Template.Components, TemplateComponent{
		Type: "header",
		Parameters: []TemplateParameter{
			{
				Type: "image",
				Image: &Media{
					Link: imageURL,
				},
			},
		},
	})

	// Add body parameters if provided
	if len(bodyParams) > 0 {
		var templateParams []TemplateParameter
		for _, param := range bodyParams {
			templateParams = append(templateParams, TemplateParameter{
				Type: "text",
				Text: param,
			})
		}

		msg.Template.Components = append(msg.Template.Components, TemplateComponent{
			Type:       "body",
			Parameters: templateParams,
		})
	}

	return c.Send(ctx, msg)
}

// SendTemplateWithCurrency sends a template message with currency parameters.
func (c *Client) SendTemplateWithCurrency(ctx context.Context, to, templateName, languageCode string, amount1000 int, currencyCode, fallbackValue string, additionalParams ...string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)

	// Build parameters array starting with currency
	var templateParams []TemplateParameter
	templateParams = append(templateParams, TemplateParameter{
		Type: "currency",
		Currency: &Currency{
			Amount1000:    amount1000,
			Code:          currencyCode,
			FallbackValue: fallbackValue,
		},
	})

	// Add additional text parameters
	for _, param := range additionalParams {
		templateParams = append(templateParams, TemplateParameter{
			Type: "text",
			Text: param,
		})
	}

	msg.Template.Components = append(msg.Template.Components, TemplateComponent{
		Type:       "body",
		Parameters: templateParams,
	})

	return c.Send(ctx, msg)
}

// SendTemplateWithDateTime sends a template message with date/time parameters.
func (c *Client) SendTemplateWithDateTime(ctx context.Context, to, templateName, languageCode, fallbackValue string, additionalParams ...string) (*SendMessageResponse, error) {
	msg := NewTemplateMessage(to, templateName, languageCode)

	// Build parameters array starting with date/time
	var templateParams []TemplateParameter
	templateParams = append(templateParams, TemplateParameter{
		Type: "date_time",
		DateTime: &DateTime{
			FallbackValue: fallbackValue,
		},
	})

	// Add additional text parameters
	for _, param := range additionalParams {
		templateParams = append(templateParams, TemplateParameter{
			Type: "text",
			Text: param,
		})
	}

	msg.Template.Components = append(msg.Template.Components, TemplateComponent{
		Type:       "body",
		Parameters: templateParams,
	})

	return c.Send(ctx, msg)
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

// GetPhoneNumberID returns the phone number ID used by this client.
func (c *Client) GetPhoneNumberID() string {
	return c.phoneNumberID
}

// SetLogger updates the client's logger.
func (c *Client) SetLogger(logger zerolog.Logger) {
	c.logger = &logger
}

// Health checks if the client can communicate with the WhatsApp API.
func (c *Client) Health(ctx context.Context) error {
	// Simple health check by trying to get phone number info
	resp, err := c.restyClient.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/%s", c.phoneNumberID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("response", resp.String()).
			Msg("Health check returned error")
		return fmt.Errorf("health check failed with status %d", resp.StatusCode())
	}

	c.logger.Info().Msg("CloudAPI health check successful")
	return nil
}
