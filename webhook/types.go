package webhook

import (
	"time"
)

// WebhookPayload represents the top-level webhook payload from WhatsApp.
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

// Entry represents an entry in the webhook payload.
type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

// Change represents a change notification.
type Change struct {
	Value Value  `json:"value"`
	Field string `json:"field"`
}

// Value contains the actual webhook data.
type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Contacts         []Contact `json:"contacts,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
	Statuses         []Status  `json:"statuses,omitempty"`
	Errors           []Error   `json:"errors,omitempty"`
}

// Metadata contains metadata about the webhook.
type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// Contact represents a contact in the webhook.
type Contact struct {
	Profile Profile `json:"profile"`
	WAID    string  `json:"wa_id"`
}

// Profile represents a user's profile information.
type Profile struct {
	Name string `json:"name"`
}

// Message represents an incoming message.
type Message struct {
	From        string           `json:"from"`
	ID          string           `json:"id"`
	Timestamp   string           `json:"timestamp"`
	Type        string           `json:"type"`
	Context     *Context         `json:"context,omitempty"`
	Text        *Text            `json:"text,omitempty"`
	Image       *Media           `json:"image,omitempty"`
	Audio       *Media           `json:"audio,omitempty"`
	Video       *Media           `json:"video,omitempty"`
	Document    *Document        `json:"document,omitempty"`
	Voice       *Media           `json:"voice,omitempty"`
	Sticker     *Media           `json:"sticker,omitempty"`
	Location    *Location        `json:"location,omitempty"`
	Contacts    []ContactMessage `json:"contacts,omitempty"`
	Interactive *Interactive     `json:"interactive,omitempty"`
	Button      *Button          `json:"button,omitempty"`
	Order       *Order           `json:"order,omitempty"`
	System      *System          `json:"system,omitempty"`
	Reaction    *Reaction        `json:"reaction,omitempty"`
	Errors      []Error          `json:"errors,omitempty"`
}

// GetTimestamp returns the message timestamp as a time.Time.
func (m *Message) GetTimestamp() time.Time {
	if timestamp, err := time.Parse("1136239445", m.Timestamp); err == nil {
		return timestamp
	}
	return time.Time{}
}

// Context represents message context (for replies).
type Context struct {
	From string `json:"from"`
	ID   string `json:"id"`
}

// Text represents a text message.
type Text struct {
	Body string `json:"body"`
}

// Media represents media content (image, audio, video, voice, sticker).
type Media struct {
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
}

// Document represents a document message.
type Document struct {
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
}

// Location represents a location message.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

// ContactMessage represents a contact message.
type ContactMessage struct {
	Addresses []Address `json:"addresses,omitempty"`
	Birthday  string    `json:"birthday,omitempty"`
	Emails    []Email   `json:"emails,omitempty"`
	Name      Name      `json:"name"`
	Org       Org       `json:"org,omitempty"`
	Phones    []Phone   `json:"phones,omitempty"`
	URLs      []URL     `json:"urls,omitempty"`
}

// Address represents a contact address.
type Address struct {
	Street      string `json:"street,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Zip         string `json:"zip,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Type        string `json:"type,omitempty"`
}

// Email represents a contact email.
type Email struct {
	Email string `json:"email,omitempty"`
	Type  string `json:"type,omitempty"`
}

// Name represents a contact name.
type Name struct {
	FormattedName string `json:"formatted_name"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	MiddleName    string `json:"middle_name,omitempty"`
	Suffix        string `json:"suffix,omitempty"`
	Prefix        string `json:"prefix,omitempty"`
}

// Org represents a contact organization.
type Org struct {
	Company    string `json:"company,omitempty"`
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`
}

// Phone represents a contact phone number.
type Phone struct {
	Phone string `json:"phone,omitempty"`
	WAID  string `json:"wa_id,omitempty"`
	Type  string `json:"type,omitempty"`
}

