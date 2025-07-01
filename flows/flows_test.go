package flows

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestFlowBuilder(t *testing.T) {
	// Test basic Flow creation
	flow := NewFlow().
		WithRouting("START", "END").
		AddScreen(
			NewScreen("START").
				WithTitle("Welcome").
				AddComponent(NewTextHeading("Hello World")).
				AddComponent(NewFooter("Continue")).
				Build(),
		).
		AddScreen(
			NewScreen("END").
				AsTerminal().
				AsSuccess().
				AddComponent(NewTextBody("Thank you!")).
				Build(),
		).
		Build()

	// Validate Flow structure
	if flow.Version != "3.1" {
		t.Errorf("Expected version 3.1, got %s", flow.Version)
	}

	if flow.DataAPIVersion != "3.0" {
		t.Errorf("Expected data API version 3.0, got %s", flow.DataAPIVersion)
	}

	if len(flow.Screens) != 2 {
		t.Errorf("Expected 2 screens, got %d", len(flow.Screens))
	}

	// Test routing model
	if routes, exists := flow.RoutingModel["START"]; !exists || len(routes) != 1 || routes[0] != "END" {
		t.Errorf("Invalid routing model: %+v", flow.RoutingModel)
	}

	// Test JSON serialization
	jsonStr, err := flow.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize Flow to JSON: %v", err)
	}

	// Test JSON deserialization
	parsedFlow, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse Flow from JSON: %v", err)
	}

	if parsedFlow.Version != flow.Version {
		t.Errorf("Version mismatch after JSON round-trip")
	}
}

func TestFlowValidation(t *testing.T) {
	// Test valid Flow
	validFlow := NewFlow().
		AddScreen(
			NewScreen("START").
				AsTerminal().
				AddComponent(NewTextHeading("Hello")).
				Build(),
		).
		Build()

	validator := NewFlowValidator()
	errors := validator.ValidateFlow(validFlow)
	if len(errors) != 0 {
		t.Errorf("Valid Flow should have no validation errors, got: %+v", errors)
	}

	// Test invalid Flow (no screens)
	invalidFlow := NewFlow().Build()
	errors = validator.ValidateFlow(invalidFlow)
	if len(errors) == 0 {
		t.Error("Flow without screens should have validation errors")
	}

	// Test invalid Flow (no terminal screen)
	noTerminalFlow := NewFlow().
		AddScreen(
			NewScreen("START").
				AddComponent(NewTextHeading("Hello")).
				Build(),
		).
		Build()

	errors = validator.ValidateFlow(noTerminalFlow)
	if len(errors) == 0 {
		t.Error("Flow without terminal screen should have validation errors")
	}
}

func TestComponentBuilders(t *testing.T) {
	// Test text input component
	textInput := NewTextInput("name", "Your Name").AsRequired().Build()
	if textInput.Type != ComponentTypeTextInput {
		t.Errorf("Expected type %s, got %s", ComponentTypeTextInput, textInput.Type)
	}
	if textInput.Name != "name" {
		t.Errorf("Expected name 'name', got %s", textInput.Name)
	}
	if !textInput.Required {
		t.Error("Component should be required")
	}

	// Test dropdown component
	options := []DataSourceItem{
		{ID: "1", Title: "Option 1"},
		{ID: "2", Title: "Option 2"},
	}
	dropdown := NewDropdown("choice", "Select Option", options).Build()
	if len(dropdown.DataSource) != 2 {
		t.Errorf("Expected 2 options, got %d", len(dropdown.DataSource))
	}

	// Test checkbox group
	checkboxGroup := NewCheckboxGroup("interests", "Select Interests").
		AddOption("tech", "Technology").
		AddOption("sports", "Sports").
		WithMinSelected(1).
		WithMaxSelected(2).
		Build()

	if checkboxGroup.Type != ComponentTypeCheckboxGroup {
		t.Errorf("Expected type %s, got %s", ComponentTypeCheckboxGroup, checkboxGroup.Type)
	}
	if len(checkboxGroup.DataSource) != 2 {
		t.Errorf("Expected 2 options, got %d", len(checkboxGroup.DataSource))
	}
	if checkboxGroup.MinSelectedItems != 1 {
		t.Errorf("Expected min selected 1, got %d", checkboxGroup.MinSelectedItems)
	}
}

func TestFlowTokenManager(t *testing.T) {
	manager := NewFlowTokenManager(1 * time.Hour)

	// Test token generation
	token, err := manager.GenerateToken("flow123", "user456", map[string]interface{}{
		"test": "data",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Test token validation
	tokenInfo, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if tokenInfo.FlowID != "flow123" {
		t.Errorf("Expected flow ID 'flow123', got %s", tokenInfo.FlowID)
	}
	if tokenInfo.UserID != "user456" {
		t.Errorf("Expected user ID 'user456', got %s", tokenInfo.UserID)
	}

	// Test token usage
	err = manager.UseToken(token)
	if err != nil {
		t.Fatalf("Failed to use token: %v", err)
	}

	// Test used token validation
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Used token should not be valid")
	}

	// Test invalid token
	_, err = manager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Invalid token should not be valid")
	}
}

