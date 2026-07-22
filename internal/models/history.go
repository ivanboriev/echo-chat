package models

type MessageHistory struct {
	buffer []ChatMessage
	head   int
	size   int
}

func (mh *MessageHistory) NewMessageHistory(size int) *MessageHistory {
	return &MessageHistory{
		buffer: make([]ChatMessage, size),
		head:   0,
		size:   size,
	}
}

func (mh *MessageHistory) Add(msg ChatMessage) {
	mh.buffer[mh.head%mh.size] = msg
	mh.head = mh.head + 1
}

func (mh *MessageHistory) GetRecent() []ChatMessage {
	return mh.buffer
}
