package flows

// FlowBuilder helps construct Flow definitions using a fluent API.
type FlowBuilder struct {
	flow *Flow
}

// NewFlow creates a new Flow builder with default values.
func NewFlow() *FlowBuilder {
	return &FlowBuilder{
		flow: &Flow{
			Version:        "3.1",
			DataAPIVersion: "3.0",
			Screens:        make([]Screen, 0),
			RoutingModel:   make(map[string][]string),
		},
	}
}

// WithVersion sets the Flow version.
func (b *FlowBuilder) WithVersion(version string) *FlowBuilder {
	b.flow.Version = version
	return b
}

// WithDataAPIVersion sets the data API version.
func (b *FlowBuilder) WithDataAPIVersion(version string) *FlowBuilder {
	b.flow.DataAPIVersion = version
	return b
}

// WithRouting adds routing rules to the Flow.
func (b *FlowBuilder) WithRouting(from string, to ...string) *FlowBuilder {
	if b.flow.RoutingModel == nil {
		b.flow.RoutingModel = make(map[string][]string)
	}
	b.flow.RoutingModel[from] = to
	return b
}

// AddScreen adds a screen to the Flow.
func (b *FlowBuilder) AddScreen(screen *Screen) *FlowBuilder {
	b.flow.Screens = append(b.flow.Screens, *screen)
	return b
}

// Build returns the constructed Flow.
func (b *FlowBuilder) Build() *Flow {
	return b.flow
}

// ScreenBuilder helps construct Flow screens.
type ScreenBuilder struct {
	screen *Screen
}

// NewScreen creates a new Screen builder.
func NewScreen(id string) *ScreenBuilder {
	return &ScreenBuilder{
		screen: &Screen{
			ID:   id,
			Data: make(map[string]DataField),
			Layout: Layout{
				Type:     LayoutTypeSingleColumn,
				Children: make([]Component, 0),
			},
		},
	}
}

// WithTitle sets the screen title.
func (b *ScreenBuilder) WithTitle(title string) *ScreenBuilder {
	b.screen.Title = title
	return b
}

// AsTerminal marks the screen as terminal.
func (b *ScreenBuilder) AsTerminal() *ScreenBuilder {
	b.screen.Terminal = true
	return b
}

// AsSuccess marks the screen as a success terminal screen.
func (b *ScreenBuilder) AsSuccess() *ScreenBuilder {
	b.screen.Terminal = true
	b.screen.Success = true
	return b
}

// WithRefreshOnBack enables refresh on back navigation.
func (b *ScreenBuilder) WithRefreshOnBack() *ScreenBuilder {
	b.screen.RefreshOnBack = true
	return b
}

// WithData adds a data field to the screen.
func (b *ScreenBuilder) WithData(name, dataType string, example interface{}) *ScreenBuilder {
	b.screen.Data[name] = DataField{
		Type:    dataType,
		Example: example,
	}
	return b
}

// WithRequiredData adds a required data field to the screen.
func (b *ScreenBuilder) WithRequiredData(name, dataType string, example interface{}) *ScreenBuilder {
	b.screen.Data[name] = DataField{
		Type:     dataType,
		Example:  example,
		Required: true,
	}
	return b
}

// AddComponent adds a component to the screen layout.
func (b *ScreenBuilder) AddComponent(component Component) *ScreenBuilder {
	b.screen.Layout.Children = append(b.screen.Layout.Children, component)
	return b
}

// Build returns the constructed Screen.
func (b *ScreenBuilder) Build() *Screen {
	return b.screen
}

// ComponentBuilder helps construct Flow components.
type ComponentBuilder struct {
	component Component
}

// NewComponent creates a new Component builder.
func NewComponent(componentType string) *ComponentBuilder {
	return &ComponentBuilder{
		component: Component{
			Type:    componentType,
			Enabled: true,
			Visible: true,
		},
	}
}

// WithText sets the component text.
func (b *ComponentBuilder) WithText(text string) *ComponentBuilder {
	b.component.Text = text
	return b
}

// WithLabel sets the component label.
func (b *ComponentBuilder) WithLabel(label string) *ComponentBuilder {
	b.component.Label = label
	return b
}

// WithName sets the component name (for input components).
func (b *ComponentBuilder) WithName(name string) *ComponentBuilder {
	b.component.Name = name
	return b
}

// WithInputType sets the input type for input components.
func (b *ComponentBuilder) WithInputType(inputType string) *ComponentBuilder {
	b.component.InputType = inputType
	return b
}

// AsRequired marks the component as required.
func (b *ComponentBuilder) AsRequired() *ComponentBuilder {
	b.component.Required = true
	return b
}

