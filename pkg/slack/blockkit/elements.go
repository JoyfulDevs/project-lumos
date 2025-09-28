package blockkit

type BlockElement interface {
	// The type of element.
	ElementType() string
}

// Allows users a direct path to performing basic actions.
type ButtonElement struct {
	// A text object that defines the button's text.
	// text may truncate with ~30 characters.
	// Maximum length for the text in this field is 75 characters.
	//
	// Can only be of type: plain_text.
	Text *TextObject `json:"text"`
	// An identifier for this action. You can use this when you receive an
	// interaction payload to identify the source of the action.
	// Should be unique among all other action_ids in the containing block.
	// Maximum length is 255 characters.
	ActionID string `json:"action_id"`
	// A URL to load in the user's browser when the button is clicked.
	// Maximum length is 3000 characters. If you're using url, you'll still receive an
	// interaction payload and will need to send an acknowledgement response.
	URL string `json:"url,omitempty"`
	// The value to send along with the interaction payload.
	// Maximum length is 2000 characters.
	Value string `json:"value,omitempty"`
	// Decorates buttons with alternative visual color schemes.
	Style ButtonStyle `json:"style,omitempty"`
	// A confirm object that defines an optional confirmation dialog after the button is clicked.
	Confirm *ConfirmObject `json:"confirm,omitempty"`
	// A label for longer descriptive text about a button element.
	// This label will be read out by screen readers instead of the button text object.
	// Maximum length is 75 characters.
	Description *TextObject `json:"accessibility_label,omitempty"`
}

func (b *ButtonElement) ElementType() string {
	return "button"
}
