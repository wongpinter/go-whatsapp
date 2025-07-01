package flows

// Advanced component builders for complex Flow components

// CheckboxGroupBuilder helps build checkbox group components.
type CheckboxGroupBuilder struct {
	component Component
}

// NewCheckboxGroup creates a new checkbox group builder.
func NewCheckboxGroup(name, label string) *CheckboxGroupBuilder {
	return &CheckboxGroupBuilder{
		component: Component{
			Type:     ComponentTypeCheckboxGroup,
			Name:     name,
			Label:    label,
			Enabled:  true,
			Visible:  true,
			DataSource: make([]DataSourceItem, 0),
		},
	}
}

// AddOption adds an option to the checkbox group.
func (b *CheckboxGroupBuilder) AddOption(id, title string) *CheckboxGroupBuilder {
	b.component.DataSource = append(b.component.DataSource, DataSourceItem{
		ID:    id,
		Title: title,
	})
	return b
}

// WithMinSelected sets the minimum number of selected items.
func (b *CheckboxGroupBuilder) WithMinSelected(min int) *CheckboxGroupBuilder {
	b.component.MinSelectedItems = min
	return b
}

// WithMaxSelected sets the maximum number of selected items.
func (b *CheckboxGroupBuilder) WithMaxSelected(max int) *CheckboxGroupBuilder {
	b.component.MaxSelectedItems = max
	return b
}

// AsRequired marks the checkbox group as required.
func (b *CheckboxGroupBuilder) AsRequired() *CheckboxGroupBuilder {
	b.component.Required = true
	return b
}

// WithHelperText sets the helper text.
func (b *CheckboxGroupBuilder) WithHelperText(text string) *CheckboxGroupBuilder {
	b.component.HelperText = text
	return b
}

// Build returns the constructed checkbox group component.
func (b *CheckboxGroupBuilder) Build() Component {
	return b.component
}

// RadioButtonsGroupBuilder helps build radio buttons group components.
type RadioButtonsGroupBuilder struct {
	component Component
}

// NewRadioButtonsGroup creates a new radio buttons group builder.
func NewRadioButtonsGroup(name, label string) *RadioButtonsGroupBuilder {
	return &RadioButtonsGroupBuilder{
		component: Component{
			Type:     ComponentTypeRadioButtonsGroup,
			Name:     name,
			Label:    label,
			Enabled:  true,
			Visible:  true,
			DataSource: make([]DataSourceItem, 0),
		},
	}
}

// AddOption adds an option to the radio buttons group.
func (b *RadioButtonsGroupBuilder) AddOption(id, title string) *RadioButtonsGroupBuilder {
	b.component.DataSource = append(b.component.DataSource, DataSourceItem{
		ID:    id,
		Title: title,
	})
	return b
}

// AsRequired marks the radio buttons group as required.
func (b *RadioButtonsGroupBuilder) AsRequired() *RadioButtonsGroupBuilder {
	b.component.Required = true
	return b
}

// WithHelperText sets the helper text.
func (b *RadioButtonsGroupBuilder) WithHelperText(text string) *RadioButtonsGroupBuilder {
	b.component.HelperText = text
	return b
}

// WithInitValue sets the initial selected value.
func (b *RadioButtonsGroupBuilder) WithInitValue(value string) *RadioButtonsGroupBuilder {
	b.component.InitValue = value
	return b
}

// Build returns the constructed radio buttons group component.
func (b *RadioButtonsGroupBuilder) Build() Component {
	return b.component
}

// DatePickerBuilder helps build date picker components.
type DatePickerBuilder struct {
	component Component
}

// NewDatePicker creates a new date picker builder.
func NewDatePicker(name, label string) *DatePickerBuilder {
	return &DatePickerBuilder{
		component: Component{
			Type:    ComponentTypeDatePicker,
			Name:    name,
			Label:   label,
			Enabled: true,
			Visible: true,
		},
	}
}

// AsRequired marks the date picker as required.
func (b *DatePickerBuilder) AsRequired() *DatePickerBuilder {
	b.component.Required = true
	return b
}

// WithHelperText sets the helper text.
func (b *DatePickerBuilder) WithHelperText(text string) *DatePickerBuilder {
	b.component.HelperText = text
	return b
}

// WithInitValue sets the initial date value (YYYY-MM-DD format).
func (b *DatePickerBuilder) WithInitValue(date string) *DatePickerBuilder {
	b.component.InitValue = date
	return b
}

