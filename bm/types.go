package bm

import (
	"fmt"
	"time"
)

// BusinessAccount represents a WhatsApp Business Account.
type BusinessAccount struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	TimezoneID       string `json:"timezone_id"`
	MessageTemplates struct {
		Data []MessageTemplate `json:"data"`
	} `json:"message_templates"`
	PhoneNumbers struct {
		Data []PhoneNumber `json:"data"`
	} `json:"phone_numbers"`
}

// PhoneNumber represents a phone number associated with a WhatsApp Business Account.
type PhoneNumber struct {
	ID                     string `json:"id"`
	DisplayPhoneNumber     string `json:"display_phone_number"`
	VerifiedName           string `json:"verified_name"`
	QualityRating          string `json:"quality_rating"`
	Status                 string `json:"status"`
	CodeVerificationStatus string `json:"code_verification_status,omitempty"`
	Certificate            string `json:"certificate,omitempty"`
}

// PhoneNumberStatus represents the possible statuses of a phone number.
type PhoneNumberStatus string

const (
	StatusConnected    PhoneNumberStatus = "CONNECTED"
	StatusDisconnected PhoneNumberStatus = "DISCONNECTED"
	StatusUnverified   PhoneNumberStatus = "UNVERIFIED"
	StatusPending      PhoneNumberStatus = "PENDING"
	StatusFlagged      PhoneNumberStatus = "FLAGGED"
	StatusRestricted   PhoneNumberStatus = "RESTRICTED"
)

// QualityRating represents the quality rating of a phone number.
type QualityRating string

const (
	QualityGreen   QualityRating = "GREEN"
	QualityYellow  QualityRating = "YELLOW"
	QualityRed     QualityRating = "RED"
	QualityUnknown QualityRating = "UNKNOWN"
)

// BusinessProfile represents the business profile information.
type BusinessProfile struct {
	MessagingProduct  string   `json:"messaging_product"`
	About             string   `json:"about"`
	Address           string   `json:"address"`
	Description       string   `json:"description"`
	Email             string   `json:"email"`
	ProfilePictureURL string   `json:"profile_picture_url"`
	Websites          []string `json:"websites"`
	Vertical          string   `json:"vertical"`
}

// MessageTemplate represents a message template.
type MessageTemplate struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	Category   string              `json:"category"`
	Language   string              `json:"language"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// TemplateStatus represents the possible statuses of a message template.
type TemplateStatus string

const (
	TemplateStatusApproved TemplateStatus = "APPROVED"
	TemplateStatusPending  TemplateStatus = "PENDING"
	TemplateStatusRejected TemplateStatus = "REJECTED"
	TemplateStatusDisabled TemplateStatus = "DISABLED"
)

// TemplateCategory represents the category of a message template.
type TemplateCategory string

const (
	CategoryMarketing      TemplateCategory = "MARKETING"
	CategoryUtility        TemplateCategory = "UTILITY"
	CategoryAuthentication TemplateCategory = "AUTHENTICATION"
)

// TemplateComponent represents a component of a message template.
type TemplateComponent struct {
	Type       string              `json:"type"`
	Format     string              `json:"format,omitempty"`
	Text       string              `json:"text,omitempty"`
	Example    *TemplateExample    `json:"example,omitempty"`
	Buttons    []TemplateButton    `json:"buttons,omitempty"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateExample represents an example for a template component.
type TemplateExample struct {
	HeaderText   []string   `json:"header_text,omitempty"`
	BodyText     [][]string `json:"body_text,omitempty"`
	HeaderHandle []string   `json:"header_handle,omitempty"`
}

// TemplateButton represents a button in a template.
type TemplateButton struct {
	Type        string   `json:"type"`
	Text        string   `json:"text"`
	URL         string   `json:"url,omitempty"`
	PhoneNumber string   `json:"phone_number,omitempty"`
	Example     []string `json:"example,omitempty"`
}

// TemplateParameter represents a parameter in a template.
type TemplateParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CreateTemplateRequest represents a request to create a new message template.
type CreateTemplateRequest struct {
	Name                string              `json:"name"`
	Language            string              `json:"language"`
	Category            TemplateCategory    `json:"category"`
	AllowCategoryChange bool                `json:"allow_category_change,omitempty"`
	Components          []TemplateComponent `json:"components"`
}