// AsDisabled marks the component as disabled.
func (b *ComponentBuilder) AsDisabled() *ComponentBuilder {
	b.component.Enabled = false
	return b
}

// AsHidden marks the component as hidden.
func (b *ComponentBuilder) AsHidden() *ComponentBuilder {
	b.component.Visible = false
	return b
}

// WithClickAction sets the on-click action for the component.
func (b *ComponentBuilder) WithClickAction(actionName string, payload map[string]interface{}) *ComponentBuilder {
	b.component.OnClickAction = &Action{
		Name:    actionName,
		Payload: payload,
	}
	return b
}

// WithSelectionChangeAction sets the on-selection-change action for the component.
func (b *ComponentBuilder) WithSelectionChangeAction(actionName string, payload map[string]interface{}) *ComponentBuilder {
	b.component.OnSelectionChange = &Action{
		Name:    actionName,
		Payload: payload,
	}
	return b
}

// WithDataSource sets the data source for dropdown/selection components.
func (b *ComponentBuilder) WithDataSource(items []DataSourceItem) *ComponentBuilder {
	b.component.DataSource = items
	return b
}

// WithInitValue sets the initial value for the component.
func (b *ComponentBuilder) WithInitValue(value interface{}) *ComponentBuilder {
	b.component.InitValue = value
	return b
}

// WithHelperText sets the helper text for the component.
func (b *ComponentBuilder) WithHelperText(text string) *ComponentBuilder {
	b.component.HelperText = text
	return b
}

// WithErrorMessage sets the error message for the component.
func (b *ComponentBuilder) WithErrorMessage(message string) *ComponentBuilder {
	b.component.ErrorMessage = message
	return b
}

// Build returns the constructed Component.
func (b *ComponentBuilder) Build() Component {
	return b.component
}

// Convenience functions for common components

// NewTextHeading creates a text heading component.
func NewTextHeading(text string) Component {
	return NewComponent(ComponentTypeTextHeading).
		WithText(text).
		Build()
}

// NewTextSubheading creates a text subheading component.
func NewTextSubheading(text string) Component {
	return NewComponent(ComponentTypeTextSubheading).
		WithText(text).
		Build()
}

// NewTextBody creates a text body component.
func NewTextBody(text string) Component {
	return NewComponent(ComponentTypeTextBody).
		WithText(text).
		Build()
}

// NewTextInput creates a text input component builder.
func NewTextInput(name, label string) *ComponentBuilder {
	return NewComponent(ComponentTypeTextInput).
		WithName(name).
		WithLabel(label).
		WithInputType(InputTypeText)
}

// NewEmailInput creates an email input component builder.
func NewEmailInput(name, label string) *ComponentBuilder {
	return NewComponent(ComponentTypeTextInput).
		WithName(name).
		WithLabel(label).
		WithInputType(InputTypeEmail)
}

// NewNumberInput creates a number input component builder.
func NewNumberInput(name, label string) *ComponentBuilder {
	return NewComponent(ComponentTypeTextInput).
		WithName(name).
		WithLabel(label).
		WithInputType(InputTypeNumber)
}

// NewPhoneInput creates a phone input component builder.
func NewPhoneInput(name, label string) *ComponentBuilder {
	return NewComponent(ComponentTypeTextInput).
		WithName(name).
		WithLabel(label).
		WithInputType(InputTypePhone)
}

// NewTextArea creates a text area component builder.
func NewTextArea(name, label string) *ComponentBuilder {
	return NewComponent(ComponentTypeTextArea).
		WithName(name).
		WithLabel(label)
}

// NewDropdown creates a dropdown component builder.
func NewDropdown(name, label string, items []DataSourceItem) *ComponentBuilder {
	return NewComponent(ComponentTypeDropdown).
		WithName(name).
		WithLabel(label).
		WithDataSource(items)
}

// NewFooter creates a footer component with navigation action.
func NewFooter(label string) Component {
	return NewComponent(ComponentTypeFooter).
		WithLabel(label).
		WithClickAction(ActionNavigate, map[string]interface{}{}).
		Build()
}

// NewDataExchangeFooter creates a footer component with data exchange action.
func NewDataExchangeFooter(label string, payload map[string]interface{}) Component {
	return NewComponent(ComponentTypeFooter).
		WithLabel(label).
		WithClickAction(ActionDataExchange, payload).
		Build()
}

// NewCompleteFooter creates a footer component with complete action.
func NewCompleteFooter(label string) Component {
	return NewComponent(ComponentTypeFooter).
		WithLabel(label).
		WithClickAction(ActionComplete, map[string]interface{}{}).
		Build()
}

// NewOptIn creates an opt-in component.
func NewOptIn(text string) Component {
	return NewComponent(ComponentTypeOptIn).
		WithText(text).
		AsRequired().
		Build()
}
