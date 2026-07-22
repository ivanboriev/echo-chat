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
	if len(mh.buffer) == 50 {
		mh.buffer = mh.buffer[1:]
	}
	mh.buffer = append(mh.buffer, msg)
}
func (mh *MessageHistory) GetRecent() []ChatMessage {
	return mh.buffer
}
