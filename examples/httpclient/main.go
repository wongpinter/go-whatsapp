package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp"
	"github.com/wongpinter/go-whatsapp/bm"
	"github.com/wongpinter/go-whatsapp/cloudapi"
	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/internal/httpclient"
)

func main() {
	// Example 1: Using the improved main client with shared HTTP client
	log.Println("=== Example 1: Main Client with Shared HTTP Client ===")
	mainClientExample()

	// Example 2: Using individual package clients with shared HTTP client manager
	log.Println("\n=== Example 2: Package Clients with Shared Manager ===")
	packageClientsExample()

	// Example 3: Custom HTTP client configuration
	log.Println("\n=== Example 3: Custom HTTP Client Configuration ===")
	customHTTPClientExample()

	// Example 4: Demonstrating client reuse
	log.Println("\n=== Example 4: Client Reuse Demonstration ===")
	clientReuseExample()
}

func mainClientExample() {
	// Create main WhatsApp client with improved HTTP client management
	client, err := whatsapp.NewClient(
		"YOUR_PHONE_NUMBER_ID",
		"YOUR_ACCESS_TOKEN",
		whatsapp.WithAPIVersion("v19.0"),
		whatsapp.WithRateLimiting(80.0), // 80 requests per second
		whatsapp.WithTimeout(30*time.Second),
		whatsapp.WithRetryConfig(3, time.Second, 10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create WhatsApp client: %v", err)
	}

	log.Printf("Main client created successfully with API version: %s", client.APIVersion)
}

func packageClientsExample() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create CloudAPI client
	cloudClient := cloudapi.NewClient(
		"YOUR_PHONE_NUMBER_ID",
		"YOUR_ACCESS_TOKEN",
		cloudapi.WithLogger(logger),
	)
	log.Printf("CloudAPI client created")

	// Create Business Management client
	bmClient := bm.NewClient(
		"YOUR_ACCESS_TOKEN",
		bm.WithLogger(logger),
		bm.WithWABAID("YOUR_WABA_ID"),
	)
	log.Printf("Business Management client created")

	// Create Flows client
	flowsClient := flows.NewClient(
		"YOUR_WABA_ID",
		"YOUR_ACCESS_TOKEN",
		flows.WithLogger(&logger),
		flows.WithAPIVersion("v18.0"),
	)
	log.Printf("Flows client created")

	// All clients are now using optimized HTTP client initialization
	_ = cloudClient
	_ = bmClient
	_ = flowsClient
}

func customHTTPClientExample() {
	// Create a custom HTTP client with specific configuration
	customHTTPClient := &http.Client{
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     60 * time.Second,
		},
	}

	// Use the custom HTTP client with WhatsApp client
	client, err := whatsapp.NewClient(
		"YOUR_PHONE_NUMBER_ID",
		"YOUR_ACCESS_TOKEN",
		whatsapp.WithHTTPClient(customHTTPClient),
		whatsapp.WithAPIVersion("v19.0"),
	)
	if err != nil {
		log.Fatalf("Failed to create WhatsApp client with custom HTTP client: %v", err)
	}

	log.Printf("Client created with custom HTTP client configuration")
	_ = client
}

func clientReuseExample() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create HTTP client manager for demonstration
	manager := httpclient.NewManager(nil, &logger)

	// Get default client options
	defaultOpts := httpclient.GetDefaultOptions("YOUR_ACCESS_TOKEN")

	// Configure rate limiting for CloudAPI
	cloudConfig := httpclient.WithRateLimiting(defaultOpts.CloudAPI, 80.0, 10)
	cloudConfig = httpclient.WithLogger(cloudConfig, &logger)

	// Create first CloudAPI client
	client1, err := manager.GetOrCreateClient(httpclient.CloudAPIClient, cloudConfig)
	if err != nil {
		log.Fatalf("Failed to create first client: %v", err)
	}

	// Create second CloudAPI client with same configuration - should reuse the first one
	client2, err := manager.GetOrCreateClient(httpclient.CloudAPIClient, cloudConfig)
	if err != nil {
		log.Fatalf("Failed to create second client: %v", err)
	}

	// Check if clients are the same instance (reused)
	if client1 == client2 {
		log.Printf("✓ Clients are reused - same instance returned")
	} else {
		log.Printf("✗ Clients are different instances")
	}

	log.Printf("Total cached clients: %d", manager.GetClientCount())

	// Clean up
	manager.CloseAll()
	log.Printf("All clients closed and cache cleared")
}

// Example of advanced HTTP client configuration
func advancedHTTPClientConfiguration() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create manager with custom configuration
	manager := httpclient.NewManager(nil, &logger)

	// Create highly customized client configuration
	config := &httpclient.ClientConfig{
		AccessToken:   "YOUR_ACCESS_TOKEN",
		APIVersion:    "v19.0",
		Timeout:       30 * time.Second,
		RetryCount:    5,
		RetryWaitTime: 2 * time.Second,
		RetryMaxWait:  20 * time.Second,
		UserAgent:     "MyApp/1.0.0 go-whatsapp-sdk/1.0.0",
		CustomHeaders: map[string]string{
			"X-Custom-Header": "MyValue",
			"X-App-Version":   "1.0.0",
		},
		Logger: &logger,
	}

	// Add rate limiting
	config = httpclient.WithRateLimiting(config, 100.0, 20)

	// Create client with advanced configuration
	client, err := manager.GetOrCreateClient(httpclient.CloudAPIClient, config)
	if err != nil {
		log.Fatalf("Failed to create advanced client: %v", err)
	}

	log.Printf("Advanced HTTP client created with custom configuration")
	_ = client
}

// Example showing different client types
func multipleClientTypesExample() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	manager := httpclient.NewManager(nil, &logger)

	// Create different types of clients
	clientTypes := []httpclient.ClientType{
		httpclient.CloudAPIClient,
		httpclient.BusinessAPIClient,
		httpclient.FlowsAPIClient,
	}

	for _, clientType := range clientTypes {
		config := &httpclient.ClientConfig{
			AccessToken: "YOUR_ACCESS_TOKEN",
			Logger:      &logger,
		}

		client, err := manager.GetOrCreateClient(clientType, config)
		if err != nil {
			log.Printf("Failed to create %s client: %v", clientType, err)
			continue
		}

		log.Printf("Created %s client with base URL: %s", clientType, client.BaseURL)
	}

	log.Printf("Total clients created: %d", manager.GetClientCount())
}