// UpdateTemplateRequest represents a request to update an existing template.
type UpdateTemplateRequest struct {
	Category   *TemplateCategory   `json:"category,omitempty"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// CreateTemplateResponse represents the response from creating a template.
type CreateTemplateResponse struct {
	ID       string           `json:"id"`
	Status   TemplateStatus   `json:"status"`
	Category TemplateCategory `json:"category"`
}

// ListTemplatesResponse represents the response from listing templates.
type ListTemplatesResponse struct {
	Data   []MessageTemplate `json:"data"`
	Paging *PagingInfo       `json:"paging,omitempty"`
}

// PagingInfo represents pagination information.
type PagingInfo struct {
	Cursors *struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	} `json:"cursors,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// DeleteTemplateResponse represents the response from deleting a template.
type DeleteTemplateResponse struct {
	Success bool `json:"success"`
}

// TemplateComponentType represents the type of a template component.
type TemplateComponentType string

const (
	ComponentTypeHeader  TemplateComponentType = "HEADER"
	ComponentTypeBody    TemplateComponentType = "BODY"
	ComponentTypeFooter  TemplateComponentType = "FOOTER"
	ComponentTypeButtons TemplateComponentType = "BUTTONS"
)

// TemplateFormat represents the format of a template component.
type TemplateFormat string

const (
	FormatText     TemplateFormat = "TEXT"
	FormatImage    TemplateFormat = "IMAGE"
	FormatVideo    TemplateFormat = "VIDEO"
	FormatDocument TemplateFormat = "DOCUMENT"
	FormatLocation TemplateFormat = "LOCATION"
)

// TemplateButtonType represents the type of a template button.
type TemplateButtonType string

const (
	ButtonTypeQuickReply  TemplateButtonType = "QUICK_REPLY"
	ButtonTypeURL         TemplateButtonType = "URL"
	ButtonTypePhoneNumber TemplateButtonType = "PHONE_NUMBER"
)

// TemplateParameterType represents the type of a template parameter.
type TemplateParameterType string

const (
	ParameterTypeText     TemplateParameterType = "text"
	ParameterTypeCurrency TemplateParameterType = "currency"
	ParameterTypeDateTime TemplateParameterType = "date_time"
	ParameterTypeImage    TemplateParameterType = "image"
	ParameterTypeDocument TemplateParameterType = "document"
	ParameterTypeVideo    TemplateParameterType = "video"
)

// TemplateValidationError represents a validation error for a template.
type TemplateValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface.
func (e *TemplateValidationError) Error() string {
	return fmt.Sprintf("template validation error in field '%s': %s (code: %s)", e.Field, e.Message, e.Code)
}

// TemplateValidationResult represents the result of template validation.
type TemplateValidationResult struct {
	Valid  bool                      `json:"valid"`
	Errors []TemplateValidationError `json:"errors,omitempty"`
}

// SendTemplateRequest represents a request to send a template message.
type SendTemplateRequest struct {
	To       string                 `json:"to"`
	Template TemplateMessagePayload `json:"template"`
}

// TemplateMessagePayload represents the template payload for sending messages.
type TemplateMessagePayload struct {
	Name       string                   `json:"name"`
	Language   TemplateLanguage         `json:"language"`
	Components []TemplateComponentParam `json:"components,omitempty"`
}

// TemplateLanguage represents the language of a template.
type TemplateLanguage struct {
	Code string `json:"code"`
}

// TemplateComponentParam represents a component parameter when sending templates.
type TemplateComponentParam struct {
	Type       TemplateComponentType `json:"type"`
	Parameters []TemplateParameter   `json:"parameters,omitempty"`
	SubType    string                `json:"sub_type,omitempty"`
	Index      int                   `json:"index,omitempty"`
}

// TemplateListParams represents parameters for listing templates.
type TemplateListParams struct {
	Fields   string `json:"fields,omitempty"`
	Status   string `json:"status,omitempty"`
	Category string `json:"category,omitempty"`
	Language string `json:"language,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	After    string `json:"after,omitempty"`
	Before   string `json:"before,omitempty"`
}

// TemplateListOption is a functional option for listing templates.
type TemplateListOption func(*TemplateListParams)

// WithTemplateFields sets the fields to retrieve for templates.
func WithTemplateFields(fields ...string) TemplateListOption {
	return func(p *TemplateListParams) {
		fieldsStr := ""
		for i, field := range fields {
			if i > 0 {
				fieldsStr += ","
			}
			fieldsStr += field
		}
		p.Fields = fieldsStr
	}
}

// WithTemplateStatus filters templates by status.
func WithTemplateStatus(status TemplateStatus) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Status = string(status)
	}
}

// WithTemplateCategory filters templates by category.
func WithTemplateCategory(category TemplateCategory) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Category = string(category)
	}
}

// WithTemplateLanguage filters templates by language.
func WithTemplateLanguage(language string) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Language = language
	}
}

// WithTemplateLimit sets the maximum number of templates to retrieve.
func WithTemplateLimit(limit int) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Limit = limit
	}
}

// WithTemplatePagination sets pagination cursors.
func WithTemplatePagination(after, before string) TemplateListOption {
	return func(p *TemplateListParams) {
		p.After = after
		p.Before = before
	}
}

// ConversationAnalytics represents conversation analytics data.
type ConversationAnalytics struct {
	Data   []DataPoint `json:"data"`
	Paging *Paging     `json:"paging,omitempty"`
}

// DataPoint represents a single data point in analytics.
type DataPoint struct {
	ConversationType      string  `json:"conversation_type"`
	ConversationDirection string  `json:"conversation_direction"`
	Country               string  `json:"country"`
	PhoneNumber           string  `json:"phone_number"`
	Cost                  float64 `json:"cost"`
	ConversationCount     int     `json:"conversation_count"`
	DataPoint             string  `json:"data_point"`
	StartTime             string  `json:"start"`
	EndTime               string  `json:"end"`
}

// GetStartTime returns the start time as a time.Time.
func (d *DataPoint) GetStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, d.StartTime)
}

// GetEndTime returns the end time as a time.Time.
func (d *DataPoint) GetEndTime() (time.Time, error) {
	return time.Parse(time.RFC3339, d.EndTime)
}

// Paging represents pagination information.
type Paging struct {
	Cursors  *Cursors `json:"cursors,omitempty"`
	Next     string   `json:"next,omitempty"`
	Previous string   `json:"previous,omitempty"`
}

// Cursors represents pagination cursors.
type Cursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// ConversationType represents the type of conversation.
type ConversationType string

const (
	ConversationTypeUserInitiated      ConversationType = "USER_INITIATED"
	ConversationTypeBusinessInitiated  ConversationType = "BUSINESS_INITIATED"
	ConversationTypeReferralConversion ConversationType = "REFERRAL_CONVERSION"
)

// ConversationDirection represents the direction of conversation.
type ConversationDirection string

const (
	DirectionInbound  ConversationDirection = "INBOUND"
	DirectionOutbound ConversationDirection = "OUTBOUND"
)

// MediaUploadResponse represents the response from media upload.
type MediaUploadResponse struct {
	ID string `json:"id"`
}

// MediaInfo represents information about uploaded media.
type MediaInfo struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	FileSize int64  `json:"file_size"`
}

// BusinessProfileUpdateRequest represents a request to update business profile.
type BusinessProfileUpdateRequest struct {
	MessagingProduct  string   `json:"messaging_product"`
	About             string   `json:"about,omitempty"`
	Address           string   `json:"address,omitempty"`
	Description       string   `json:"description,omitempty"`
	Email             string   `json:"email,omitempty"`
	ProfilePictureURL string   `json:"profile_picture_url,omitempty"`
	Websites          []string `json:"websites,omitempty"`
	Vertical          string   `json:"vertical,omitempty"`
}

// TemplateCreateRequest represents a request to create a message template.
type TemplateCreateRequest struct {
	Name       string              `json:"name"`
	Category   TemplateCategory    `json:"category"`
	Language   string              `json:"language"`
	Components []TemplateComponent `json:"components"`
}

// TemplateDeleteResponse represents the response from deleting a template.
type TemplateDeleteResponse struct {
	Success bool `json:"success"`
}

// PhoneNumberInfo represents detailed information about a phone number.
type PhoneNumberInfo struct {
	PhoneNumber
	TwoStepVerificationStatus string `json:"two_step_verification_status,omitempty"`
	AccountMode               string `json:"account_mode,omitempty"`
	CertificateStatus         string `json:"certificate_status,omitempty"`
	NameStatus                string `json:"name_status,omitempty"`
	NewNameStatus             string `json:"new_name_status,omitempty"`
	SearchVisibility          string `json:"search_visibility,omitempty"`
	IsOfficialBusinessAccount bool   `json:"is_official_business_account,omitempty"`
}

// BusinessVertical represents business verticals.
type BusinessVertical string

const (
	VerticalUndefined     BusinessVertical = "UNDEFINED"
	VerticalOther         BusinessVertical = "OTHER"
	VerticalAutoMotive    BusinessVertical = "AUTO"
	VerticalBeauty        BusinessVertical = "BEAUTY"
	VerticalApparel       BusinessVertical = "APPAREL"
	VerticalEducation     BusinessVertical = "EDU"
	VerticalEntertainment BusinessVertical = "ENTERTAIN"
	VerticalEventPlanning BusinessVertical = "EVENT_PLAN"
	VerticalFinance       BusinessVertical = "FINANCE"
	VerticalGrocery       BusinessVertical = "GROCERY"
	VerticalGovernment    BusinessVertical = "GOVT"
	VerticalHotel         BusinessVertical = "HOTEL"
	VerticalHealth        BusinessVertical = "HEALTH"
	VerticalNonProfit     BusinessVertical = "NONPROFIT"
	VerticalProfessional  BusinessVertical = "PROF_SERVICES"
	VerticalShopping      BusinessVertical = "SHOPPING"
	VerticalTravel        BusinessVertical = "TRAVEL"
	VerticalRestaurant    BusinessVertical = "RESTAURANT"
)

// IsValidStatus checks if a phone number status is valid.
func (s PhoneNumberStatus) IsValid() bool {
	switch s {
	case StatusConnected, StatusDisconnected, StatusUnverified, StatusPending, StatusFlagged, StatusRestricted:
		return true
	default:
		return false
	}
}

// IsValidQualityRating checks if a quality rating is valid.
func (q QualityRating) IsValid() bool {
	switch q {
	case QualityGreen, QualityYellow, QualityRed, QualityUnknown:
		return true
	default:
		return false
	}
}

// IsValidTemplateStatus checks if a template status is valid.
func (s TemplateStatus) IsValid() bool {
	switch s {
	case TemplateStatusApproved, TemplateStatusPending, TemplateStatusRejected, TemplateStatusDisabled:
		return true
	default:
		return false
	}
}

// IsValidTemplateCategory checks if a template category is valid.
func (c TemplateCategory) IsValid() bool {
	switch c {
	case CategoryMarketing, CategoryUtility, CategoryAuthentication:
		return true
	default:
		return false
	}
}

// MessageMetrics represents message analytics data.
type MessageMetrics struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Sent      int       `json:"sent"`
	Delivered int       `json:"delivered"`
	Read      int       `json:"read"`
	Failed    int       `json:"failed"`
}

// GetStartTime returns the start time as a time.Time.
func (m *MessageMetrics) GetStartTime() time.Time {
	return m.Start
}

// GetEndTime returns the end time as a time.Time.
func (m *MessageMetrics) GetEndTime() time.Time {
	return m.End
}

// DeliveryRate calculates the delivery rate as a percentage.
func (m *MessageMetrics) DeliveryRate() float64 {
	if m.Sent == 0 {
		return 0
	}
	return float64(m.Delivered) / float64(m.Sent) * 100
}

// ReadRate calculates the read rate as a percentage.
func (m *MessageMetrics) ReadRate() float64 {
	if m.Delivered == 0 {
		return 0
	}
	return float64(m.Read) / float64(m.Delivered) * 100
}

// FailureRate calculates the failure rate as a percentage.
func (m *MessageMetrics) FailureRate() float64 {
	if m.Sent == 0 {
		return 0
	}
	return float64(m.Failed) / float64(m.Sent) * 100
}
