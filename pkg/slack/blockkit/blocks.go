package blockkit

type Block interface {
	// The type of block.
	BlockType() string
}

// Holds multiple interactive elements.
type ActionBlock struct {
	// An array of interactive element objects -
	// buttons, select menus, overflow menus, or date pickers.
	// There is a maximum of 25 elements in each action block.
	Elements []BlockElement `json:"elements"`
}

func (a *ActionBlock) BlockType() string {
	return "actions"
}

// Visually separates pieces of info inside of a message.
type DividerBlock struct {
}

func (d *DividerBlock) BlockType() string {
	return "divider"
}

// Displays a larger-sized text.
type HeaderBlock struct {
	// The text for the block, in the form of a plain_text text object.
	// Maximum length for the text in this field is 150 characters.
	Text *TextObject `json:"text"`
}

func (h *HeaderBlock) BlockType() string {
	return "header"
}

// Displays text, possibly alongside elements.
type SectionBlock struct {
	// The text for the block, in the form of a text object.
	// Minimum length for the text in this field is 1 and maximum length is 3000 characters.
	// This field is not required if a valid array of fields objects is provided instead.
	Text *TextObject `json:"text,omitempty"`
	// Required if no text is provided. An array of text objects.
	// Any text objects included with fields will be rendered in a compact format that
	// allows for 2 columns of side-by-side text.
	// Maximum number of items is 10.
	// Maximum length for the text in each item is 2000 characters.
	Fields []TextObject `json:"fields,omitempty"`
	// One of the compatible element objects noted above.
	// Be sure to confirm the desired element works with section.
	Accessory BlockElement `json:"accessory,omitempty"`
	// Whether or not this section block's text should always expand when rendered.
	// If false or not provided, it may be rendered with a 'see more' option to expand and show the full text.
	// For AI Assistant apps, this allows the app to post long messages without
	// users needing to click 'see more' to expand the message.
	Expand bool `json:"expand,omitempty"`
}

func (s *SectionBlock) BlockType() string {
	return "section"
}
