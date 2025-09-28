package blockkit

type ButtonStyle string

const (
	ButtonStyleDefault = ""
	// primary gives buttons a green outline and text,
	// ideal for affirmation or confirmation actions.
	// primary should only be used for one button within a set.
	ButtonStylePrimary = "primary"
	// danger gives buttons a red outline and text,
	// and should be used when the action is destructive.
	// Use danger even more sparingly than primary.
	ButtonStyleDanger = "Danger"
)
