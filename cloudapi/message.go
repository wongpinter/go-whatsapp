package cloudapi

import (
	"encoding/json"
	"fmt"
)

// Message represents any object that can be sent as a WhatsApp message.
type Message interface {
	// MessageType returns the string identifier for the message type (e.g., "text", "template").
	MessageType() string
	// GetTo returns the recipient phone number.
	GetTo() string
	// Validate checks if the message is valid for sending.
	Validate() error
}

// MessageBase contains fields common to all message types.
type MessageBase struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	RecipientType    string `json:"recipient_type"`
}

// GetTo returns the recipient phone number.
func (m *MessageBase) GetTo() string {
	return m.To
}

// Validate performs basic validation on the message base.
func (m *MessageBase) Validate() error {
	if m.To == "" {
		return fmt.Errorf("recipient phone number is required")
	}
	if m.MessagingProduct == "" {
		m.MessagingProduct = "whatsapp" // Set default
	}
	if m.RecipientType == "" {
		m.RecipientType = "individual" // Set default
	}
	return nil
}

// TextMessage represents a plain text message.
type TextMessage struct {
	MessageBase
	Text struct {
		Body       string `json:"body"`
		PreviewURL bool   `json:"preview_url,omitempty"`
	} `json:"text"`
}

// MessageType returns the message type identifier.
func (t *TextMessage) MessageType() string {
	return "text"
}

// Validate checks if the text message is valid.
func (t *TextMessage) Validate() error {
	if err := t.MessageBase.Validate(); err != nil {
		return err
	}
	if t.Text.Body == "" {
		return fmt.Errorf("text message body cannot be empty")
	}
	if len(t.Text.Body) > 4096 {
		return fmt.Errorf("text message body cannot exceed 4096 characters")
	}
	return nil
}

// NewTextMessage creates a new text message.
func NewTextMessage(to, body string) *TextMessage {
	msg := &TextMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Text.Body = body
	return msg
}

// WithPreviewURL enables URL preview for the text message.
func (t *TextMessage) WithPreviewURL(enable bool) *TextMessage {
	t.Text.PreviewURL = enable
	return t
}

// ImageMessage represents an image message.
type ImageMessage struct {
	MessageBase
	Image struct {
		ID      string `json:"id,omitempty"`      // Media ID from upload
		Link    string `json:"link,omitempty"`    // Public URL
		Caption string `json:"caption,omitempty"` // Optional caption
	} `json:"image"`
}

// MessageType returns the message type identifier.
func (i *ImageMessage) MessageType() string {
	return "image"
}

// Validate checks if the image message is valid.
func (i *ImageMessage) Validate() error {
	if err := i.MessageBase.Validate(); err != nil {
		return err
	}
	if i.Image.ID == "" && i.Image.Link == "" {
		return fmt.Errorf("image message must have either media ID or link")
	}
	if i.Image.ID != "" && i.Image.Link != "" {
		return fmt.Errorf("image message cannot have both media ID and link")
	}
	if len(i.Image.Caption) > 1024 {
		return fmt.Errorf("image caption cannot exceed 1024 characters")
	}
	return nil
}

// NewImageMessageFromID creates a new image message using a media ID.
func NewImageMessageFromID(to, mediaID string) *ImageMessage {
	msg := &ImageMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Image.ID = mediaID
	return msg
}

// NewImageMessageFromURL creates a new image message using a public URL.
func NewImageMessageFromURL(to, url string) *ImageMessage {
	msg := &ImageMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Image.Link = url
	return msg
}

// WithCaption adds a caption to the image message.
func (i *ImageMessage) WithCaption(caption string) *ImageMessage {
	i.Image.Caption = caption
	return i
}

// DocumentMessage represents a document message.
type DocumentMessage struct {
	MessageBase
	Document struct {
		ID       string `json:"id,omitempty"`       // Media ID from upload
		Link     string `json:"link,omitempty"`     // Public URL
		Caption  string `json:"caption,omitempty"`  // Optional caption
		Filename string `json:"filename,omitempty"` // Optional filename
	} `json:"document"`
}

// MessageType returns the message type identifier.
func (d *DocumentMessage) MessageType() string {
	return "document"
}

// Validate checks if the document message is valid.
func (d *DocumentMessage) Validate() error {
	if err := d.MessageBase.Validate(); err != nil {
		return err
	}
	if d.Document.ID == "" && d.Document.Link == "" {
		return fmt.Errorf("document message must have either media ID or link")
	}
	if d.Document.ID != "" && d.Document.Link != "" {
		return fmt.Errorf("document message cannot have both media ID and link")
	}
	if len(d.Document.Caption) > 1024 {
		return fmt.Errorf("document caption cannot exceed 1024 characters")
	}
	return nil
}

// NewDocumentMessageFromID creates a new document message using a media ID.
func NewDocumentMessageFromID(to, mediaID string) *DocumentMessage {
	msg := &DocumentMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Document.ID = mediaID
	return msg
}

// NewDocumentMessageFromURL creates a new document message using a public URL.
func NewDocumentMessageFromURL(to, url string) *DocumentMessage {
	msg := &DocumentMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Document.Link = url
	return msg
}

