package blockkit

import "encoding/json"

// MarshalBlock marshals a Block into JSON format.
func MarshalBlock(block Block) ([]byte, error) {
	raw := struct {
		Block
		Type string `json:"type"`
	}{
		Block: block,
		Type:  block.BlockType(),
	}
	return json.Marshal(raw)
}

// MarshalElement marshals a BlockElement into JSON format.
func MarshalElement(element BlockElement) ([]byte, error) {
	raw := struct {
		BlockElement
		Type string `json:"type"`
	}{
		BlockElement: element,
		Type:         element.ElementType(),
	}
	return json.Marshal(raw)
}
