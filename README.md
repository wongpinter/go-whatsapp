# WhatsApp Go Library

[![Go Reference](https://pkg.go.dev/badge/github.com/wongpinter/go-whatsapp.svg)](https://pkg.go.dev/github.com/wongpinter/go-whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/wongpinter/go-whatsapp)](https://goreportcard.com/report/github.com/wongpinter/go-whatsapp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go library for the WhatsApp Business Cloud API. Send messages, handle webhooks, manage templates, and build interactive Flows with a clean, type-safe interface.

## ✨ Features

* 📱 **Complete API Coverage** - Messages, webhooks, business management, Flows
* 🔒 **Security Built-in** - Webhook signature verification, secure token handling
* 🏗️ **Framework Agnostic** - Works with Gin, Echo, standard HTTP, and more
* 📝 **Type Safe** - Full type definitions with comprehensive error handling
* 🧪 **Testable** - Interface-based design for easy mocking

## 🚀 Quick Start

### Installation

```bash
go get github.com/wongpinter/go-whatsapp
```

### Send a Message

```go
package main

import (
    "context"
    "log"

    "github.com/wongpinter/go-whatsapp/cloudapi"
)

func main() {
    client := cloudapi.NewClient("PHONE_NUMBER_ID", "ACCESS_TOKEN")

    resp, err := client.SendText(context.Background(), "+1234567890", "Hello from Go!")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent: %s", resp.GetMessageID())
}
```

### Handle Webhooks

```go
package main

import (
    "context"
    "net/http"

    "github.com/wongpinter/go-whatsapp/webhook"
)

type Handler struct{}

func (h *Handler) OnTextMessage(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
    log.Printf("Received: %s from %s", msg.Text.Body, msg.From)
    return nil
}

func main() {
    handler := webhook.NewHandler("VERIFY_TOKEN", "APP_SECRET", nil)
    handler.SetMessageHandler(&Handler{})

    http.HandleFunc("/webhook", handler.ServeHTTP)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Build WhatsApp Flows

```go
package main

import (
    "github.com/wongpinter/go-whatsapp/flows"
)

func main() {
    flow := flows.NewFlowBuilder().
        SetVersion("3.0").
        AddScreen("contact_form").
            SetTitle("Contact Information").
            AddTextInput("name", "Full Name").SetRequired(true).Done().
            AddEmailInput("email", "Email").SetRequired(true).Done().
            SetTerminal(true).
            Done().
        Build()

    flowJSON, _ := flow.ToJSON()
    log.Printf("Flow created: %s", flowJSON)
}
```

## 📦 Packages

| Package | Purpose | Documentation |
|---------|---------|---------------|
| `cloudapi` | Send messages (text, media, templates) | [API Reference](docs/COMPLETE_GUIDE.md#cloudapi-package) |
| `webhook` | Handle incoming events and messages | [Webhook Guide](docs/COMPLETE_GUIDE.md#webhook-package) |
| `flows` | Build and manage WhatsApp Flows | [Flows Guide](docs/COMPLETE_GUIDE.md#flows-package) |
| `bm` | Business Management and templates | [BM Guide](docs/COMPLETE_GUIDE.md#business-management-package) |

## 🛠️ Configuration

### Environment Variables

```bash
# Required
WHATSAPP_ACCESS_TOKEN=your_access_token
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id

# For webhooks
WHATSAPP_VERIFY_TOKEN=your_verify_token
WHATSAPP_APP_SECRET=your_app_secret
```

### Client Options

```go
client := cloudapi.NewClient(phoneNumberID, accessToken,
    cloudapi.WithLogger(logger),
    cloudapi.WithRateLimiting(80.0),
    cloudapi.WithRetryConfig(3, time.Second),
)
```

## 📚 Documentation

* **[Complete Guide](docs/COMPLETE_GUIDE.md)** - Comprehensive documentation with examples
* **[API Reference](docs/COMPLETE_GUIDE.md#api-reference)** - Detailed API documentation
* **[Getting Started](docs/INDEX.md)** - Quick navigation and setup guide
* **[Examples](examples/)** - Working code examples

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

* **[WhatsApp Business API](https://developers.facebook.com/docs/whatsapp/cloud-api)** - Official API documentation
* **[Issues](https://github.com/wongpinter/go-whatsapp/issues)** - Report bugs or request features
* **[Discussions](https://github.com/wongpinter/go-whatsapp/discussions)** - Community discussions

---

<div align="center">
Made with ❤️ for the Go community
</div>