// WithCaption adds a caption to the document message.
func (d *DocumentMessage) WithCaption(caption string) *DocumentMessage {
	d.Document.Caption = caption
	return d
}

// WithFilename sets the filename for the document message.
func (d *DocumentMessage) WithFilename(filename string) *DocumentMessage {
	d.Document.Filename = filename
	return d
}

// SendMessageRequest represents the request payload for sending a message.
type SendMessageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	RecipientType    string      `json:"recipient_type"`
	Type             string      `json:"type"`
	MessageBody      interface{} `json:"-"` // Will be marshaled under the type key
}

// MarshalJSON implements custom JSON marshaling for the send message request.
func (r *SendMessageRequest) MarshalJSON() ([]byte, error) {
	// Create a map to hold the final JSON structure
	result := map[string]interface{}{
		"messaging_product": r.MessagingProduct,
		"to":                r.To,
		"recipient_type":    r.RecipientType,
		"type":              r.Type,
	}

	// Add the message body under the type key
	result[r.Type] = r.MessageBody

	return json.Marshal(result)
}

// SendMessageResponse represents the response from sending a message.
type SendMessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WAID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// GetMessageID returns the WhatsApp message ID from the response.
func (r *SendMessageResponse) GetMessageID() string {
	if len(r.Messages) > 0 {
		return r.Messages[0].ID
	}
	return ""
}

// GetWAID returns the WhatsApp ID of the recipient.
func (r *SendMessageResponse) GetWAID() string {
	if len(r.Contacts) > 0 {
		return r.Contacts[0].WAID
	}
	return ""
}

// TemplateMessage represents a template message.
type TemplateMessage struct {
	MessageBase
	Template struct {
		Name       string              `json:"name"`
		Language   TemplateLanguage    `json:"language"`
		Components []TemplateComponent `json:"components,omitempty"`
	} `json:"template"`
}

// MessageType returns the message type identifier.
func (t *TemplateMessage) MessageType() string {
	return "template"
}

// Validate checks if the template message is valid.
func (t *TemplateMessage) Validate() error {
	if err := t.MessageBase.Validate(); err != nil {
		return err
	}
	if t.Template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if t.Template.Language.Code == "" {
		return fmt.Errorf("template language code is required")
	}
	return nil
}

// TemplateLanguage represents a template language.
type TemplateLanguage struct {
	Code string `json:"code"`
}

