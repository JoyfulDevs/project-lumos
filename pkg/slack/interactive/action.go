package interactive

type ActionType string

const (
	// When the button is clicked.
	Button ActionType = "button"
	// When a checkbox is ticked or unticked.
	Checkbox ActionType = "checkboxes"
	// When a date is chosen and the date picker closes.
	Datepicker ActionType = "datepicker"
	// When an item from the overflow menu is clicked.
	Overflow ActionType = "overflow"
	// Determined by the dispatch_action_config field in the element.
	PlainTextInput ActionType = "plain_text_input"
	// When the selected radio in a group of radio buttons is changed.
	Radio ActionType = "radio_buttons"
	// Determined by the dispatch_action_config field in the element.
	RichTextInput ActionType = "rich_text_input"

	StaticSelect        ActionType = "static_select"
	ExternalSelect      ActionType = "external_select"
	UsersSelect         ActionType = "users_select"
	ConversationsSelect ActionType = "conversations_select"
	ChannelsSelect      ActionType = "channels_select"

	MultiStaticSelect        ActionType = "multi_static_select"
	MultiExternalSelect      ActionType = "multi_external_select"
	MultiUsersSelect         ActionType = "multi_users_select"
	MultiConversationsSelect ActionType = "multi_conversations_select"
	MultiChannelsSelect      ActionType = "multi_channels_select"
)

type Action struct {
	Type     ActionType `json:"type"`
	ActionID string     `json:"action_id"`
	BlockID  string     `json:"block_id"`
	Value    string     `json:"value,omitempty"`
}