func TestDataExchangeRequest(t *testing.T) {
	// Test data exchange request creation
	request := &DataExchangeRequest{
		Version:   "3.0",
		Action:    "submit_form",
		Screen:    "FORM_SCREEN",
		FlowToken: "test-token",
		Data: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Test JSON deserialization
	var parsedRequest DataExchangeRequest
	err = json.Unmarshal(jsonData, &parsedRequest)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if parsedRequest.Action != request.Action {
		t.Errorf("Action mismatch after JSON round-trip")
	}
}

func TestActionRouter(t *testing.T) {
	router := NewActionRouter()

	// Register test handler
	router.RegisterHandlerFunc("test_action", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "SUCCESS_SCREEN",
			Data: map[string]interface{}{
				"result": "success",
			},
		}, nil
	})

	// Test action routing
	request := &DataExchangeRequest{
		Version:   "3.0",
		Action:    "test_action",
		Screen:    "TEST_SCREEN",
		FlowToken: "test-token",
		Data:      map[string]interface{}{},
	}

	response, err := router.RouteAction(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to route action: %v", err)
	}

	if response.Screen != "SUCCESS_SCREEN" {
		t.Errorf("Expected screen 'SUCCESS_SCREEN', got %s", response.Screen)
	}

	// Test unknown action
	request.Action = "unknown_action"
	response, err = router.RouteAction(context.Background(), request)
	if err != nil {
		t.Fatalf("Router should handle unknown actions gracefully: %v", err)
	}

	if response.ErrorDetails == nil {
		t.Error("Unknown action should return error details")
	}
}

func TestFlowStateManager(t *testing.T) {
	manager := NewFlowStateManager()

	// Test state creation
	state := &FlowState{
		FlowToken: "test-token",
		FlowID:    "flow123",
		UserID:    "user456",
		Screen:    "START",
		Data: map[string]interface{}{
			"step": 1,
		},
	}

	manager.SetState(state)

	// Test state retrieval
	retrievedState, err := manager.GetState("test-token")
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if retrievedState.FlowID != "flow123" {
		t.Errorf("Expected flow ID 'flow123', got %s", retrievedState.FlowID)
	}

	// Test state update
	err = manager.UpdateState("test-token", map[string]interface{}{
		"screen": "STEP2",
		"data": map[string]interface{}{
			"step": 2,
		},
	})
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	updatedState, err := manager.GetState("test-token")
	if err != nil {
		t.Fatalf("Failed to get updated state: %v", err)
	}

	if updatedState.Screen != "STEP2" {
		t.Errorf("Expected screen 'STEP2', got %s", updatedState.Screen)
	}

	// Test state deletion
	manager.DeleteState("test-token")
	_, err = manager.GetState("test-token")
	if err == nil {
		t.Error("Deleted state should not be retrievable")
	}
}

func TestFlowCompletion(t *testing.T) {
	// Test Flow completion parsing
	completionJSON := `{
		"flow_token": "test-token",
		"response": {
			"name": "John Doe",
			"email": "john@example.com",
			"satisfaction": "very_satisfied"
		}
	}`

	completion, err := ParseFlowCompletionFromJSON([]byte(completionJSON))
	if err != nil {
		t.Fatalf("Failed to parse Flow completion: %v", err)
	}

	if completion.FlowToken != "test-token" {
		t.Errorf("Expected flow token 'test-token', got %s", completion.FlowToken)
	}

	if completion.Response["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", completion.Response["name"])
	}
}

func TestInMemoryCompletionStore(t *testing.T) {
	store := NewInMemoryCompletionStore()
	ctx := context.Background()

	// Test storing completion
	result := &FlowCompletionResult{
		FlowID:      "flow123",
		UserID:      "user456",
		FlowToken:   "token789",
		Response:    map[string]interface{}{"test": "data"},
		CompletedAt: time.Now(),
		Success:     true,
	}

	err := store.StoreCompletion(ctx, result)
	if err != nil {
		t.Fatalf("Failed to store completion: %v", err)
	}

	// Test retrieving completion
	retrieved, err := store.GetCompletion(ctx, "token789")
	if err != nil {
		t.Fatalf("Failed to get completion: %v", err)
	}

	if retrieved.FlowID != "flow123" {
		t.Errorf("Expected flow ID 'flow123', got %s", retrieved.FlowID)
	}

	// Test getting completions by user
	userCompletions, err := store.GetCompletionsByUser(ctx, "user456")
	if err != nil {
		t.Fatalf("Failed to get completions by user: %v", err)
	}

	if len(userCompletions) != 1 {
		t.Errorf("Expected 1 completion for user, got %d", len(userCompletions))
	}

	// Test getting completions by Flow
	flowCompletions, err := store.GetCompletionsByFlow(ctx, "flow123")
	if err != nil {
		t.Fatalf("Failed to get completions by Flow: %v", err)
	}

	if len(flowCompletions) != 1 {
		t.Errorf("Expected 1 completion for Flow, got %d", len(flowCompletions))
	}
}
