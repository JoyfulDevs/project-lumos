package blockkit

import "encoding/json"

type BlockElementType string

const (
	BlockElementTypeButton BlockElementType = "button"
)

type BlockElement struct {
	Type BlockElementType `json:"type"`

	OfButtonElement *ButtonElement `json:"-"`
}

func NewBlockElementWithButtonElement(buttonElement *ButtonElement) *BlockElement {
	return &BlockElement{
		Type:            BlockElementTypeButton,
		OfButtonElement: buttonElement,
	}
}

func (e *BlockElement) MarshalJSON() ([]byte, error) {
	type Alias BlockElement

	switch e.Type {
	case BlockElementTypeButton:
		raw := struct {
			ButtonElement
			Alias
		}{ButtonElement: *e.OfButtonElement, Alias: (Alias)(*e)}
		return json.Marshal(raw)
	}

	return nil, nil
}

func (e *BlockElement) UnmarshalJSON(data []byte) error {
	type alias BlockElement

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	e.Type = raw.Type
	switch raw.Type {
	case BlockElementTypeButton:
		e.OfButtonElement = &ButtonElement{}
		if err := json.Unmarshal(data, e.OfButtonElement); err != nil {
			return err
		}
	}

	return nil
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
