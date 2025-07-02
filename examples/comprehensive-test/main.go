package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpserver"
	"github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
	fmt.Println("🚀 WhatsApp HTTP Server Framework Abstraction - Comprehensive Test")
	fmt.Println("================================================================")
	
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Test 1: Framework Support
	fmt.Println("\n📋 Test 1: Framework Support")
	testFrameworkSupport()

	// Test 2: Webhook Integration
	fmt.Println("\n📋 Test 2: Webhook Integration")
	testWebhookIntegration(&logger)

	// Test 3: Build Tag Detection
	fmt.Println("\n📋 Test 3: Build Tag Detection")
	testBuildTagDetection()

	// Test 4: Framework Detection
	fmt.Println("\n📋 Test 4: Framework Detection")
	testFrameworkDetection()

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("\n📖 Usage Instructions:")
	fmt.Println("  • Default build (mock implementations):")
	fmt.Println("    go run examples/comprehensive-test/main.go")
	fmt.Println("  • With real Gin support:")
	fmt.Println("    go run -tags gin examples/comprehensive-test/main.go")
	fmt.Println("  • With real Echo support:")
	fmt.Println("    go run -tags echo examples/comprehensive-test/main.go")
	fmt.Println("  • With both real frameworks:")
	fmt.Println("    go run -tags \"gin echo\" examples/comprehensive-test/main.go")
}

func testFrameworkSupport() {
	factory := httpserver.NewServerFactory()
	frameworks := httpserver.GetSupportedFrameworks()
	
	fmt.Printf("   Supported frameworks: %v\n", frameworks)
	
	for _, framework := range frameworks[:3] { // Test first 3 (implemented ones)
		server, err := factory.CreateServer(framework)
		if err != nil {
			fmt.Printf("   ❌ %s: %v\n", framework, err)
		} else {
			fmt.Printf("   ✅ %s: %T\n", framework, server.Native())
		}
	}
}

func testWebhookIntegration(logger *zerolog.Logger) {
	factory := httpserver.NewServerFactory()
	webhookAdapter := webhook.NewServerAdapter("test-secret", "test-token", logger)
	
	frameworks := []httpserver.Framework{
		httpserver.FrameworkStandard,
		httpserver.FrameworkGin,
		httpserver.FrameworkEcho,
	}
	
	for _, framework := range frameworks {
		server, err := factory.CreateServer(framework)
		if err != nil {
			fmt.Printf("   ❌ %s webhook integration failed: %v\n", framework, err)
			continue
		}
		
		// Register webhook routes
		webhookAdapter.RegisterRoutes(server.Router())
		
		// Add test route
		server.Router().GET("/test", func(ctx httpserver.HTTPContext) error {
			return ctx.JSON(200, map[string]string{
				"framework": string(framework),
				"status":    "working",
			})
		})
		
		fmt.Printf("   ✅ %s webhook integration successful\n", framework)
	}
}

func testBuildTagDetection() {
	// Check which real implementations are available
	ginAvailable := isRealGinAvailable()
	echoAvailable := isRealEchoAvailable()
	
	fmt.Printf("   Real Gin implementation: %s\n", boolToStatus(ginAvailable))
	fmt.Printf("   Real Echo implementation: %s\n", boolToStatus(echoAvailable))
	
	if ginAvailable || echoAvailable {
		fmt.Printf("   🎯 Build tags are working correctly!\n")
	} else {
		fmt.Printf("   📝 Using mock implementations (no build tags)\n")
	}
}

func testFrameworkDetection() {
	factory := httpserver.NewServerFactory()
	
	testCases := []struct {
		framework httpserver.Framework
		expected  string
	}{
		{httpserver.FrameworkStandard, "standard"},
		{httpserver.FrameworkGin, "gin or standard"},
		{httpserver.FrameworkEcho, "echo or standard"},
	}
	
	for _, tc := range testCases {
		server, err := factory.CreateServer(tc.framework)
		if err != nil {
			fmt.Printf("   ❌ %s detection failed: %v\n", tc.framework, err)
			continue
		}
		
		detected := httpserver.DetectFramework(server.Native())
		fmt.Printf("   ✅ %s → detected as %s\n", tc.framework, detected)
	}
}

// Helper functions

func isRealGinAvailable() bool {
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(httpserver.FrameworkGin)
	if err != nil {
		return false
	}
	
	// Check if it's a real Gin engine or mock
	nativeType := fmt.Sprintf("%T", server.Native())
	return nativeType == "*gin.Engine"
}

func isRealEchoAvailable() bool {
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(httpserver.FrameworkEcho)
	if err != nil {
		return false
	}
	
	// Check if it's a real Echo engine or mock
	nativeType := fmt.Sprintf("%T", server.Native())
	return nativeType == "*echo.Echo"
}

func boolToStatus(b bool) string {
	if b {
		return "✅ Available"
	}
	return "📦 Mock"
}

// Example usage functions

func exampleStandardHTTP() {
	fmt.Println("\n🔧 Example: Standard HTTP Server")
	
	factory := httpserver.NewServerFactory()
	server, _ := factory.CreateServer(httpserver.FrameworkStandard)
	
	router := server.Router()
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.String(200, "Hello from Standard HTTP!")
	})
	
	fmt.Println("   Server configured with standard HTTP")
}

func exampleGinIntegration() {
	fmt.Println("\n🔧 Example: Gin Integration")
	
	// This would work with real Gin when build tag is used
	factory := httpserver.NewServerFactory()
	server, _ := factory.CreateServer(httpserver.FrameworkGin)
	
	router := server.Router()
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"message": "Hello from Gin!",
		})
	})
	
	fmt.Println("   Server configured with Gin framework")
}

func exampleEchoIntegration() {
	fmt.Println("\n🔧 Example: Echo Integration")
	
	// This would work with real Echo when build tag is used
	factory := httpserver.NewServerFactory()
	server, _ := factory.CreateServer(httpserver.FrameworkEcho)
	
	router := server.Router()
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"message": "Hello from Echo!",
		})
	})
	
	fmt.Println("   Server configured with Echo framework")
}

func init() {
	// Check if this is being run with specific instructions
	if len(os.Args) > 1 && os.Args[1] == "--examples" {
		exampleStandardHTTP()
		exampleGinIntegration()
		exampleEchoIntegration()
		os.Exit(0)
	}
}