// Build returns the constructed date picker component.
func (b *DatePickerBuilder) Build() Component {
	return b.component
}

// ImageBuilder helps build image components.
type ImageBuilder struct {
	component Component
}

// NewImage creates a new image builder.
func NewImage(src string) *ImageBuilder {
	return &ImageBuilder{
		component: Component{
			Type:    ComponentTypeImage,
			Text:    src, // Image source URL is stored in Text field
			Enabled: true,
			Visible: true,
		},
	}
}

// WithAltText sets the alt text for the image.
func (b *ImageBuilder) WithAltText(altText string) *ImageBuilder {
	b.component.Label = altText // Alt text is stored in Label field
	return b
}

// Build returns the constructed image component.
func (b *ImageBuilder) Build() Component {
	return b.component
}

// EmbeddedLinkBuilder helps build embedded link components.
type EmbeddedLinkBuilder struct {
	component Component
}

// NewEmbeddedLink creates a new embedded link builder.
func NewEmbeddedLink(text, url string) *EmbeddedLinkBuilder {
	return &EmbeddedLinkBuilder{
		component: Component{
			Type:    ComponentTypeEmbeddedLink,
			Text:    text,
			Label:   url, // URL is stored in Label field
			Enabled: true,
			Visible: true,
		},
	}
}

// Build returns the constructed embedded link component.
func (b *EmbeddedLinkBuilder) Build() Component {
	return b.component
}

// Helper functions for creating data source items

// NewDataSourceItem creates a new data source item.
func NewDataSourceItem(id, title string) DataSourceItem {
	return DataSourceItem{
		ID:    id,
		Title: title,
	}
}

// NewDataSourceItems creates multiple data source items from a map.
func NewDataSourceItems(items map[string]string) []DataSourceItem {
	result := make([]DataSourceItem, 0, len(items))
	for id, title := range items {
		result = append(result, DataSourceItem{
			ID:    id,
			Title: title,
		})
	}
	return result
}

// Helper functions for creating actions

// NewNavigateAction creates a navigate action.
func NewNavigateAction(payload map[string]interface{}) *Action {
	return &Action{
		Name:    ActionNavigate,
		Payload: payload,
	}
}

// NewDataExchangeAction creates a data exchange action.
func NewDataExchangeAction(payload map[string]interface{}) *Action {
	return &Action{
		Name:    ActionDataExchange,
		Payload: payload,
	}
}

// NewCompleteAction creates a complete action.
func NewCompleteAction() *Action {
	return &Action{
		Name:    ActionComplete,
		Payload: make(map[string]interface{}),
	}
}

// Validation helpers for components

// ValidateComponent performs validation on a component.
func ValidateComponent(component Component) error {
	// Check if component type is valid
	validTypes := map[string]bool{
		ComponentTypeTextHeading:       true,
		ComponentTypeTextSubheading:    true,
		ComponentTypeTextBody:          true,
		ComponentTypeTextCaption:       true,
		ComponentTypeTextInput:         true,
		ComponentTypeTextArea:          true,
		ComponentTypeCheckboxGroup:     true,
		ComponentTypeRadioButtonsGroup: true,
		ComponentTypeDropdown:          true,
		ComponentTypeFooter:            true,
		ComponentTypeOptIn:             true,
		ComponentTypeEmbeddedLink:      true,
		ComponentTypeDatePicker:        true,
		ComponentTypeImage:             true,
	}

	if !validTypes[component.Type] {
		return ErrInvalidComponentType
	}

	// Check if input components have names
	inputTypes := map[string]bool{
		ComponentTypeTextInput:         true,
		ComponentTypeTextArea:          true,
		ComponentTypeCheckboxGroup:     true,
		ComponentTypeRadioButtonsGroup: true,
		ComponentTypeDropdown:          true,
		ComponentTypeDatePicker:        true,
	}

	if inputTypes[component.Type] && component.Name == "" {
		return ErrMissingComponentName
	}

	// Validate actions
	if component.OnClickAction != nil {
		if err := ValidateAction(*component.OnClickAction); err != nil {
			return err
		}
	}

	if component.OnSelectionChange != nil {
		if err := ValidateAction(*component.OnSelectionChange); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAction performs validation on an action.
func ValidateAction(action Action) error {
	validActions := map[string]bool{
		ActionNavigate:     true,
		ActionDataExchange: true,
		ActionComplete:     true,
	}

	if !validActions[action.Name] {
		return ErrInvalidActionName
	}

	return nil
}