// URL represents a contact URL.
type URL struct {
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

// Interactive represents an interactive message.
type Interactive struct {
	Type        string       `json:"type"`
	ButtonReply *ButtonReply `json:"button_reply,omitempty"`
	ListReply   *ListReply   `json:"list_reply,omitempty"`
}

// ButtonReply represents a button reply.
type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListReply represents a list reply.
type ListReply struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// Button represents a button message.
type Button struct {
	Payload string `json:"payload"`
	Text    string `json:"text"`
}

// Order represents an order message.
type Order struct {
	CatalogID    string        `json:"catalog_id"`
	ProductItems []ProductItem `json:"product_items"`
	Text         string        `json:"text,omitempty"`
}

// ProductItem represents a product item in an order.
type ProductItem struct {
	ProductRetailerID string `json:"product_retailer_id"`
	Quantity          int    `json:"quantity"`
	ItemPrice         string `json:"item_price"`
	Currency          string `json:"currency"`
}

// System represents a system message.
type System struct {
	Body     string `json:"body"`
	Identity string `json:"identity"`
	WAID     string `json:"wa_id"`
	Type     string `json:"type"`
	Customer string `json:"customer"`
}

// Status represents a message status update.
type Status struct {
	ID           string        `json:"id"`
	Status       string        `json:"status"`
	Timestamp    string        `json:"timestamp"`
	RecipientID  string        `json:"recipient_id"`
	Conversation *Conversation `json:"conversation,omitempty"`
	Pricing      *Pricing      `json:"pricing,omitempty"`
	Errors       []Error       `json:"errors,omitempty"`
}

// GetTimestamp returns the status timestamp as a time.Time.
func (s *Status) GetTimestamp() time.Time {
	if timestamp, err := time.Parse("1136239445", s.Timestamp); err == nil {
		return timestamp
	}
	return time.Time{}
}

// Conversation represents conversation information in status updates.
type Conversation struct {
	ID                  string  `json:"id"`
	ExpirationTimestamp string  `json:"expiration_timestamp,omitempty"`
	Origin              *Origin `json:"origin"`
}

// Origin represents the conversation origin.
type Origin struct {
	Type string `json:"type"`
}

// Pricing represents pricing information.
type Pricing struct {
	Billable     bool   `json:"billable"`
	PricingModel string `json:"pricing_model"`
	Category     string `json:"category"`
}

// Error represents an error in the webhook.
type Error struct {
	Code      int    `json:"code"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	ErrorData struct {
		Details string `json:"details"`
	} `json:"error_data"`
}

// EventType represents the type of webhook event.
type EventType string

const (
	EventTypeTextMessage     EventType = "message.text"
	EventTypeImageMessage    EventType = "message.image"
	EventTypeAudioMessage    EventType = "message.audio"
	EventTypeVideoMessage    EventType = "message.video"
	EventTypeDocumentMessage EventType = "message.document"
	EventTypeVoiceMessage    EventType = "message.voice"
	EventTypeStickerMessage  EventType = "message.sticker"
	EventTypeLocationMessage EventType = "message.location"
	EventTypeContactMessage  EventType = "message.contact"
	EventTypeButtonReply     EventType = "interactive.button_reply"
	EventTypeListReply       EventType = "interactive.list_reply"
	EventTypeOrderMessage    EventType = "message.order"
	EventTypeSystemMessage   EventType = "message.system"
	EventTypeStatusUpdate    EventType = "status.update"
	EventTypeError           EventType = "error"
)

// GetEventType determines the event type from a message.
func (m *Message) GetEventType() EventType {
	switch m.Type {
	case "text":
		return EventTypeTextMessage
	case "image":
		return EventTypeImageMessage
	case "audio":
		return EventTypeAudioMessage
	case "video":
		return EventTypeVideoMessage
	case "document":
		return EventTypeDocumentMessage
	case "voice":
		return EventTypeVoiceMessage
	case "sticker":
		return EventTypeStickerMessage
	case "location":
		return EventTypeLocationMessage
	case "contacts":
		return EventTypeContactMessage
	case "interactive":
		if m.Interactive != nil {
			switch m.Interactive.Type {
			case "button_reply":
				return EventTypeButtonReply
			case "list_reply":
				return EventTypeListReply
			}
		}
	case "order":
		return EventTypeOrderMessage
	case "system":
		return EventTypeSystemMessage
	case "reaction":
		return EventTypeReactionMessage
	case "unknown":
		return EventTypeUnknownMessage
	}
	return EventType(m.Type)
}

// Additional message types from the technical reference
const (
	EventTypeReactionMessage    EventType = "message.reaction"
	EventTypeUnknownMessage     EventType = "message.unknown"
	EventTypeUnsupportedMessage EventType = "message.unsupported"
)

// Reaction represents a reaction to a message
type Reaction struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

// UnsupportedMessage represents an unsupported message type
type UnsupportedMessage struct {
	Errors []Error `json:"errors"`
}
