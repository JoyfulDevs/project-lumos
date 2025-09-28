package blockkit

import "encoding/json"

type BlockType string

const (
	BlockTypeActions BlockType = "actions"
	BlockTypeDivider BlockType = "divider"
	BlockTypeHeader  BlockType = "header"
	BlockTypeSection BlockType = "section"
)

type Block struct {
	Type BlockType `json:"type"`
	ID   string    `json:"block_id,omitempty"`

	OfActionBlock  *ActionBlock  `json:"-"`
	OfDividerBlock *DividerBlock `json:"-"`
	OfHeaderBlock  *HeaderBlock  `json:"-"`
	OfSectionBlock *SectionBlock `json:"-"`
}

func NewBlockWithActionBlock(actionBlock *ActionBlock) *Block {
	return &Block{
		Type:          BlockTypeActions,
		OfActionBlock: actionBlock,
	}
}

func NewBlockWithDividerBlock() *Block {
	return &Block{
		Type:           BlockTypeDivider,
		OfDividerBlock: &DividerBlock{},
	}
}

func NewBlockWithHeaderBlock(headerBlock *HeaderBlock) *Block {
	return &Block{
		Type:          BlockTypeHeader,
		OfHeaderBlock: headerBlock,
	}
}

func NewBlockWithSectionBlock(sectionBlock *SectionBlock) *Block {
	return &Block{
		Type:           BlockTypeSection,
		OfSectionBlock: sectionBlock,
	}
}

func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block

	switch b.Type {
	case BlockTypeActions:
		raw := struct {
			ActionBlock
			Alias
		}{ActionBlock: *b.OfActionBlock, Alias: (Alias)(*b)}
		return json.Marshal(raw)
	case BlockTypeDivider:
		raw := struct {
			DividerBlock
			Alias
		}{DividerBlock: *b.OfDividerBlock, Alias: (Alias)(*b)}
		return json.Marshal(raw)
	case BlockTypeHeader:
		raw := struct {
			HeaderBlock
			Alias
		}{HeaderBlock: *b.OfHeaderBlock, Alias: (Alias)(*b)}
		return json.Marshal(raw)
	case BlockTypeSection:
		raw := struct {
			SectionBlock
			Alias
		}{SectionBlock: *b.OfSectionBlock, Alias: (Alias)(*b)}
		return json.Marshal(raw)
	}

	return nil, nil
}

func (b *Block) UnmarshalJSON(data []byte) error {
	type alias Block

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	b.Type = raw.Type
	b.ID = raw.ID
	switch raw.Type {
	case BlockTypeActions:
		b.OfActionBlock = &ActionBlock{}
		if err := json.Unmarshal(data, b.OfActionBlock); err != nil {
			return err
		}
	case BlockTypeDivider:
		b.OfDividerBlock = &DividerBlock{}
		if err := json.Unmarshal(data, b.OfDividerBlock); err != nil {
			return err
		}
	case BlockTypeHeader:
		b.OfHeaderBlock = &HeaderBlock{}
		if err := json.Unmarshal(data, b.OfHeaderBlock); err != nil {
			return err
		}
	case BlockTypeSection:
		b.OfSectionBlock = &SectionBlock{}
		if err := json.Unmarshal(data, b.OfSectionBlock); err != nil {
			return err
		}
	}

	return nil
}

// Holds multiple interactive elements.
type ActionBlock struct {
	// An array of interactive element objects -
	// buttons, select menus, overflow menus, or date pickers.
	// There is a maximum of 25 elements in each action block.
	Elements []*BlockElement `json:"elements"`
}

// Visually separates pieces of info inside of a message.
type DividerBlock struct {
}

// Displays a larger-sized text.
type HeaderBlock struct {
	// The text for the block, in the form of a plain_text text object.
	// Maximum length for the text in this field is 150 characters.
	Text *TextObject `json:"text"`
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
	Accessory *BlockElement `json:"accessory,omitempty"`
	// Whether or not this section block's text should always expand when rendered.
	// If false or not provided, it may be rendered with a 'see more' option to expand and show the full text.
	// For AI Assistant apps, this allows the app to post long messages without
	// users needing to click 'see more' to expand the message.
	Expand bool `json:"expand,omitempty"`
}
