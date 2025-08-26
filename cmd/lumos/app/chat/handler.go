package chat

type Handler interface {
	HandleChat(chat *Chat)
}

type HandlerFunc func(chat *Chat)

func (f HandlerFunc) HandleChat(chat *Chat) {
	f(chat)
}
