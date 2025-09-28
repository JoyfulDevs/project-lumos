package blockkit

type TextObjectType string

const (
	PlainText TextObjectType = "plain_text"
	Markdown  TextObjectType = "mrkdwn"
)

// Formatted either as plain_text or using mrkdwn, our proprietary contribution to the much beloved Markdown standard.
type TextObject struct {
	// The formatting to use for this text object.
	Type TextObjectType `json:"type"`
	// The text for the block. This field accepts any of
	// the standard text formatting markup when type is mrkdwn.
	// The minimum length is 1 and maximum length is 3000 characters.
	Text string `json:"text"`
	// Indicates whether emojis in a text field should be escaped into the colon emoji format.
	// This field is only usable when type is plain_text.
	Emoji bool `json:"emoji,omitempty"`
	// When set to false (as is default) URLs will be auto-converted into links,
	// conversation names will be link-ified, and certain mentions will be automatically parsed.
	//
	// When set to true, Slack will continue to process all markdown formatting and
	// manual parsing strings, but it won’t modify any plain-text content.
	//
	// For example, channel names will not be hyperlinked. This field is only usable when type is mrkdwn.
	Verbatim bool `json:"verbatim,omitempty"`
}

// NewPlainText creates a new plain text object.
func NewPlainText(text string, emoji bool) *TextObject {
	return &TextObject{
		Type:  PlainText,
		Text:  text,
		Emoji: emoji,
	}
}

// NewMarkdownText creates a new markdown text object.
func NewMarkdownText(text string, verbatim bool) *TextObject {
	return &TextObject{
		Type:     Markdown,
		Text:     text,
		Verbatim: verbatim,
	}
}

// Defines a dialog that adds a confirmation step to interactive elements.
type ConfirmObject struct {
	// A plain_text text object that defines the dialog's title.
	// Maximum length for this field is 100 characters.
	Title *TextObject `json:"title"`
	// A plain_text text object that defines the explanatory text that appears in the confirm dialog.
	// Maximum length for the text in this field is 300 characters.
	Text *TextObject `json:"text"`
	// A plain_text text object to define the text of the button that confirms the action.
	// Maximum length for the text in this field is 30 characters.
	Confirm *TextObject `json:"confirm"`
	// A plain_text text object to define the text of the button that cancels the action.
	// Maximum length for the text in this field is 30 characters.
	Deny *TextObject `json:"deny"`
	// Defines the color scheme applied to the confirm button.
	Style ButtonStyle `json:"style,omitempty"`
}