// TemplateComponent represents a component in a template.
type TemplateComponent struct {
	Type       string              `json:"type"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateParameter represents a parameter in a template component.
type TemplateParameter struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	Currency *Currency `json:"currency,omitempty"`
	DateTime *DateTime `json:"date_time,omitempty"`
	Image    *Media    `json:"image,omitempty"`
	Document *Media    `json:"document,omitempty"`
	Video    *Media    `json:"video,omitempty"`
}

// Currency represents a currency parameter.
type Currency struct {
	FallbackValue string `json:"fallback_value"`
	Code          string `json:"code"`
	Amount1000    int    `json:"amount_1000"`
}

// DateTime represents a date/time parameter.
type DateTime struct {
	FallbackValue string `json:"fallback_value"`
}

// Media represents a media parameter.
type Media struct {
	Link string `json:"link,omitempty"`
}

// NewTemplateMessage creates a new template message.
func NewTemplateMessage(to, name, languageCode string) *TemplateMessage {
	msg := &TemplateMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Template.Name = name
	msg.Template.Language.Code = languageCode
	return msg
}

// WithTextParameter adds a text parameter to the template.
func (t *TemplateMessage) WithTextParameter(text string) *TemplateMessage {
	if len(t.Template.Components) == 0 {
		t.Template.Components = append(t.Template.Components, TemplateComponent{
			Type: "body",
		})
	}

	// Add to the first (body) component
	t.Template.Components[0].Parameters = append(t.Template.Components[0].Parameters, TemplateParameter{
		Type: "text",
		Text: text,
	})
	return t
}

// WithCurrencyParameter adds a currency parameter to the template.
func (t *TemplateMessage) WithCurrencyParameter(amount1000 int, code, fallbackValue string) *TemplateMessage {
	if len(t.Template.Components) == 0 {
		t.Template.Components = append(t.Template.Components, TemplateComponent{
			Type: "body",
		})
	}

	t.Template.Components[0].Parameters = append(t.Template.Components[0].Parameters, TemplateParameter{
		Type: "currency",
		Currency: &Currency{
			Amount1000:    amount1000,
			Code:          code,
			FallbackValue: fallbackValue,
		},
	})
	return t
}

// InteractiveMessage represents an interactive message with buttons or lists.
type InteractiveMessage struct {
	MessageBase
	Interactive struct {
		Type   string      `json:"type"`
		Header *Header     `json:"header,omitempty"`
		Body   Body        `json:"body"`
		Footer *Footer     `json:"footer,omitempty"`
		Action interface{} `json:"action"`
	} `json:"interactive"`
}

// MessageType returns the message type identifier.
func (i *InteractiveMessage) MessageType() string {
	return "interactive"
}

// Validate checks if the interactive message is valid.
func (i *InteractiveMessage) Validate() error {
	if err := i.MessageBase.Validate(); err != nil {
		return err
	}
	if i.Interactive.Type == "" {
		return fmt.Errorf("interactive type is required")
	}
	if i.Interactive.Body.Text == "" {
		return fmt.Errorf("interactive body text is required")
	}
	return nil
}

// Header represents a header in an interactive message.
type Header struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Document *Media `json:"document,omitempty"`
	Image    *Media `json:"image,omitempty"`
	Video    *Media `json:"video,omitempty"`
}

// Body represents a body in an interactive message.
type Body struct {
	Text string `json:"text"`
}

// Footer represents a footer in an interactive message.
type Footer struct {
	Text string `json:"text"`
}

// Button represents a button in an interactive message.
type Button struct {
	Type  string `json:"type"`
	Reply Reply  `json:"reply"`
}

// Reply represents a reply in a button.
type Reply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ButtonAction represents the action for button interactive messages.
type ButtonAction struct {
	Buttons []Button `json:"buttons"`
}

// ListAction represents the action for list interactive messages.
type ListAction struct {
	Button   string    `json:"button"`
	Sections []Section `json:"sections"`
}

// Section represents a section in a list.
type Section struct {
	Title string `json:"title,omitempty"`
	Rows  []Row  `json:"rows"`
}

// Row represents a row in a list section.
type Row struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// NewInteractiveButtonMessage creates a new interactive message with buttons.
func NewInteractiveButtonMessage(to, bodyText string) *InteractiveMessage {
	msg := &InteractiveMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Interactive.Type = "button"
	msg.Interactive.Body.Text = bodyText
	msg.Interactive.Action = &ButtonAction{
		Buttons: make([]Button, 0),
	}
	return msg
}

// NewInteractiveListMessage creates a new interactive message with a list.
func NewInteractiveListMessage(to, bodyText, buttonText string) *InteractiveMessage {
	msg := &InteractiveMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Interactive.Type = "list"
	msg.Interactive.Body.Text = bodyText
	msg.Interactive.Action = &ListAction{
		Button:   buttonText,
		Sections: make([]Section, 0),
	}
	return msg
}

// WithHeader adds a text header to the interactive message.
func (i *InteractiveMessage) WithHeader(text string) *InteractiveMessage {
	i.Interactive.Header = &Header{
		Type: "text",
		Text: text,
	}
	return i
}

// WithFooter adds a footer to the interactive message.
func (i *InteractiveMessage) WithFooter(text string) *InteractiveMessage {
	i.Interactive.Footer = &Footer{
		Text: text,
	}
	return i
}

// AddButton adds a button to the interactive message (for button type).
func (i *InteractiveMessage) AddButton(id, title string) *InteractiveMessage {
	if action, ok := i.Interactive.Action.(*ButtonAction); ok {
		action.Buttons = append(action.Buttons, Button{
			Type: "reply",
			Reply: Reply{
				ID:    id,
				Title: title,
			},
		})
	}
	return i
}

// AddSection adds a section to the interactive message (for list type).
func (i *InteractiveMessage) AddSection(title string, rows []Row) *InteractiveMessage {
	if action, ok := i.Interactive.Action.(*ListAction); ok {
		action.Sections = append(action.Sections, Section{
			Title: title,
			Rows:  rows,
		})
	}
	return i
}

// AddRow adds a row to the last section of the interactive message (for list type).
func (i *InteractiveMessage) AddRow(id, title, description string) *InteractiveMessage {
	if action, ok := i.Interactive.Action.(*ListAction); ok {
		if len(action.Sections) == 0 {
			action.Sections = append(action.Sections, Section{
				Rows: make([]Row, 0),
			})
		}
		lastSection := &action.Sections[len(action.Sections)-1]
		lastSection.Rows = append(lastSection.Rows, Row{
			ID:          id,
			Title:       title,
			Description: description,
		})
	}
	return i
}

// LocationMessage represents a location message.
type LocationMessage struct {
	MessageBase
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name,omitempty"`
		Address   string  `json:"address,omitempty"`
	} `json:"location"`
}

// MessageType returns the message type identifier.
func (l *LocationMessage) MessageType() string {
	return "location"
}

// Validate checks if the location message is valid.
func (l *LocationMessage) Validate() error {
	if err := l.MessageBase.Validate(); err != nil {
		return err
	}
	if l.Location.Latitude == 0 && l.Location.Longitude == 0 {
		return fmt.Errorf("location coordinates are required")
	}
	return nil
}

// NewLocationMessage creates a new location message.
func NewLocationMessage(to string, latitude, longitude float64) *LocationMessage {
	msg := &LocationMessage{}
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	msg.To = to
	msg.Location.Latitude = latitude
	msg.Location.Longitude = longitude
	return msg
}

// WithName adds a name to the location message.
func (l *LocationMessage) WithName(name string) *LocationMessage {
	l.Location.Name = name
	return l
}

// WithAddress adds an address to the location message.
func (l *LocationMessage) WithAddress(address string) *LocationMessage {
	l.Location.Address = address
	return l
}
