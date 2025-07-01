package webhook

import (
	"context"
	"fmt"
)

// Event handler interfaces following the Interface Segregation Principle (ISP)

// TextMessageHandler defines the interface for handling text messages.
type TextMessageHandler interface {
	OnTextMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// ImageMessageHandler defines the interface for handling image messages.
type ImageMessageHandler interface {
	OnImageMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// AudioMessageHandler defines the interface for handling audio messages.
type AudioMessageHandler interface {
	OnAudioMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// VideoMessageHandler defines the interface for handling video messages.
type VideoMessageHandler interface {
	OnVideoMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// DocumentMessageHandler defines the interface for handling document messages.
type DocumentMessageHandler interface {
	OnDocumentMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// VoiceMessageHandler defines the interface for handling voice messages.
type VoiceMessageHandler interface {
	OnVoiceMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// StickerMessageHandler defines the interface for handling sticker messages.
type StickerMessageHandler interface {
	OnStickerMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// LocationMessageHandler defines the interface for handling location messages.
type LocationMessageHandler interface {
	OnLocationMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// ContactMessageHandler defines the interface for handling contact messages.
type ContactMessageHandler interface {
	OnContactMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// ButtonReplyHandler defines the interface for handling button replies.
type ButtonReplyHandler interface {
	OnButtonReply(ctx context.Context, msg *Message, metadata *Metadata) error
}

// ListReplyHandler defines the interface for handling list replies.
type ListReplyHandler interface {
	OnListReply(ctx context.Context, msg *Message, metadata *Metadata) error
}

// OrderMessageHandler defines the interface for handling order messages.
type OrderMessageHandler interface {
	OnOrderMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// SystemMessageHandler defines the interface for handling system messages.
type SystemMessageHandler interface {
	OnSystemMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// StatusUpdateHandler defines the interface for handling message status updates.
type StatusUpdateHandler interface {
	OnStatusUpdate(ctx context.Context, status *Status, metadata *Metadata) error
}

// ErrorHandler defines the interface for handling error notifications.
type ErrorHandler interface {
	OnError(ctx context.Context, err *Error, metadata *Metadata) error
}

// UnknownMessageHandler defines the interface for handling unknown message types.
type UnknownMessageHandler interface {
	OnUnknownMessage(ctx context.Context, msg *Message, metadata *Metadata) error
}

// EventDispatcher routes events to registered handlers.
type EventDispatcher struct {
	textHandler        TextMessageHandler
	imageHandler       ImageMessageHandler
	audioHandler       AudioMessageHandler
	videoHandler       VideoMessageHandler
	documentHandler    DocumentMessageHandler
	voiceHandler       VoiceMessageHandler
	stickerHandler     StickerMessageHandler
	locationHandler    LocationMessageHandler
	contactHandler     ContactMessageHandler
	buttonReplyHandler ButtonReplyHandler
	listReplyHandler   ListReplyHandler
	orderHandler       OrderMessageHandler
	systemHandler      SystemMessageHandler
	statusHandler      StatusUpdateHandler
	errorHandler       ErrorHandler
	unknownHandler     UnknownMessageHandler
}

// NewEventDispatcher creates a new event dispatcher.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

// Registration methods for each handler type

func (d *EventDispatcher) RegisterTextMessageHandler(h TextMessageHandler) {
	d.textHandler = h
}

func (d *EventDispatcher) RegisterImageMessageHandler(h ImageMessageHandler) {
	d.imageHandler = h
}

func (d *EventDispatcher) RegisterAudioMessageHandler(h AudioMessageHandler) {
	d.audioHandler = h
}

func (d *EventDispatcher) RegisterVideoMessageHandler(h VideoMessageHandler) {
	d.videoHandler = h
}

func (d *EventDispatcher) RegisterDocumentMessageHandler(h DocumentMessageHandler) {
	d.documentHandler = h
}

func (d *EventDispatcher) RegisterVoiceMessageHandler(h VoiceMessageHandler) {
	d.voiceHandler = h
}

func (d *EventDispatcher) RegisterStickerMessageHandler(h StickerMessageHandler) {
	d.stickerHandler = h
}

func (d *EventDispatcher) RegisterLocationMessageHandler(h LocationMessageHandler) {
	d.locationHandler = h
}

func (d *EventDispatcher) RegisterContactMessageHandler(h ContactMessageHandler) {
	d.contactHandler = h
}

func (d *EventDispatcher) RegisterButtonReplyHandler(h ButtonReplyHandler) {
	d.buttonReplyHandler = h
}

func (d *EventDispatcher) RegisterListReplyHandler(h ListReplyHandler) {
	d.listReplyHandler = h
}

func (d *EventDispatcher) RegisterOrderMessageHandler(h OrderMessageHandler) {
	d.orderHandler = h
}

func (d *EventDispatcher) RegisterSystemMessageHandler(h SystemMessageHandler) {
	d.systemHandler = h
}

func (d *EventDispatcher) RegisterStatusUpdateHandler(h StatusUpdateHandler) {
	d.statusHandler = h
}

func (d *EventDispatcher) RegisterErrorHandler(h ErrorHandler) {
	d.errorHandler = h
}

func (d *EventDispatcher) RegisterUnknownMessageHandler(h UnknownMessageHandler) {
	d.unknownHandler = h
}

// DispatchMessage dispatches a message to the appropriate handler.
func (d *EventDispatcher) DispatchMessage(ctx context.Context, msg *Message, metadata *Metadata) error {
	switch msg.Type {
	case "text":
		if d.textHandler != nil {
			return d.textHandler.OnTextMessage(ctx, msg, metadata)
		}
	case "image":
		if d.imageHandler != nil {
			return d.imageHandler.OnImageMessage(ctx, msg, metadata)
		}
	case "audio":
		if d.audioHandler != nil {
			return d.audioHandler.OnAudioMessage(ctx, msg, metadata)
		}
	case "video":
		if d.videoHandler != nil {
			return d.videoHandler.OnVideoMessage(ctx, msg, metadata)
		}
	case "document":
		if d.documentHandler != nil {
			return d.documentHandler.OnDocumentMessage(ctx, msg, metadata)
		}
	case "voice":
		if d.voiceHandler != nil {
			return d.voiceHandler.OnVoiceMessage(ctx, msg, metadata)
		}
	case "sticker":
		if d.stickerHandler != nil {
			return d.stickerHandler.OnStickerMessage(ctx, msg, metadata)
		}
	case "location":
		if d.locationHandler != nil {
			return d.locationHandler.OnLocationMessage(ctx, msg, metadata)
		}
	case "contacts":
		if d.contactHandler != nil {
			return d.contactHandler.OnContactMessage(ctx, msg, metadata)
		}
	case "interactive":
		if msg.Interactive != nil {
			switch msg.Interactive.Type {
			case "button_reply":
				if d.buttonReplyHandler != nil {
					return d.buttonReplyHandler.OnButtonReply(ctx, msg, metadata)
				}
			case "list_reply":
				if d.listReplyHandler != nil {
					return d.listReplyHandler.OnListReply(ctx, msg, metadata)
				}
			}
		}
	case "order":
		if d.orderHandler != nil {
			return d.orderHandler.OnOrderMessage(ctx, msg, metadata)
		}
	case "system":
		if d.systemHandler != nil {
			return d.systemHandler.OnSystemMessage(ctx, msg, metadata)
		}
	default:
		// Handle unknown message types
		if d.unknownHandler != nil {
			return d.unknownHandler.OnUnknownMessage(ctx, msg, metadata)
		}
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	// No handler registered for this message type
	return nil
}

// DispatchStatus dispatches a status update to the registered handler.
func (d *EventDispatcher) DispatchStatus(ctx context.Context, status *Status, metadata *Metadata) error {
	if d.statusHandler != nil {
		return d.statusHandler.OnStatusUpdate(ctx, status, metadata)
	}
	return nil
}

// DispatchError dispatches an error notification to the registered handler.
func (d *EventDispatcher) DispatchError(ctx context.Context, err *Error, metadata *Metadata) error {
	if d.errorHandler != nil {
		return d.errorHandler.OnError(ctx, err, metadata)
	}
	return nil
}

// HasHandler checks if a handler is registered for the given message type.
func (d *EventDispatcher) HasHandler(messageType string) bool {
	switch messageType {
	case "text":
		return d.textHandler != nil
	case "image":
		return d.imageHandler != nil
	case "audio":
		return d.audioHandler != nil
	case "video":
		return d.videoHandler != nil
	case "document":
		return d.documentHandler != nil
	case "voice":
		return d.voiceHandler != nil
	case "sticker":
		return d.stickerHandler != nil
	case "location":
		return d.locationHandler != nil
	case "contacts":
		return d.contactHandler != nil
	case "interactive":
		return d.buttonReplyHandler != nil || d.listReplyHandler != nil
	case "order":
		return d.orderHandler != nil
	case "system":
		return d.systemHandler != nil
	default:
		return d.unknownHandler != nil
	}
}

// GetRegisteredHandlers returns a list of registered handler types.
func (d *EventDispatcher) GetRegisteredHandlers() []string {
	var handlers []string

	if d.textHandler != nil {
		handlers = append(handlers, "text")
	}
	if d.imageHandler != nil {
		handlers = append(handlers, "image")
	}
	if d.audioHandler != nil {
		handlers = append(handlers, "audio")
	}
	if d.videoHandler != nil {
		handlers = append(handlers, "video")
	}
	if d.documentHandler != nil {
		handlers = append(handlers, "document")
	}
	if d.voiceHandler != nil {
		handlers = append(handlers, "voice")
	}
	if d.stickerHandler != nil {
		handlers = append(handlers, "sticker")
	}
	if d.locationHandler != nil {
		handlers = append(handlers, "location")
	}
	if d.contactHandler != nil {
		handlers = append(handlers, "contacts")
	}
	if d.buttonReplyHandler != nil {
		handlers = append(handlers, "button_reply")
	}
	if d.listReplyHandler != nil {
		handlers = append(handlers, "list_reply")
	}
	if d.orderHandler != nil {
		handlers = append(handlers, "order")
	}
	if d.systemHandler != nil {
		handlers = append(handlers, "system")
	}
	if d.statusHandler != nil {
		handlers = append(handlers, "status")
	}
	if d.errorHandler != nil {
		handlers = append(handlers, "error")
	}
	if d.unknownHandler != nil {
		handlers = append(handlers, "unknown")
	}

	return handlers
}
